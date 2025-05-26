package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PhotoPlugin is the implementation of the PhotoPlugin service
type PhotoPlugin struct {
	// Implement the UnimplementedPhotoPluginServer to satisfy the interface
	UnimplementedPhotoPluginServer
}

// Photo represents a photo in the system
type Photo struct {
	ID     string
	S3Key  string
	Mime   string
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query string
}

// SearchResult represents a search result
type SearchResult struct {
	IDs []string
}

// UnimplementedPhotoPluginServer must be embedded to have forward compatible implementations
type UnimplementedPhotoPluginServer struct{}

func (*UnimplementedPhotoPluginServer) ProcessPhoto(context.Context, *Photo) (*emptypb.Empty, error) {
	return nil, nil
}

func (*UnimplementedPhotoPluginServer) Search(context.Context, *SearchRequest) (*SearchResult, error) {
	return nil, nil
}

// ProcessPhoto processes a photo (noop implementation)
func (p *PhotoPlugin) ProcessPhoto(ctx context.Context, photo *Photo) (*emptypb.Empty, error) {
	log.Printf("Received ProcessPhoto request for photo ID: %s", photo.ID)
	return &emptypb.Empty{}, nil
}

// Search searches for photos (noop implementation)
func (p *PhotoPlugin) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	log.Printf("Received Search request with query: %s", req.Query)
	return &SearchResult{IDs: []string{}}, nil
}

// PhotoUploaded represents a photo uploaded event
type PhotoUploaded struct {
	Id        string `json:"id"`
	Filename  string `json:"filename"`
	Mime      string `json:"mime"`
	S3Key     string `json:"s3_key"`
	CreatedAt string `json:"created_at"`
}

// Config holds the configuration for the thumbnailer
type Config struct {
	NatsURL      string
	S3Endpoint   string
	S3AccessKey  string
	S3SecretKey  string
	S3Bucket     string
	DatabaseURL  string
	NumWorkers   int
	MaxRetries   int
	ThumbnailSize int
}

// Thumbnailer is the main struct for the thumbnailer plugin
type Thumbnailer struct {
	config     Config
	s3Client   *s3.Client
	dbPool     *pgxpool.Pool
	js         nats.JetStreamContext
	sub        *nats.Subscription
	workerPool chan struct{}
	wg         sync.WaitGroup
}

// NewThumbnailer creates a new thumbnailer
func NewThumbnailer(config Config) (*Thumbnailer, error) {
	// Create a custom resolver that routes all requests to the specified endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               config.S3Endpoint,
			HostnameImmutable: true,
			SigningRegion:     "us-east-1", // MinIO doesn't care about the region
		}, nil
	})

	// Create a custom AWS config
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.S3AccessKey,
			config.S3SecretKey,
			"",
		)),
		awsconfig.WithRegion("us-east-1"), // MinIO doesn't care about the region
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create an S3 client
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // MinIO requires path-style addressing
	})

	// Create a connection pool
	poolConfig, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set some reasonable defaults for the connection pool
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Connect to NATS
	nc, err := nats.Connect(config.NatsURL)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		pool.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// Create a worker pool
	workerPool := make(chan struct{}, config.NumWorkers)
	for i := 0; i < config.NumWorkers; i++ {
		workerPool <- struct{}{}
	}

	return &Thumbnailer{
		config:     config,
		s3Client:   s3Client,
		dbPool:     pool,
		js:         js,
		workerPool: workerPool,
	}, nil
}

// Start starts the thumbnailer
func (t *Thumbnailer) Start() error {
	// Create a consumer
	sub, err := t.js.QueueSubscribe(
		"photo.uploaded",
		"thumbnailer",
		t.handlePhotoUploaded,
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.DeliverNew(),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to photo.uploaded: %w", err)
	}
	t.sub = sub

	log.Printf("Subscribed to photo.uploaded")
	return nil
}

// Stop stops the thumbnailer
func (t *Thumbnailer) Stop() {
	if t.sub != nil {
		t.sub.Unsubscribe()
	}
	t.wg.Wait()
	t.dbPool.Close()
}

// handlePhotoUploaded handles a photo.uploaded event
func (t *Thumbnailer) handlePhotoUploaded(msg *nats.Msg) {
	// Get a worker from the pool
	<-t.workerPool
	t.wg.Add(1)

	// Process the message in a goroutine
	go func() {
		defer func() {
			// Return the worker to the pool
			t.workerPool <- struct{}{}
			t.wg.Done()
		}()

		var retries int
		var backoff = 1 * time.Second

		for retries <= t.config.MaxRetries {
			err := t.processMessage(msg)
			if err == nil {
				// Message processed successfully
				if err := msg.Ack(); err != nil {
					log.Printf("Failed to acknowledge message: %v", err)
				}
				return
			}

			// Message processing failed
			log.Printf("Failed to process message (attempt %d/%d): %v", retries+1, t.config.MaxRetries+1, err)

			if retries == t.config.MaxRetries {
				log.Printf("Max retries reached, giving up: %v", err)
				if err := msg.Nak(); err != nil {
					log.Printf("Failed to negative acknowledge message: %v", err)
				}
				return
			}

			// Sleep with exponential backoff
			time.Sleep(backoff)
			backoff *= 2
			retries++
		}
	}()
}

// processMessage processes a message
func (t *Thumbnailer) processMessage(msg *nats.Msg) error {
	// Parse the message
	var event PhotoUploaded
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("Processing photo: %s, key: %s, mime: %s", event.Id, event.S3Key, event.Mime)

	// Skip non-image MIME types
	if !strings.HasPrefix(event.Mime, "image/") {
		log.Printf("Skipping non-image MIME type: %s", event.Mime)
		return nil
	}

	// Download the original image from S3
	result, err := t.s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(t.config.S3Bucket),
		Key:    aws.String(event.S3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the image
	imgData, err := io.ReadAll(result.Body)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize the image
	resized := imaging.Fit(img, t.config.ThumbnailSize, t.config.ThumbnailSize, imaging.Lanczos)

	// Encode the image as JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Upload the thumbnail to S3
	thumbnailKey := fmt.Sprintf("thumb/%d/%s.jpg", t.config.ThumbnailSize, event.Id)
	_, err = t.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(t.config.S3Bucket),
		Key:         aws.String(thumbnailKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload thumbnail to S3: %w", err)
	}

	// Update the database
	sizeStr := strconv.Itoa(t.config.ThumbnailSize)
	
	query := `
		UPDATE photos
		SET meta = jsonb_set(
			COALESCE(meta, '{}'::jsonb),
			'{thumbnails}',
			COALESCE(meta->'thumbnails', '{}'::jsonb) || jsonb_build_object($1, $2)
		)
		WHERE id = $3
	`
	_, err = t.dbPool.Exec(context.Background(), query, sizeStr, thumbnailKey, event.Id)
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	log.Printf("Successfully created thumbnail for photo %s: %s", event.Id, thumbnailKey)
	return nil
}

func main() {
	// Parse command line flags
	port := flag.Int("port", 0, "The server port")
	flag.Parse()

	// Load configuration from environment variables
	config := Config{
		NatsURL:       getEnv("NATS_URL", "nats://nats:4222"),
		S3Endpoint:    getEnv("S3_ENDPOINT", "http://minio:9000"),
		S3AccessKey:   getEnv("S3_ACCESS_KEY", "minio"),
		S3SecretKey:   getEnv("S3_SECRET_KEY", "minio123"),
		S3Bucket:      getEnv("S3_BUCKET", "pixie"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://pixie:pixiepass@postgres:5432/pixiedb?sslmode=disable"),
		NumWorkers:    getIntEnv("THUMB_WORKERS", 4),
		MaxRetries:    3,
		ThumbnailSize: 512,
	}

	// Create a thumbnailer
	thumbnailer, err := NewThumbnailer(config)
	if err != nil {
		log.Fatalf("Failed to create thumbnailer: %v", err)
	}

	// Start the thumbnailer
	if err := thumbnailer.Start(); err != nil {
		log.Fatalf("Failed to start thumbnailer: %v", err)
	}

	// If port is 0, let the OS choose a port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Print the port for the plugin loader to read
	fmt.Printf("PORT=%d\n", lis.Addr().(*net.TCPAddr).Port)

	// Create a gRPC server
	s := grpc.NewServer()

	// Register the health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register the plugin service
	RegisterPhotoPluginServer(s, &PhotoPlugin{})

	// Handle shutdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		s.GracefulStop()
		thumbnailer.Stop()
	}()

	// Start the server
	log.Printf("Starting thumbnailer plugin server on port %d with %d workers", lis.Addr().(*net.TCPAddr).Port, config.NumWorkers)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// RegisterPhotoPluginServer registers the PhotoPluginServer to the gRPC server
func RegisterPhotoPluginServer(s *grpc.Server, srv *PhotoPlugin) {
	s.RegisterService(&_PhotoPlugin_serviceDesc, srv)
}

// Service descriptor for the PhotoPlugin service
var _PhotoPlugin_serviceDesc = grpc.ServiceDesc{
	ServiceName: "plugin.v1.PhotoPlugin",
	HandlerType: (*interface{})(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessPhoto",
			Handler:    _PhotoPlugin_ProcessPhoto_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _PhotoPlugin_Search_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "plugin/v1/plugin.proto",
}

// Handler for ProcessPhoto
func _PhotoPlugin_ProcessPhoto_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Photo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*PhotoPlugin).ProcessPhoto(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.PhotoPlugin/ProcessPhoto",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*PhotoPlugin).ProcessPhoto(ctx, req.(*Photo))
	}
	return interceptor(ctx, in, info, handler)
}

// Handler for Search
func _PhotoPlugin_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*PhotoPlugin).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.PhotoPlugin/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*PhotoPlugin).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv gets an environment variable as an integer or returns a default value
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

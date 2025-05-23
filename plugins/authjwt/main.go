package main

import (
	"context"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Config holds the plugin configuration
type Config struct {
	JWTAlgo          string
	JWTSecret        string
	JWTPublicKeyFile string
}

// PhotoPlugin is the implementation of the PhotoPlugin service
type PhotoPlugin struct {
	// Implement the UnimplementedPhotoPluginServer to satisfy the interface
	UnimplementedPhotoPluginServer
	config Config
	pubKey *rsa.PublicKey
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

// ValidateTokenRequest represents a token validation request
type ValidateTokenRequest struct {
	Token string
}

// ValidateTokenResponse represents a token validation response
type ValidateTokenResponse struct {
	Ok     bool
	UserId string
	Error  string
}

// UnimplementedPhotoPluginServer must be embedded to have forward compatible implementations
type UnimplementedPhotoPluginServer struct{}

func (*UnimplementedPhotoPluginServer) ProcessPhoto(context.Context, *Photo) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (*UnimplementedPhotoPluginServer) Search(context.Context, *SearchRequest) (*SearchResult, error) {
	return &SearchResult{IDs: []string{}}, nil
}

func (*UnimplementedPhotoPluginServer) ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	return &ValidateTokenResponse{Ok: false, Error: "Not implemented"}, nil
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

// ValidateToken validates a JWT token
func (p *PhotoPlugin) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	log.Printf("Validating token")

	// Parse the token
	var token *jwt.Token
	var err error

	switch p.config.JWTAlgo {
	case "HS256":
		// Parse the token with HMAC signing method
		token, err = jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(p.config.JWTSecret), nil
		})
	case "RS256":
		// Parse the token with RSA signing method
		token, err = jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return p.pubKey, nil
		})
	default:
		return &ValidateTokenResponse{
			Ok:    false,
			Error: fmt.Sprintf("unsupported algorithm: %s", p.config.JWTAlgo),
		}, nil
	}

	// Check for parsing errors
	if err != nil {
		return &ValidateTokenResponse{
			Ok:    false,
			Error: fmt.Sprintf("failed to parse token: %v", err),
		}, nil
	}

	// Check if the token is valid
	if !token.Valid {
		return &ValidateTokenResponse{
			Ok:    false,
			Error: "invalid token",
		}, nil
	}

	// Get the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &ValidateTokenResponse{
			Ok:    false,
			Error: "failed to parse claims",
		}, nil
	}

	// Check if the token has expired
	exp, ok := claims["exp"]
	if !ok {
		return &ValidateTokenResponse{
			Ok:    false,
			Error: "missing exp claim",
		}, nil
	}
	
	// Convert exp to float64 and then to int64
	expFloat, ok := exp.(float64)
	if !ok {
		return &ValidateTokenResponse{
			Ok:    false,
			Error: "invalid exp claim format",
		}, nil
	}
	
	// Check if token has expired
	if time.Now().Unix() > int64(expFloat) {
		return &ValidateTokenResponse{
			Ok:    false,
			Error: "token has expired",
		}, nil
	}

	// Get the subject claim
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return &ValidateTokenResponse{
			Ok:    false,
			Error: "missing or invalid sub claim",
		}, nil
	}

	// Return success
	return &ValidateTokenResponse{
		Ok:     true,
		UserId: sub,
	}, nil
}

func main() {
	// Parse command line flags
	port := flag.Int("port", 0, "The server port")
	flag.Parse()

	// Load configuration from environment variables
	config := Config{
		JWTAlgo:          getEnv("JWT_ALGO", "HS256"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTPublicKeyFile: getEnv("JWT_PUBLIC_KEY_FILE", ""),
	}

	// Validate configuration
	if config.JWTAlgo != "HS256" && config.JWTAlgo != "RS256" {
		log.Fatalf("Invalid JWT_ALGO: %s. Must be HS256 or RS256", config.JWTAlgo)
	}

	if config.JWTAlgo == "HS256" && config.JWTSecret == "" {
		log.Fatalf("JWT_SECRET is required when JWT_ALGO is HS256")
	}

	if config.JWTAlgo == "RS256" && config.JWTPublicKeyFile == "" {
		log.Fatalf("JWT_PUBLIC_KEY_FILE is required when JWT_ALGO is RS256")
	}

	// Create the plugin
	plugin := &PhotoPlugin{
		config: config,
	}

	// Load the public key if using RS256
	if config.JWTAlgo == "RS256" {
		pubKeyBytes, err := os.ReadFile(config.JWTPublicKeyFile)
		if err != nil {
			log.Fatalf("Failed to read public key file: %v", err)
		}

		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
		if err != nil {
			log.Fatalf("Failed to parse public key: %v", err)
		}

		plugin.pubKey = pubKey
	}

	// If port is 0, let the OS choose a port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Print the port for the plugin loader to read
	fmt.Printf("PORT=%d\n", lis.Addr().(*net.TCPAddr).Port)

	// Create a gRPC server with middleware
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	// Register the health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register the plugin service
	RegisterPhotoPluginServer(s, plugin)

	// Handle shutdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()

	// Start the server
	log.Printf("Starting auth-jwt plugin server on port %d", lis.Addr().(*net.TCPAddr).Port)
	log.Printf("Using JWT algorithm: %s", config.JWTAlgo)
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
		{
			MethodName: "ValidateToken",
			Handler:    _PhotoPlugin_ValidateToken_Handler,
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

// Handler for ValidateToken
func _PhotoPlugin_ValidateToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*PhotoPlugin).ValidateToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.PhotoPlugin/ValidateToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*PhotoPlugin).ValidateToken(ctx, req.(*ValidateTokenRequest))
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

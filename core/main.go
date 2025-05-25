package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"pixie/auth"
	"pixie/db"
	"pixie/events"
	"pixie/photo/v1"
	"pixie/plugin/loader"
	"pixie/storage"
)

// Config holds the application configuration
type Config struct {
	S3Endpoint      string
	S3AccessKey     string
	S3SecretKey     string
	S3Bucket        string
	DatabaseURL     string
	JWTAlgo         string
	JWTSecret       string
	TokenExpiration time.Duration
}

// App holds the application state
type App struct {
	Config  Config
	DB      interface {
		SavePhoto(ctx context.Context, id, s3Key, filename, mime string) error
		GetPhoto(ctx context.Context, id string) (string, string, error)
		DeletePhoto(ctx context.Context, id string) error
		ListPhotos(ctx context.Context) ([]db.Photo, error)
	}
	Storage interface {
		UploadObject(ctx context.Context, key string, data io.Reader, contentType string) error
		GetObject(ctx context.Context, key string) (*s3.GetObjectOutput, error)
		DeleteObject(ctx context.Context, key string) error
	}
	Auth *auth.Service
}

func main() {
	// Load configuration from environment variables
	config := Config{
		S3Endpoint:      getEnv("S3_ENDPOINT", "http://minio:9000"),
		S3AccessKey:     getEnv("S3_ACCESS_KEY", "minio"),
		S3SecretKey:     getEnv("S3_SECRET_KEY", "minio123"),
		S3Bucket:        getEnv("S3_BUCKET", "pixie"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://pixie:pixiepass@postgres:5432/pixiedb?sslmode=disable"),
		JWTAlgo:         getEnv("JWT_ALGO", "HS256"),
		JWTSecret:       getEnv("JWT_SECRET", "supersecret123"),
		TokenExpiration: 24 * time.Hour, // Default to 24 hours
	}


	// Initialize implementations
	log.Println("Initializing implementations")
	
	// Initialize database
	dbInstance, err := db.New(context.Background(), db.Config{
		URL: config.DatabaseURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Initialize the database schema
	if err := dbInstance.InitSchema(context.Background()); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
	
	// Initialize S3 storage
	s3Storage, err := storage.New(context.Background(), storage.Config{
		Endpoint:  config.S3Endpoint,
		AccessKey: config.S3AccessKey,
		SecretKey: config.S3SecretKey,
		Bucket:    config.S3Bucket,
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}
	
	// Initialize real NATS events
	if err := events.Init(); err != nil {
		log.Printf("Failed to initialize NATS: %v", err)
		// Continue even if NATS initialization fails
	}
	
	// Initialize plugin loader (for non-auth plugins)
	if err := loader.Init(); err != nil {
		log.Printf("Failed to initialize plugin loader: %v", err)
		// Continue even if plugin loading fails
	}

	// Initialize authentication service
	authService, err := auth.NewService(auth.Config{
		JWTAlgo:           config.JWTAlgo,
		JWTSecret:         config.JWTSecret,
		JWTPublicKeyFile:  getEnv("JWT_PUBLIC_KEY_FILE", ""),
		JWTPrivateKeyFile: getEnv("JWT_PRIVATE_KEY_FILE", ""),
		TokenExpiration:   config.TokenExpiration,
	})
	if err != nil {
		log.Fatalf("Failed to initialize authentication service: %v", err)
	}

	// Create the application with DB and Storage
	app := &App{
		Config:  config,
		DB:      dbInstance,
		Storage: s3Storage,
		Auth:    authService,
	}

	// Create a router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/healthz", app.healthzHandler).Methods("GET")
	
	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	
	// Auth endpoints
	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/health", app.authHealthHandler).Methods("GET")
	authRouter.HandleFunc("/token", app.generateTokenHandler).Methods("POST")
	authRouter.HandleFunc("/revoke", app.revokeTokenHandler).Methods("POST")
	
	// Protected endpoints
	protectedRouter := apiRouter.PathPrefix("").Subrouter()
	protectedRouter.Use(app.Auth.Middleware)
	
	protectedRouter.HandleFunc("/upload", app.uploadHandler).Methods("POST")
	protectedRouter.HandleFunc("/photo/{id}", app.photoHandler).Methods("GET")
	protectedRouter.HandleFunc("/photo/{id}", app.deletePhotoHandler).Methods("DELETE")
	protectedRouter.HandleFunc("/photos", app.listPhotosHandler).Methods("GET")

	// Serve static files from the UI React plugin
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("/plugins/ui-react/dist")))

	// Start the server
	log.Println("Starting Pixie Core server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// healthzHandler handles the /healthz endpoint
func (app *App) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pixie core")
}

// uploadHandler handles the /upload endpoint
func (app *App) uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Parse the multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a UUID for the file
	id := uuid.New().String()

	// Determine the MIME type
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	// Create the S3 key
	s3Key := fmt.Sprintf("photos/%s", id)

	// Upload the file to S3
	if err := app.Storage.UploadObject(ctx, s3Key, file, mime); err != nil {
		log.Printf("Failed to upload file to S3: %v", err)
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// Save the metadata to the database
	if err := app.DB.SavePhoto(ctx, id, s3Key, header.Filename, mime); err != nil {
		log.Printf("Failed to save metadata to database: %v", err)
		http.Error(w, "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	// Create and publish the photo.uploaded event
	go func() {
		// Create a new context with timeout for publishing
		publishCtx, publishCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer publishCancel()

		// Create the event data using protobuf
		event := &photo.PhotoUploaded{
			Id:        id,
			Filename:  header.Filename,
			Mime:      mime,
			S3Key:     s3Key,
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		// Marshal the event to JSON
		eventData, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event data: %v", err)
			return
		}

		// Publish the event
		if err := events.Publish(publishCtx, "photo.uploaded", eventData); err != nil {
			log.Printf("Failed to publish photo.uploaded event: %v", err)
		}
	}()

	// Return the ID as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// photoHandler handles the /photo/:id endpoint
func (app *App) photoHandler(w http.ResponseWriter, r *http.Request) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get the ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	// Get the metadata from the database
	s3Key, mime, err := app.DB.GetPhoto(ctx, id)
	if err != nil {
		log.Printf("Failed to get metadata from database: %v", err)
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Get the object from S3
	result, err := app.Storage.GetObject(ctx, s3Key)
	if err != nil {
		log.Printf("Failed to get object from S3: %v", err)
		http.Error(w, "Failed to get photo", http.StatusInternalServerError)
		return
	}
	defer result.Body.Close()

	// Set the content type
	w.Header().Set("Content-Type", mime)

	// Stream the object to the response
	if _, err := io.Copy(w, result.Body); err != nil {
		log.Printf("Failed to stream object: %v", err)
		// Can't send an error response here as we've already started writing the response
	}
}

// deletePhotoHandler handles the DELETE /photo/:id endpoint
func (app *App) deletePhotoHandler(w http.ResponseWriter, r *http.Request) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get the ID from the URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	// Get the metadata from the database
	s3Key, _, err := app.DB.GetPhoto(ctx, id)
	if err != nil {
		log.Printf("Failed to get metadata from database: %v", err)
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Delete the object from S3
	if err := app.Storage.DeleteObject(ctx, s3Key); err != nil {
		log.Printf("Failed to delete object from S3: %v", err)
		http.Error(w, "Failed to delete photo from storage", http.StatusInternalServerError)
		return
	}

	// Delete the metadata from the database
	if err := app.DB.DeletePhoto(ctx, id); err != nil {
		log.Printf("Failed to delete metadata from database: %v", err)
		http.Error(w, "Failed to delete photo metadata", http.StatusInternalServerError)
		return
	}

	// Create and publish the photo.deleted event
	go func() {
		// Create a new context with timeout for publishing
		publishCtx, publishCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer publishCancel()

		// Create the event data using protobuf
		event := &photo.PhotoDeleted{
			Id:        id,
			DeletedAt: time.Now().Format(time.RFC3339),
		}

		// Marshal the event to JSON
		eventData, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event data: %v", err)
			return
		}

		// Publish the event
		if err := events.Publish(publishCtx, "photo.deleted", eventData); err != nil {
			log.Printf("Failed to publish photo.deleted event: %v", err)
		}
	}()

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// listPhotosHandler handles the /photos endpoint
func (app *App) listPhotosHandler(w http.ResponseWriter, r *http.Request) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get all photos from the database
	photos, err := app.DB.ListPhotos(ctx)
	if err != nil {
		log.Printf("Failed to list photos from database: %v", err)
		http.Error(w, "Failed to list photos", http.StatusInternalServerError)
		return
	}

	// Return the photos as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"photos": photos,
	})
}

// authHealthHandler handles the /auth/health endpoint
func (app *App) authHealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.Auth.HealthCheck(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Auth service unhealthy: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Auth service healthy")
}

// TokenRequest represents a request for a new token
type TokenRequest struct {
	Subject      string                 `json:"subject"`
	CustomClaims map[string]interface{} `json:"custom_claims,omitempty"`
}

// generateTokenHandler handles the /auth/token endpoint
func (app *App) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.Subject == "" {
		http.Error(w, "Subject is required", http.StatusBadRequest)
		return
	}

	// Generate the token
	token, err := app.Auth.GenerateToken(req.Subject, req.CustomClaims)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return the token as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

// RevokeRequest represents a request to revoke a token
type RevokeRequest struct {
	Token string `json:"token"`
}

// revokeTokenHandler handles the /auth/revoke endpoint
func (app *App) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Revoke the token
	app.Auth.RevokeToken(req.Token)

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

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

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourname/pixie/core/db"
	"github.com/yourname/pixie/core/storage"
)

// Config holds the application configuration
type Config struct {
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	DatabaseURL string
}

// App holds the application state
type App struct {
	Config  Config
	DB      *db.DB
	Storage *storage.S3
}

func main() {
	// Load configuration from environment variables
	config := Config{
		S3Endpoint:  getEnv("S3_ENDPOINT", "http://minio:9000"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minio"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minio123"),
		S3Bucket:    getEnv("S3_BUCKET", "pixie"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/pixie?sslmode=disable"),
	}

	// Create a context with timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize the database
	database, err := db.New(ctx, db.Config{
		URL: config.DatabaseURL,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize the database schema
	if err := database.InitSchema(ctx); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Initialize the storage
	s3Client, err := storage.New(ctx, storage.Config{
		Endpoint:  config.S3Endpoint,
		AccessKey: config.S3AccessKey,
		SecretKey: config.S3SecretKey,
		Bucket:    config.S3Bucket,
	})
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create the application
	app := &App{
		Config:  config,
		DB:      database,
		Storage: s3Client,
	}

	// Create a router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/healthz", app.healthzHandler).Methods("GET")
	router.HandleFunc("/upload", app.uploadHandler).Methods("POST")
	router.HandleFunc("/photo/{id}", app.photoHandler).Methods("GET")

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

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

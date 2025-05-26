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
		TrashPhoto(ctx context.Context, id string) error
		RestorePhoto(ctx context.Context, id string) error
		ListTrashedPhotos(ctx context.Context) ([]db.Photo, error)
		EmptyTrash(ctx context.Context) (int64, error)
		PermanentlyDeletePhoto(ctx context.Context, id string) error
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
	
	// Trash functionality endpoints
	protectedRouter.HandleFunc("/photos/trash/{id}", app.trashPhotoHandler).Methods("PUT")
	protectedRouter.HandleFunc("/photos/trash/{id}/restore", app.restorePhotoHandler).Methods("PUT")
	protectedRouter.HandleFunc("/photos/trash", app.listTrashHandler).Methods("GET")
	protectedRouter.HandleFunc("/photos/trash", app.emptyTrashHandler).Methods("DELETE")
	protectedRouter.HandleFunc("/photos/trash/{id}", app.permanentDeletePhotoHandler).Methods("DELETE")

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
		log.Printf("Missing photo ID in request")
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	// Check if a thumbnail size was requested
	thumbnailSize := r.URL.Query().Get("thumbnail")
	log.Printf("Request for photo %s with thumbnail size: %s", id, thumbnailSize)

	// Check for token in query parameter (for image requests from <img> tags)
	tokenParam := r.URL.Query().Get("token")
	if tokenParam != "" {
		// Validate the token
		userId, _, err := app.Auth.ValidateToken(tokenParam)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Printf("User authenticated via query param: %s", userId)
	} else {
		// Check if the request is already authenticated via middleware
		if r.Header.Get("X-User-Id") == "" {
			log.Printf("Unauthorized access attempt for photo %s", id)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Get the photo from the database
	photos, err := app.DB.ListPhotos(ctx)
	if err != nil {
		log.Printf("Failed to list photos from database: %v", err)
		http.Error(w, "Failed to get photo metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d photos in database", len(photos))

	// Find the photo with the matching ID
	var photo *db.Photo
	for i := range photos {
		if photos[i].ID == id {
			photo = &photos[i]
			break
		}
	}

	if photo == nil {
		log.Printf("Photo not found in database: %s", id)
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	log.Printf("Found photo: %+v", *photo)

	// Determine which S3 key to use
	s3Key := photo.S3Key
	mime := photo.Mime

	// If a thumbnail was requested and exists, use it
	if thumbnailSize != "" && photo.Meta != nil {
		log.Printf("Thumbnail requested: %s for photo ID: %s", thumbnailSize, id)
		log.Printf("Photo meta: %+v", photo.Meta)

		if thumbnails, ok := photo.Meta["thumbnails"].(map[string]interface{}); ok {
			log.Printf("Thumbnails found: %+v", thumbnails)

			if thumbKey, ok := thumbnails[thumbnailSize].(string); ok {
				s3Key = thumbKey
				mime = "image/jpeg" // Thumbnails are always JPEG
				log.Printf("Using thumbnail: %s", thumbKey)
			} else {
				log.Printf("Thumbnail of size %s not found for photo %s", thumbnailSize, id)
			}
		} else {
			log.Printf("No thumbnails found in meta for photo %s", id)
		}
	}

	log.Printf("Attempting to retrieve S3 object with key: %s", s3Key)

	// Get the object from S3
	result, err := app.Storage.GetObject(ctx, s3Key)
	if err != nil {
		log.Printf("Failed to get object from S3: %v", err)
		// Return a more detailed error to help with debugging
		http.Error(w, fmt.Sprintf("Failed to get photo from storage: %v", err), http.StatusInternalServerError)
		return
	}
	defer result.Body.Close()

	log.Printf("Successfully retrieved S3 object with key: %s", s3Key)

	// Set the content type
	w.Header().Set("Content-Type", mime)

	// Add cache control headers
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Stream the object to the response
	bytesCopied, err := io.Copy(w, result.Body)
	if err != nil {
		log.Printf("Failed to stream object: %v", err)
		// Can't send an error response here as we've already started writing the response
	} else {
		log.Printf("Successfully streamed %d bytes to client for photo %s", bytesCopied, id)
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

// trashPhotoHandler handles the PUT /photos/trash/:id endpoint (move to trash)
func (app *App) trashPhotoHandler(w http.ResponseWriter, r *http.Request) {
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

	// Move the photo to trash
	if err := app.DB.TrashPhoto(ctx, id); err != nil {
		log.Printf("Failed to trash photo: %v", err)
		http.Error(w, fmt.Sprintf("Failed to trash photo: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Photo moved to trash",
		"id":      id,
	})
}

// restorePhotoHandler handles the PUT /photos/trash/:id/restore endpoint
func (app *App) restorePhotoHandler(w http.ResponseWriter, r *http.Request) {
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

	// Restore the photo from trash
	if err := app.DB.RestorePhoto(ctx, id); err != nil {
		log.Printf("Failed to restore photo: %v", err)
		http.Error(w, fmt.Sprintf("Failed to restore photo: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Photo restored from trash",
		"id":      id,
	})
}

// listTrashHandler handles the GET /photos/trash endpoint
func (app *App) listTrashHandler(w http.ResponseWriter, r *http.Request) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get all trashed photos from the database
	photos, err := app.DB.ListTrashedPhotos(ctx)
	if err != nil {
		log.Printf("Failed to list trashed photos: %v", err)
		http.Error(w, "Failed to list trashed photos", http.StatusInternalServerError)
		return
	}

	// Return the photos as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"photos": photos,
	})
}

// emptyTrashHandler handles the DELETE /photos/trash endpoint
func (app *App) emptyTrashHandler(w http.ResponseWriter, r *http.Request) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Get all trashed photos to delete their S3 objects
	photos, err := app.DB.ListTrashedPhotos(ctx)
	if err != nil {
		log.Printf("Failed to list trashed photos: %v", err)
		http.Error(w, "Failed to empty trash", http.StatusInternalServerError)
		return
	}

	// Delete all S3 objects for trashed photos
	for _, photo := range photos {
		if err := app.Storage.DeleteObject(ctx, photo.S3Key); err != nil {
			log.Printf("Failed to delete S3 object for photo %s: %v", photo.ID, err)
			// Continue deleting other photos even if one fails
		}

		// Check for thumbnails and delete them too
		if photo.Meta != nil {
			if thumbnails, ok := photo.Meta["thumbnails"].(map[string]interface{}); ok {
				for _, thumbKey := range thumbnails {
					if thumbKeyStr, ok := thumbKey.(string); ok {
						if err := app.Storage.DeleteObject(ctx, thumbKeyStr); err != nil {
							log.Printf("Failed to delete thumbnail S3 object: %v", err)
							// Continue deleting other thumbnails even if one fails
						}
					}
				}
			}
		}
	}

	// Empty the trash in the database
	count, err := app.DB.EmptyTrash(ctx)
	if err != nil {
		log.Printf("Failed to empty trash in database: %v", err)
		http.Error(w, "Failed to empty trash", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Trash emptied",
		"count":   count,
	})
}

// permanentDeletePhotoHandler handles the DELETE /photos/trash/:id endpoint
func (app *App) permanentDeletePhotoHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get the photo from the database to get its S3 key
	trashedPhotos, err := app.DB.ListTrashedPhotos(ctx)
	if err != nil {
		log.Printf("Failed to list trashed photos: %v", err)
		http.Error(w, "Failed to get photo metadata", http.StatusInternalServerError)
		return
	}

	// Find the photo with the matching ID
	var photo *db.Photo
	for i := range trashedPhotos {
		if trashedPhotos[i].ID == id {
			photo = &trashedPhotos[i]
			break
		}
	}

	if photo == nil {
		log.Printf("Photo not found in trash: %s", id)
		http.Error(w, "Photo not found in trash", http.StatusNotFound)
		return
	}

	// Delete the S3 object
	if err := app.Storage.DeleteObject(ctx, photo.S3Key); err != nil {
		log.Printf("Failed to delete S3 object: %v", err)
		http.Error(w, "Failed to delete photo from storage", http.StatusInternalServerError)
		return
	}

	// Check for thumbnails and delete them too
	if photo.Meta != nil {
		if thumbnails, ok := photo.Meta["thumbnails"].(map[string]interface{}); ok {
			for _, thumbKey := range thumbnails {
				if thumbKeyStr, ok := thumbKey.(string); ok {
					if err := app.Storage.DeleteObject(ctx, thumbKeyStr); err != nil {
						log.Printf("Failed to delete thumbnail S3 object: %v", err)
						// Continue deleting the photo even if thumbnail deletion fails
					}
				}
			}
		}
	}

	// Permanently delete the photo from the database
	if err := app.DB.PermanentlyDeletePhoto(ctx, id); err != nil {
		log.Printf("Failed to permanently delete photo: %v", err)
		http.Error(w, fmt.Sprintf("Failed to permanently delete photo: %v", err), http.StatusInternalServerError)
		return
	}

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

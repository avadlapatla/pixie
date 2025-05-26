package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a PostgreSQL database connection pool
type DB struct {
	Pool *pgxpool.Pool
}

// Config holds the database configuration
type Config struct {
	URL string
}

// New creates a new database connection pool
func New(ctx context.Context, config Config) (*DB, error) {
	// Create a connection pool
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set some reasonable defaults for the connection pool
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// InitSchema initializes the database schema
func (db *DB) InitSchema(ctx context.Context) error {
	// Create the photos table if it doesn't exist
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS photos (
			id UUID PRIMARY KEY,
			s3_key TEXT NOT NULL,
			filename TEXT,
			mime TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			meta JSONB
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create photos table: %w", err)
	}
	
	// Note: User table creation is now handled by the user.Manager

	// Check if deleted_at column exists, add if not
	var columnExists bool
	err = db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'photos' AND column_name = 'deleted_at'
		);
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check for deleted_at column: %w", err)
	}

	if !columnExists {
		_, err = db.Pool.Exec(ctx, `
			ALTER TABLE photos 
			ADD COLUMN deleted_at TIMESTAMPTZ;
		`)
		if err != nil {
			return fmt.Errorf("failed to add deleted_at column: %w", err)
		}
		log.Println("Added deleted_at column to photos table")
	}

	// Check if status column exists, add if not
	err = db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'photos' AND column_name = 'status'
		);
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check for status column: %w", err)
	}

	if !columnExists {
		_, err = db.Pool.Exec(ctx, `
			ALTER TABLE photos 
			ADD COLUMN status TEXT DEFAULT 'active';
		`)
		if err != nil {
			return fmt.Errorf("failed to add status column: %w", err)
		}
		log.Println("Added status column to photos table")
	}

	return nil
}

// SavePhoto saves a photo's metadata to the database
func (db *DB) SavePhoto(ctx context.Context, id, s3Key, filename, mime string) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO photos (id, s3_key, filename, mime)
		VALUES ($1, $2, $3, $4)
	`, id, s3Key, filename, mime)
	if err != nil {
		return fmt.Errorf("failed to insert photo: %w", err)
	}

	return nil
}

// GetPhoto retrieves a photo's metadata from the database
func (db *DB) GetPhoto(ctx context.Context, id string) (string, string, error) {
	var s3Key, mime string
	err := db.Pool.QueryRow(ctx, `
		SELECT s3_key, mime FROM photos WHERE id = $1
	`, id).Scan(&s3Key, &mime)
	if err != nil {
		return "", "", fmt.Errorf("failed to get photo: %w", err)
	}

	return s3Key, mime, nil
}

// DeletePhoto deletes a photo's metadata from the database
func (db *DB) DeletePhoto(ctx context.Context, id string) error {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM photos WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("photo not found: %s", id)
	}

	return nil
}

// Photo represents a photo in the database
type Photo struct {
	ID        string                 `json:"id"`
	S3Key     string                 `json:"s3_key"`
	Filename  string                 `json:"filename"`
	Mime      string                 `json:"mime"`
	CreatedAt time.Time              `json:"created_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
	Status    string                 `json:"status"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// TrashPhoto moves a photo to trash by updating its status and setting deleted_at
func (db *DB) TrashPhoto(ctx context.Context, id string) error {
	result, err := db.Pool.Exec(ctx, `
		UPDATE photos 
		SET status = 'trashed', deleted_at = NOW() 
		WHERE id = $1 AND status = 'active'
	`, id)
	if err != nil {
		return fmt.Errorf("failed to trash photo: %w", err)
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("photo not found or already trashed: %s", id)
	}

	return nil
}

// RestorePhoto restores a photo from trash by updating its status and clearing deleted_at
func (db *DB) RestorePhoto(ctx context.Context, id string) error {
	result, err := db.Pool.Exec(ctx, `
		UPDATE photos 
		SET status = 'active', deleted_at = NULL 
		WHERE id = $1 AND status = 'trashed'
	`, id)
	if err != nil {
		return fmt.Errorf("failed to restore photo: %w", err)
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("photo not found in trash: %s", id)
	}

	return nil
}

// EmptyTrash permanently deletes all photos in trash
func (db *DB) EmptyTrash(ctx context.Context) (int64, error) {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM photos 
		WHERE status = 'trashed'
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to empty trash: %w", err)
	}

	// Return number of photos deleted
	return result.RowsAffected(), nil
}

// PermanentlyDeletePhoto permanently deletes a specific photo from trash
func (db *DB) PermanentlyDeletePhoto(ctx context.Context, id string) error {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM photos 
		WHERE id = $1 AND status = 'trashed'
	`, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete photo: %w", err)
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("photo not found in trash: %s", id)
	}

	return nil
}

// ListTrashedPhotos retrieves all photos from trash
func (db *DB) ListTrashedPhotos(ctx context.Context) ([]Photo, error) {
	// Get column information from the database
	columnsInfo, err := db.Pool.Query(ctx, `
		SELECT column_name FROM information_schema.columns 
		WHERE table_name = 'photos'
		ORDER BY ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns info: %w", err)
	}
	defer columnsInfo.Close()

	// Build a list of existing columns
	columns := make(map[string]bool)
	var columnNames []string
	for columnsInfo.Next() {
		var colName string
		if err := columnsInfo.Scan(&colName); err != nil {
			return nil, fmt.Errorf("failed to scan column name: %w", err)
		}
		columns[colName] = true
		columnNames = append(columnNames, colName)
	}

	// Check if we have the status column, if not return empty list
	// since we can't identify trashed photos without status
	hasStatus := columns["status"]
	if !hasStatus {
		log.Println("Warning: status column doesn't exist yet, returning empty trash list")
		return []Photo{}, nil
	}

	hasDeletedAt := columns["deleted_at"]

	// Select trashed photos based on the available columns
	var sqlQuery string
	sqlQuery = `
		SELECT id, s3_key, filename, mime, created_at
	`
	// Add optional columns if they exist
	if hasDeletedAt {
		sqlQuery += `, deleted_at`
	} else {
		sqlQuery += `, NULL as deleted_at`
	}
	
	sqlQuery += `, status, meta
		FROM photos
		WHERE status = 'trashed'
	`
	
	// Order by deleted_at if it exists, otherwise by created_at
	if hasDeletedAt {
		sqlQuery += ` ORDER BY deleted_at DESC NULLS LAST`
	} else {
		sqlQuery += ` ORDER BY created_at DESC`
	}

	rows, err := db.Pool.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query trashed photos: %w", err)
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var photo Photo
		var meta []byte
		if err := rows.Scan(&photo.ID, &photo.S3Key, &photo.Filename, &photo.Mime, &photo.CreatedAt, 
			&photo.DeletedAt, &photo.Status, &meta); err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		
		// Parse the meta JSON if it's not null
		if meta != nil {
			if err := json.Unmarshal(meta, &photo.Meta); err != nil {
				return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
			}
		}
		
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trashed photos: %w", err)
	}

	return photos, nil
}

// ListPhotos retrieves all active photos from the database
func (db *DB) ListPhotos(ctx context.Context) ([]Photo, error) {
	// Get column information from the database
	columnsInfo, err := db.Pool.Query(ctx, `
		SELECT column_name FROM information_schema.columns 
		WHERE table_name = 'photos'
		ORDER BY ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns info: %w", err)
	}
	defer columnsInfo.Close()

	// Build a list of existing columns
	columns := make(map[string]bool)
	var columnNames []string
	for columnsInfo.Next() {
		var colName string
		if err := columnsInfo.Scan(&colName); err != nil {
			return nil, fmt.Errorf("failed to scan column name: %w", err)
		}
		columns[colName] = true
		columnNames = append(columnNames, colName)
	}

	var rows pgx.Rows
	// Check if we have the status and deleted_at columns
	hasStatus := columns["status"]
	hasDeletedAt := columns["deleted_at"]

	// Select photos based on the available columns
	var sqlQuery string
	if hasStatus {
		// If status column exists, only select active photos
		sqlQuery = `
			SELECT id, s3_key, filename, mime, created_at
		`
		// Add optional columns if they exist
		if hasDeletedAt {
			sqlQuery += `, deleted_at`
		} else {
			sqlQuery += `, NULL as deleted_at`
		}
		
		sqlQuery += `, status, meta
			FROM photos
			WHERE status = 'active' OR status IS NULL
			ORDER BY created_at DESC
		`
	} else {
		// If status column doesn't exist, select all photos
		sqlQuery = `
			SELECT id, s3_key, filename, mime, created_at, NULL as deleted_at, 'active' as status, meta
			FROM photos
			ORDER BY created_at DESC
		`
	}
	rows, err = db.Pool.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query photos: %w", err)
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var photo Photo
		var meta []byte
		if err := rows.Scan(&photo.ID, &photo.S3Key, &photo.Filename, &photo.Mime, &photo.CreatedAt, 
			&photo.DeletedAt, &photo.Status, &meta); err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		
		// Set default status if null
		if photo.Status == "" {
			photo.Status = "active"
		}
		
		// Parse the meta JSON if it's not null
		if meta != nil {
			if err := json.Unmarshal(meta, &photo.Meta); err != nil {
				return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
			}
		}
		
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating photos: %w", err)
	}

	return photos, nil
}

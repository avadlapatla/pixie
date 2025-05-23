package db

import (
	"context"
	"fmt"
	"time"

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
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create photos table: %w", err)
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

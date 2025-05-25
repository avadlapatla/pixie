package events

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	js nats.JetStreamContext
	// DefaultPublish is the default implementation of the Publish function
	DefaultPublish PublishFunc
)

// Init initializes the NATS connection and JetStream
func Init() error {
	// Get NATS URL from environment or use default
	natsURL := getEnv("NATS_URL", "nats://nats:4222")

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	jsc, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// Create a context with timeout for stream creation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure the PHOTO stream exists
	cfg := &nats.StreamConfig{
		Name:      "PHOTO",
		Subjects:  []string{"photo.*"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour, // 7 days
	}

	_, err = jsc.AddStream(cfg, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to create or update stream: %w", err)
	}

	// Store the JetStream context for later use
	js = jsc

	log.Println("NATS JetStream initialized successfully")
	return nil
}

// DefaultPublishImpl is the default implementation of the Publish function
func DefaultPublishImpl(ctx context.Context, subj string, data []byte) error {
	// Create a context with timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// If js is nil, log a warning and return nil
	if js == nil {
		log.Printf("Warning: NATS JetStream not initialized, message to %s not published", subj)
		return nil
	}

	// Publish the message
	_, err := js.Publish(subj, data, nats.Context(publishCtx))
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Initialize the Publish variable with the default implementation
func init() {
	// Set the default implementation
	DefaultPublish = DefaultPublishImpl
	// Set the current implementation to the default
	Publish = DefaultPublish
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

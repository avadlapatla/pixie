package events

import (
	"context"
	"log"
)

// PublishFunc is the function type for publishing events
type PublishFunc func(ctx context.Context, subj string, data []byte) error

// Publish is the function used to publish events
var Publish PublishFunc

// InitMock initializes a mock NATS connection
func InitMock() {
	log.Println("Mock NATS initialized")
	// Replace the Publish function with our mock implementation
	Publish = MockPublish
}

// MockPublish is a mock implementation of the Publish function
func MockPublish(ctx context.Context, subj string, data []byte) error {
	log.Printf("Mock publish to %s: %s", subj, string(data))
	return nil
}

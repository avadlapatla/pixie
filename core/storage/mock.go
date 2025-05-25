package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MockS3 is a mock implementation of the S3 interface for testing
type MockS3 struct {
	S3
	objects map[string][]byte
	mimes   map[string]string
	mutex   sync.RWMutex
}

// NewMock creates a new mock S3 client
func NewMock() *MockS3 {
	return &MockS3{
		objects: make(map[string][]byte),
		mimes:   make(map[string]string),
	}
}

// UploadObject uploads an object to the mock S3
func (m *MockS3) UploadObject(ctx context.Context, key string, data io.Reader, contentType string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Read all data from the reader
	content, err := ioutil.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	// Store the object and its content type
	m.objects[key] = content
	m.mimes[key] = contentType

	return nil
}

// GetObject retrieves an object from the mock S3
func (m *MockS3) GetObject(ctx context.Context, key string) (*s3.GetObjectOutput, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if the object exists
	content, ok := m.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}

	// Get the content type
	contentType := m.mimes[key]

	// Create a reader for the content
	reader := bytes.NewReader(content)

	// Create a mock GetObjectOutput
	output := &s3.GetObjectOutput{
		Body:        io.NopCloser(reader),
		ContentType: &contentType,
	}

	return output, nil
}

// DeleteObject deletes an object from the mock S3
func (m *MockS3) DeleteObject(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if the object exists
	if _, ok := m.objects[key]; !ok {
		return fmt.Errorf("object not found: %s", key)
	}

	// Delete the object and its content type
	delete(m.objects, key)
	delete(m.mimes, key)

	return nil
}

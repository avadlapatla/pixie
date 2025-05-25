package db

import (
	"context"
	"sync"
	"time"
)

// MockDB is a mock implementation of the DB interface for testing
type MockDB struct {
	DB
	photos map[string]*Photo
	mutex  sync.RWMutex
}

// NewMock creates a new mock database
func NewMock() *MockDB {
	return &MockDB{
		photos: make(map[string]*Photo),
	}
}

// SavePhoto saves a photo's metadata to the mock database
func (m *MockDB) SavePhoto(ctx context.Context, id, s3Key, filename, mime string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.photos[id] = &Photo{
		ID:        id,
		S3Key:     s3Key,
		Filename:  filename,
		Mime:      mime,
		CreatedAt: time.Now(),
	}

	return nil
}

// GetPhoto retrieves a photo's metadata from the mock database
func (m *MockDB) GetPhoto(ctx context.Context, id string) (string, string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	photo, ok := m.photos[id]
	if !ok {
		return "", "", ErrPhotoNotFound{ID: id}
	}

	return photo.S3Key, photo.Mime, nil
}

// DeletePhoto deletes a photo's metadata from the mock database
func (m *MockDB) DeletePhoto(ctx context.Context, id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.photos[id]; !ok {
		return ErrPhotoNotFound{ID: id}
	}

	delete(m.photos, id)
	return nil
}

// ListPhotos retrieves all photos from the mock database
func (m *MockDB) ListPhotos(ctx context.Context) ([]Photo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	photos := make([]Photo, 0, len(m.photos))
	for _, photo := range m.photos {
		photos = append(photos, *photo)
	}

	return photos, nil
}

// ErrPhotoNotFound is returned when a photo is not found
type ErrPhotoNotFound struct {
	ID string
}

// Error implements the error interface
func (e ErrPhotoNotFound) Error() string {
	return "photo not found: " + e.ID
}

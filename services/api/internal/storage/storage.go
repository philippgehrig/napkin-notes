package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Storage defines the file storage interface.
type Storage interface {
	Save(path string, reader io.Reader) error
	Get(path string) (io.ReadCloser, error)
	Delete(path string) error
}

// LocalStorage implements Storage using the local filesystem.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new LocalStorage with the given base directory.
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// Save writes data from reader to the given path relative to basePath.
func (s *LocalStorage) Save(path string, reader io.Reader) error {
	fullPath := filepath.Join(s.basePath, path)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("storage: failed to create directory: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("storage: failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("storage: failed to write file: %w", err)
	}

	return nil
}

// Get opens the file at the given path relative to basePath for reading.
func (s *LocalStorage) Get(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)

	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage: file not found: %s", path)
		}
		return nil, fmt.Errorf("storage: failed to open file: %w", err)
	}

	return f, nil
}

// Delete removes the file at the given path relative to basePath.
func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("storage: file not found: %s", path)
		}
		return fmt.Errorf("storage: failed to delete file: %w", err)
	}

	return nil
}

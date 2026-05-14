package storage

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLocalStorage_SaveAndGet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewLocalStorage(tmpDir)

	content := "hello, storage!"
	err = s.Save("fonts/test-file.bin", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reader, err := s.Get("fonts/test-file.bin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if !bytes.Equal(data, []byte(content)) {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewLocalStorage(tmpDir)

	content := "delete me"
	err = s.Save("fonts/delete-me.bin", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err = s.Delete("fonts/delete-me.bin")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify file is gone
	_, err = s.Get("fonts/delete-me.bin")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestLocalStorage_GetAfterDeleteFails(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewLocalStorage(tmpDir)

	err = s.Save("fonts/ephemeral.bin", strings.NewReader("ephemeral"))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	err = s.Delete("fonts/ephemeral.bin")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = s.Get("fonts/ephemeral.bin")
	if err == nil {
		t.Fatal("expected error getting deleted file, got nil")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("expected 'file not found' error, got: %v", err)
	}
}

func TestLocalStorage_GetNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewLocalStorage(tmpDir)

	_, err = s.Get("does-not-exist.bin")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

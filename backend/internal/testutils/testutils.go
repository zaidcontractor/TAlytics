package testutils

import (
	"os"
	"path/filepath"
	"testing"
)

// SetupTestDBPath returns a temporary database path for testing
func SetupTestDBPath(t *testing.T) string {
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "test.db")
}

// CleanupTestDB removes all data from test database
// This function should be called with the database connection from the test
func CleanupTestDB(t *testing.T, db interface{}) {
	// This is a helper that tests can use with their own DB connection
	// We don't import database here to avoid cycles
	t.Log("CleanupTestDB called - implement cleanup in test if needed")
}

// GetEnvOrDefault returns environment variable or default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


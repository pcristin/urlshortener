package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	opts := NewOptions()
	assert.NotNil(t, opts)
}

func TestGetServerURL(t *testing.T) {
	opts := NewOptions()

	// Test environment variable
	os.Setenv("SERVER_ADDRESS", "localhost:8888")
	opts.LoadEnvVariables()
	assert.Equal(t, "localhost:8888", opts.GetServerURL())

	// Test default value
	os.Unsetenv("SERVER_ADDRESS")
	opts = NewOptions()
	opts.LoadEnvVariables()
	assert.Equal(t, "localhost:8080", opts.GetServerURL())
}

func TestGetBaseURL(t *testing.T) {
	opts := NewOptions()

	// Test environment variable
	os.Setenv("BASE_URL", "http://example.com")
	opts.LoadEnvVariables()
	assert.Equal(t, "http://example.com", opts.GetBaseURL())

	// Test empty value
	os.Unsetenv("BASE_URL")
	opts = NewOptions()
	opts.LoadEnvVariables()
	assert.Equal(t, "", opts.GetBaseURL())
}

func TestGetPathToSavedData(t *testing.T) {
	opts := NewOptions()

	// Test environment variable
	os.Setenv("FILE_STORAGE_PATH", "/tmp/data.json")
	opts.LoadEnvVariables()
	assert.Equal(t, "/tmp/data.json", opts.GetPathToSavedData())

	// Test default value
	os.Unsetenv("FILE_STORAGE_PATH")
	opts = NewOptions()
	opts.LoadEnvVariables()
	assert.Equal(t, "saved_data.json", opts.GetPathToSavedData())
}

func TestGetDatabaseDSN(t *testing.T) {
	opts := NewOptions()

	// Test environment variable
	os.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
	opts.LoadEnvVariables()
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", opts.GetDatabaseDSN())

	// Test empty value
	os.Unsetenv("DATABASE_DSN")
	opts = NewOptions()
	opts.LoadEnvVariables()
	assert.Equal(t, "", opts.GetDatabaseDSN())
}

func TestGetSecret(t *testing.T) {
	opts := NewOptions()

	// Test environment variable
	os.Setenv("SECRET_URL_SERVICE", "test-secret")
	opts.LoadEnvVariables()
	assert.Equal(t, "test-secret", opts.GetSecret())

	// Test empty value
	os.Unsetenv("SECRET_URL_SERVICE")
	opts = NewOptions()
	opts.LoadEnvVariables()
	assert.Equal(t, "", opts.GetSecret())
}

func TestLoadEnvVariables(t *testing.T) {
	// Set up test environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:9999")
	os.Setenv("BASE_URL", "http://test.com")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/test.json")
	os.Setenv("DATABASE_DSN", "postgres://test:test@localhost:5432/testdb")
	os.Setenv("SECRET_URL_SERVICE", "super-secret")

	// Create options and load environment variables
	opts := NewOptions()
	opts.LoadEnvVariables()

	// Verify all values were loaded correctly
	assert.Equal(t, "localhost:9999", opts.GetServerURL())
	assert.Equal(t, "http://test.com", opts.GetBaseURL())
	assert.Equal(t, "/tmp/test.json", opts.GetPathToSavedData())
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", opts.GetDatabaseDSN())
	assert.Equal(t, "super-secret", opts.GetSecret())

	// Clean up
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("SECRET_URL_SERVICE")
}

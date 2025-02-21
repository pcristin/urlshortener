package config

import (
	"flag"
	"os"
)

type Options struct {
	serverURL       string
	baseURL         string
	pathToSavedData string
	databaseDSN     string
}

// NewOptions creates a new Options instance
func NewOptions() *Options {
	return &Options{
		serverURL:       "localhost:8080",
		baseURL:         "",
		pathToSavedData: "saved_data.json",
		databaseDSN:     "",
	}
}

// ParseFlags parses command line flags and environment variables
func (o *Options) ParseFlags() {
	flag.StringVar(&o.serverURL, "a", o.serverURL, "address and port to run server")
	flag.StringVar(&o.baseURL, "b", o.baseURL, "server url and short url path to redirect")
	flag.StringVar(&o.pathToSavedData, "f", o.pathToSavedData, "path to json file with saved data")
	flag.StringVar(&o.databaseDSN, "d", o.databaseDSN, "string of db connection params")

	flag.Parse()

	if valueEnvServerURL, foundEnvServerURL := os.LookupEnv("SERVER_ADDRESS"); foundEnvServerURL && valueEnvServerURL != "" {
		o.serverURL = os.Getenv("SERVER_ADDRESS")
	}

	if valueBaseURL, foundBaseURL := os.LookupEnv("BASE_URL"); foundBaseURL && valueBaseURL != "" {
		o.baseURL = os.Getenv("BASE_URL")
	}

	if valueEnvPathToJSON, foundEnvPathToJSON := os.LookupEnv("FILE_STORAGE_PATH"); foundEnvPathToJSON && valueEnvPathToJSON != "" {
		o.pathToSavedData = os.Getenv("FILE_STORAGE_PATH")
	}

	if valueDatabaseDSN, foundDatabaseDSN := os.LookupEnv("DATABASE_DSN"); foundDatabaseDSN && valueDatabaseDSN != "" {
		o.databaseDSN = os.Getenv("DATABASE_DSN")
	}
}

// GetServerURL returns the server URL
func (o *Options) GetServerURL() string {
	return o.serverURL
}

// GetBaseURL returns the base URL
func (o *Options) GetBaseURL() string {
	return o.baseURL
}

// GetPathToSavedData returns the path to JSON file with saved data
func (o *Options) GetPathToSavedData() string {
	return o.pathToSavedData
}

// GetDatabaseDSN returns the
func (o *Options) GetDatabaseDSN() string {
	return o.databaseDSN
}

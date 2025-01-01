package config

import (
	"flag"
	"log"
	"os"
)

// OptionsConfigger defines the interface for configuration
type OptionsConfigger interface {
	GetServerURL() string
	GetBaseURL() string
	ParseFlags()
}

type Options struct {
	serverURL string
	baseURL   string
}

// NewOptions creates a new Options instance
func NewOptions() OptionsConfigger {
	return &Options{
		serverURL: "localhost:8080",
		baseURL:   "",
	}
}

// ParseFlags parses command line flags and environment variables
func (o *Options) ParseFlags() {
	flag.StringVar(&o.serverURL, "a", o.serverURL, "address and port to run server")
	flag.StringVar(&o.baseURL, "b", o.baseURL, "server url and short url path to redirect")

	flag.Parse()

	if valueEnvServerURL, foundEnvServerURL := os.LookupEnv("SERVER_ADDRESS"); foundEnvServerURL && valueEnvServerURL != "" {
		o.serverURL = os.Getenv("SERVER_ADDRESS")
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		o.baseURL = baseURL
	}

	log.Printf("Configuration: ServerURL=%s, BaseURL=%s", o.serverURL, o.baseURL)
}

// GetServerURL returns the server URL
func (o *Options) GetServerURL() string {
	return o.serverURL
}

// GetBaseURL returns the base URL
func (o *Options) GetBaseURL() string {
	return o.baseURL
}

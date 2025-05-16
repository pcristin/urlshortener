package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

// Options holds configuration settings for the URL shortener service
type Options struct {
	serverURL       string
	baseURL         string
	pathToSavedData string
	databaseDSN     string
	secret          string
	enableHTTPS     bool
	config          string
}

type ConfigFile struct {
	ServerURL       string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStorageData string `json:"file_storage_data"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// NewOptions creates a new Options instance
func NewOptions() *Options {
	return &Options{
		serverURL:       "localhost:8080",
		baseURL:         "",
		pathToSavedData: "saved_data.json",
		databaseDSN:     "",
		secret:          "",
		enableHTTPS:     false,
		config:          "",
	}
}

// ParseFlags parses command line flags and environment variables
func (o *Options) ParseFlags() {
	flag.StringVar(&o.serverURL, "a", o.serverURL, "address and port to run server")
	flag.StringVar(&o.baseURL, "b", o.baseURL, "server url and short url path to redirect")
	flag.StringVar(&o.pathToSavedData, "f", o.pathToSavedData, "path to json file with saved data")
	flag.StringVar(&o.databaseDSN, "d", o.databaseDSN, "string of db connection params")
	flag.BoolVar(&o.enableHTTPS, "s", o.enableHTTPS, "enable https")

	flag.Parse()

	o.LoadEnvVariables()
}

// LoadEnvVariables loads configuration from environment variables
func (o *Options) LoadEnvVariables() {

	if valueConfig, foundConfig := os.LookupEnv("CONFIG"); foundConfig && valueConfig != "" {
		o.config = os.Getenv("CONFIG")
	}
	if o.config != "" {
		jsonFile, err := os.ReadFile(o.config)
		if err != nil {
			log.Fatalf("yamlFile.Get err #%v ", err)
		}
		var configFile ConfigFile
		err = json.Unmarshal(jsonFile, &configFile)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	}

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

	if valueSecret, foundSecret := os.LookupEnv("SECRET_URL_SERVICE"); foundSecret && valueSecret != "" {
		o.secret = os.Getenv("SECRET_URL_SERVICE")
	}

	if valueEnableHTTPS, foundEnableHTTPS := os.LookupEnv("ENABLE_HTTPS"); foundEnableHTTPS && valueEnableHTTPS != "" {
		o.enableHTTPS = os.Getenv("ENABLE_HTTPS") == "true"
	}
}

// ParseConfigJSONFile parses the config file into ENV variables
func (o *Options) ParseConfigJSONFile() {
	var configFile ConfigFile
	jsonFile, err := os.ReadFile(o.config)
	if err != nil {
		log.Fatalf("yamlFile.Get err #%v ", err)
	}
	err = json.Unmarshal(jsonFile, &configFile)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	// Set ENV based on config file
	os.Setenv("SERVER_ADDRESS", configFile.ServerURL)
	os.Setenv("BASE_URL", configFile.BaseURL)
	os.Setenv("FILE_STORAGE_PATH", configFile.FileStorageData)
	os.Setenv("DATABASE_DSN", configFile.DatabaseDSN)
	os.Setenv("ENABLE_HTTPS", strconv.FormatBool(configFile.EnableHTTPS))
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

// GetDatabaseDSN returns the database connection string
func (o *Options) GetDatabaseDSN() string {
	return o.databaseDSN
}

// GetSecret returns the secret key for URL service
func (o *Options) GetSecret() string {
	return o.secret
}

// GetEnableHTTPS returns the enable HTTPS flag
func (o *Options) GetEnableHTTPS() bool {
	return o.enableHTTPS
}

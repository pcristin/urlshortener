package config

import (
	"flag"
	"log"
	"os"
)

type Options struct {
	ServerURL string
	BaseURL   string
}

var OptionsFlag Options

func FlagParse() {
	flag.StringVar(&OptionsFlag.ServerURL, "a", OptionsFlag.ServerURL, "address and port to run server")
	flag.StringVar(&OptionsFlag.BaseURL, "b", "", "server url and short url path to redirect")

	flag.Parse()

	if valueEnvServerURL, foundEnvServerURL := os.LookupEnv("SERVER_ADDRESS"); foundEnvServerURL && valueEnvServerURL != "" {
		OptionsFlag.ServerURL = os.Getenv("SERVER_ADDRESS")
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		OptionsFlag.BaseURL = envBaseURL
	}

	log.Printf("After parsing - ServerURL: %s\r\n", OptionsFlag.ServerURL)
	log.Printf("After parsing - BaseURL: %s\r\n", OptionsFlag.BaseURL)
}

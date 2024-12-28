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

var OptionsFlag = Options{
	ServerURL: "localhost:8080", // Set default value here
}

func FlagParse() {
	// Print before parsing
	log.Printf("Before parsing - ServerURL: %s\r\n", OptionsFlag.ServerURL)
	log.Printf("Before parsing - BaseURL: %s\r\n", OptionsFlag.BaseURL)

	flag.StringVar(&OptionsFlag.ServerURL, "a", OptionsFlag.ServerURL, "address and port to run server")
	flag.StringVar(&OptionsFlag.BaseURL, "b", "", "server url and short url path to redirect")

	flag.Parse()

	// For cases when ENV variable is set, change the priority
	for envServerURL := os.Getenv("SERVER_ADDRESS"); envServerURL != ""; {
		OptionsFlag.ServerURL = envServerURL
	}

	for envBaseURL := os.Getenv("BASE_URL"); envBaseURL != ""; {
		OptionsFlag.BaseURL = envBaseURL
	}

	// Print after parsing
	log.Printf("After parsing - ServerURL: %s\r\n", OptionsFlag.ServerURL)
	log.Printf("After parsing - BaseURL: %s\r\n", OptionsFlag.BaseURL)
}

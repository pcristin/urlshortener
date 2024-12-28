package config

import (
	"flag"
	"log"
	"os"
)

type Options struct {
	ServerURL string
	BaseUrl   string
}

var OptionsFlag = Options{
	ServerURL: "localhost:8080", // Set default value here
}

func FlagParse() {
	// Print before parsing
	log.Printf("Before parsing - ServerURL: %s\n", OptionsFlag.ServerURL)
	log.Printf("Before parsing - BaseURL: %s\n", OptionsFlag.BaseUrl)

	flag.StringVar(&OptionsFlag.ServerURL, "a", OptionsFlag.ServerURL, "address and port to run server")
	flag.StringVar(&OptionsFlag.BaseUrl, "b", "", "server url and short url path to redirect")

	flag.Parse()

	// For cases when ENV variable is set, change the priority
	for envServerUrl := os.Getenv("SERVER_ADDRESS"); envServerUrl != ""; {
		OptionsFlag.ServerURL = envServerUrl
	}

	for envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != ""; {
		OptionsFlag.BaseUrl = envBaseUrl
	}

	// Print after parsing
	log.Printf("After parsing - ServerURL: %s\n", OptionsFlag.ServerURL)
	log.Printf("After parsing - BaseURL: %s\n", OptionsFlag.BaseUrl)
}

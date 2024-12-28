package config

import (
	"flag"
	"log"
	"os"
	"strings"
)

type Options struct {
	ServerURL string
	BaseURL   string
}

var OptionsFlag = Options{
	ServerURL: "localhost:8080", // Default value
}

func FlagParse() {
	flag.StringVar(&OptionsFlag.ServerURL, "a", OptionsFlag.ServerURL, "address and port to run server")
	flag.StringVar(&OptionsFlag.BaseURL, "b", "", "server url and short url path to redirect")

	flag.Parse()

	if envServerURL := os.Getenv("SERVER_ADDRESS"); envServerURL != "" {
		OptionsFlag.ServerURL = cleanServerURL(envServerURL)
	} else {
		OptionsFlag.ServerURL = cleanServerURL(OptionsFlag.ServerURL)
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		OptionsFlag.BaseURL = envBaseURL
	}

	log.Printf("After parsing - ServerURL: %s\r\n", OptionsFlag.ServerURL)
	log.Printf("After parsing - BaseURL: %s\r\n", OptionsFlag.BaseURL)
}

// Removes protocol prefix if present
func cleanServerURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}

package config

import (
	"flag"
	"log"
)

type Options struct {
	ServerURL string
	ShortURL  string
}

var OptionsFlag = Options{
	ServerURL: "localhost:8080", // Set default value here
}

func FlagParse() {
	// Print before parsing
	log.Printf("Before parsing - ServerURL: %s\n", OptionsFlag.ServerURL)

	flag.StringVar(&OptionsFlag.ServerURL, "a", OptionsFlag.ServerURL, "address and port to run server")
	flag.StringVar(&OptionsFlag.ShortURL, "b", "", "server url and short url path to redirect")

	flag.Parse()

	// Print after parsing
	log.Printf("After parsing - ServerURL: %s\n", OptionsFlag.ServerURL)
}

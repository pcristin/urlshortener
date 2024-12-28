package config

import (
	"flag"
)

type Options struct {
	ServerURL string
	ShortURL  string
}

var OptionsFlag Options

func FlagParse() {
	options := new(Options)

	flag.StringVar(&options.ServerURL, "a", "localhost:8888", "address and port to run server")
	flag.StringVar(&options.ShortURL, "b", "http://localhost:8000/qsd54gFg", "server url and short url path to redirect")

	flag.Parse()
}

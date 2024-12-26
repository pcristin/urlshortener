package main

import (
	"log"

	"github.com/fasthttp/router"
	"github.com/pcristin/urlshortener/internal/app"
	"github.com/valyala/fasthttp"
)

func main() {
	r := router.New()
	r.POST("/", app.FastHTTPEncodeURLHandler)
	r.GET("/{id}", app.FastHTTPDecodeURLHandler)

	if err := fasthttp.ListenAndServe("localhost:8080", r.Handler); err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}

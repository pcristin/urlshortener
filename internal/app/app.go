package app

import (
	"log"

	uu "github.com/pcristin/urlshortener/internal/urlutils"
	"github.com/valyala/fasthttp"
)

func FastHTTPEncodeURLHandler(ctx *fasthttp.RequestCtx) {
	// Setting up response headers
	ctx.Response.Header.Reset()
	ctx.SetContentType("text/plain")

	longURL := string(ctx.PostBody())

	if string(ctx.Method()) != "POST" || string(ctx.Host()) != "localhost:8080" || !uu.URLCheck(longURL) {
		ctx.Error("bad request", fasthttp.StatusBadRequest)
		return
	}

	token := uu.EncodeURL(longURL)
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetBody([]byte("http://" + string(ctx.Host()) + "/" + string(token)))
}

func FastHTTPDecodeURLHandler(ctx *fasthttp.RequestCtx) {
	// Reseting headers
	ctx.Response.Header.Reset()

	if string(ctx.Method()) != "GET" || string(ctx.Host()) != "localhost:8080" {
		ctx.Error("bad request", fasthttp.StatusBadRequest)
		return
	}
	token := ctx.UserValue("id").(string)
	log.Println(token)
	longURL, err := uu.DecodeURL(token)
	if err != nil {
		ctx.Error("bad request", fasthttp.StatusBadRequest)
		return
	}
	ctx.Response.Header.Set("Location", longURL)
	ctx.SetStatusCode(fasthttp.StatusTemporaryRedirect)
}

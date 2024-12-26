package app

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestFastHTTPEncodeURLHandler(t *testing.T) {
	type reqParams struct {
		method       string
		sentData     string
		sentHost     string
		expectedCode int
	}

	tt := []struct {
		name      string
		reqParams reqParams
	}{
		{
			name: "post wo data",
			reqParams: reqParams{
				method:       "POST",
				sentData:     "",
				sentHost:     "localhost:8080",
				expectedCode: fasthttp.StatusBadRequest,
			},
		},
		{
			name: "post with data",
			reqParams: reqParams{
				method:       "POST",
				sentData:     "https://google.com",
				sentHost:     "localhost:8080",
				expectedCode: fasthttp.StatusCreated,
			},
		},
		{
			name: "post with strange data",
			reqParams: reqParams{
				method:       "POST",
				sentData:     "app",
				sentHost:     "localhost:8080",
				expectedCode: fasthttp.StatusBadRequest,
			},
		},
		{
			name: "get request",
			reqParams: reqParams{
				method:       "GET",
				sentHost:     "localhost:8080",
				expectedCode: fasthttp.StatusBadRequest,
			},
		},
		{
			name: "wrong host request",
			reqParams: reqParams{
				method:       "POST",
				sentData:     "yandex.mail.ru",
				sentHost:     "",
				expectedCode: fasthttp.StatusBadRequest,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(tc.reqParams.method)
			ctx.Request.SetHost(tc.reqParams.sentHost)
			ctx.Request.SetBody([]byte(tc.reqParams.sentData))

			// Call handler
			FastHTTPEncodeURLHandler(ctx)

			// Test status codes
			assert.Equal(t, tc.reqParams.expectedCode, ctx.Response.StatusCode())
			if tc.reqParams.sentData != "" {
				fmt.Printf("POST data (long url): %s\n", tc.reqParams.sentData)
			}
		})
	}
}

func TestFastHTTPDecodeURLHandler(t *testing.T) {
	type reqParams struct {
		method       string
		sentHost     string
		token        string
		expectedCode int
	}

	tt := []struct {
		name      string
		reqParams reqParams
	}{
		{
			name: "post method",
			reqParams: reqParams{
				method:       "POST",
				sentHost:     "localhost:8080",
				token:        "2gr",
				expectedCode: fasthttp.StatusBadRequest,
			},
		},
		{
			name: "empty id",
			reqParams: reqParams{
				method:       "GET",
				sentHost:     "localhost:8080",
				token:        "",
				expectedCode: fasthttp.StatusBadRequest,
			},
		},
		{
			name: "wrong host",
			reqParams: reqParams{
				method:       "GET",
				sentHost:     "",
				token:        "greg1451",
				expectedCode: fasthttp.StatusBadRequest,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(tc.reqParams.method)
			ctx.Request.SetHost(tc.reqParams.sentHost)
			ctx.SetUserValue("id", tc.reqParams.token)

			// Call handler
			FastHTTPDecodeURLHandler(ctx)

			// Test status codes
			assert.Equal(t, tc.reqParams.expectedCode, ctx.Response.StatusCode())
			fmt.Printf("Token value: %v\n", tc.reqParams.token)
		})
	}
}

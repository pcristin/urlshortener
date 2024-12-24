package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeURLHandler(t *testing.T) {
	type reqParams struct {
		method          string
		sentData        string
		sentContentType string
		sentHost        string
		expectedCode    int
	}

	tt := []struct {
		name      string
		reqParams reqParams
	}{
		{
			name: "post wo data",
			reqParams: reqParams{
				method:          http.MethodPost,
				sentData:        "",
				sentContentType: "text/plain; charset=utf-8",
				sentHost:        "localhost:8080",
				expectedCode:    http.StatusBadRequest,
			},
		},
		{
			name: "post with data",
			reqParams: reqParams{
				method:          http.MethodPost,
				sentData:        "https://google.com",
				sentContentType: "text/plain; charset=utf-8",
				sentHost:        "localhost:8080",
				expectedCode:    http.StatusCreated,
			},
		},
		{
			name: "post with strange data",
			reqParams: reqParams{
				method:          http.MethodPost,
				sentData:        "app",
				sentContentType: "text/plain; charset=utf-8",
				sentHost:        "localhost:8080",
				expectedCode:    http.StatusBadRequest,
			},
		},
		{
			name: "get request",
			reqParams: reqParams{
				method:          http.MethodGet,
				sentHost:        "localhost:8080",
				sentContentType: "text/plain; charset=utf-8",
				expectedCode:    http.StatusBadRequest,
			},
		},
		{
			name: "put request",
			reqParams: reqParams{
				method:          http.MethodPut,
				sentHost:        "localhost:8080",
				sentContentType: "text/plain; charset=utf-8",
				expectedCode:    http.StatusBadRequest,
			},
		},
		{
			name: "wrong host request",
			reqParams: reqParams{
				method:          http.MethodPost,
				sentData:        "yandex.mail.ru",
				sentHost:        "",
				sentContentType: "text/plain; charset=utf-8",
				expectedCode:    http.StatusBadRequest,
			},
		},
		{
			name: "wrong content type",
			reqParams: reqParams{
				method:          http.MethodPost,
				sentData:        "yandex.mail.ru",
				sentHost:        "localhost:8080",
				sentContentType: "application/json",
				expectedCode:    http.StatusBadRequest,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init request
			req := httptest.NewRequest(http.MethodPost, "localhost:8080", strings.NewReader(tc.reqParams.sentData))
			req.Host = tc.reqParams.sentHost
			req.Header.Set("Content-Type", tc.reqParams.sentContentType)

			// Init recorder (response writer)
			wr := httptest.NewRecorder()
			EncodeURLHandler(wr, req)

			// Init result
			res := wr.Result()

			// Close request connection
			defer req.Body.Close()

			//Test status codes
			if tc.reqParams.sentData == "" {
				assert.Equal(t, res.StatusCode, tc.reqParams.expectedCode)
				return
			} else {
				fmt.Printf("POST data (long url): %s\n", tc.reqParams.sentData)
				assert.Equal(t, res.StatusCode, tc.reqParams.expectedCode)
				return
			}

		})
	}
}

func TestDecodeURLHandler(t *testing.T) {
	type reqParams struct {
		method       string
		sentHost     string
		sentPath     string
		expectedCode int
		urlStorage   map[string]string
	}

	var URLStorage = make(map[string]string)

	URLStorage["f12rw2t"] = "https://dzen.ru"

	tt := []struct {
		name      string
		reqParams reqParams
	}{
		{
			name: "post method",
			reqParams: reqParams{
				method:       http.MethodPost,
				sentHost:     "localhost:8080",
				sentPath:     "/2gr",
				expectedCode: http.StatusBadRequest,
				urlStorage:   URLStorage,
			},
		},
		{
			name: "empty id",
			reqParams: reqParams{
				method:       http.MethodGet,
				sentHost:     "localhost:8080",
				sentPath:     "/",
				expectedCode: http.StatusBadRequest,
				urlStorage:   URLStorage,
			},
		},
		{
			name: "empty id",
			reqParams: reqParams{
				method:       http.MethodGet,
				sentHost:     "",
				sentPath:     "/greg1451",
				expectedCode: http.StatusBadRequest,
				urlStorage:   URLStorage,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init request
			req := httptest.NewRequest(tc.reqParams.method, "localhost:8080", nil)
			req.URL.Path = tc.reqParams.sentPath
			req.Host = tc.reqParams.sentHost

			// Init recorder (response writer)
			wr := httptest.NewRecorder()
			DecodeURLHandler(wr, req)

			// Init result object
			res := wr.Result()

			// Closing connection
			defer req.Body.Close()

			// Test status codes
			assert.Equal(t, tc.reqParams.expectedCode, res.StatusCode)
			fmt.Printf("Raw Path value: %v", req.URL.Path)

		})
	}
}

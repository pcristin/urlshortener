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
		method       string
		sentData     string
		expectedCode int
	}

	tt := []struct {
		name      string
		reqParams reqParams
	}{
		{
			name: "post wo data",
			reqParams: reqParams{
				method:       http.MethodPost,
				sentData:     "",
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name: "post with data",
			reqParams: reqParams{
				method:       http.MethodPost,
				sentData:     "https://google.com",
				expectedCode: http.StatusCreated,
			},
		},
		{
			name: "post with strange data",
			reqParams: reqParams{
				method:       http.MethodPost,
				sentData:     "app",
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name: "get request",
			reqParams: reqParams{
				method:       http.MethodGet,
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name: "put request",
			reqParams: reqParams{
				method:       http.MethodPut,
				expectedCode: http.StatusBadRequest,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init request
			req := httptest.NewRequest(http.MethodPost, "localhost:8080", strings.NewReader(tc.reqParams.sentData))

			// Init recorder (response writer)
			wr := httptest.NewRecorder()
			EncodeURLHandler(wr, req)

			// Init result
			res := wr.Result()

			// Close request connection
			defer res.Body.Close()

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
		method string
		// sentHost     string
		sentPath     string
		expectedCode int
	}

	tt := []struct {
		name      string
		reqParams reqParams
	}{
		{
			name: "post method",
			reqParams: reqParams{
				method:       http.MethodPost,
				sentPath:     "/2gr",
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty id",
			reqParams: reqParams{
				method:       http.MethodGet,
				sentPath:     "/",
				expectedCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty id",
			reqParams: reqParams{
				method:       http.MethodGet,
				sentPath:     "/greg1451",
				expectedCode: http.StatusBadRequest,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init request
			req := httptest.NewRequest(tc.reqParams.method, "localhost:8080", nil)
			req.URL.Path = tc.reqParams.sentPath

			// Init recorder (response writer)
			wr := httptest.NewRecorder()
			DecodeURLHandler(wr, req)

			// Init result object
			res := wr.Result()

			// Closing connection
			defer res.Body.Close()

			// Test status codes
			assert.Equal(t, tc.reqParams.expectedCode, res.StatusCode)
			fmt.Printf("Raw Path value: %v\n", req.URL.Path)

		})
	}
}

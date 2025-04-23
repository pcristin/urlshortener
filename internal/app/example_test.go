package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pcristin/urlshortener/internal/models"
)

func ExampleHandlerInterface_EncodeURLHandler() {
	// This example demonstrates how to shorten a URL using the plain text endpoint

	// The URL to be shortened
	longURL := "https://github.com/pcristin/urlshortener"

	// Create a new request to the URL shortening endpoint
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/",
		bytes.NewBuffer([]byte(longURL)))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Send the request
	client := &http.Client{
		// Don't follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and display the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// The response will be the shortened URL
	fmt.Printf("Status: %d, Shortened URL: %s\n", resp.StatusCode, string(body))

	// Output:
	// Status: 201, Shortened URL: http://localhost:8080/AbCdEf
}

func ExampleHandler_APIEncodeHandler() {
	// This example demonstrates how to shorten a URL using the JSON API endpoint

	// Create the request body
	reqBody := models.Request{
		URL: "https://github.com/pcristin/urlshortener",
	}

	// Convert the request to JSON
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Error creating JSON: %v", err)
	}

	// Create a new request to the API endpoint
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var response models.Response
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	fmt.Printf("Status: %d, Shortened URL: %s\n", resp.StatusCode, response.Result)

	// Output:
	// Status: 201, Shortened URL: http://localhost:8080/AbCdEf
}

func ExampleHandler_APIEncodeBatchHandler() {
	// This example demonstrates how to shorten multiple URLs in a batch

	// Create the batch request
	batchReq := models.BatchRequest{
		{
			CorrelationID: "1",
			OriginalURL:   "https://github.com/pcristin/urlshortener",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://golang.org",
		},
	}

	// Convert the request to JSON
	jsonBody, err := json.Marshal(batchReq)
	if err != nil {
		log.Fatalf("Error creating JSON: %v", err)
	}

	// Create a new request to the batch API endpoint
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/batch",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var response models.BatchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	// Display the batch results
	fmt.Printf("Status: %d\n", resp.StatusCode)
	for _, item := range response {
		fmt.Printf("Correlation ID: %s, Shortened URL: %s\n",
			item.CorrelationID, item.ShortURL)
	}

	// Output:
	// Status: 201
	// Correlation ID: 1, Shortened URL: http://localhost:8080/AbCdEf
	// Correlation ID: 2, Shortened URL: http://localhost:8080/GhIjKl
}

func ExampleHandler_DecodeURLHandler() {
	// This example demonstrates how to access a shortened URL and get redirected

	// Create a request to access a shortened URL
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/AbCdEf", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Send the request with a client that doesn't follow redirects automatically
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status and location header
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Location: %s\n", resp.Header.Get("Location"))

	// Output:
	// Status: 307
	// Location: https://github.com/pcristin/urlshortener
}

func ExampleHandler_GetUserURLsHandler() {
	// This example demonstrates how to get all URLs shortened by a user

	// Create a request to get the user's URLs
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/api/user/urls", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Add authentication cookie if needed
	// req.AddCookie(&http.Cookie{Name: "auth_token", Value: "example-token"})

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// If the response is successful and not empty
	if resp.StatusCode == http.StatusOK && len(body) > 0 {
		var urls []models.URLStorageNode
		if err := json.Unmarshal(body, &urls); err != nil {
			log.Fatalf("Error parsing response: %v", err)
		}

		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Number of URLs: %d\n", len(urls))
		for i, url := range urls {
			if i < 2 { // Show just a couple of examples
				fmt.Printf("Original: %s, Short: %s\n",
					url.OriginalURL, url.ShortURL)
			}
		}
	} else if resp.StatusCode == http.StatusNoContent {
		fmt.Printf("Status: %d (No URLs found)\n", resp.StatusCode)
	}

	// Output:
	// Status: 200
	// Number of URLs: 2
	// Original: https://github.com/pcristin/urlshortener, Short: http://localhost:8080/AbCdEf
	// Original: https://golang.org, Short: http://localhost:8080/GhIjKl
}

func ExampleHandler_DeleteUserURLsHandler() {
	// This example demonstrates how to delete URLs

	// List of URL IDs to delete
	urlIDs := []string{"AbCdEf", "GhIjKl"}

	// Convert the list to JSON
	jsonBody, err := json.Marshal(urlIDs)
	if err != nil {
		log.Fatalf("Error creating JSON: %v", err)
	}

	// Create a request to delete the URLs
	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/user/urls",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add authentication cookie if needed
	// req.AddCookie(&http.Cookie{Name: "auth_token", Value: "example-token"})

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", strings.TrimSpace(string(body)))
	} else {
		fmt.Printf("Successfully marked %d URLs for deletion\n", len(urlIDs))
	}

	// Output:
	// Status: 202
	// Successfully marked 2 URLs for deletion
}

func ExampleHandler_PingHandler() {
	// This example demonstrates how to check the database connection

	// Create a request to ping the database
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/ping", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	fmt.Printf("Status: %d\n", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Database connection is healthy")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Database connection error: %s\n", strings.TrimSpace(string(body)))
	}

	// Output:
	// Status: 200
	// Database connection is healthy
}

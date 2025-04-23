package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestModel(t *testing.T) {
	// Create a request
	req := Request{
		URL: "https://github.com/pcristin/urlshortener",
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Unmarshal back
	var newReq Request
	err = json.Unmarshal(data, &newReq)
	require.NoError(t, err)

	// Verify the result
	assert.Equal(t, req.URL, newReq.URL)
}

func TestResponseModel(t *testing.T) {
	// Create a response
	resp := Response{
		Result: "http://localhost:8080/abc123",
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// Unmarshal back
	var newResp Response
	err = json.Unmarshal(data, &newResp)
	require.NoError(t, err)

	// Verify the result
	assert.Equal(t, resp.Result, newResp.Result)
}

func TestBatchRequestItems(t *testing.T) {
	// Create a batch request
	req := BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://github.com/pcristin/urlshortener"},
		{CorrelationID: "2", OriginalURL: "https://golang.org"},
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Unmarshal back
	var newReq BatchRequest
	err = json.Unmarshal(data, &newReq)
	require.NoError(t, err)

	// Verify the result
	assert.Equal(t, len(req), len(newReq))
	assert.Equal(t, req[0].CorrelationID, newReq[0].CorrelationID)
	assert.Equal(t, req[0].OriginalURL, newReq[0].OriginalURL)
	assert.Equal(t, req[1].CorrelationID, newReq[1].CorrelationID)
	assert.Equal(t, req[1].OriginalURL, newReq[1].OriginalURL)
}

func TestBatchResponseItems(t *testing.T) {
	// Create a batch response
	resp := BatchResponse{
		{CorrelationID: "1", ShortURL: "http://localhost:8080/abc123"},
		{CorrelationID: "2", ShortURL: "http://localhost:8080/def456"},
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// Unmarshal back
	var newResp BatchResponse
	err = json.Unmarshal(data, &newResp)
	require.NoError(t, err)

	// Verify the result
	assert.Equal(t, len(resp), len(newResp))
	assert.Equal(t, resp[0].CorrelationID, newResp[0].CorrelationID)
	assert.Equal(t, resp[0].ShortURL, newResp[0].ShortURL)
	assert.Equal(t, resp[1].CorrelationID, newResp[1].CorrelationID)
	assert.Equal(t, resp[1].ShortURL, newResp[1].ShortURL)
}

func TestURLStorageNode(t *testing.T) {
	// Create a storage node
	userID := "test-user"
	id := uuid.New()
	node := URLStorageNode{
		UUID:        id,
		ShortURL:    "abc123",
		OriginalURL: "https://github.com/pcristin/urlshortener",
		UserID:      userID,
		IsDeleted:   false,
	}

	// Marshal to JSON
	data, err := json.Marshal(node)
	require.NoError(t, err)

	// Unmarshal back
	var newNode URLStorageNode
	err = json.Unmarshal(data, &newNode)
	require.NoError(t, err)

	// Verify the result
	assert.Equal(t, node.UUID, newNode.UUID)
	assert.Equal(t, node.ShortURL, newNode.ShortURL)
	assert.Equal(t, node.OriginalURL, newNode.OriginalURL)
	assert.Equal(t, node.UserID, newNode.UserID)
	assert.Equal(t, node.IsDeleted, newNode.IsDeleted)
}

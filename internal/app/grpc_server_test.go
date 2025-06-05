package app

import (
	"context"
	"net"
	"testing"

	"github.com/pcristin/urlshortener/internal/proto"
	"github.com/pcristin/urlshortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupGRPCTest(t *testing.T) (*grpc.Server, proto.URLShortenerServiceClient, func()) {
	lis := bufconn.Listen(bufSize)

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	// Create storage and service
	storage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)
	service := NewURLService(storage, "http://localhost:8080", "")
	grpcServer := NewGRPCServer(service, "http://localhost:8080")

	// Create auth interceptor
	authInterceptor := NewAuthInterceptor("test-secret")

	// Create gRPC server with interceptors
	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryServerInterceptor()),
		grpc.StreamInterceptor(authInterceptor.StreamServerInterceptor()),
	)
	proto.RegisterURLShortenerServiceServer(s, grpcServer)

	// Start server
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	// Create client
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := proto.NewURLShortenerServiceClient(conn)

	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return s, client, cleanup
}

func TestGRPCServer_ShortenURL(t *testing.T) {
	_, client, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	var header metadata.MD

	// Test shortening a new URL
	resp, err := client.ShortenURL(ctx, &proto.ShortenURLRequest{
		Url: "https://example.com",
	}, grpc.Header(&header))

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Result)
	assert.False(t, resp.AlreadyExists)
	assert.Contains(t, resp.Result, "http://localhost:8080/")

	// Check auth headers
	userIDs := header.Get("user-id")
	signatures := header.Get("signature")
	assert.NotEmpty(t, userIDs)
	assert.NotEmpty(t, signatures)

	// Test shortening the same URL with auth
	md := metadata.Pairs("user-id", userIDs[0], "signature", signatures[0])
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp2, err := client.ShortenURL(ctx, &proto.ShortenURLRequest{
		Url: "https://example.com",
	})

	require.NoError(t, err)
	assert.Equal(t, resp.Result, resp2.Result)
	assert.True(t, resp2.AlreadyExists)
}

func TestGRPCServer_GetOriginalURL(t *testing.T) {
	_, client, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()

	// First shorten a URL
	shortenResp, err := client.ShortenURL(ctx, &proto.ShortenURLRequest{
		Url: "https://example.com",
	})
	require.NoError(t, err)

	// Extract token from result
	token := shortenResp.Result[len("http://localhost:8080/"):]

	// Get original URL
	getResp, err := client.GetOriginalURL(ctx, &proto.GetOriginalURLRequest{
		Token: token,
	})

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", getResp.OriginalUrl)
	assert.False(t, getResp.IsDeleted)
}

func TestGRPCServer_ShortenBatch(t *testing.T) {
	_, client, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()

	resp, err := client.ShortenBatch(ctx, &proto.ShortenBatchRequest{
		Items: []*proto.ShortenBatchRequest_BatchItem{
			{CorrelationId: "1", OriginalUrl: "https://example1.com"},
			{CorrelationId: "2", OriginalUrl: "https://example2.com"},
			{CorrelationId: "3", OriginalUrl: "https://example3.com"},
		},
	})

	require.NoError(t, err)
	assert.Len(t, resp.Results, 3)

	for i, result := range resp.Results {
		assert.Equal(t, string(rune('1'+i)), result.CorrelationId)
		assert.Contains(t, result.ShortUrl, "http://localhost:8080/")
	}
}

func TestGRPCServer_GetUserURLs(t *testing.T) {
	_, client, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	var header metadata.MD

	// First request to get auth headers
	_, err := client.ShortenURL(ctx, &proto.ShortenURLRequest{
		Url: "https://example1.com",
	}, grpc.Header(&header))
	require.NoError(t, err)

	// Use the auth from the first request for all subsequent requests
	userIDs := header.Get("user-id")
	signatures := header.Get("signature")
	md := metadata.Pairs("user-id", userIDs[0], "signature", signatures[0])
	authCtx := metadata.NewOutgoingContext(ctx, md)

	// Shorten second URL with same user
	_, err = client.ShortenURL(authCtx, &proto.ShortenURLRequest{
		Url: "https://example2.com",
	})
	require.NoError(t, err)

	// Get user URLs
	resp, err := client.GetUserURLs(authCtx, &proto.GetUserURLsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Urls, 2)
}

func TestGRPCServer_Ping(t *testing.T) {
	_, client, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()

	resp, err := client.Ping(ctx, &proto.PingRequest{})
	require.NoError(t, err)
	assert.False(t, resp.DatabaseOk) // No database in test
}

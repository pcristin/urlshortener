package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/pcristin/urlshortener/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	var (
		grpcAddr = flag.String("addr", "localhost:50051", "gRPC server address")
		action   = flag.String("action", "shorten", "Action to perform: shorten, get, batch, list, delete, ping, stats")
		url      = flag.String("url", "https://example.com", "URL to shorten")
		token    = flag.String("token", "", "Token to get original URL")
		userID   = flag.String("user", "", "User ID for authentication")
		sig      = flag.String("sig", "", "Signature for authentication")
	)
	flag.Parse()

	// Create connection
	conn, err := grpc.Dial(*grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := proto.NewURLShortenerServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add authentication metadata if provided
	if *userID != "" && *sig != "" {
		md := metadata.Pairs("user-id", *userID, "signature", *sig)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Create a variable to store response headers
	var header metadata.MD

	switch *action {
	case "shorten":
		resp, err := client.ShortenURL(ctx, &proto.ShortenURLRequest{
			Url: *url,
		}, grpc.Header(&header))
		if err != nil {
			log.Fatalf("ShortenURL failed: %v", err)
		}
		fmt.Printf("Shortened URL: %s\n", resp.Result)
		fmt.Printf("Already exists: %v\n", resp.AlreadyExists)
		printAuthHeaders(header)

	case "get":
		if *token == "" {
			log.Fatal("Token is required for get action")
		}
		resp, err := client.GetOriginalURL(ctx, &proto.GetOriginalURLRequest{
			Token: *token,
		})
		if err != nil {
			log.Fatalf("GetOriginalURL failed: %v", err)
		}
		fmt.Printf("Original URL: %s\n", resp.OriginalUrl)
		fmt.Printf("Is deleted: %v\n", resp.IsDeleted)

	case "batch":
		resp, err := client.ShortenBatch(ctx, &proto.ShortenBatchRequest{
			Items: []*proto.ShortenBatchRequest_BatchItem{
				{CorrelationId: "1", OriginalUrl: "https://example1.com"},
				{CorrelationId: "2", OriginalUrl: "https://example2.com"},
				{CorrelationId: "3", OriginalUrl: "https://example3.com"},
			},
		}, grpc.Header(&header))
		if err != nil {
			log.Fatalf("ShortenBatch failed: %v", err)
		}
		fmt.Println("Batch results:")
		for _, result := range resp.Results {
			fmt.Printf("  %s: %s\n", result.CorrelationId, result.ShortUrl)
		}
		printAuthHeaders(header)

	case "list":
		resp, err := client.GetUserURLs(ctx, &proto.GetUserURLsRequest{}, grpc.Header(&header))
		if err != nil {
			log.Fatalf("GetUserURLs failed: %v", err)
		}
		fmt.Println("User URLs:")
		for _, url := range resp.Urls {
			fmt.Printf("  %s -> %s\n", url.ShortUrl, url.OriginalUrl)
		}
		printAuthHeaders(header)

	case "delete":
		if *token == "" {
			log.Fatal("Token is required for delete action")
		}
		_, err := client.DeleteUserURLs(ctx, &proto.DeleteUserURLsRequest{
			Tokens: []string{*token},
		}, grpc.Header(&header))
		if err != nil {
			log.Fatalf("DeleteUserURLs failed: %v", err)
		}
		fmt.Println("URLs marked for deletion")
		printAuthHeaders(header)

	case "ping":
		resp, err := client.Ping(ctx, &proto.PingRequest{})
		if err != nil {
			log.Fatalf("Ping failed: %v", err)
		}
		fmt.Printf("Database OK: %v\n", resp.DatabaseOk)

	case "stats":
		// Add X-Real-IP header for stats
		md := metadata.Pairs("x-real-ip", "127.0.0.1")
		ctx = metadata.NewOutgoingContext(ctx, md)

		resp, err := client.GetStats(ctx, &proto.GetStatsRequest{})
		if err != nil {
			log.Fatalf("GetStats failed: %v", err)
		}
		fmt.Printf("Stats:\n")
		fmt.Printf("  URLs: %d\n", resp.Urls)
		fmt.Printf("  Users: %d\n", resp.Users)

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func printAuthHeaders(header metadata.MD) {
	if userIDs := header.Get("user-id"); len(userIDs) > 0 {
		fmt.Printf("\nReceived auth headers:\n")
		fmt.Printf("  user-id: %s\n", userIDs[0])
		if sigs := header.Get("signature"); len(sigs) > 0 {
			fmt.Printf("  signature: %s\n", sigs[0])
		}
	}
}

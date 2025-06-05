package app

import (
	"context"
	"strings"

	"github.com/pcristin/urlshortener/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCServer implements the gRPC URLShortenerService
type GRPCServer struct {
	proto.UnimplementedURLShortenerServiceServer
	service *URLService
	baseURL string
}

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(service *URLService, baseURL string) *GRPCServer {
	return &GRPCServer{
		service: service,
		baseURL: baseURL,
	}
}

// getUserIDFromMetadata extracts user ID from gRPC metadata
func getUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing user-id in metadata")
	}

	return userIDs[0], nil
}

// getClientIPFromMetadata extracts client IP from gRPC metadata
func getClientIPFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Try different headers
	for _, header := range []string{"x-real-ip", "x-forwarded-for"} {
		ips := md.Get(header)
		if len(ips) > 0 {
			// For X-Forwarded-For, take the first IP
			if header == "x-forwarded-for" {
				parts := strings.Split(ips[0], ",")
				return strings.TrimSpace(parts[0])
			}
			return ips[0]
		}
	}

	return ""
}

// constructURL builds the full URL for a shortened link
func (s *GRPCServer) constructURL(token string) string {
	if s.baseURL != "" {
		return s.baseURL + "/" + token
	}
	return "http://localhost:8080/" + token
}

// ShortenURL creates a shortened URL from a long URL
func (s *GRPCServer) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	token, alreadyExists, err := s.service.ShortenURL(ctx, req.Url, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &proto.ShortenURLResponse{
		Result:        s.constructURL(token),
		AlreadyExists: alreadyExists,
	}, nil
}

// GetOriginalURL retrieves the original URL by its token
func (s *GRPCServer) GetOriginalURL(ctx context.Context, req *proto.GetOriginalURLRequest) (*proto.GetOriginalURLResponse, error) {
	originalURL, isDeleted, err := s.service.GetOriginalURL(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &proto.GetOriginalURLResponse{
		OriginalUrl: originalURL,
		IsDeleted:   isDeleted,
	}, nil
}

// ShortenBatch creates multiple shortened URLs in a single request
func (s *GRPCServer) ShortenBatch(ctx context.Context, req *proto.ShortenBatchRequest) (*proto.ShortenBatchResponse, error) {
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Convert proto items to service items
	items := make([]BatchItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = BatchItem{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		}
	}

	results, err := s.service.ShortenBatch(ctx, items, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert results to proto format
	protoResults := make([]*proto.ShortenBatchResponse_BatchResult, len(results))
	for i, result := range results {
		protoResults[i] = &proto.ShortenBatchResponse_BatchResult{
			CorrelationId: result.CorrelationID,
			ShortUrl:      s.constructURL(result.ShortURL),
		}
	}

	return &proto.ShortenBatchResponse{
		Results: protoResults,
	}, nil
}

// GetUserURLs returns all URLs shortened by the authenticated user
func (s *GRPCServer) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	urls, err := s.service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to proto format
	protoURLs := make([]*proto.GetUserURLsResponse_UserURL, len(urls))
	for i, url := range urls {
		protoURLs[i] = &proto.GetUserURLsResponse_UserURL{
			ShortUrl:    s.constructURL(url.ShortURL),
			OriginalUrl: url.OriginalURL,
		}
	}

	return &proto.GetUserURLsResponse{
		Urls: protoURLs,
	}, nil
}

// DeleteUserURLs marks specified URLs as deleted for the authenticated user
func (s *GRPCServer) DeleteUserURLs(ctx context.Context, req *proto.DeleteUserURLsRequest) (*proto.DeleteUserURLsResponse, error) {
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	err = s.service.DeleteUserURLs(ctx, userID, req.Tokens)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.DeleteUserURLsResponse{}, nil
}

// Ping checks the health of the service and database connectivity
func (s *GRPCServer) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	dbOk, err := s.service.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.PingResponse{
		DatabaseOk: dbOk,
	}, nil
}

// GetStats returns statistics about the URL shortener service
func (s *GRPCServer) GetStats(ctx context.Context, req *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	clientIP := getClientIPFromMetadata(ctx)

	stats, err := s.service.GetStats(ctx, clientIP)
	if err != nil {
		if err.Error() == "unauthorized: IP not in trusted subnet" {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.GetStatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}

package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor provides authentication for gRPC requests
type AuthInterceptor struct {
	secret []byte
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(secret string) *AuthInterceptor {
	return &AuthInterceptor{
		secret: []byte(secret),
	}
}

// UnaryServerInterceptor returns a unary server interceptor for authentication
func (a *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for methods that don't require it
		if info.FullMethod == "/urlshortener.URLShortenerService/GetOriginalURL" ||
			info.FullMethod == "/urlshortener.URLShortenerService/Ping" {
			return handler(ctx, req)
		}

		// For GetStats, we don't require user auth, just IP check
		if info.FullMethod == "/urlshortener.URLShortenerService/GetStats" {
			return handler(ctx, req)
		}

		// Authenticate user
		newCtx, err := a.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// authenticate validates or creates user credentials
func (a *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	userIDs := md.Get("user-id")
	signatures := md.Get("signature")

	var userID string
	var needNewAuth bool

	if len(userIDs) > 0 && len(signatures) > 0 {
		// Validate existing credentials
		userID = userIDs[0]
		signature := signatures[0]
		if !a.validateSignature(userID, signature) {
			needNewAuth = true
		}
	} else {
		needNewAuth = true
	}

	if needNewAuth {
		// Generate new credentials
		userID = uuid.New().String()
		signature := a.generateSignature(userID)

		// Add to outgoing context for the client
		header := metadata.Pairs(
			"user-id", userID,
			"signature", signature,
		)
		grpc.SendHeader(ctx, header)
	}

	// Add user ID to incoming context for handlers
	md.Set("user-id", userID)
	newCtx := metadata.NewIncomingContext(ctx, md)

	return newCtx, nil
}

// generateSignature creates an HMAC signature for the given user ID
func (a *AuthInterceptor) generateSignature(userID string) string {
	h := hmac.New(sha256.New, a.secret)
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

// validateSignature validates the HMAC signature for the given user ID
func (a *AuthInterceptor) validateSignature(userID, signature string) bool {
	expectedSignature := a.generateSignature(userID)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// StreamServerInterceptor returns a stream server interceptor for authentication
func (a *AuthInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip auth for methods that don't require it
		if info.FullMethod == "/urlshortener.URLShortenerService/GetOriginalURL" ||
			info.FullMethod == "/urlshortener.URLShortenerService/Ping" {
			return handler(srv, ss)
		}

		// Authenticate user
		newCtx, err := a.authenticate(ss.Context())
		if err != nil {
			return err
		}

		// Wrap the stream with authenticated context
		wrappedStream := &authenticatedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrappedStream)
	}
}

// authenticatedServerStream wraps grpc.ServerStream with authenticated context
type authenticatedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authenticatedServerStream) Context() context.Context {
	return s.ctx
}

// ExtractUserIDFromContext extracts user ID from context (for use in handlers)
func ExtractUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing user-id")
	}

	return userIDs[0], nil
}

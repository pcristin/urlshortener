package urlutils

import (
	"testing"

	"github.com/pcristin/urlshortener/internal/storage"
)

func BenchmarkEncodeURL(b *testing.B) {
	memStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := "https://example.com/" + string(rune(i))
		_, _ = EncodeURL(url, memStorage, "user1")
	}
}

func BenchmarkDecodeURL(b *testing.B) {
	memStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)

	// Prepare by adding URLs
	var tokens []string
	for i := 0; i < 1000; i++ {
		url := "https://example.com/" + string(rune(i))
		token, _ := EncodeURL(url, memStorage, "user1")
		tokens = append(tokens, token)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(tokens)
		_, _ = DecodeURL(tokens[idx], memStorage)
	}
}

func BenchmarkURLCheck(b *testing.B) {
	urls := []string{
		"https://example.com",
		"http://example.com/path",
		"https://subdomain.example.co.uk/path?query=value",
		"example.com",
		"invalid-url",
		"http://localhost:8080",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(urls)
		_ = URLCheck(urls[idx])
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateToken()
	}
}

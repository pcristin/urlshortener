package storage

import (
	"testing"
)

func BenchmarkMemoryStorage_AddURL(b *testing.B) {
	storage := NewURLStorage(MemoryStorageType, "", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := "token" + string(rune(i))
		longURL := "https://example.com/" + string(rune(i))
		userID := "user1"
		_ = storage.AddURL(token, longURL, userID)
	}
}

func BenchmarkMemoryStorage_GetURL(b *testing.B) {
	storage := NewURLStorage(MemoryStorageType, "", nil)

	// Prepare by saving URLs
	var tokens []string
	for i := 0; i < 1000; i++ {
		token := "token" + string(rune(i))
		longURL := "https://example.com/" + string(rune(i))
		userID := "user1"
		_ = storage.AddURL(token, longURL, userID)
		tokens = append(tokens, token)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(tokens)
		_, _ = storage.GetURL(tokens[idx])
	}
}

func BenchmarkMemoryStorage_GetUserURLs(b *testing.B) {
	storage := NewURLStorage(MemoryStorageType, "", nil)

	// Prepare by saving URLs
	for i := 0; i < 1000; i++ {
		token := "token" + string(rune(i))
		longURL := "https://example.com/" + string(rune(i))
		userID := "user1"
		_ = storage.AddURL(token, longURL, userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetUserURLs("user1")
	}
}

func BenchmarkMemoryStorage_GetTokenByURL(b *testing.B) {
	storage := NewURLStorage(MemoryStorageType, "", nil)

	// Prepare by saving URLs
	var urls []string
	for i := 0; i < 1000; i++ {
		token := "token" + string(rune(i))
		longURL := "https://example.com/" + string(rune(i))
		urls = append(urls, longURL)
		_ = storage.AddURL(token, longURL, "user1")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(urls)
		_, _ = storage.GetTokenByURL(urls[idx])
	}
}

func BenchmarkMemoryStorage_DeleteURLs(b *testing.B) {
	storage := NewURLStorage(MemoryStorageType, "", nil)

	// Prepare by saving URLs
	var tokens []string
	for i := 0; i < 1000; i++ {
		token := "token" + string(rune(i))
		longURL := "https://example.com/" + string(rune(i))
		tokens = append(tokens, token)
		_ = storage.AddURL(token, longURL, "user1")
	}

	b.ResetTimer()
	b.Run("Delete10URLs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := (i * 10) % (len(tokens) - 10)
			batchTokens := tokens[start : start+10]
			_ = storage.DeleteURLs("user1", batchTokens)
		}
	})
}

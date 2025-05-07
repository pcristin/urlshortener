package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/pcristin/urlshortener/internal/storage"
	"github.com/pcristin/urlshortener/internal/urlutils"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func run() error {
	// Create profiles directory if it doesn't exist
	profilesDir := "../../profiles"
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		if err := os.Mkdir(profilesDir, 0755); err != nil {
			return fmt.Errorf("failed to create profiles directory: %w", err)
		}
	}

	// Run benchmarks and create memory profile
	return runBenchmarksAndProfile(filepath.Join(profilesDir, "result.pprof"))
}

func runBenchmarksAndProfile(profilePath string) error {
	// Force garbage collection before profiling
	runtime.GC()

	// Create profile file
	f, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer f.Close()

	// Run operations that we want to profile
	memStorage := storage.NewURLStorage(storage.MemoryStorageType, "", nil)

	// Add 10,000 URLs
	fmt.Println("Adding 10,000 URLs to memory storage...")
	for i := 0; i < 10000; i++ {
		url := fmt.Sprintf("https://example.com/page%d", i)
		token, _ := urlutils.EncodeURL(url, memStorage, "user1")

		// Perform some lookups to simulate real usage
		if i%100 == 0 {
			_, _ = memStorage.GetURL(token)
			_, _ = urlutils.DecodeURL(token, memStorage)
		}
	}

	// Get user URLs
	fmt.Println("Getting user URLs...")
	urls, _ := memStorage.GetUserURLs("user1")
	fmt.Printf("Found %d URLs for user1\n", len(urls))

	// Force GC again before writing profile
	runtime.GC()

	// Write heap profile
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}

	fmt.Printf("Memory profile written to %s\n", profilePath)
	return nil
}

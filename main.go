package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"searchapi/models"
	"searchapi/search"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	categories = []string{
		"Electronics", "Clothing", "Books", "Home & Kitchen", "Sports",
		"Toys", "Beauty", "Automotive", "Garden", "Health",
	}
	
	adjectives = []string{
		"Premium", "Deluxe", "Essential", "Basic", "Professional",
		"Advanced", "Smart", "Classic", "Modern", "Eco-friendly",
	}
	
	nouns = []string{
		"Laptop", "Phone", "Camera", "Watch", "Headphones",
		"Speaker", "Tablet", "Monitor", "Keyboard", "Mouse",
	}
)

// generateProduct generates a single product
func generateProduct(id int, rng *rand.Rand) models.Product {
	adj := adjectives[rng.Intn(len(adjectives))]
	noun := nouns[rng.Intn(len(nouns))]
	category := categories[rng.Intn(len(categories))]
	
	return models.Product{
		ID:       id,
		Name:     fmt.Sprintf("%s %s", adj, noun),
		Category: category,
	}
}

// printMemUsage outputs the current memory usage
func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Memory usage: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB",
		m.Alloc/1024/1024,
		m.TotalAlloc/1024/1024,
		m.Sys/1024/1024,
	)
}

func main() {
	// Lower memory target to help GC be more aggressive
	debug.SetGCPercent(20)
	
	// Free OS memory more aggressively
	debug.SetMemoryLimit(1024 * 1024 * 1024) 

	// Create a new search index
	searchIndex, err := search.NewSearchIndex()
	if err != nil {
		log.Fatalf("Failed to create search index: %v", err)
	}
	defer searchIndex.Close()

	// Reduce the total number of products to stay within memory constraints
	const totalProducts = 300000
	const batchSize = 1000     
	
	log.Printf("Starting to generate and index %d products in batches of %d...\n", totalProducts, batchSize)
	
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	startTime := time.Now()
	
	// Process in batches
	for i := 0; i < totalProducts; i += batchSize {
		batchEnd := i + batchSize
		if batchEnd > totalProducts {
			batchEnd = totalProducts
		}
		
		// Generate and index batch
		for j := i; j < batchEnd; j++ {
			product := generateProduct(j+1, rng)
			if err := searchIndex.IndexProduct(product); err != nil {
				log.Printf("Error indexing product %d: %v", product.ID, err)
			}
		}
		
		// Print progress
		progress := float64(batchEnd) / float64(totalProducts) * 100
		elapsed := time.Since(startTime)
		estimatedTotal := elapsed * time.Duration(totalProducts) / time.Duration(batchEnd)
		remaining := estimatedTotal - elapsed
		
		log.Printf("Progress: %.2f%% (%d/%d products) | Est. remaining time: %v",
			progress, batchEnd, totalProducts, remaining.Round(time.Second))
		
		// Print memory usage more frequently (every 5%)
		if batchEnd%(totalProducts/20) == 0 {
			printMemUsage()
		}
		
		// Force garbage collection after each batch
		runtime.GC()
	}

	log.Printf("Indexing complete! Total time: %v\n", time.Since(startTime).Round(time.Second))
	printMemUsage()

	// Create router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Routes
	router.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if strings.TrimSpace(query) == "" {
			http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		// Perform search
		products, total, err := searchIndex.Search(query, 50)
		if err != nil {
			http.Error(w, fmt.Sprintf("Search error: %v", err), http.StatusInternalServerError)
			return
		}

		// Return results
		response := models.ProductResponse{
			Products: products,
			Total:    total,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Create server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Channel for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		log.Printf("\nServer starting on http://localhost%s", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Server is shutting down...")

	// Create shutdown context with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
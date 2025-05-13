package search

import (
	"fmt"
	"searchapi/models"

	"github.com/blevesearch/bleve/v2"
)

type SearchIndex struct {
	index bleve.Index
}

// NewSearchIndex creates a new search index
func NewSearchIndex() (*SearchIndex, error) {
	// Create a mapping for our product
	mapping := bleve.NewIndexMapping()
	
	// Create an in-memory index
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, fmt.Errorf("error creating index: %v", err)
	}

	return &SearchIndex{index: index}, nil
}

// IndexProduct adds a product to the search index
func (s *SearchIndex) IndexProduct(product models.Product) error {
	// Create a document ID from the product ID
	docID := fmt.Sprintf("%d", product.ID)
	
	// Index the product
	err := s.index.Index(docID, product)
	if err != nil {
		return fmt.Errorf("error indexing product %d: %v", product.ID, err)
	}
	
	return nil
}

// Search performs a search query and returns matching products
func (s *SearchIndex) Search(query string, limit int) ([]models.Product, int, error) {
	// Create a search query
	q := bleve.NewMatchQuery(query)
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Size = limit
	searchRequest.Fields = []string{"id", "name", "category"}
	
	// Execute the search
	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing search: %v", err)
	}

	// Convert results back to products
	products := make([]models.Product, 0, len(searchResults.Hits))
	for _, hit := range searchResults.Hits {
		var product models.Product
		
		// Extract fields from the stored fields
		if idStr, ok := hit.Fields["id"].(float64); ok {
			product.ID = int(idStr)
		}
		if name, ok := hit.Fields["name"].(string); ok {
			product.Name = name
		}
		if category, ok := hit.Fields["category"].(string); ok {
			product.Category = category
		}
		
		products = append(products, product)
	}

	return products, int(searchResults.Total), nil
}

// Close closes the index
func (s *SearchIndex) Close() error {
	return s.index.Close()
} 
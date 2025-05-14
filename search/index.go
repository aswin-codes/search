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
	
	// Create a document mapping that indexes name and category fields
	productMapping := bleve.NewDocumentMapping()
	
	// Set up the name field mapping
	nameFieldMapping := bleve.NewTextFieldMapping()
	productMapping.AddFieldMappingsAt("name", nameFieldMapping)
	
	// Set up the category field mapping - enable indexing for this field
	categoryFieldMapping := bleve.NewTextFieldMapping()
	productMapping.AddFieldMappingsAt("category", categoryFieldMapping)
	
	// Disable indexing for ID field (it'll still be stored)
	idFieldMapping := bleve.NewNumericFieldMapping()
	idFieldMapping.Index = false
	productMapping.AddFieldMappingsAt("id", idFieldMapping	)
	
	// Add the document mapping to the index mapping
	mapping.AddDocumentMapping("_default", productMapping)
	
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

func (s *SearchIndex) Search(query string, limit int) ([]models.Product, int, error) {
	fuzzyName := bleve.NewFuzzyQuery(query)
	fuzzyName.SetField("name")
	fuzzyName.SetFuzziness(1)

	prefixName := bleve.NewPrefixQuery(query)
	prefixName.SetField("name")

	fuzzyCategory := bleve.NewFuzzyQuery(query)
	fuzzyCategory.SetField("category")
	fuzzyCategory.SetFuzziness(1)

	prefixCategory := bleve.NewPrefixQuery(query)
	prefixCategory.SetField("category")

	wildcardName := bleve.NewWildcardQuery("*" + query + "*")
	wildcardName.SetField("name")

	wildcardCategory := bleve.NewWildcardQuery("*" + query + "*")
	wildcardCategory.SetField("category")

	q := bleve.NewDisjunctionQuery(fuzzyName, prefixName, wildcardName, fuzzyCategory, prefixCategory, wildcardCategory)

	

	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Size = limit
	searchRequest.Fields = []string{"id", "name", "category"}

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing search: %v", err)
	}

	products := make([]models.Product, 0, len(searchResults.Hits))
	for _, hit := range searchResults.Hits {
		var product models.Product

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
package models

// Product represents a product in our system
type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// ProductResponse represents the search response format
type ProductResponse struct {
	Products []Product `json:"products"`
	Total    int      `json:"total"`
} 
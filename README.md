# Product Search API

A simple and efficient product search API built with Go that performs full-text search operations on 1 million product records.

## Demo



https://github.com/user-attachments/assets/34704c59-22b1-434f-b121-26b34c6a4ce0


## Decision Log

[View the Decision Log on Google Docs](https://docs.google.com/document/d/17wd7sXPdkPsMiDlwdIoZ-lD0wqmpwxXfr8PGldKifpA/edit?usp=sharing)

## Features

- In-memory full-text search using Bleve
- RESTful API using Chi router
- Handles 1 million product records
- Graceful shutdown
- Memory-efficient implementation

## Requirements

- Go 1.21 or later

## Installation

1. Clone the repository
2. Install dependencies:
```bash
go mod download
```

## Running the Server

```bash
go run main.go
```

The server will start on `http://localhost:8080`.

## API Endpoints

### Search Products

```
GET /search?q=<query>
```

Parameters:
- `q`: Search query (required)

Example:
```bash
curl "http://localhost:8080/search?q=premium laptop"
```

Response:
```json
{
    "products": [
        {
            "id": 123,
            "name": "Premium Laptop",
            "category": "Electronics"
        }
    ],
    "total": 1
}
```

## Implementation Details

- Uses Bleve's in-memory index for efficient full-text search
- Returns up to 50 matching products per query
- Products are generated with random combinations of adjectives and nouns
- Includes 10 different product categories

## Memory Usage

The implementation is optimized for systems with limited RAM (8GB) by:
- Using an in-memory index
- Limiting search results to 50 items
- Efficient product generation
- Minimal memory allocations during search operations 

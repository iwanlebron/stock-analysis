# AGENTS.md - Stock Analysis Fear & Greed Index

This file provides guidance for AI agents working on this Go-based Fear & Greed Index analysis tool.

## Project Overview

A Go-based Fear & Greed Index analysis tool that computes sentiment scores (0-100) for stocks across US, Hong Kong, China A-share, and cryptocurrency markets. It fetches OHLCV data from Yahoo Finance, calculates 8 technical indicators, normalizes them using rolling window percentiles, and serves results via a web dashboard and JSON API.

## Build & Development Commands

### Basic Commands
```bash
# Run directly (development)
go run main.go

# Build binary
go build -o server main.go
./server

# Cross-compilation
GOOS=linux GOARCH=amd64 go build -o server main.go
GOOS=linux GOARCH=arm64 go build -o server main.go

# Docker
docker build -t stock-analysis .
docker run -d -p 8000:8000 --name stock-analysis stock-analysis

# Push to DockerHub (uses docker.sh script)
./docker.sh <docker-username>
```

### Quality & Testing
```bash
# Format code (must pass in CI)
go fmt ./...

# Vet code (must pass in CI)
go vet ./...

# Run tests (if any exist)
go test ./...

# Run single test file
go test ./internal/calc/...

# Run specific test
go test -run TestFunctionName ./...
```

### Linting & Code Quality
- Use `go fmt` to format code (standard Go formatting)
- Use `go vet` to check for suspicious constructs
- No external linters configured - rely on Go's built-in tools
- CI runs: `gofmt -l .` (should output nothing), `go test ./...`, `go vet ./...`

## Code Style Guidelines

### Imports
```go
import (
    "context"
    "embed"
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "math"
    "net"
    "net/http"
    "strconv"
    "strings"
    "time"

    "stock-analysis/internal/calc"
    "stock-analysis/internal/data"
    "stock-analysis/internal/models"

    "github.com/patrickmn/go-cache"
)
```
- Group imports: standard library, internal packages, external dependencies
- Use absolute imports for internal packages (`stock-analysis/internal/...`)
- Sort imports alphabetically within groups

### Naming Conventions
- **Packages**: lowercase, single word, descriptive (e.g., `calc`, `api`, `models`)
- **Types**: PascalCase (e.g., `PriceFrame`, `ScoreResult`, `Config`)
- **Variables**: camelCase (e.g., `priceFrame`, `scoreResult`, `normWindow`)
- **Constants**: PascalCase or ALL_CAPS depending on scope
- **Interfaces**: PascalCase ending with "er" if appropriate (not currently used)
- **Methods**: camelCase, receiver variable should be short (e.g., `(p *Price)`, `(pf *PriceFrame)`)

### Error Handling
- Use `log.Fatal()` for unrecoverable startup errors only
- Return errors from functions that can fail
- Use `if err != nil { return err }` pattern
- For HTTP handlers, log errors and return appropriate HTTP status codes
- Use context for cancellation and timeouts in HTTP operations

### Types & Structs
```go
// Price represents a single candle
type Price struct {
    Date   time.Time
    Open   float64
    High   float64
    Low    float64
    Close  float64
    Volume float64
}

// ScoreResult represents the fear & greed score for a single day
type ScoreResult struct {
    Date   time.Time `json:"date"`
    Score  float64   `json:"score"`
    Label  string    `json:"label"`
    Price  float64   `json:"price"`
    Values struct {
        Trend    float64 `json:"trend"`
        Momentum float64 `json:"momentum"`
        RSI      float64 `json:"rsi"`
        MACD     float64 `json:"macd"`
        Drawdown float64 `json:"drawdown"`
        Vol      float64 `json:"volatility"`
        VolSent  float64 `json:"volume_sentiment"`
        MFI      float64 `json:"mfi"`
        BB       float64 `json:"bb_pct_b"`
    } `json:"values"`
}
```
- Use JSON tags for structs that will be serialized
- Embed structs for organization (e.g., `Values`, `Raw` in `ScoreResult`)
- Use `float64` for all financial calculations
- Use `time.Time` for dates

### Functions & Methods
- Keep functions focused and single-purpose
- Use descriptive names that indicate what they return
- For indicator calculations, accept slices and return slices
- Handle edge cases (e.g., insufficient data returns `NaN` or empty results)
- Document public functions with comments

### Comments
- Document exported types, functions, and methods
- Use `//` for single-line comments
- Use `/* */` for multi-line comments only when necessary
- Comment complex algorithms or business logic
- Keep comments up-to-date with code changes

### File Organization
- One type per file unless types are closely related
- Group related functionality in the same package
- Keep files under 500 lines when possible
- Use `internal/` package structure for implementation details

## Architecture Patterns

### Indicator Calculations
- All calculations in `internal/calc/`
- Functions accept `[]float64` slices and return `[]float64`
- Handle `NaN` values appropriately
- Use rolling window algorithms for efficiency
- Normalize indicators using `RollingScore()` function

### HTTP API
- Use `net/http` standard library
- Implement middleware for logging and rate limiting
- Use `go-cache` for response caching (5-minute TTL)
- Embed templates with `go:embed`
- Return JSON for API endpoints, HTML for web interface

### Data Flow
1. Fetch OHLCV data from Yahoo Finance (`internal/data/yahoo.go`)
2. Calculate indicators (`internal/calc/indicators.go`)
3. Normalize scores using rolling percentiles (`internal/calc/engine.go`)
4. Weight and combine sub-scores
5. Serve via HTTP API or web dashboard

### Configuration
- Use `Config` struct for tunable parameters
- Provide `DefaultConfig` with sensible defaults
- Make weights configurable via `Weights` map
- Allow API parameters to override defaults

## Testing Guidelines

### Test Structure
- Create `_test.go` files alongside implementation
- Use table-driven tests for indicator calculations
- Test edge cases: empty data, insufficient data, NaN values
- Mock external dependencies (Yahoo Finance) for unit tests

### Test Commands
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/calc

# Run benchmark tests
go test -bench=. ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance Considerations

- Use slices instead of arrays for dynamic data
- Pre-allocate slices when size is known
- Avoid unnecessary allocations in hot paths
- Use rolling window algorithms to avoid O(n²) complexity
- Cache expensive calculations (already implemented with go-cache)

## Security Best Practices

- Validate and sanitize all user input (ticker symbols, dates)
- Implement rate limiting to prevent abuse
- Use context timeouts for external API calls
- Never log sensitive data
- Use HTTPS in production (not implemented in current code)

## Common Patterns

### Indicator Normalization
```go
// Normalize using rolling window percentiles
func RollingScore(values []float64, window int, direction float64) []float64 {
    // Implementation in internal/calc/indicators.go
}
```

### Error Handling in HTTP Handlers
```go
func handleFearGreed(w http.ResponseWriter, r *http.Request) {
    // Parse parameters
    // Fetch data
    // Calculate scores
    // Return JSON or error
}
```

### Cache Usage
```go
// Check cache first
if cached, found := memCache.Get(cacheKey); found {
    // Return cached response
}

// Calculate and cache
memCache.Set(cacheKey, result, cache.DefaultExpiration)
```

## When Making Changes

1. **Run tests**: `go test ./...`
2. **Format code**: `go fmt ./...`
3. **Check vetting**: `go vet ./...`
4. **Build binary**: `go build -o server main.go`
5. **Test manually**: `./server` and verify functionality
6. **Update documentation**: Keep README.md and CLAUDE.md current

## Dependencies

- Go 1.22+ required
- Single external dependency: `github.com/patrickmn/go-cache`
- No database or complex external services
- Yahoo Finance API as data source

## File Structure Reference

```
.
├── main.go                 # HTTP server entry point
├── internal/
│   ├── api/               # HTTP handlers, templates, caching
│   │   └── handler.go
│   ├── calc/              # Indicator calculations, score normalization
│   │   ├── engine.go      # Main scoring engine
│   │   └── indicators.go  # Technical indicator implementations
│   └── models/            # Data structures
│       └── types.go
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── Dockerfile            # Docker build configuration
└── docker.sh             # Docker push script
```
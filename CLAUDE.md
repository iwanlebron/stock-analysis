# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go-based Fear & Greed Index analysis tool that computes sentiment scores (0-100) for stocks across US, Hong Kong, China A-share, and cryptocurrency markets. It fetches OHLCV data from Yahoo Finance, calculates 8 technical indicators, normalizes them using rolling window percentiles, and serves results via a web dashboard and JSON API.

## Development Commands

```bash
# Run directly (development)
go run main.go

# Build binary
go build -o server main.go
./server

# Build for specific architecture
GOOS=linux GOARCH=amd64 go build -o server main.go
GOOS=linux GOARCH=arm64 go build -o server main.go

# Docker
docker build -t stock-analysis .
docker run -d -p 8000:8000 --name stock-analysis stock-analysis

# Push to DockerHub (uses docker.sh script)
./docker.sh <docker-username>
```

## Architecture

The application follows a layered architecture:

```
main.go (HTTP server)
    ↓
internal/api/ (HTTP handlers, templates, caching)
    ↓
    ├─→ internal/data/ (Yahoo Finance API client)
    └─→ internal/calc/ (Indicator calculations, score normalization)
```

### Key Components

**`internal/api/handler.go`**: HTTP request handling
- Serves embedded HTML templates via `go:embed`
- `/` route: Serves the web dashboard
- `/fear-greed` route: JSON API endpoint for scores
- In-memory caching (5 min expiration, 10 min purge)
- Handles date buffering: fetches ~2 years of historical data to support MA60/252-day calculations

**`internal/calc/engine.go`**: Core scoring engine
- `Compute()`: Main entry point, orchestrates indicator calculation
- `DefaultConfig`: Uses 252-day normalization window, MA20/MA60 for trend, 20-day momentum/volatility, 14-day RSI/MFI
- Weights (total = 1.0): trend (15%), momentum (15%), rsi (10%), macd (10%), drawdown (10%), volatility (10%), mfi (15%), bb_pct_b (15%)
- Final score is weighted average of available sub-scores

**`internal/calc/indicators.go`**: Technical indicator implementations
- `MFI()`: Money Flow Index with rolling window (14-day default)
- `BollingerPercentB()`: %B indicator (20-window, 2 std dev)
- `SMA()`, `EMA()`: Moving averages
- `RSI()`: Wilder's smoothing method
- `MACD()`: 12-26-9 histogram
- `RealizedVol()`: Annualized volatility (std * sqrt(252))
- `Momentum()`: N-period return
- `RollingScore()`: Percentile normalization - ranks values in rolling window, maps to 0-100

**`internal/data/yahoo.go`**: Data provider
- Fetches from `https://query1.finance.yahoo.com/v8/finance/chart/`
- 10-second timeout, requires User-Agent header
- Supports daily ("1d") and hourly ("1h") intervals
- Returns sorted OHLCV data, null values are filtered out

**`internal/models/types.go`**: Data structures
- `Price`: OHLCV with timestamp
- `PriceFrame`: Container for price data
- `ScoreResult`: Date, Price, Score (0-100), Label, Raw/Values sub-scores

### Score Normalization

Indicators are normalized using rolling window percentiles:
- For each data point, collect values from the past `norm_window` periods (default 252)
- Calculate the percentile rank of the current value within that window
- Map to 0-100 scale
- "Bad" indicators (volatility, drawdown) use `direction=-1` to invert the score

### Label Ranges

| Score | Chinese | English |
|-------|---------|---------|
| 0-24 | 极度恐惧 | Extreme Fear |
| 25-44 | 恐惧 | Fear |
| 45-55 | 中性 | Neutral |
| 56-75 | 贪婪 | Greed |
| 76-100 | 极度贪婪 | Extreme Greed |

## API Endpoint

`GET /fear-greed`

Query Parameters:
- `ticker` (required): Stock symbol (e.g., AAPL, NVDA, 0700.HK, BTC-USD)
- `freq`: "1d" (default) or "1h"
- `start`: Start date in YYYY-MM-DD format
- `window`: Normalization window (default 252)
- `tail`: Number of recent data points (default 600)
- `lang`: "zh" (default) or "en"

## Supported Markets

- **US**: SPY, QQQ, NVDA, AAPL, TSLA, AMD, MSFT
- **Hong Kong**: ^HSI, 0700.HK, 9988.HK, 3690.HK
- **China A-Share**: 000001.SS, 399001.SZ, 600519.SS
- **Crypto**: BTC-USD, ETH-USD, SOL-USD, BNB-USD, DOGE-USD

## Configuration

- **Port**: `PORT` environment variable (default 8000)
- **Cache**: Configured in `internal/api/handler.go:36` (default 5min expiration)
- **Weights**: Defined in `internal/calc/engine.go:28`
- **Normalization Window**: Configurable via API, default 252 (1 year of trading days)

## Important Notes

- When adding new indicators, implement the raw calculation in `indicators.go`, add normalization in `engine.go`, and update the `Weights` map
- The web templates are embedded via `go:embed` - changes require recompilation
- Yahoo Finance API may rate-limit; the cache helps mitigate this
- NaN handling: API returns null for NaN scores (`*float64` pointer)
- Date buffering: The handler fetches extra historical data (2 years for daily, 2 months for hourly) to ensure indicators have sufficient warmup period

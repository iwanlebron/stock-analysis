package models

import "time"

// Price represents a single candle
type Price struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// PriceFrame holds a series of prices
type PriceFrame struct {
	Ticker    string
	Frequency string
	Prices    []Price
}

// ScoreResult represents the fear & greed score for a single day
type ScoreResult struct {
	Date   time.Time `json:"date"`
	Score  float64   `json:"score"`
	Label  string    `json:"label"`
	Price  float64   `json:"price"` // Added Price
	Values struct {
		Trend    float64 `json:"trend"`
		Momentum float64 `json:"momentum"`
		RSI      float64 `json:"rsi"`
		MACD     float64 `json:"macd"`
		Drawdown float64 `json:"drawdown"`
		Vol      float64 `json:"volatility"`
		VolSent  float64 `json:"volume_sentiment"`
		MFI      float64 `json:"mfi"`      // New
		BB       float64 `json:"bb_pct_b"` // New
	} `json:"values"`
	Raw struct {
		Trend    float64 `json:"trend_raw"`
		Momentum float64 `json:"momentum_raw"`
		RSI      float64 `json:"rsi_raw"`
		MACD     float64 `json:"macd_raw"`
		Drawdown float64 `json:"drawdown_raw"`
		Vol      float64 `json:"volatility_raw"`
		VolSent  float64 `json:"volume_sentiment_raw"`
		MFI      float64 `json:"mfi_raw"`      // New
		BB       float64 `json:"bb_pct_b_raw"` // New
	} `json:"raw"`
}

type Component struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Detail      string  `json:"detail"`
	Weight      float64 `json:"weight"`
}

type APIResponse struct {
	Ticker          string                 `json:"ticker"`
	Frequency       string                 `json:"frequency"`
	Latest          *SimpleScore           `json:"latest"`
	Series          []SimpleScore          `json:"series"`
	Method          map[string]interface{} `json:"method"`
	Components      []Component            `json:"components"`
	LatestSubscores map[string]float64     `json:"latest_subscores"`
}

type SimpleScore struct {
	Date  string  `json:"date"`
	Score float64 `json:"score"`
	Label string  `json:"label"`
	Price float64 `json:"price"` // Added Price
}

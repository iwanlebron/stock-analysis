package calc

import (
	"math"
	"stock-analysis/internal/models"
)

type Config struct {
	NormWindow int
	MAFast     int
	MASlow     int
	MomWindow  int
	VolWindow  int
	RSIWindow  int
	DDWindow   int
}

var DefaultConfig = Config{
	NormWindow: 252,
	MAFast:     20,
	MASlow:     60,
	MomWindow:  20,
	VolWindow:  20,
	RSIWindow:  14,
	DDWindow:   252,
}

var Weights = map[string]float64{
	"trend":      0.15, // Reduced from 0.20
	"momentum":   0.15,
	"rsi":        0.10, // Reduced from 0.15
	"macd":       0.10,
	"drawdown":   0.10, // Reduced from 0.15
	"volatility": 0.10, // Reduced from 0.15
	"mfi":        0.15, // New: Replaces volume_sentiment (0.10) + extra
	"bb_pct_b":   0.15, // New: Replaces part of trend/volatility
}

func Compute(pf *models.PriceFrame, cfg Config, lang string) []models.ScoreResult {
	n := len(pf.Prices)
	if n == 0 {
		return nil
	}

	// Extract series
	closes := make([]float64, n)
	volumes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)

	for i, p := range pf.Prices {
		closes[i] = p.Close
		volumes[i] = p.Volume
		highs[i] = p.High
		lows[i] = p.Low
	}

	// Adjust norm window if not enough data
	normWindow := cfg.NormWindow
	if n < normWindow {
		normWindow = n
	}
	if normWindow < 10 {
		normWindow = 10 // minimum
	}

	// 1. Calculate Raw Indicators
	maFast := SMA(closes, cfg.MAFast)
	maSlow := SMA(closes, cfg.MASlow)
	trendRaw := make([]float64, n)
	for i := 0; i < n; i++ {
		t1 := 0.0
		if maFast[i] > 0 {
			t1 = (closes[i]/maFast[i] - 1.0)
		}
		t2 := 0.0
		if maSlow[i] > 0 {
			t2 = (closes[i]/maSlow[i] - 1.0)
		}
		trendRaw[i] = 0.5*t1 + 0.5*t2
	}

	momRaw := Momentum(closes, cfg.MomWindow)
	rsiRaw := RSI(closes, cfg.RSIWindow)
	macdRaw := MACD(closes)
	volRaw := RealizedVol(closes, cfg.VolWindow)

	// New Indicators
	mfiRaw := MFI(highs, lows, closes, volumes, 14) // Standard 14
	bbRaw := BollingerPercentB(closes, 20, 2.0)     // Standard 20, 2

	// Drawdown (from 252 high)
	ddRaw := make([]float64, n)
	// Rolling max
	for i := 0; i < n; i++ {
		start := i - cfg.DDWindow + 1
		if start < 0 {
			start = 0
		}
		maxP := 0.0
		for j := start; j <= i; j++ {
			if closes[j] > maxP {
				maxP = closes[j]
			}
		}
		if maxP > 0 {
			ddRaw[i] = (closes[i] / maxP) - 1.0
		}
	}

	// 2. Normalize to Scores (0-100)
	sTrend := RollingScore(trendRaw, normWindow, 1)
	sMom := RollingScore(momRaw, normWindow, 1)
	sRSI := RollingScore(rsiRaw, normWindow, 1)
	sMACD := RollingScore(macdRaw, normWindow, 1)
	sDD := RollingScore(ddRaw, normWindow, 1)
	sVol := RollingScore(volRaw, normWindow, -1) // Lower vol is better (greedier)
	sMFI := RollingScore(mfiRaw, normWindow, 1)
	sBB := RollingScore(bbRaw, normWindow, 1)

	// 3. Aggregate
	results := make([]models.ScoreResult, n)

	for i := 0; i < n; i++ {
		res := models.ScoreResult{
			Date:  pf.Prices[i].Date,
			Price: pf.Prices[i].Close, // Fill price
		}
		// Raw
		res.Raw.Trend = trendRaw[i]
		res.Raw.Momentum = momRaw[i]
		res.Raw.RSI = rsiRaw[i]
		res.Raw.MACD = macdRaw[i]
		res.Raw.Drawdown = ddRaw[i]
		res.Raw.Vol = volRaw[i]
		// res.Raw.VolSent = volSentRaw[i] // Removed

		// Scores
		res.Values.Trend = sTrend[i]
		res.Values.Momentum = sMom[i]
		res.Values.RSI = sRSI[i]
		res.Values.MACD = sMACD[i]
		res.Values.Drawdown = sDD[i]
		res.Values.Vol = sVol[i]
		res.Values.MFI = sMFI[i]
		res.Values.BB = sBB[i]

		// Weighted Sum
		wSum := 0.0
		scoreSum := 0.0

		add := func(key string, val float64) {
			if !math.IsNaN(val) {
				w := Weights[key]
				scoreSum += val * w
				wSum += w
			}
		}

		add("trend", sTrend[i])
		add("momentum", sMom[i])
		add("rsi", sRSI[i])
		add("macd", sMACD[i])
		add("drawdown", sDD[i])
		add("volatility", sVol[i])
		add("mfi", sMFI[i])
		add("bb_pct_b", sBB[i])

		if wSum > 0 {
			res.Score = scoreSum / wSum
			res.Label = LabelFromScore(res.Score, lang)
		} else {
			res.Score = math.NaN()
			res.Label = "-"
		}

		results[i] = res
	}

	return results
}

func LabelFromScore(s float64, lang string) string {
	if math.IsNaN(s) {
		return "-"
	}

	if lang == "en" {
		if s < 25 {
			return "Extreme Fear"
		}
		if s < 45 {
			return "Fear"
		}
		if s <= 55 {
			return "Neutral"
		}
		if s <= 75 {
			return "Greed"
		}
		return "Extreme Greed"
	}

	// Default Chinese
	if s < 25 {
		return "极度恐惧"
	}
	if s < 45 {
		return "恐惧"
	}
	if s <= 55 {
		return "中性"
	}
	if s <= 75 {
		return "贪婪"
	}
	return "极度贪婪"
}

package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"stock-analysis/internal/calc"
	"stock-analysis/internal/data"
	"stock-analysis/internal/models"

	"github.com/patrickmn/go-cache"
)

//go:embed templates/*.html templates/partials/*.html
var templateFS embed.FS

var templates *template.Template
var memCache *cache.Cache

func init() {
	var err error
	templates, err = template.ParseFS(templateFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	memCache = cache.New(5*time.Minute, 10*time.Minute)
}

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/fear-greed", handleFearGreed)
	return loggingMiddleware(mux)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Wrap ResponseWriter to capture status code
		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.String(), ww.statusCode, time.Since(start))
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", 500)
	}
}

func handleFearGreed(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ticker := q.Get("ticker")
	if ticker == "" {
		http.Error(w, `{"detail":"Ticker required"}`, 400)
		return
	}

	freq := q.Get("freq")
	if freq == "" {
		freq = "1d"
	}

	startStr := q.Get("start")
	var start time.Time
	if startStr != "" {
		start, _ = time.Parse("2006-01-02", startStr)
	}

	// Adjust start date to fetch earlier data for warmup (e.g. MA60 needs 60 bars)
	// We add buffer based on window.
	// To be safe, we fetch (window * 2) extra days.
	// For daily: 252 * 2 = ~500 days (approx 2 years)
	fetchStart := start
	if !start.IsZero() {
		if freq == "1d" {
			// Approx 2 years buffer for daily
			// This ensures even if user asks for data starting today, we have enough history to compute indicators
			fetchStart = start.AddDate(-2, 0, 0)
		} else {
			// For hourly, maybe 2 months buffer
			fetchStart = start.AddDate(0, -2, 0)
		}
	}

	lang := q.Get("lang")
	if lang == "" {
		lang = "zh"
	}

	// Window
	windowStr := q.Get("window")
	window := 252
	if windowStr != "" {
		if v, err := strconv.Atoi(windowStr); err == nil {
			window = v
		}
	}

	tail := 600
	if tailStr := q.Get("tail"); tailStr != "" {
		if v, err := strconv.Atoi(tailStr); err == nil {
			tail = v
		}
	}

	// Cache Key
	cacheKey := fmt.Sprintf("%s-%s-%s-%s-%d-%d", ticker, freq, startStr, lang, window, tail)
	if cachedResp, found := memCache.Get(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cachedResp)
		return
	}

	// Fetch Data
	log.Printf("Fetching data for %s (Start: %s, Freq: %s)", ticker, startStr, freq)
	provider := data.NewYahooProvider()
	pf, err := provider.GetPrices(ticker, fetchStart, time.Time{}, freq)
	if err != nil {
		log.Printf("Error fetching data for %s: %v", ticker, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "未找到股票代码或暂无数据：" + ticker + " (" + err.Error() + ")",
		})
		return
	}

	// Compute
	log.Printf("Computing indicators for %s (%d bars)", ticker, len(pf.Prices))
	cfg := calc.DefaultConfig
	cfg.NormWindow = window
	results := calc.Compute(pf, cfg, lang)

	// Find start index based on user request "start" time
	// We calculated results starting from fetchStart (which is start - buffer).
	// Now we want to find the index that corresponds to the user's requested 'start'.
	startIdx := 0
	if !start.IsZero() {
		for i, r := range results {
			// Find the first date >= user requested start
			if !r.Date.Before(start) {
				startIdx = i
				break
			}
		}
	} else {
		// If no start provided, default logic (e.g. tail)
		startIdx = len(results) - tail
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// Ensure bounds
	if startIdx >= len(results) {
		startIdx = len(results) - 1
	}
	if startIdx < 0 {
		startIdx = 0
	}

	// Fix NaNs in series
	// We need custom serialization or pre-process.
	// Since struct has float64, we can't put null.
	// Let's use 0 for simplicity or modify struct to *float64
	// Let's modify struct logic inline:

	type SafeScore struct {
		Date  string   `json:"date"`
		Score *float64 `json:"score"` // Pointer allows null
		Label string   `json:"label"`
		Price float64  `json:"price"` // Added Price
	}

	safeSeries := make([]SafeScore, 0, len(results)-startIdx)
	for i := startIdx; i < len(results); i++ {
		r := results[i]
		var s *float64
		if !math.IsNaN(r.Score) {
			v := r.Score
			s = &v
		}
		safeSeries = append(safeSeries, SafeScore{
			Date:  r.Date.Format("2006-01-02"),
			Score: s,
			Label: r.Label,
			Price: r.Price, // Pass Price to frontend
		})
	}

	// Latest
	var latest *models.SimpleScore
	var latestSubscores map[string]float64

	if len(results) > 0 {
		last := results[len(results)-1]
		if !math.IsNaN(last.Score) {
			latest = &models.SimpleScore{
				Date:  last.Date.Format("2006-01-02"),
				Score: last.Score,
				Label: last.Label,
				Price: last.Price, // Added Price
			}

			latestSubscores = map[string]float64{
				"trend":      last.Values.Trend,
				"momentum":   last.Values.Momentum,
				"rsi":        last.Values.RSI,
				"macd":       last.Values.MACD,
				"drawdown":   last.Values.Drawdown,
				"volatility": last.Values.Vol,
				"mfi":        last.Values.MFI,
				"bb_pct_b":   last.Values.BB,
			}
		}
	}

	var components []models.Component
	var method map[string]interface{}

	if lang == "en" {
		components = []models.Component{
			{
				ID:          "trend",
				Name:        "Trend Strength",
				Description: "Price vs MA20/60 position",
				Detail:      "Trend Strength measures the current price relative to long-term (60-day) and medium-term (20-day) moving averages. Price above MAs indicates strong uptrend (Greed).",
				Weight:      0.15,
			},
			{
				ID:          "momentum",
				Name:        "Momentum",
				Description: "20-day return, short-term power",
				Detail:      "Momentum is based on the cumulative return over the past 20 trading days. Higher positive returns indicate stronger upward momentum (Greed).",
				Weight:      0.15,
			},
			{
				ID:          "rsi",
				Name:        "RSI",
				Description: "Relative Strength Index (14D)",
				Detail:      "RSI measures the speed and change of price movements. RSI > 70 is considered overbought (Extreme Greed), while RSI < 30 is oversold (Extreme Fear).",
				Weight:      0.10,
			},
			{
				ID:          "macd",
				Name:        "MACD",
				Description: "MACD Histogram, momentum shift",
				Detail:      "The MACD histogram reflects the convergence and divergence of trends. Expanding positive values indicate strengthening upward momentum.",
				Weight:      0.10,
			},
			{
				ID:          "drawdown",
				Name:        "Drawdown",
				Description: "Drop from 252-day high",
				Detail:      "Drawdown calculates the percentage drop from the highest price in the past 252 trading days. Smaller drawdown indicates a stronger market.",
				Weight:      0.10,
			},
			{
				ID:          "volatility",
				Name:        "Volatility",
				Description: "20-day realized volatility",
				Detail:      "Volatility is based on the standard deviation of returns over 20 days. Spikes in volatility often accompany market panic (Fear).",
				Weight:      0.10,
			},
			{
				ID:          "mfi",
				Name:        "Money Flow (MFI)",
				Description: "Volume-weighted RSI (14D)",
				Detail:      "MFI incorporates both price and volume to measure buying and selling pressure. It is often a leading indicator for reversals compared to standard RSI.",
				Weight:      0.15,
			},
			{
				ID:          "bb_pct_b",
				Name:        "Bollinger %B",
				Description: "Price vs Bollinger Bands",
				Detail:      "Bollinger %B quantifies a security's price relative to the upper and lower Bollinger Bands. %B > 1 indicates price is above the upper band (Greed/Overbought).",
				Weight:      0.15,
			},
		}
		method = map[string]interface{}{
			"reference_window_trading_days": window,
			"normalize":                     "For each sub-indicator, calculate its rolling percentile within the reference window and map it to a 0-100 score.",
			"aggregate":                     "Total score is the weighted average of available sub-scores: sum(score_i * w_i) / sum(w_i).",
		}
	} else {
		// Default Chinese
		components = []models.Component{
			{
				ID:          "trend",
				Name:        "趋势强度",
				Description: "价格相对均线(MA20/60)的位置，越高越强",
				Detail:      "趋势强度衡量当前价格相对于长期（60日）和中期（20日）均线的位置。价格位于均线上方表明上升趋势强劲（贪婪），反之则为下降趋势（恐惧）。",
				Weight:      0.15,
			},
			{
				ID:          "momentum",
				Name:        "动量",
				Description: "20日收益率，反映短期冲力",
				Detail:      "动量指标基于过去 20 个交易日的累计收益率。正收益率越高表示上涨动能越强，可能引发贪婪情绪；负收益率表示下跌动能。",
				Weight:      0.15,
			},
			{
				ID:          "rsi",
				Name:        "RSI",
				Description: "相对强弱指标(14日)，反映超买超卖",
				Detail:      "相对强弱指数（RSI）衡量价格变动的速度和幅度。RSI > 70 通常被视为超买（极度贪婪），而 RSI < 30 则被视为超卖（极度恐惧）。",
				Weight:      0.10,
			},
			{
				ID:          "macd",
				Name:        "MACD",
				Description: "MACD柱状图，反映动能变化",
				Detail:      "MACD 柱状图反映了短期和长期趋势的聚合与分离。正值扩大表示上涨动能增强，负值扩大表示下跌动能增强。",
				Weight:      0.10,
			},
			{
				ID:          "drawdown",
				Name:        "回撤压力",
				Description: "距离252日高点的跌幅，越小越好",
				Detail:      "回撤压力计算当前价格距离过去 252 个交易日（一年）最高点的跌幅。回撤越小，市场越强势；回撤越大，市场恐慌情绪越重。",
				Weight:      0.10,
			},
			{
				ID:          "volatility",
				Name:        "波动率",
				Description: "20日实现波动率，越低越稳定",
				Detail:      "波动率基于 20 日历史价格的标准差。波动率飙升通常伴随着市场恐慌（恐惧），而低波动率通常对应市场的温和上涨（贪婪）。",
				Weight:      0.10,
			},
			{
				ID:          "mfi",
				Name:        "资金流量 (MFI)",
				Description: "结合成交量的RSI，反映资金进出",
				Detail:      "MFI 指标综合了价格和成交量来衡量买卖压力。相比普通的 RSI，MFI 往往能更早地发现顶背离和底背离信号。",
				Weight:      0.15,
			},
			{
				ID:          "bb_pct_b",
				Name:        "布林带位置 (%B)",
				Description: "价格在布林带中的相对位置",
				Detail:      "布林带 %B 量化了当前价格相对于布林带上下轨的位置。%B > 1 表示股价突破上轨（贪婪/超买），%B < 0 表示跌破下轨（恐惧/超卖）。",
				Weight:      0.15,
			},
		}
		method = map[string]interface{}{
			"reference_window_trading_days": window,
			"normalize":                     "对每个子指标，计算其在参考周期内的滚动分位数，并映射为 0–100 分。",
			"aggregate":                     "总分为可用子分数的加权平均：sum(score_i * w_i) / sum(w_i)。",
		}
	}

	resp := map[string]interface{}{
		"ticker":           ticker,
		"frequency":        freq,
		"latest":           latest,
		"series":           safeSeries,
		"method":           method,
		"components":       components,
		"latest_subscores": latestSubscores,
	}

	// Set Cache
	memCache.Set(cacheKey, resp, cache.DefaultExpiration)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(resp)
}

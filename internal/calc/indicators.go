package calc

import (
	"math"
)

// Money Flow Index (MFI)
func MFI(high, low, close, volume []float64, window int) []float64 {
	out := make([]float64, len(close))
	for i := range out {
		out[i] = math.NaN()
	}
	if len(close) < window+1 {
		return out
	}

	// Typical Price = (High + Low + Close) / 3
	// Raw Money Flow = Typical Price * Volume
	typicalPrice := make([]float64, len(close))
	rawMoneyFlow := make([]float64, len(close))

	for i := range close {
		typicalPrice[i] = (high[i] + low[i] + close[i]) / 3.0
		rawMoneyFlow[i] = typicalPrice[i] * volume[i]
	}

	posFlow := make([]float64, len(close))
	negFlow := make([]float64, len(close))

	for i := 1; i < len(close); i++ {
		if typicalPrice[i] > typicalPrice[i-1] {
			posFlow[i] = rawMoneyFlow[i]
		} else if typicalPrice[i] < typicalPrice[i-1] {
			negFlow[i] = rawMoneyFlow[i]
		}
		// if equal, discard
	}

	// Calculate initial 
	sumPos := 0.0
	sumNeg := 0.0
	for i := 1; i <= window; i++ {
		sumPos += posFlow[i]
		sumNeg += negFlow[i]
	}

	// Rolling
	for i := window; i < len(close); i++ {
		if i > window {
			sumPos = sumPos - posFlow[i-window] + posFlow[i]
			sumNeg = sumNeg - negFlow[i-window] + negFlow[i]
		}

		mfr := 0.0
		if sumNeg != 0 {
			mfr = sumPos / sumNeg
		} else if sumPos > 0 {
			mfr = 1e9 // Max
		} else {
			mfr = 0 // Both 0?
		}

		out[i] = 100.0 - (100.0 / (1.0 + mfr))
	}

	return out
}

// Bollinger Bands %B
func BollingerPercentB(close []float64, window int, numStdDev float64) []float64 {
	out := make([]float64, len(close))
	for i := range out {
		out[i] = math.NaN()
	}
	if len(close) < window {
		return out
	}

	sma := SMA(close, window)

	// Rolling StdDev
	for i := window - 1; i < len(close); i++ {
		sumSq := 0.0
		for j := 0; j < window; j++ {
			val := close[i-j]
			d := val - sma[i]
			sumSq += d * d
		}
		stdDev := math.Sqrt(sumSq / float64(window)) // Population std dev usually used in BB

		upper := sma[i] + (stdDev * numStdDev)
		lower := sma[i] - (stdDev * numStdDev)

		if upper != lower {
			out[i] = (close[i] - lower) / (upper - lower)
		} else {
			out[i] = 0.5
		}
	}

	return out
}

// Simple Moving Average
func SMA(values []float64, window int) []float64 {
	out := make([]float64, len(values))
	for i := range values {
		out[i] = math.NaN()
	}
	if len(values) < window {
		return out
	}
	
	sum := 0.0
	for i := 0; i < window; i++ {
		sum += values[i]
	}
	out[window-1] = sum / float64(window)
	
	for i := window; i < len(values); i++ {
		sum = sum - values[i-window] + values[i]
		out[i] = sum / float64(window)
	}
	return out
}

// Momentum (Return over N periods)
func Momentum(values []float64, window int) []float64 {
	out := make([]float64, len(values))
	for i := range out {
		out[i] = math.NaN()
	}
	for i := window; i < len(values); i++ {
		if values[i-window] != 0 {
			out[i] = (values[i] / values[i-window]) - 1.0
		}
	}
	return out
}

// RSI
func RSI(values []float64, window int) []float64 {
	out := make([]float64, len(values))
	for i := range out {
		out[i] = math.NaN()
	}
	if len(values) < window+1 {
		return out
	}

	gains := make([]float64, len(values))
	losses := make([]float64, len(values))

	for i := 1; i < len(values); i++ {
		diff := values[i] - values[i-1]
		if diff > 0 {
			gains[i] = diff
		} else {
			losses[i] = -diff
		}
	}

	// First average
	avgGain := 0.0
	avgLoss := 0.0
	for i := 1; i <= window; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(window)
	avgLoss /= float64(window)

	// Wilder's Smoothing
	for i := window + 1; i < len(values); i++ {
		avgGain = (avgGain*float64(window-1) + gains[i]) / float64(window)
		avgLoss = (avgLoss*float64(window-1) + losses[i]) / float64(window)
		
		rs := 0.0
		if avgLoss != 0 {
			rs = avgGain / avgLoss
		} else if avgGain == 0 {
			rs = 0 
		} else {
			rs = 1e9 // Max
		}
		
		out[i] = 100 - (100 / (1 + rs))
	}
	
	// Initial value
	rs := 0.0
	if avgLoss != 0 {
		rs = avgGain / avgLoss
	}
	if avgLoss == 0 && avgGain > 0 {
		out[window] = 100
	} else {
		out[window] = 100 - (100 / (1 + rs))
	}

	return out
}

// MACD Histogram
func MACD(values []float64) []float64 {
	fast := EMA(values, 12)
	slow := EMA(values, 26)
	macdLine := make([]float64, len(values))
	for i := range values {
		macdLine[i] = fast[i] - slow[i]
	}
	signal := EMA(macdLine, 9)
	hist := make([]float64, len(values))
	for i := range values {
		hist[i] = macdLine[i] - signal[i]
	}
	return hist
}

func EMA(values []float64, span int) []float64 {
	out := make([]float64, len(values))
	k := 2.0 / float64(span+1)
	
	// Start with SMA or first value
	// Standard Pandas ewm adjust=False starts with first value as mean
	if len(values) == 0 {
		return out
	}
	out[0] = values[0]
	for i := 1; i < len(values); i++ {
		out[i] = values[i]*k + out[i-1]*(1-k)
	}
	return out
}

// Realized Volatility (Annualized)
func RealizedVol(values []float64, window int) []float64 {
	out := make([]float64, len(values))
	for i := range out {
		out[i] = math.NaN()
	}
	
	returns := make([]float64, len(values))
	for i := 1; i < len(values); i++ {
		if values[i-1] != 0 {
			returns[i] = (values[i] / values[i-1]) - 1.0
		}
	}

	for i := window; i < len(values); i++ {
		// Sample std dev of returns over window
		sum := 0.0
		for j := 0; j < window; j++ {
			sum += returns[i-j]
		}
		mean := sum / float64(window)
		
		sqSum := 0.0
		for j := 0; j < window; j++ {
			d := returns[i-j] - mean
			sqSum += d * d
		}
		std := math.Sqrt(sqSum / float64(window)) // Pandas default ddof=1? No, usually ddof=1. Python code used ddof=0? 
		// Python code: .std(ddof=0) * np.sqrt(252)
		
		out[i] = std * math.Sqrt(252)
	}
	return out
}

// Rolling Percentile (0-100)
// For each point, look back `window` periods (including current), find rank of current value.
func RollingScore(values []float64, window int, direction int) []float64 {
	out := make([]float64, len(values))
	for i := range out {
		out[i] = math.NaN()
	}

	// Naive implementation O(N*W) is fine for N=600, W=252 => 150k ops
	buf := make([]float64, 0, window)

	for i := 0; i < len(values); i++ {
		if math.IsNaN(values[i]) {
			continue
		}
		
		// Collect valid history window
		buf = buf[:0]
		start := i - window + 1
		if start < 0 {
			start = 0
		}
		
		currentVal := values[i] * float64(direction)

		// We need to verify min_periods? Python code used min_periods=window
		// But here we might want to be flexible. Let's strictly require window for stability
		// EXCEPT if total data is short, handled outside.
		// Let's implement min_periods logic: if valid count < window, return NaN?
		// The python code: rolling(window=norm_window, min_periods=norm_window)
		
		// However, to prevent NaN start, we often allow smaller window at start? 
		// Python code strictly waits for window.
		if i < window-1 {
			out[i] = math.NaN()
			continue
		}

		for j := start; j <= i; j++ {
			if !math.IsNaN(values[j]) {
				buf = append(buf, values[j]*float64(direction))
			}
		}

		if len(buf) == 0 {
			continue
		}

		// Rank
		countSmaller := 0
		for _, v := range buf {
			if v <= currentVal {
				countSmaller++
			}
		}
		
		pct := float64(countSmaller) / float64(len(buf))
		out[i] = pct * 100.0
	}
	
	return out
}

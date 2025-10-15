package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Candle represents a single kline/candlestick from Binance
type Candle struct {
	OpenTime                 time.Time
	Open                     float64
	High                     float64
	Low                      float64
	Close                    float64
	Volume                   float64
	CloseTime                time.Time
	QuoteAssetVolume         float64
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  float64
	TakerBuyQuoteAssetVolume float64
}

// Global flag to determine if we're using futures or spot market
var USE_FUTURES bool = false

// GetBaseURL returns the appropriate base URL based on market type
func GetBaseURL() string {
	if USE_FUTURES {
		return "https://fapi.binance.com"
	}
	return "https://api.binance.com"
}

// GetKlinesEndpoint returns the appropriate klines endpoint based on market type
func GetKlinesEndpoint() string {
	if USE_FUTURES {
		return "/fapi/v1/klines"
	}
	return "/api/v3/klines"
}

func fetchKlines(symbol, interval string, limit int) ([]Candle, error) {
	baseURL := GetBaseURL()
	endpoint := GetKlinesEndpoint()
	url := fmt.Sprintf("%s%s?symbol=%s&interval=%s&limit=%d", baseURL, endpoint, symbol, interval, limit)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var raw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	candles := make([]Candle, 0, len(raw))
	for _, k := range raw {
		if len(k) < 11 { // basic sanity
			continue
		}
		// Helper to parse float strings
		parseF := func(v interface{}) float64 {
			s, _ := v.(string)
			f, _ := strconv.ParseFloat(s, 64)
			return f
		}
		openTimeMs, _ := k[0].(float64)
		closeTimeMs, _ := k[6].(float64)
		c := Candle{
			OpenTime:                 time.UnixMilli(int64(openTimeMs)),
			Open:                     parseF(k[1]),
			High:                     parseF(k[2]),
			Low:                      parseF(k[3]),
			Close:                    parseF(k[4]),
			Volume:                   parseF(k[5]),
			CloseTime:                time.UnixMilli(int64(closeTimeMs)),
			QuoteAssetVolume:         parseF(k[7]),
			NumberOfTrades:           int64(parseF(k[8])),
			TakerBuyBaseAssetVolume:  parseF(k[9]),
			TakerBuyQuoteAssetVolume: parseF(k[10]),
		}
		candles = append(candles, c)
	}
	return candles, nil
}

func main() {
	symbol := flag.String("symbol", "BTCUSDT", "Trading pair symbol (e.g., BTCUSDT, ETHUSDT)")
	interval := flag.String("interval", "4h", "Timeframe interval (e.g., 1m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M)")
	limit := flag.Int("limit", 1000, "Number of candles to fetch (max 1000)")
	paperMode := flag.Bool("paper", false, "Enable paper trading mode (simulated trades)")
	balance := flag.Float64("balance", 10000.0, "Starting balance for paper trading")

	// Multi-symbol analysis flags
	multiSymbol := flag.Bool("multi", false, "Enable multi-symbol analysis")
	topN := flag.Int("top", 50, "Number of top symbols by volume to analyze (use with --multi)")
	allSymbols := flag.Bool("all", false, "Analyze ALL USDT pairs (500+ symbols, use with caution)")

	// Multi-symbol paper trading flags
	multiPaper := flag.Bool("multi-paper", false, "Enable multi-symbol paper trading (trade multiple coins simultaneously)")
	maxPositions := flag.Int("max-pos", 5, "Maximum simultaneous positions (use with --multi-paper)")

	// Display mode flags
	quiet := flag.Bool("quiet", false, "Quiet mode - only show trading signals and P/L (no technical details)")

	// Market type flag
	futures := flag.Bool("futures", false, "Use Binance Futures market (default: spot market)")

	flag.Parse()

	// Set market type
	USE_FUTURES = *futures
	
	// Display market type
	marketType := "SPOT"
	if USE_FUTURES {
		marketType = "FUTURES"
	}
	fmt.Printf("📊 Market Type: %s\n", marketType)

	// Apply quiet mode settings
	if *quiet {
		SetQuietMode(true)
	}

	*symbol = strings.ToUpper(*symbol)

	// Multi-symbol paper trading mode
	if *multiPaper {
		var symbols []string
		var err error

		if *allSymbols {
			fmt.Println("🔍 Fetching all USDT trading pairs from Binance...")
			symbols, err = FetchAllBinanceSymbols()
		} else {
			fmt.Printf("🔍 Fetching top %d symbols by 24h volume...\n", *topN)
			symbols, err = FilterTopSymbolsByVolume(*topN)
		}

		if err != nil {
			fmt.Printf("❌ Failed to fetch symbols: %v\n", err)
			return
		}

		fmt.Printf("✅ Found %d symbols\n", len(symbols))
		fmt.Println()

		engine := NewMultiPaperTradingEngine(symbols, *interval, *limit, *balance, *maxPositions)
		if err := engine.RunMultiPaperTrading(); err != nil {
			fmt.Printf("❌ Multi-symbol paper trading error: %v\n", err)
		}

		return
	}

	// Multi-symbol analysis mode
	if *multiSymbol || *allSymbols {
		var symbols []string
		var err error

		if *allSymbols {
			fmt.Println("🔍 Fetching all USDT trading pairs from Binance...")
			symbols, err = FetchAllBinanceSymbols()
		} else {
			fmt.Printf("🔍 Fetching top %d symbols by 24h volume...\n", *topN)
			symbols, err = FilterTopSymbolsByVolume(*topN)
		}

		if err != nil {
			fmt.Printf("❌ Failed to fetch symbols: %v\n", err)
			return
		}

		fmt.Printf("✅ Found %d symbols\n", len(symbols))
		fmt.Println()

		if ENABLE_LIVE_MODE {
			if err := RunMultiSymbolLiveMode(symbols, *interval, *limit); err != nil {
				fmt.Printf("❌ Multi-symbol live mode error: %v\n", err)
			}
		} else {
			results := RunMultiSymbolAnalysis(symbols, *interval, *limit)
			PrintMultiSymbolResults(results)
		}

		return
	}

	// Single symbol modes
	if *paperMode {
		engine := NewPaperTradingEngine(*symbol, *interval, *limit, *balance)
		if err := engine.RunPaperTrading(); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	} else {
		RunEngine(*symbol, *interval, *limit)
	}
}

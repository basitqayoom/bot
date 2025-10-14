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

func fetchKlines(symbol, interval string, limit int) ([]Candle, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, limit)
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
	flag.Parse()

	*symbol = strings.ToUpper(*symbol)

	if *paperMode {
		engine := NewPaperTradingEngine(*symbol, *interval, *limit, *balance)
		if err := engine.RunPaperTrading(); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	} else {
		RunEngine(*symbol, *interval, *limit)
	}
}

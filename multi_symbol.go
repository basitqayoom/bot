package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ==================== BINANCE SYMBOL FETCHING ====================

type BinanceSymbolInfo struct {
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
}

type BinanceExchangeInfo struct {
	Symbols []BinanceSymbolInfo `json:"symbols"`
}

// FetchAllBinanceSymbols retrieves all trading pairs from Binance
func FetchAllBinanceSymbols() ([]string, error) {
	url := "https://api.binance.com/api/v3/exchangeInfo"

	fmt.Println("🔄 Fetching exchange info from Binance...")

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var exchangeInfo BinanceExchangeInfo
	if err := json.Unmarshal(body, &exchangeInfo); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var symbols []string
	for _, s := range exchangeInfo.Symbols {
		// Only include actively trading USDT pairs
		if s.Status == "TRADING" && s.QuoteAsset == "USDT" {
			symbols = append(symbols, s.Symbol)
		}
	}

	return symbols, nil
}

// FilterTopSymbolsByVolume gets the top N symbols by 24h volume
func FilterTopSymbolsByVolume(limit int) ([]string, error) {
	url := "https://api.binance.com/api/v3/ticker/24hr"

	fmt.Printf("🔄 Fetching 24h volume data for ranking...\n")

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tickers []struct {
		Symbol      string `json:"symbol"`
		QuoteVolume string `json:"quoteVolume"`
	}

	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, err
	}

	type SymbolVolume struct {
		Symbol string
		Volume float64
	}

	var volumeList []SymbolVolume
	for _, t := range tickers {
		if strings.HasSuffix(t.Symbol, "USDT") {
			var vol float64
			fmt.Sscanf(t.QuoteVolume, "%f", &vol)
			volumeList = append(volumeList, SymbolVolume{t.Symbol, vol})
		}
	}

	// Sort by volume (descending) - bubble sort for simplicity
	for i := 0; i < len(volumeList)-1; i++ {
		for j := i + 1; j < len(volumeList); j++ {
			if volumeList[j].Volume > volumeList[i].Volume {
				volumeList[i], volumeList[j] = volumeList[j], volumeList[i]
			}
		}
	}

	var result []string
	for i := 0; i < limit && i < len(volumeList); i++ {
		result = append(result, volumeList[i].Symbol)
	}

	return result, nil
}

// ==================== MULTI-SYMBOL ANALYSIS ====================

type MultiSymbolResult struct {
	Symbol      string
	Divergences int
	SRZones     int
	CurrentRSI  float64
	HasSignal   bool
	SignalType  string
	Error       error
	Duration    time.Duration
}

// RunMultiSymbolAnalysis analyzes multiple symbols in parallel
func RunMultiSymbolAnalysis(symbols []string, interval string, limit int) []MultiSymbolResult {
	fmt.Printf("\n╔════════════════════════════════════════╗\n")
	fmt.Printf("║   MULTI-SYMBOL PARALLEL ANALYSIS       ║\n")
	fmt.Printf("╚════════════════════════════════════════╝\n")
	fmt.Printf("\n🚀 Analyzing %d symbols on %s timeframe\n", len(symbols), interval)
	fmt.Printf("⚡ Using %d parallel workers\n", NUM_WORKERS)
	fmt.Println()

	var wg sync.WaitGroup
	resultsChan := make(chan MultiSymbolResult, len(symbols))
	semaphore := make(chan struct{}, NUM_WORKERS)

	startTime := time.Now()

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()

			// Limit concurrent API calls to avoid rate limiting
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			start := time.Now()
			result := MultiSymbolResult{Symbol: sym}

			// Create engine and run analysis
			engine := NewOptimizedEngine(sym, interval, limit)

			// Fetch data
			if err := engine.FetchData(); err != nil {
				result.Error = err
				resultsChan <- result
				return
			}

			// Calculate indicators
			engine.CalculateIndicators()
			engine.FindDivergences()
			engine.IdentifySupportResistance()

			// Store results
			result.Divergences = len(engine.Divergences)
			result.SRZones = len(engine.SRZones)

			if len(engine.RSI) > 0 {
				result.CurrentRSI = engine.RSI[len(engine.RSI)-1]
			}

			// Check for trading signals
			recentDivergences := 0
			for _, div := range engine.Divergences {
				divTime, _ := time.Parse("2006-01-02 15:04", div.EndTime)
				hoursSince := time.Since(divTime).Hours()
				if hoursSince < 72 {
					recentDivergences++
				}
			}

			if recentDivergences >= MIN_DIVERGENCES_FOR_SIGNAL && result.CurrentRSI > 70 {
				result.HasSignal = true
				result.SignalType = "SHORT"
			}

			result.Duration = time.Since(start)
			resultsChan <- result
		}(symbol)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results and show progress
	var results []MultiSymbolResult
	completed := 0

	fmt.Println("Progress:")
	for result := range resultsChan {
		results = append(results, result)
		completed++

		status := "✅"
		if result.Error != nil {
			status = "❌"
		}

		signal := " "
		if result.HasSignal {
			signal = "🔔"
		}

		fmt.Printf("\r[%d/%d] %s %s %s - RSI: %.2f, Div: %d, S/R: %d (%.2fs)   ",
			completed, len(symbols), status, signal, result.Symbol,
			result.CurrentRSI, result.Divergences, result.SRZones, result.Duration.Seconds())
	}

	totalDuration := time.Since(startTime)

	fmt.Printf("\n\n⚡ Completed in %.2f seconds\n", totalDuration.Seconds())
	fmt.Printf("📊 Average per symbol: %.2f seconds\n", totalDuration.Seconds()/float64(len(symbols)))

	return results
}

// PrintMultiSymbolResults displays analysis results and highlights signals
func PrintMultiSymbolResults(results []MultiSymbolResult) {
	fmt.Println("\n════════════════════════════════════════")
	fmt.Println("🎯 SYMBOLS WITH TRADE SIGNALS")
	fmt.Println("════════════════════════════════════════")

	signalCount := 0
	for _, r := range results {
		if r.HasSignal {
			signalCount++
			fmt.Printf("\n🔔 %s\n", r.Symbol)
			fmt.Printf("   📊 RSI: %.2f (Overbought)\n", r.CurrentRSI)
			fmt.Printf("   📈 Divergences: %d\n", r.Divergences)
			fmt.Printf("   🎯 S/R Zones: %d\n", r.SRZones)
			fmt.Printf("   📉 Signal: %s\n", r.SignalType)
		}
	}

	if signalCount == 0 {
		fmt.Println("\n⚠️  No trading signals found")
	} else {
		fmt.Printf("\n✅ Found signals in %d/%d symbols\n", signalCount, len(results))
	}

	// Show errors if any
	errorCount := 0
	for _, r := range results {
		if r.Error != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		fmt.Printf("\n⚠️  %d symbols had errors (API limits or data issues)\n", errorCount)
	}

	// Top RSI symbols
	fmt.Println("\n════════════════════════════════════════")
	fmt.Println("📊 TOP 10 OVERBOUGHT SYMBOLS (RSI > 60)")
	fmt.Println("════════════════════════════════════════")

	// Sort by RSI
	sortedByRSI := make([]MultiSymbolResult, len(results))
	copy(sortedByRSI, results)

	for i := 0; i < len(sortedByRSI)-1; i++ {
		for j := i + 1; j < len(sortedByRSI); j++ {
			if sortedByRSI[j].CurrentRSI > sortedByRSI[i].CurrentRSI {
				sortedByRSI[i], sortedByRSI[j] = sortedByRSI[j], sortedByRSI[i]
			}
		}
	}

	count := 0
	for _, r := range sortedByRSI {
		if r.CurrentRSI > 60 && r.Error == nil && count < 10 {
			fmt.Printf("  %s: RSI %.2f (Div: %d, S/R: %d)\n",
				r.Symbol, r.CurrentRSI, r.Divergences, r.SRZones)
			count++
		}
	}

	if count == 0 {
		fmt.Println("  No symbols with RSI > 60")
	}

	fmt.Println("════════════════════════════════════════")
}

// RunMultiSymbolLiveMode continuously monitors multiple symbols
func RunMultiSymbolLiveMode(symbols []string, interval string, limit int) error {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   MULTI-SYMBOL LIVE MONITOR            ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n🚀 Monitoring %d symbols on %s timeframe\n", len(symbols), interval)
	fmt.Println("🔄 Live mode: Scanning on each candle close")
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println()

	// Use first symbol to track candle timing
	engine := NewTradingEngine(symbols[0], interval, limit)

	scanCount := 0

	for {
		if WAIT_FOR_CANDLE_CLOSE {
			engine.WaitForCandleClose()
		}

		scanCount++

		fmt.Println("\n" + strings.Repeat("═", 60))
		fmt.Printf("🔔 CANDLE CLOSED - Multi-Symbol Scan #%d\n", scanCount)
		fmt.Printf("⏰ %s IST (%s UTC)\n",
			getIST().Format("2006-01-02 15:04:05"),
			time.Now().UTC().Format("2006-01-02 15:04:05"))
		fmt.Println(strings.Repeat("═", 60))

		results := RunMultiSymbolAnalysis(symbols, interval, limit)
		PrintMultiSymbolResults(results)

		if !ENABLE_LIVE_MODE {
			break
		}

		fmt.Printf("\n⏳ Next scan in %d seconds (or at next candle close)...\n", CHECK_INTERVAL)
		time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
	}

	return nil
}

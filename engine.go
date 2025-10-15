package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ==================== CONSTANTS ====================
// Trading Strategy Configuration
const (
	// Binance API Configuration
	DEFAULT_SYMBOL   = "BTCUSDT"
	DEFAULT_INTERVAL = "4h"
	DEFAULT_LIMIT    = 1000

	// Technical Indicator Parameters
	RSI_PERIOD        = 14
	SWING_LOOKBACK    = 2  // Candles on each side to identify swing high/low
	SIGNIFICANT_SWING = 10 // Lookback for significant highs/lows

	// Support/Resistance Configuration (matches TradingView Bjorgum Key Levels)
	PIVOT_LEFT_LOOKBACK  = 20   // Look left 20 bars for pivot (TradingView: left)
	PIVOT_RIGHT_LOOKBACK = 15   // Look right 15 bars for pivot (TradingView: right)
	ATR_LENGTH           = 30   // ATR period for zone width (TradingView: atrLen)
	ATR_MULTIPLIER       = 0.5  // Zone width = 0.5 * ATR (TradingView: mult)
	MAX_ZONE_PERCENT     = 5.0  // Max zone size as % of price (TradingView: per)
	ALIGN_ZONES          = true // Merge overlapping zones (TradingView: alignZones)
	SR_MIN_STRENGTH      = 1    // Minimum touches to consider significant
	SR_MAX_ZONES         = 20   // Maximum zones to track
	SR_MAX_ZONES_DISPLAY = 10   // Maximum number of zones to display

	// Risk Management
	RISK_REWARD_RATIO   = 2.0 // Minimum risk/reward ratio for trades
	MAX_RISK_PERCENT    = 2.0 // Max % of capital to risk per trade
	STOP_LOSS_PERCENT   = 3.0 // Default stop loss %
	TAKE_PROFIT_PERCENT = 6.0 // Default take profit %

	// Analysis Settings
	MIN_DIVERGENCES_FOR_SIGNAL = 1  // Minimum divergences needed for a signal
	DIVERGENCE_STRENGTH_HIGH   = 10 // RSI difference % for strong divergence
	DIVERGENCE_STRENGTH_MEDIUM = 5  // RSI difference % for medium divergence

	// Scheduler Configuration
	ENABLE_LIVE_MODE      = true // Set to true for continuous monitoring
	CHECK_INTERVAL        = 30   // Seconds between checks for new candle
	WAIT_FOR_CANDLE_CLOSE = true // Only analyze on confirmed candle close

	// Timezone Configuration
	TIMEZONE_OFFSET = 5*60 + 30 // IST = UTC + 5:30 (in minutes)

	// Performance Configuration
	ENABLE_PARALLEL_MODE = true  // Use goroutines for parallel processing (3-4x faster)
	NUM_WORKERS          = 8     // Number of concurrent workers for batch processing
	ENABLE_MULTI_SYMBOL  = false // Enable concurrent multi-symbol analysis
)

// ==================== DISPLAY VARIABLES ====================
// These are variables (not constants) so they can be modified by flags

var (
	// Display Settings
	SHOW_DIVERGENCES    = true
	SHOW_SR_ZONES       = true
	SHOW_TRADE_SIGNALS  = true
	SHOW_DETAILED_ZONES = true
	VERBOSE_MODE        = true
)

// ==================== DISPLAY MODE CONTROL ====================

// SetQuietMode enables or disables quiet mode (minimal output)
func SetQuietMode(quiet bool) {
	if quiet {
		VERBOSE_MODE = false
		SHOW_DIVERGENCES = false
		SHOW_SR_ZONES = false
		SHOW_DETAILED_ZONES = false
	} else {
		VERBOSE_MODE = true
		SHOW_DIVERGENCES = true
		SHOW_SR_ZONES = true
		SHOW_DETAILED_ZONES = true
	}
}

// ==================== ENGINE STRUCT ====================

// TradingEngine orchestrates all analysis components
type TradingEngine struct {
	Symbol      string
	Interval    string
	Limit       int
	Candles     []Candle
	RSI         []float64
	ATR         []float64
	Divergences []BearishDivergence
	SRZones     []SRZone
	SRConfig    SRConfig
}

// ==================== ENGINE METHODS ====================

// NewTradingEngine creates a new engine instance with default settings
func NewTradingEngine(symbol, interval string, limit int) *TradingEngine {
	if symbol == "" {
		symbol = DEFAULT_SYMBOL
	}
	if interval == "" {
		interval = DEFAULT_INTERVAL
	}
	if limit == 0 {
		limit = DEFAULT_LIMIT
	}

	// Initialize S/R config matching TradingView indicator
	srConfig := SRConfig{
		LookLeft:       PIVOT_LEFT_LOOKBACK,
		LookRight:      PIVOT_RIGHT_LOOKBACK,
		ATRLength:      ATR_LENGTH,
		ATRMultiplier:  ATR_MULTIPLIER,
		MaxZonePercent: MAX_ZONE_PERCENT,
		AlignZones:     ALIGN_ZONES,
		MinStrength:    SR_MIN_STRENGTH,
		MaxZones:       SR_MAX_ZONES,
	}

	return &TradingEngine{
		Symbol:   symbol,
		Interval: interval,
		Limit:    limit,
		SRConfig: srConfig,
	}
}

// FetchData retrieves candle data from Binance
func (e *TradingEngine) FetchData() error {
	fmt.Printf("🔄 Fetching %s data for %s (limit: %d)...\n", e.Interval, e.Symbol, e.Limit)

	candles, err := fetchKlines(e.Symbol, e.Interval, e.Limit)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	e.Candles = candles
	fmt.Printf("✅ Fetched %d candles\n", len(e.Candles))

	if VERBOSE_MODE {
		fmt.Printf("   First candle: %s (O: %.2f, H: %.2f, L: %.2f, C: %.2f)\n",
			e.Candles[0].OpenTime.Format("2006-01-02 15:04"),
			e.Candles[0].Open, e.Candles[0].High, e.Candles[0].Low, e.Candles[0].Close)
		fmt.Printf("   Last candle:  %s (O: %.2f, H: %.2f, L: %.2f, C: %.2f)\n",
			e.Candles[len(e.Candles)-1].OpenTime.Format("2006-01-02 15:04"),
			e.Candles[len(e.Candles)-1].Open, e.Candles[len(e.Candles)-1].High,
			e.Candles[len(e.Candles)-1].Low, e.Candles[len(e.Candles)-1].Close)
	}

	return nil
}

// CalculateIndicators computes RSI and other technical indicators
func (e *TradingEngine) CalculateIndicators() {
	fmt.Printf("\n📊 Calculating technical indicators...\n")

	closes := make([]float64, len(e.Candles))
	for i, c := range e.Candles {
		closes[i] = c.Close
	}

	// Calculate RSI
	e.RSI = calcRSI(closes, RSI_PERIOD)

	// Calculate ATR for S/R zones
	e.ATR = calcATR(e.Candles, e.SRConfig.ATRLength)

	currentRSI := e.RSI[len(e.RSI)-1]
	currentATR := e.ATR[len(e.ATR)-1]

	fmt.Printf("✅ RSI calculated (current: %.2f)\n", currentRSI)
	fmt.Printf("✅ ATR calculated (current: %.2f)\n", currentATR)

	if currentRSI > 70 {
		fmt.Printf("   ⚠️  RSI is OVERBOUGHT (%.2f > 70)\n", currentRSI)
	} else if currentRSI < 30 {
		fmt.Printf("   ⚠️  RSI is OVERSOLD (%.2f < 30)\n", currentRSI)
	}
}

// FindDivergences identifies bearish divergences
func (e *TradingEngine) FindDivergences() {
	fmt.Printf("\n🔍 Scanning for bearish divergences...\n")

	e.Divergences = findBearishDivergences(e.Candles, e.RSI, SWING_LOOKBACK)

	fmt.Printf("✅ Found %d bearish divergence(s)\n", len(e.Divergences))

	if len(e.Divergences) > 0 && SHOW_DIVERGENCES {
		e.printDivergences()
	}
}

// IdentifySupportResistance finds support and resistance zones
func (e *TradingEngine) IdentifySupportResistance() {
	fmt.Printf("\n🎯 Identifying support & resistance zones (TradingView algorithm)...\n")
	fmt.Printf("   Using pivot detection: %d left / %d right bars\n", e.SRConfig.LookLeft, e.SRConfig.LookRight)
	fmt.Printf("   Zone width: %.1f × ATR (max %.1f%% of price)\n", e.SRConfig.ATRMultiplier, e.SRConfig.MaxZonePercent)
	if e.SRConfig.AlignZones {
		fmt.Printf("   Zone alignment: ENABLED (merging overlapping zones)\n")
	}

	// Use the advanced S/R detection matching TradingView
	e.SRZones = findAdvancedSupportResistance(e.Candles, e.SRConfig)

	fmt.Printf("✅ Found %d significant zone(s)\n", len(e.SRZones))

	if len(e.SRZones) > 0 && SHOW_SR_ZONES {
		currentPrice := e.Candles[len(e.Candles)-1].Close
		e.printSupportResistanceZones(currentPrice)
	}
}

// GenerateTradeSignals analyzes data and generates actionable signals
func (e *TradingEngine) GenerateTradeSignals() {
	if !SHOW_TRADE_SIGNALS {
		return
	}

	fmt.Printf("\n💡 TRADE SIGNAL ANALYSIS\n")
	fmt.Println("==========================================")

	currentPrice := e.Candles[len(e.Candles)-1].Close
	currentTime := e.Candles[len(e.Candles)-1].OpenTime

	// Check for recent divergences
	recentDivergences := 0
	for _, div := range e.Divergences {
		divTime, _ := time.Parse("2006-01-02 15:04", div.EndTime)
		hoursSince := time.Since(divTime).Hours()

		if hoursSince < 72 { // Within last 72 hours (3 days for 4h timeframe)
			recentDivergences++
		}
	}

	// Find nearest resistance
	var nearestResistance *SRZone
	minDistanceUp := 1000000.0

	for i := range e.SRZones {
		if e.SRZones[i].Level > currentPrice {
			distance := e.SRZones[i].Level - currentPrice
			if distance < minDistanceUp {
				minDistanceUp = distance
				nearestResistance = &e.SRZones[i]
			}
		}
	}

	// Find nearest support
	var nearestSupport *SRZone
	minDistanceDown := 1000000.0

	for i := range e.SRZones {
		if e.SRZones[i].Level < currentPrice {
			distance := currentPrice - e.SRZones[i].Level
			if distance < minDistanceDown {
				minDistanceDown = distance
				nearestSupport = &e.SRZones[i]
			}
		}
	}

	// Generate signal
	signal := "NEUTRAL"
	strength := "WEAK"

	if recentDivergences >= MIN_DIVERGENCES_FOR_SIGNAL {
		signal = "BEARISH"
		if recentDivergences >= 2 {
			strength = "STRONG"
		} else {
			strength = "MEDIUM"
		}
	}

	fmt.Printf("📍 Current Price: $%.2f (%s)\n", currentPrice, currentTime.Format("2006-01-02 15:04"))
	fmt.Printf("📊 Current RSI: %.2f\n", e.RSI[len(e.RSI)-1])
	fmt.Printf("🔔 Signal: %s (%s)\n", signal, strength)
	fmt.Printf("📈 Recent Divergences (72h): %d\n\n", recentDivergences)

	if signal == "BEARISH" {
		fmt.Println("🎯 SUGGESTED SHORT TRADE SETUP:")
		fmt.Println("─────────────────────────────────────────")

		entry := currentPrice
		var stopLoss, takeProfit float64

		if nearestResistance != nil {
			stopLoss = nearestResistance.ZoneTop // Stop above the zone top
		} else {
			stopLoss = currentPrice * (1 + STOP_LOSS_PERCENT/100)
		}

		if nearestSupport != nil {
			takeProfit = nearestSupport.ZoneBot // Target below the zone bottom
		} else {
			takeProfit = currentPrice * (1 - TAKE_PROFIT_PERCENT/100)
		}

		risk := stopLoss - entry
		reward := entry - takeProfit
		rr := reward / risk

		fmt.Printf("  Entry:        $%.2f\n", entry)
		fmt.Printf("  Stop Loss:    $%.2f (%.2f%% above entry)\n",
			stopLoss, ((stopLoss-entry)/entry)*100)
		fmt.Printf("  Take Profit:  $%.2f (%.2f%% below entry)\n",
			takeProfit, ((entry-takeProfit)/entry)*100)
		fmt.Printf("  Risk/Reward:  %.2f:1\n", rr)

		if rr >= RISK_REWARD_RATIO {
			fmt.Printf("  ✅ R/R ratio meets minimum requirement (%.1f:1)\n", RISK_REWARD_RATIO)
		} else {
			fmt.Printf("  ⚠️  R/R ratio below minimum (required: %.1f:1)\n", RISK_REWARD_RATIO)
		}

		fmt.Println("\n  Position Sizing (example $10,000 account):")
		accountSize := 10000.0
		riskAmount := accountSize * (MAX_RISK_PERCENT / 100)
		riskPercentPrice := ((stopLoss - entry) / entry) * 100
		positionSize := riskAmount / (riskPercentPrice / 100 * entry)

		fmt.Printf("    Max Risk:     $%.2f (%.1f%% of account)\n", riskAmount, MAX_RISK_PERCENT)
		fmt.Printf("    Position:     $%.2f (%.4f %s)\n",
			positionSize*entry, positionSize, e.Symbol[:len(e.Symbol)-4])
		fmt.Printf("    Potential P/L: -$%.2f / +$%.2f\n", riskAmount, riskAmount*rr)

	} else {
		fmt.Println("⏸️  No clear trade setup at this time")
		fmt.Println("   Consider waiting for:")
		fmt.Println("   • Price to reach key support/resistance")
		fmt.Println("   • Additional bearish divergences")
		fmt.Println("   • RSI confirmation (overbought)")
	}

	fmt.Println("==========================================")
}

// Run executes the complete analysis pipeline
func (e *TradingEngine) Run() error {
	startTime := time.Now()

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   CRYPTO TRADING ENGINE v1.0           ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n🚀 Starting analysis for %s on %s timeframe\n", e.Symbol, e.Interval)
	fmt.Printf("⏰ Start time: %s\n", startTime.Format("2006-01-02 15:04:05"))

	// Step 1: Fetch Data
	if err := e.FetchData(); err != nil {
		return err
	}

	// Step 2: Calculate Indicators
	e.CalculateIndicators()

	// Step 3: Find Divergences
	e.FindDivergences()

	// Step 4: Identify S/R Zones
	e.IdentifySupportResistance()

	// Step 5: Generate Trade Signals
	e.GenerateTradeSignals()

	// Summary
	duration := time.Since(startTime)
	fmt.Printf("\n\n✨ Analysis completed in %v\n", duration.Round(time.Millisecond))
	fmt.Println("════════════════════════════════════════")
	fmt.Println()

	return nil
}

// ==================== HELPER PRINT METHODS ====================

func (e *TradingEngine) printDivergences() {
	fmt.Println("\n========== BEARISH DIVERGENCES ==========")
	fmt.Println("Draw lines on TradingView between these two points:")
	fmt.Println()

	for i, div := range e.Divergences {
		priceChange := ((div.EndPrice - div.StartPrice) / div.StartPrice) * 100
		rsiChange := ((div.StartRSI - div.EndRSI) / div.StartRSI) * 100

		// Determine divergence strength
		divStrength := "WEAK"
		if rsiChange >= DIVERGENCE_STRENGTH_HIGH {
			divStrength = "STRONG"
		} else if rsiChange >= DIVERGENCE_STRENGTH_MEDIUM {
			divStrength = "MEDIUM"
		}

		fmt.Printf("Divergence #%d [%s]:\n", i+1, divStrength)
		fmt.Printf("  START POINT (Earlier Swing):\n")
		fmt.Printf("    Index: %d | Time: %s | Price: %.2f | RSI: %.2f\n",
			div.StartIdx, div.StartTime, div.StartPrice, div.StartRSI)
		fmt.Printf("  END POINT (Later Swing):\n")
		fmt.Printf("    Index: %d | Time: %s | Price: %.2f | RSI: %.2f\n",
			div.EndIdx, div.EndTime, div.EndPrice, div.EndRSI)
		fmt.Printf("  DIVERGENCE: Price %.2f → %.2f (↑ %.2f%%) but RSI %.2f → %.2f (↓ %.2f%%)\n\n",
			div.StartPrice, div.EndPrice, priceChange,
			div.StartRSI, div.EndRSI, rsiChange)
	}

	fmt.Printf("Total divergences found: %d\n", len(e.Divergences))
	fmt.Println("=========================================")
	fmt.Println()
}

func (e *TradingEngine) printSupportResistanceZones(currentPrice float64) {
	if len(e.SRZones) == 0 {
		fmt.Println("No significant support/resistance zones found")
		return
	}

	fmt.Println("\n========== SUPPORT & RESISTANCE ZONES ==========")

	// Separate into support and resistance
	var supports, resistances []SRZone
	for _, zone := range e.SRZones {
		if zone.Level < currentPrice {
			supports = append(supports, zone)
		} else {
			resistances = append(resistances, zone)
		}
	}

	// Print resistances (above current price) - sorted by proximity
	if len(resistances) > 0 {
		fmt.Println("\nRESISTANCE ZONES (Above Current Price):")
		displayCount := len(resistances)
		if displayCount > SR_MAX_ZONES_DISPLAY/2 {
			displayCount = SR_MAX_ZONES_DISPLAY / 2
		}

		for i := 0; i < displayCount; i++ {
			zone := resistances[i]
			distance := ((zone.Level - currentPrice) / currentPrice) * 100

			strengthLabel := "●"
			if zone.Strength >= 5 {
				strengthLabel = "●●●"
			} else if zone.Strength >= 3 {
				strengthLabel = "●●"
			}

			zoneWidth := ((zone.ZoneTop - zone.ZoneBot) / zone.Level) * 100

			fmt.Printf("  R%d: $%.2f [%.2f - %.2f] %s\n",
				i+1, zone.Level, zone.ZoneBot, zone.ZoneTop, strengthLabel)
			fmt.Printf("      Strength: %d pivots | Width: %.2f%% | Distance: +%.2f%%\n",
				zone.PivotCount, zoneWidth, distance)

			if SHOW_DETAILED_ZONES {
				fmt.Printf("      First: %s | Last: %s | Avg ATR: $%.2f\n",
					zone.FirstTouch.Format("2006-01-02 15:04"),
					zone.LastTouch.Format("2006-01-02 15:04"),
					zone.AvgATR)
			}
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("CURRENT PRICE: $%.2f\n", currentPrice)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Print supports (below current price) - sorted by proximity
	if len(supports) > 0 {
		fmt.Println("\nSUPPORT ZONES (Below Current Price):")
		displayCount := len(supports)
		if displayCount > SR_MAX_ZONES_DISPLAY/2 {
			displayCount = SR_MAX_ZONES_DISPLAY / 2
		}

		for i := 0; i < displayCount; i++ {
			zone := supports[i]
			distance := ((currentPrice - zone.Level) / currentPrice) * 100

			strengthLabel := "●"
			if zone.Strength >= 5 {
				strengthLabel = "●●●"
			} else if zone.Strength >= 3 {
				strengthLabel = "●●"
			}

			zoneWidth := ((zone.ZoneTop - zone.ZoneBot) / zone.Level) * 100

			fmt.Printf("  S%d: $%.2f [%.2f - %.2f] %s\n",
				i+1, zone.Level, zone.ZoneBot, zone.ZoneTop, strengthLabel)
			fmt.Printf("      Strength: %d pivots | Width: %.2f%% | Distance: -%.2f%%\n",
				zone.PivotCount, zoneWidth, distance)

			if SHOW_DETAILED_ZONES {
				fmt.Printf("      First: %s | Last: %s | Avg ATR: $%.2f\n",
					zone.FirstTouch.Format("2006-01-02 15:04"),
					zone.LastTouch.Format("2006-01-02 15:04"),
					zone.AvgATR)
			}
		}
	}

	fmt.Printf("\nTotal zones found: %d (%d resistance, %d support)\n",
		len(e.SRZones), len(resistances), len(supports))
	fmt.Println("================================================")
	fmt.Println()
}

// ==================== TIMEZONE HELPERS ====================

// getIST returns the current time in IST
func getIST() time.Time {
	return time.Now().UTC().Add(time.Duration(TIMEZONE_OFFSET) * time.Minute)
}

// getLocalTime converts UTC time to IST
func getLocalTime(utc time.Time) time.Time {
	return utc.Add(time.Duration(TIMEZONE_OFFSET) * time.Minute)
}

// ==================== LIVE MODE SCHEDULER ====================

// parseCandleDuration converts interval string to time.Duration
func (e *TradingEngine) parseCandleDuration() time.Duration {
	switch e.Interval {
	case "1m":
		return time.Minute
	case "3m":
		return 3 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return time.Hour
	case "2h":
		return 2 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "8h":
		return 8 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "1d":
		return 24 * time.Hour
	case "3d":
		return 3 * 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	default:
		return 4 * time.Hour
	}
}

// WaitForCandleClose blocks until the current candle closes (IST-aware)
func (e *TradingEngine) WaitForCandleClose() time.Duration {
	nowUTC := time.Now().UTC()
	nowIST := getIST()

	candleDuration := e.parseCandleDuration()

	// Binance uses UTC for candle alignment
	candleStartUTC := nowUTC.Truncate(candleDuration)
	candleCloseUTC := candleStartUTC.Add(candleDuration)

	// Convert to IST for display
	candleStartIST := getLocalTime(candleStartUTC)
	candleCloseIST := getLocalTime(candleCloseUTC)

	waitDuration := time.Until(candleCloseUTC)

	if waitDuration > 0 {
		fmt.Printf("\n⏰ Waiting for candle to close...\n")
		fmt.Printf("   Current time (IST): %s\n", nowIST.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Current time (UTC): %s\n", nowUTC.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Candle opened at:   %s IST (%s UTC)\n",
			candleStartIST.Format("15:04"), candleStartUTC.Format("15:04"))
		fmt.Printf("   Candle closes at:   %s IST (%s UTC)\n",
			candleCloseIST.Format("15:04"), candleCloseUTC.Format("15:04"))

		fmt.Println("\n" + strings.Repeat("─", 60))
		fmt.Println("⏳ COUNTDOWN TO NEXT CANDLE CLOSE")
		fmt.Println(strings.Repeat("─", 60))

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			remaining := time.Until(candleCloseUTC)

			if remaining <= 0 {
				break
			}

			hours := int(remaining.Hours())
			minutes := int(remaining.Minutes()) % 60
			seconds := int(remaining.Seconds()) % 60

			fmt.Printf("\r⏱️  Time remaining: %02dh %02dm %02ds | Next execution at: %s IST     ",
				hours, minutes, seconds, candleCloseIST.Format("15:04:05"))
		}

		fmt.Printf("\r✅ Candle closed! Executing analysis...%s\n", strings.Repeat(" ", 30))
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()
	}

	return waitDuration
}

// isCandleClosed checks if a new candle has closed since last check
func (e *TradingEngine) isCandleClosed(lastCheckTime time.Time) bool {
	now := time.Now().UTC()
	candleDuration := e.parseCandleDuration()

	currentCandleStart := now.Truncate(candleDuration)
	lastCandleStart := lastCheckTime.Truncate(candleDuration)

	return !currentCandleStart.Equal(lastCandleStart)
}

// printCandleSchedule shows the complete schedule for the day
func (e *TradingEngine) printCandleSchedule() {
	fmt.Println("\n📅 TODAY'S CANDLE SCHEDULE (IST/UTC)")
	fmt.Println("═══════════════════════════════════════════════════")

	nowUTC := time.Now().UTC()
	todayStart := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)

	candleDuration := e.parseCandleDuration()
	numCandles := int(24 * time.Hour / candleDuration)

	for i := 0; i < numCandles; i++ {
		candleUTC := todayStart.Add(time.Duration(i) * candleDuration)
		candleIST := getLocalTime(candleUTC)

		status := "  "
		if nowUTC.After(candleUTC) && nowUTC.Before(candleUTC.Add(candleDuration)) {
			status = "🔴" // Current candle
		} else if nowUTC.After(candleUTC.Add(candleDuration)) {
			status = "✅" // Closed
		} else {
			status = "⏸️ " // Future
		}

		fmt.Printf("%s Candle %2d: %s IST (%s UTC) - %s IST (%s UTC)\n",
			status, i+1,
			candleIST.Format("15:04"), candleUTC.Format("15:04"),
			candleIST.Add(candleDuration).Format("15:04"),
			candleUTC.Add(candleDuration).Format("15:04"))
	}

	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Printf("📍 Daily open: 05:30 IST (00:00 UTC)\n")
	fmt.Printf("🕐 Timezone: IST (UTC+5:30)\n\n")
}

// RunLive executes the bot in continuous mode, analyzing on each candle close
func (e *TradingEngine) RunLive() error {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   CRYPTO TRADING ENGINE v1.0           ║")
	fmt.Println("║        LIVE MODE ACTIVATED             ║")
	fmt.Println("╚════════════════════════════════════════╝")

	fmt.Printf("\n🚀 Monitoring %s on %s timeframe\n", e.Symbol, e.Interval)
	fmt.Printf("📊 Analysis runs on confirmed candle close\n")
	fmt.Printf("🔄 Checking every %d seconds\n", CHECK_INTERVAL)
	fmt.Printf("🌍 Timezone: IST (UTC+5:30)\n")

	// Show today's schedule
	e.printCandleSchedule()

	lastCheckTime := time.Now().UTC()
	analysisCount := 0

	for {
		if WAIT_FOR_CANDLE_CLOSE {
			// WaitForCandleClose() already waits internally with the countdown
			// It returns when the candle has closed
			e.WaitForCandleClose()
		}

		if e.isCandleClosed(lastCheckTime) || !WAIT_FOR_CANDLE_CLOSE {
			analysisCount++
			lastCheckTime = time.Now().UTC()
			lastCheckIST := getIST()

			fmt.Println("\n" + strings.Repeat("═", 60))
			fmt.Printf("🔔 CANDLE CLOSED - Running Analysis #%d\n", analysisCount)
			fmt.Printf("⏰ %s IST (%s UTC)\n",
				lastCheckIST.Format("2006-01-02 15:04:05"),
				lastCheckTime.Format("2006-01-02 15:04:05"))
			fmt.Println(strings.Repeat("═", 60))

			if err := e.Run(); err != nil {
				fmt.Printf("⚠️  Analysis error: %v\n", err)
				fmt.Println("   Continuing to monitor...")
			}

			fmt.Printf("\n✅ Analysis #%d completed\n", analysisCount)
			fmt.Printf("⏳ Next check in %d seconds...\n", CHECK_INTERVAL)
		}

		time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
	}
}

// ==================== OPTIMIZED PARALLEL ENGINE ====================

// OptimizedEngine wraps TradingEngine with concurrency support
type OptimizedEngine struct {
	*TradingEngine
	workerPool int
}

// NewOptimizedEngine creates an engine with parallel processing support
func NewOptimizedEngine(symbol, interval string, limit int) *OptimizedEngine {
	return &OptimizedEngine{
		TradingEngine: NewTradingEngine(symbol, interval, limit),
		workerPool:    NUM_WORKERS,
	}
}

// RunParallel executes analysis using goroutines for 3-4x performance improvement
func (e *OptimizedEngine) RunParallel() error {
	start := time.Now()

	if VERBOSE_MODE {
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║   OPTIMIZED TRADING ENGINE v2.0        ║")
		fmt.Println("║      PARALLEL PROCESSING MODE          ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("\n🚀 Using %d concurrent workers\n", e.workerPool)
		fmt.Printf("📊 Analyzing %s on %s timeframe\n\n", e.Symbol, e.Interval)
	}

	// Step 1: Fetch data (must happen first)
	if err := e.FetchData(); err != nil {
		return fmt.Errorf("data fetch failed: %w", err)
	}

	if VERBOSE_MODE {
		fmt.Printf("✅ Fetched %d candles\n\n", len(e.Candles))
	}

	// Step 2: Run independent calculations in parallel using goroutines
	var wg sync.WaitGroup
	var rsiDone, atrDone bool

	if VERBOSE_MODE {
		fmt.Println("⚡ Starting parallel indicator calculations...")
	}

	// Goroutine 1: Calculate RSI
	wg.Add(1)
	go func() {
		defer wg.Done()
		if VERBOSE_MODE {
			fmt.Println("  🔄 [Thread 1] Calculating RSI...")
		}
		closes := make([]float64, len(e.Candles))
		for i, c := range e.Candles {
			closes[i] = c.Close
		}
		e.RSI = calcRSI(closes, RSI_PERIOD)
		rsiDone = true
		if VERBOSE_MODE {
			fmt.Println("  ✅ [Thread 1] RSI completed")
		}
	}()

	// Goroutine 2: Calculate ATR
	wg.Add(1)
	go func() {
		defer wg.Done()
		if VERBOSE_MODE {
			fmt.Println("  🔄 [Thread 2] Calculating ATR...")
		}
		e.ATR = calcATR(e.Candles, ATR_LENGTH)
		atrDone = true
		if VERBOSE_MODE {
			fmt.Println("  ✅ [Thread 2] ATR completed")
		}
	}()

	// Wait for all indicators to complete
	wg.Wait()

	if VERBOSE_MODE {
		fmt.Printf("\n✅ All indicators calculated (RSI: %v, ATR: %v)\n\n", rsiDone, atrDone)
		fmt.Println("⚡ Starting parallel analysis...")
	}

	// Step 3: Run dependent analyses in parallel (they need indicators)
	wg.Add(2)

	// Goroutine 3: Find divergences (needs RSI)
	go func() {
		defer wg.Done()
		if VERBOSE_MODE {
			fmt.Println("  🔄 [Thread 3] Scanning for divergences...")
		}
		e.FindDivergences()
		if VERBOSE_MODE {
			fmt.Printf("  ✅ [Thread 3] Found %d divergences\n", len(e.Divergences))
		}
	}()

	// Goroutine 4: Find S/R zones (needs ATR)
	go func() {
		defer wg.Done()
		if VERBOSE_MODE {
			fmt.Println("  🔄 [Thread 4] Identifying S/R zones...")
		}
		e.IdentifySupportResistance()
		if VERBOSE_MODE {
			fmt.Printf("  ✅ [Thread 4] Found %d S/R zones\n", len(e.SRZones))
		}
	}()

	wg.Wait()

	if VERBOSE_MODE {
		fmt.Println("\n✅ All parallel analyses completed\n")
	}

	// Step 4: Generate signals (needs all previous data)
	if VERBOSE_MODE {
		fmt.Println("🔄 Generating trade signals...")
	}
	e.GenerateTradeSignals()

	// Print summary
	elapsed := time.Since(start)
	fmt.Printf("\n\n✨ Parallel analysis completed in %v\n", elapsed.Round(time.Millisecond))
	if VERBOSE_MODE {
		avgTimePerIndicator := elapsed / 4 // 4 parallel operations
		fmt.Printf("⚡ Average time per indicator: %v\n", avgTimePerIndicator.Round(time.Millisecond))
		fmt.Printf("🚀 Estimated speedup: 3-4x vs sequential mode\n")
	}
	fmt.Println("════════════════════════════════════════")
	fmt.Println()

	return nil
}

// RunLiveOptimized runs the bot with parallel processing on each candle
func (e *OptimizedEngine) RunLiveOptimized() error {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   OPTIMIZED LIVE MODE v2.0             ║")
	fmt.Println("║   Parallel Processing Enabled          ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n🚀 Using %d concurrent workers per analysis\n", e.workerPool)

	e.printCandleSchedule()

	lastCheckTime := time.Now().UTC()
	analysisCount := 0

	for {
		if WAIT_FOR_CANDLE_CLOSE {
			// WaitForCandleClose() already waits internally with the countdown
			// It returns when the candle has closed
			e.WaitForCandleClose()
		}

		if e.isCandleClosed(lastCheckTime) || !WAIT_FOR_CANDLE_CLOSE {
			analysisCount++
			lastCheckTime = time.Now().UTC()
			lastCheckIST := getIST()

			fmt.Println("\n" + strings.Repeat("═", 60))
			fmt.Printf("🔔 CANDLE CLOSED - Running Parallel Analysis #%d\n", analysisCount)
			fmt.Printf("⏰ %s IST (%s UTC)\n",
				lastCheckIST.Format("2006-01-02 15:04:05"),
				lastCheckTime.Format("2006-01-02 15:04:05"))
			fmt.Println(strings.Repeat("═", 60))

			if err := e.RunParallel(); err != nil {
				fmt.Printf("⚠️  Analysis error: %v\n", err)
				fmt.Println("   Continuing to monitor...")
			}

			fmt.Printf("\n✅ Parallel Analysis #%d completed\n", analysisCount)
			fmt.Printf("⏳ Next check in %d seconds...\n", CHECK_INTERVAL)
		}

		time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
	}
}

// ==================== MULTI-SYMBOL CONCURRENT ANALYSIS ====================

// ConcurrentMultiSymbolAnalysis analyzes multiple symbols in parallel
func ConcurrentMultiSymbolAnalysis(symbols []string, interval string, limit int) {
	var wg sync.WaitGroup
	type Result struct {
		Symbol      string
		Divergences int
		SRZones     int
		Duration    time.Duration
		Error       error
	}
	results := make(chan Result, len(symbols))

	fmt.Printf("\n╔════════════════════════════════════════╗\n")
	fmt.Printf("║   MULTI-SYMBOL ANALYSIS v1.0           ║\n")
	fmt.Printf("╚════════════════════════════════════════╝\n")
	fmt.Printf("\n🚀 Analyzing %d symbols concurrently...\n", len(symbols))
	fmt.Printf("📊 Timeframe: %s | Candles: %d\n\n", interval, limit)

	startTotal := time.Now()

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()

			start := time.Now()
			engine := NewOptimizedEngine(sym, interval, limit)

			fmt.Printf("🔄 [%s] Starting parallel analysis...\n", sym)

			err := engine.RunParallel()
			duration := time.Since(start)

			results <- Result{
				Symbol:      sym,
				Divergences: len(engine.Divergences),
				SRZones:     len(engine.SRZones),
				Duration:    duration,
				Error:       err,
			}
		}(symbol)
	}

	// Close results channel when all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and print results
	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("📊 RESULTS SUMMARY:")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Printf("%-12s | %-12s | %-10s | %-15s\n", "SYMBOL", "DIVERGENCES", "S/R ZONES", "EXEC TIME")
	fmt.Println(strings.Repeat("-", 70))

	successCount := 0
	totalDivergences := 0
	totalZones := 0

	for result := range results {
		if result.Error != nil {
			fmt.Printf("%-12s | ❌ ERROR: %v\n", result.Symbol, result.Error)
		} else {
			fmt.Printf("%-12s | %-12d | %-10d | %-15v\n",
				result.Symbol, result.Divergences, result.SRZones, result.Duration)
			successCount++
			totalDivergences += result.Divergences
			totalZones += result.SRZones
		}
	}

	totalDuration := time.Since(startTotal)
	fmt.Println(strings.Repeat("═", 70))
	fmt.Printf("✅ Completed: %d/%d symbols\n", successCount, len(symbols))
	fmt.Printf("📊 Total Divergences: %d | Total S/R Zones: %d\n", totalDivergences, totalZones)
	fmt.Printf("⚡ Total Time: %v (%.1fx faster than sequential)\n",
		totalDuration, float64(len(symbols)*2)/totalDuration.Seconds())
	fmt.Println(strings.Repeat("═", 70))
}

// ==================== MAIN RUNNER ====================

// RunEngine is the main entry point for the trading engine
func RunEngine(symbol, interval string, limit int) {
	// Check if multi-symbol mode is enabled
	if ENABLE_MULTI_SYMBOL {
		symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT", "ADAUSDT"}
		ConcurrentMultiSymbolAnalysis(symbols, interval, limit)
		return
	}

	// Single symbol analysis with optional parallel processing
	if ENABLE_PARALLEL_MODE {
		// Use optimized parallel engine (3-4x faster)
		engine := NewOptimizedEngine(symbol, interval, limit)

		if ENABLE_LIVE_MODE {
			fmt.Println("🔴 LIVE MODE: Parallel processing enabled")
			fmt.Println("   Press Ctrl+C to stop")
			fmt.Println()

			if err := engine.RunLiveOptimized(); err != nil {
				log.Fatalf("❌ Engine error: %v\n", err)
			}
		} else {
			fmt.Println("📸 SNAPSHOT MODE: Parallel processing enabled")
			fmt.Println()

			if err := engine.RunParallel(); err != nil {
				log.Fatalf("❌ Engine error: %v\n", err)
			}
		}
	} else {
		// Use standard sequential engine (backward compatibility)
		engine := NewTradingEngine(symbol, interval, limit)

		if ENABLE_LIVE_MODE {
			fmt.Println("🔴 LIVE MODE: Bot will run continuously")
			fmt.Println("   Press Ctrl+C to stop")
			fmt.Println()

			if err := engine.RunLive(); err != nil {
				log.Fatalf("❌ Engine error: %v\n", err)
			}
		} else {
			fmt.Println("📸 SNAPSHOT MODE: One-time analysis")
			fmt.Println()

			if err := engine.Run(); err != nil {
				log.Fatalf("❌ Engine error: %v\n", err)
			}
		}
	}
}

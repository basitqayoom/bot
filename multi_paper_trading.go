package main

import (
	"fmt"
	"sync"
	"time"
)

// ==================== MULTI-SYMBOL PAPER TRADING ====================

type MultiPaperTradingEngine struct {
	Symbols         []string
	Interval        string
	Limit           int
	StartingBalance float64
	CurrentBalance  float64
	Trades          []PaperTrade
	ActiveTrades    map[string]*PaperTrade // symbol -> trade
	TradeCounter    int
	WinCount        int
	LossCount       int
	TotalProfit     float64
	TotalLoss       float64
	mutex           sync.Mutex
	MaxPositions    int // Maximum simultaneous positions
	Logger          *TradeLogger
}

func NewMultiPaperTradingEngine(symbols []string, interval string, limit int, startingBalance float64, maxPositions int) *MultiPaperTradingEngine {
	if maxPositions == 0 {
		maxPositions = 5 // Default to 5 simultaneous positions
	}

	// Initialize multi-symbol trade logger
	logger, err := NewMultiTradeLogger()
	if err != nil {
		fmt.Printf("⚠️  Failed to create trade logger: %v\n", err)
		logger = nil
	}

	return &MultiPaperTradingEngine{
		Symbols:         symbols,
		Interval:        interval,
		Limit:           limit,
		StartingBalance: startingBalance,
		CurrentBalance:  startingBalance,
		Trades:          make([]PaperTrade, 0),
		ActiveTrades:    make(map[string]*PaperTrade),
		TradeCounter:    0,
		MaxPositions:    maxPositions,
		Logger:          logger,
	}
}

func (mp *MultiPaperTradingEngine) OpenTrade(symbol, side string, entryPrice, stopLoss, takeProfit, size float64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Check if already have a trade for this symbol
	if _, exists := mp.ActiveTrades[symbol]; exists {
		if VERBOSE_MODE {
			fmt.Printf("⚠️  Already have an open trade for %s\n", symbol)
		}
		return
	}

	// Check if we've reached max positions
	if len(mp.ActiveTrades) >= mp.MaxPositions {
		if VERBOSE_MODE {
			fmt.Printf("⚠️  Max positions reached (%d). Skipping %s\n", mp.MaxPositions, symbol)
		}
		return
	}

	mp.TradeCounter++
	trade := PaperTrade{
		ID:         mp.TradeCounter,
		Symbol:     symbol,
		Side:       side,
		EntryPrice: entryPrice,
		EntryTime:  time.Now(),
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
		Size:       size,
		Status:     "OPEN",
	}

	risk := 0.0
	if side == "SHORT" {
		risk = stopLoss - entryPrice
	} else {
		risk = entryPrice - stopLoss
	}

	reward := 0.0
	if side == "SHORT" {
		reward = entryPrice - takeProfit
	} else {
		reward = takeProfit - entryPrice
	}

	if risk > 0 {
		trade.RiskReward = reward / risk
	}

	mp.ActiveTrades[symbol] = &trade

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   📝 NEW POSITION OPENED               ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n🔔 Trade #%d: %s %s\n", trade.ID, side, symbol)
	fmt.Printf("💰 Entry:       $%.2f\n", entryPrice)
	fmt.Printf("🛑 Stop Loss:   $%.2f (%.2f%%)\n", stopLoss, (risk/entryPrice)*100)
	fmt.Printf("🎯 Take Profit: $%.2f (%.2f%%)\n", takeProfit, (reward/entryPrice)*100)
	fmt.Printf("📊 Size:        $%.2f\n", size)
	fmt.Printf("⚖️  Risk/Reward: %.2f:1\n", trade.RiskReward)
	fmt.Printf("📈 Active Positions: %d/%d\n", len(mp.ActiveTrades), mp.MaxPositions)
	fmt.Println("════════════════════════════════════════")
}

func (mp *MultiPaperTradingEngine) CheckAndClosePositions(currentPrices map[string]float64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	for symbol, trade := range mp.ActiveTrades {
		currentPrice, exists := currentPrices[symbol]
		if !exists {
			continue
		}

		shouldClose := false
		closeReason := ""

		if trade.Side == "SHORT" {
			if currentPrice >= trade.StopLoss {
				shouldClose = true
				closeReason = "STOP_LOSS"
			} else if currentPrice <= trade.TakeProfit {
				shouldClose = true
				closeReason = "TAKE_PROFIT"
			}
		} else if trade.Side == "LONG" {
			if currentPrice <= trade.StopLoss {
				shouldClose = true
				closeReason = "STOP_LOSS"
			} else if currentPrice >= trade.TakeProfit {
				shouldClose = true
				closeReason = "TAKE_PROFIT"
			}
		}

		if shouldClose {
			mp.closeTradeInternal(symbol, currentPrice, closeReason)
		}
	}
}

func (mp *MultiPaperTradingEngine) closeTradeInternal(symbol string, exitPrice float64, reason string) {
	trade, exists := mp.ActiveTrades[symbol]
	if !exists {
		return
	}

	trade.ExitPrice = exitPrice
	trade.ExitTime = time.Now()

	if trade.Side == "SHORT" {
		trade.ProfitLoss = (trade.EntryPrice - exitPrice) * (trade.Size / trade.EntryPrice)
	} else {
		trade.ProfitLoss = (exitPrice - trade.EntryPrice) * (trade.Size / trade.EntryPrice)
	}

	trade.ProfitLossPct = (trade.ProfitLoss / trade.Size) * 100

	if trade.ProfitLoss > 0 {
		mp.WinCount++
		mp.TotalProfit += trade.ProfitLoss
		if reason == "STOP_LOSS" {
			trade.Status = "CLOSED_SL_WIN"
		} else if reason == "TAKE_PROFIT" {
			trade.Status = "CLOSED_TP"
		} else {
			trade.Status = "CLOSED_WIN"
		}
	} else {
		mp.LossCount++
		mp.TotalLoss += trade.ProfitLoss
		if reason == "STOP_LOSS" {
			trade.Status = "CLOSED_SL"
		} else {
			trade.Status = "CLOSED_LOSS"
		}
	}

	mp.CurrentBalance += trade.ProfitLoss
	mp.Trades = append(mp.Trades, *trade)

	// Log trade to CSV
	if mp.Logger != nil {
		if err := mp.Logger.LogTrade(trade); err != nil {
			fmt.Printf("⚠️  Failed to log trade to CSV: %v\n", err)
		} else {
			fmt.Printf("💾 Trade logged to CSV\n")
		}
	}

	duration := trade.ExitTime.Sub(trade.EntryTime)

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   ❌ POSITION CLOSED                   ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n📝 Trade #%d: %s %s\n", trade.ID, trade.Side, symbol)
	fmt.Printf("📍 Entry:  $%.2f → Exit: $%.2f\n", trade.EntryPrice, exitPrice)
	fmt.Printf("📊 Reason: %s\n", reason)
	fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Second))

	if trade.ProfitLoss > 0 {
		fmt.Printf("💰 P/L: +$%.2f (+%.2f%%) ✅\n", trade.ProfitLoss, trade.ProfitLossPct)
	} else {
		fmt.Printf("💰 P/L: -$%.2f (%.2f%%) ❌\n", -trade.ProfitLoss, trade.ProfitLossPct)
	}

	totalPL := mp.CurrentBalance - mp.StartingBalance
	totalPLPct := (totalPL / mp.StartingBalance) * 100
	if totalPL > 0 {
		fmt.Printf("💵 Portfolio: $%.2f (+%.2f%%) ✅\n", mp.CurrentBalance, totalPLPct)
	} else {
		fmt.Printf("💵 Portfolio: $%.2f (%.2f%%) ❌\n", mp.CurrentBalance, totalPLPct)
	}
	fmt.Printf("📈 Active Positions: %d/%d\n", len(mp.ActiveTrades)-1, mp.MaxPositions)
	fmt.Println("════════════════════════════════════════")

	delete(mp.ActiveTrades, symbol)
}

// ShowUnrealizedPL displays unrealized P/L for all active positions
func (mp *MultiPaperTradingEngine) ShowUnrealizedPL(currentPrices map[string]float64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if len(mp.ActiveTrades) == 0 {
		return
	}

	fmt.Println("\n📊 UNREALIZED P/L (Active Positions):")
	fmt.Println("────────────────────────────────────────")

	totalUnrealized := 0.0

	for symbol, trade := range mp.ActiveTrades {
		currentPrice, exists := currentPrices[symbol]
		if !exists {
			continue
		}

		var unrealizedPL float64
		if trade.Side == "SHORT" {
			unrealizedPL = (trade.EntryPrice - currentPrice) * (trade.Size / trade.EntryPrice)
		} else {
			unrealizedPL = (currentPrice - trade.EntryPrice) * (trade.Size / trade.EntryPrice)
		}

		unrealizedPct := (unrealizedPL / trade.Size) * 100
		totalUnrealized += unrealizedPL

		status := "❌"
		if unrealizedPL > 0 {
			status = "✅"
			fmt.Printf("  %s %s: +$%.2f (+%.2f%%) %s\n",
				symbol, trade.Side, unrealizedPL, unrealizedPct, status)
		} else {
			fmt.Printf("  %s %s: -$%.2f (%.2f%%) %s\n",
				symbol, trade.Side, -unrealizedPL, unrealizedPct, status)
		}

		// Distance to SL/TP
		slDist := 0.0
		tpDist := 0.0
		if trade.Side == "SHORT" {
			slDist = ((trade.StopLoss - currentPrice) / currentPrice) * 100
			tpDist = ((currentPrice - trade.TakeProfit) / currentPrice) * 100
		} else {
			slDist = ((currentPrice - trade.StopLoss) / currentPrice) * 100
			tpDist = ((trade.TakeProfit - currentPrice) / currentPrice) * 100
		}
		fmt.Printf("    Price: $%.2f | SL: %.2f%% | TP: %.2f%%\n",
			currentPrice, slDist, tpDist)
	}

	fmt.Println("────────────────────────────────────────")
	if totalUnrealized > 0 {
		fmt.Printf("💰 Total Unrealized: +$%.2f ✅\n", totalUnrealized)
	} else {
		fmt.Printf("💰 Total Unrealized: -$%.2f ❌\n", -totalUnrealized)
	}

	potentialBalance := mp.CurrentBalance + totalUnrealized
	fmt.Printf("💵 Potential Balance: $%.2f\n", potentialBalance)
}

func (mp *MultiPaperTradingEngine) fetchPricesParallel(symbols []string) map[string]float64 {
	type priceResult struct {
		symbol string
		price  float64
		err    error
	}

	var wg sync.WaitGroup
	resultsChan := make(chan priceResult, len(symbols))
	semaphore := make(chan struct{}, NUM_WORKERS) // Limit concurrent API calls

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			engine := NewTradingEngine(sym, mp.Interval, mp.Limit)
			if err := engine.FetchData(); err == nil && len(engine.Candles) > 0 {
				resultsChan <- priceResult{
					symbol: sym,
					price:  engine.Candles[len(engine.Candles)-1].Close,
					err:    nil,
				}
			} else {
				resultsChan <- priceResult{
					symbol: sym,
					price:  0,
					err:    err,
				}
			}
		}(symbol)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	prices := make(map[string]float64)
	for result := range resultsChan {
		if result.err == nil && result.price > 0 {
			prices[result.symbol] = result.price
		}
	}

	return prices
}

func (mp *MultiPaperTradingEngine) PrintPortfolio() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	totalPL := mp.CurrentBalance - mp.StartingBalance
	totalPLPct := (totalPL / mp.StartingBalance) * 100

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║      PORTFOLIO SUMMARY                 ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n💰 Starting Balance: $%.2f\n", mp.StartingBalance)
	fmt.Printf("💰 Current Balance:  $%.2f", mp.CurrentBalance)

	if totalPL > 0 {
		fmt.Printf(" (+$%.2f, +%.2f%%) ✅\n", totalPL, totalPLPct)
	} else if totalPL < 0 {
		fmt.Printf(" (-$%.2f, %.2f%%) ❌\n", -totalPL, totalPLPct)
	} else {
		fmt.Printf("\n")
	}

	fmt.Printf("📊 Active Positions: %d/%d\n", len(mp.ActiveTrades), mp.MaxPositions)
	fmt.Printf("📈 Total Trades: %d (W: %d, L: %d)\n", len(mp.Trades), mp.WinCount, mp.LossCount)

	if len(mp.Trades) > 0 {
		winRate := (float64(mp.WinCount) / float64(len(mp.Trades))) * 100
		fmt.Printf("✅ Win Rate: %.1f%%\n", winRate)

		profitFactor := 0.0
		if mp.TotalLoss != 0 {
			profitFactor = mp.TotalProfit / -mp.TotalLoss
		}
		fmt.Printf("⚖️  Profit Factor: %.2f\n", profitFactor)
	}

	if len(mp.ActiveTrades) > 0 {
		fmt.Println("\n📋 Active Positions:")
		for symbol, trade := range mp.ActiveTrades {
			duration := time.Since(trade.EntryTime)
			fmt.Printf("  %s: %s @ $%.2f (%.0f ago)\n",
				symbol, trade.Side, trade.EntryPrice, duration.Minutes())
		}
	}

	fmt.Println("════════════════════════════════════════")
}

func (mp *MultiPaperTradingEngine) RunMultiPaperTrading() error {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   MULTI-SYMBOL PAPER TRADING v1.0      ║")
	fmt.Println("║   SIMULATED TRADING - NO REAL MONEY    ║")
	fmt.Println("╚════════════════════════════════════════╝")

	// Print full configuration at startup
	PrintMultiSymbolConfig(mp.Symbols, mp.Interval, mp.StartingBalance, mp.MaxPositions, "MULTI-SYMBOL PAPER TRADING")

	// Enable interactive mode with portfolio display on status
	StartInteractiveMode(func() {
		PrintMultiSymbolConfig(mp.Symbols, mp.Interval, mp.StartingBalance, mp.MaxPositions, "MULTI-SYMBOL PAPER TRADING")
	}, func() {
		// Show portfolio when 's' is pressed
		mp.PrintPortfolio()
	})

	fmt.Println()

	if ENABLE_LIVE_MODE {
		engine := NewTradingEngine(mp.Symbols[0], mp.Interval, mp.Limit)
		engine.printCandleSchedule()
	}

	lastCheckTime := time.Now().UTC()
	scanCount := 0

	for {
		if ENABLE_LIVE_MODE {
			if WAIT_FOR_CANDLE_CLOSE {
				engine := NewTradingEngine(mp.Symbols[0], mp.Interval, mp.Limit)
				engine.WaitForCandleClose()
			}

			engine := NewTradingEngine(mp.Symbols[0], mp.Interval, mp.Limit)
			if !engine.isCandleClosed(lastCheckTime) && WAIT_FOR_CANDLE_CLOSE {
				time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
				continue
			}
		}

		scanCount++
		lastCheckTime = time.Now().UTC()
		lastCheckIST := getIST()

		if ENABLE_LIVE_MODE {
			fmt.Println("\n" + "═══════════════════════════════════════════════════════════")
			fmt.Printf("🔔 CANDLE CLOSED - Multi-Symbol Scan #%d\n", scanCount)
			fmt.Printf("⏰ %s IST (%s UTC)\n",
				lastCheckIST.Format("2006-01-02 15:04:05"),
				lastCheckTime.Format("2006-01-02 15:04:05"))
			fmt.Println("═══════════════════════════════════════════════════════════")
		}

		// Analyze all symbols in parallel
		results := RunMultiSymbolAnalysis(mp.Symbols, mp.Interval, mp.Limit)

		// Collect current prices for position management IN PARALLEL
		currentPrices := mp.fetchPricesParallel(mp.Symbols)

		// Check and close positions that hit SL/TP
		mp.CheckAndClosePositions(currentPrices)

		// Look for new trade signals
		newSignals := 0
		for _, result := range results {
			if result.HasSignal && result.Error == nil {
				// Check if we already have a position for this symbol
				mp.mutex.Lock()
				_, hasPosition := mp.ActiveTrades[result.Symbol]
				canOpenMore := len(mp.ActiveTrades) < mp.MaxPositions
				mp.mutex.Unlock()

				if !hasPosition && canOpenMore {
					// Fetch detailed data for this symbol
					engine := NewOptimizedEngine(result.Symbol, mp.Interval, mp.Limit)
					if err := engine.FetchData(); err != nil {
						continue
					}

					engine.CalculateIndicators()
					engine.FindDivergences()
					engine.IdentifySupportResistance()

					if len(engine.Candles) == 0 || len(engine.RSI) == 0 {
						continue
					}

					currentPrice := engine.Candles[len(engine.Candles)-1].Close
					currentRSI := engine.RSI[len(engine.RSI)-1]

					// Find nearest resistance and support
					var nearestResistance *SRZone
					minDistanceUp := 1000000.0
					for i := range engine.SRZones {
						if engine.SRZones[i].Level > currentPrice {
							distance := engine.SRZones[i].Level - currentPrice
							if distance < minDistanceUp {
								minDistanceUp = distance
								nearestResistance = &engine.SRZones[i]
							}
						}
					}

					var nearestSupport *SRZone
					minDistanceDown := 1000000.0
					for i := range engine.SRZones {
						if engine.SRZones[i].Level < currentPrice {
							distance := currentPrice - engine.SRZones[i].Level
							if distance < minDistanceDown {
								minDistanceDown = distance
								nearestSupport = &engine.SRZones[i]
							}
						}
					}

					entry := currentPrice
					var stopLoss, takeProfit float64

					if nearestResistance != nil {
						stopLoss = nearestResistance.ZoneTop
					} else {
						stopLoss = currentPrice * (1 + STOP_LOSS_PERCENT/100)
					}

					if nearestSupport != nil {
						takeProfit = nearestSupport.ZoneBot
					} else {
						takeProfit = currentPrice * (1 - TAKE_PROFIT_PERCENT/100)
					}

					risk := stopLoss - entry
					reward := entry - takeProfit
					rr := reward / risk

					if rr >= RISK_REWARD_RATIO && currentRSI > 70 {
						riskAmount := mp.CurrentBalance * (MAX_RISK_PERCENT / 100)
						riskPercentPrice := (risk / entry) * 100
						positionSize := riskAmount / (riskPercentPrice / 100)

						fmt.Printf("\n🎯 SIGNAL: %s (RSI: %.2f, Div: %d, R/R: %.2f:1)\n",
							result.Symbol, currentRSI, result.Divergences, rr)

						mp.OpenTrade(result.Symbol, "SHORT", entry, stopLoss, takeProfit, positionSize)
						newSignals++
					}
				}
			}
		}

		if newSignals > 0 {
			fmt.Printf("\n✅ Opened %d new position(s)\n", newSignals)
		}

		// Show unrealized P/L for active positions
		if len(mp.ActiveTrades) > 0 {
			mp.ShowUnrealizedPL(currentPrices)
		}

		// Print portfolio summary
		mp.PrintPortfolio()

		if !ENABLE_LIVE_MODE {
			break
		}

		fmt.Printf("\n⏳ Next scan in %d seconds...\n", CHECK_INTERVAL)
		time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
	}

	// Close logger on exit
	if mp.Logger != nil {
		mp.Logger.Close()
		fmt.Printf("\n📁 All trades saved to: %s\n", mp.Logger.filename)
	}

	return nil
}

func RunMultiPaperTrading() {
	// This is called from main when --multi-paper flag is used
	// Implementation will be added to binance_fetcher.go
}

package main

import (
	"fmt"
	"sync"
	"time"

	"example.com/bot/internal/trademanager"
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
	TradeManager    *trademanager.Manager // 3-Tier trade management system
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

	// Initialize 3-Tier trade management system
	tmConfig := trademanager.DefaultConfig()
	tradeManager := trademanager.NewManager(tmConfig, VERBOSE_MODE)

	engine := &MultiPaperTradingEngine{
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
		TradeManager:    tradeManager,
	}

	// Setup trade manager callbacks
	tradeManager.SetCallbacks(
		engine.handlePartialExit,
		engine.handleStopUpdate,
		nil, // position close callback (optional)
	)

	// Display 3-Tier configuration
	if VERBOSE_MODE {
		fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
		fmt.Println("║        🛡️  3-TIER TRADE MANAGEMENT: ACTIVE 🛡️             ║")
		fmt.Println("╚═══════════════════════════════════════════════════════════╝")
		fmt.Printf("\n📊 Engine Configuration:\n")
		fmt.Printf("   Stop Loss:    %.1f%%\n", STOP_LOSS_PERCENT)
		fmt.Printf("   Take Profit:  %.1f%%\n", TAKE_PROFIT_PERCENT)
		fmt.Printf("   Timeframe:    %s\n", interval)

		fmt.Println("\n🎯 3-Tier Protection Layers:")
		fmt.Printf("   Tier 1: %.1f%% (Breakeven Lock)\n", tmConfig.Tier1BreakevenThreshold)
		fmt.Printf("   Tier 2: %.1f%% (Partial Exit %.0f%%)\n",
			tmConfig.Tier2PartialExitThreshold, tmConfig.Tier2PartialExitPercent)
		fmt.Printf("   Tier 3: %ds (Trailing Stop - Locks %.0f%% of max profit)\n",
			tmConfig.Tier3TimeThreshold, tmConfig.Tier3ProfitLockPercent)

		fmt.Println("\n💡 Expected Impact:")
		fmt.Println("   • Reduced give-back: ~67%")
		fmt.Println("   • Protected breakeven after +0.3%")
		fmt.Println("   • Profit secured before TP hit")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()
	} else {
		fmt.Println("\n✅ 3-Tier Trade Management: ACTIVE")
		fmt.Printf("   Engine: %.1f%% SL / %.1f%% TP | %s\n", STOP_LOSS_PERCENT, TAKE_PROFIT_PERCENT, interval)
		fmt.Printf("   Tiers: %.1f%% BE | %.1f%% Partial | %ds Trailing\n",
			tmConfig.Tier1BreakevenThreshold, tmConfig.Tier2PartialExitThreshold, tmConfig.Tier3TimeThreshold)
	}

	return engine
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
		ID:           mp.TradeCounter,
		Symbol:       symbol,
		Interval:     mp.Interval,
		Side:         side,
		EntryPrice:   entryPrice,
		EntryTime:    time.Now(),
		StopLoss:     stopLoss,
		TakeProfit:   takeProfit,
		Size:         size,
		Status:       "OPEN",
		HighestPrice: entryPrice, // Initialize to entry price
		LowestPrice:  entryPrice, // Initialize to entry price
		MaxProfit:    0,
		MaxProfitPct: 0,
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

	if VERBOSE_MODE {
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║   📝 NEW POSITION OPENED               ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("\n🔔 Trade #%d: %s %s\n", trade.ID, side, symbol)
		fmt.Printf("💰 Entry:       $%.2f\n", entryPrice)
		fmt.Printf("🛑 Stop Loss:   $%.2f (%.2f%%)\n", stopLoss, (risk/entryPrice)*100)
		fmt.Printf("🎯 Take Profit: $%.2f (%.2f%%)\n", takeProfit, (reward/entryPrice)*100)
		fmt.Printf("📊 Size:        $%.2f\n", size)
		fmt.Printf("⚖️  Risk/Reward: %.2f:1\n", trade.RiskReward)
	} else {
		fmt.Printf("\n🎯 [%s] %s OPENED @ $%.2f | SL: $%.2f | TP: $%.2f\n",
			symbol, side, entryPrice, stopLoss, takeProfit)
	}
	fmt.Printf("📈 Active Positions: %d/%d\n", len(mp.ActiveTrades), mp.MaxPositions)
	fmt.Println("════════════════════════════════════════")

	// Add position to trade manager for 3-Tier protection (ADAPTIVE MODE)
	if mp.TradeManager != nil && mp.TradeManager.IsEnabled() {
		// Use adaptive config that adjusts thresholds based on actual SL/TP distances
		mp.TradeManager.AddPositionWithAdaptiveConfig(
			trade.ID,
			symbol,
			side,
			entryPrice,
			stopLoss,
			takeProfit,
			size,
		)
	}
}

func (mp *MultiPaperTradingEngine) CheckAndClosePositions(currentPrices map[string]float64) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	for symbol, trade := range mp.ActiveTrades {
		currentPrice, exists := currentPrices[symbol]
		if !exists {
			continue
		}

		// Update trade manager with current price (evaluates 3-Tier rules)
		if mp.TradeManager != nil && mp.TradeManager.IsEnabled() {
			if err := mp.TradeManager.UpdatePrice(symbol, currentPrice); err != nil {
				if VERBOSE_MODE {
					fmt.Printf("⚠️  Trade manager error for %s: %v\n", symbol, err)
				}
			}
		}

		// Track highest and lowest prices
		if currentPrice > trade.HighestPrice {
			trade.HighestPrice = currentPrice
		}
		if currentPrice < trade.LowestPrice {
			trade.LowestPrice = currentPrice
		}

		// Calculate current profit/loss
		var currentProfit float64
		if trade.Side == "SHORT" {
			currentProfit = (trade.EntryPrice - currentPrice) * (trade.Size / trade.EntryPrice)
		} else {
			currentProfit = (currentPrice - trade.EntryPrice) * (trade.Size / trade.EntryPrice)
		}

		currentProfitPct := (currentProfit / trade.Size) * 100

		// Track maximum profit
		if currentProfit > trade.MaxProfit {
			trade.MaxProfit = currentProfit
			trade.MaxProfitPct = currentProfitPct
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

	// Calculate "give back" (difference between max profit and actual profit)
	giveBack := trade.MaxProfit - trade.ProfitLoss
	giveBackPct := trade.MaxProfitPct - trade.ProfitLossPct

	if VERBOSE_MODE {
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║   ❌ POSITION CLOSED                   ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("\n📝 Trade #%d: %s %s\n", trade.ID, trade.Side, symbol)
		fmt.Printf("📍 Entry:  $%.2f → Exit: $%.2f\n", trade.EntryPrice, exitPrice)
		fmt.Printf("📊 Reason: %s\n", reason)
		fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Second))

		// Show price extremes
		fmt.Printf("\n📈 Highest Price: $%.2f\n", trade.HighestPrice)
		fmt.Printf("📉 Lowest Price:  $%.2f\n", trade.LowestPrice)

		// Show final P/L
		if trade.ProfitLoss > 0 {
			fmt.Printf("\n💰 Final P/L: +$%.2f (+%.2f%%) ✅\n", trade.ProfitLoss, trade.ProfitLossPct)
		} else {
			fmt.Printf("\n💰 Final P/L: -$%.2f (%.2f%%) ❌\n", -trade.ProfitLoss, trade.ProfitLossPct)
		}

		// Show maximum profit reached and "give back"
		if trade.MaxProfit > 0 {
			fmt.Printf("🎯 Max Profit: +$%.2f (+%.2f%%)\n", trade.MaxProfit, trade.MaxProfitPct)
			if giveBack > 0 {
				fmt.Printf("⚠️  Give Back:  -$%.2f (-%.2f%%) 📉\n", giveBack, giveBackPct)
			}
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
	} else {
		// Quiet mode: concise output
		if trade.ProfitLoss > 0 {
			fmt.Printf("\n✅ [%s] %s CLOSED @ $%.2f | %s | P/L: +$%.2f (+%.2f%%)\n",
				symbol, trade.Side, exitPrice, reason, trade.ProfitLoss, trade.ProfitLossPct)
		} else {
			fmt.Printf("\n❌ [%s] %s CLOSED @ $%.2f | %s | P/L: -$%.2f (%.2f%%)\n",
				symbol, trade.Side, exitPrice, reason, -trade.ProfitLoss, trade.ProfitLossPct)
		}

		// Show give back in quiet mode if significant
		if giveBack > 0 && giveBack > 1.0 {
			fmt.Printf("   ⚠️  Max Profit: +$%.2f | Give Back: -$%.2f\n", trade.MaxProfit, giveBack)
		}
	}

	delete(mp.ActiveTrades, symbol)

	// Remove from trade manager
	if mp.TradeManager != nil {
		mp.TradeManager.RemovePosition(symbol)
	}
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
						if VERBOSE_MODE {
							fmt.Printf("   🎯 [%s] Using resistance zone SL: $%.4f (zone: $%.4f-$%.4f)\n",
								result.Symbol, stopLoss, nearestResistance.ZoneBot, nearestResistance.ZoneTop)
						}
					} else {
						stopLoss = currentPrice * (1 + STOP_LOSS_PERCENT/100)
						if VERBOSE_MODE {
							fmt.Printf("   🎯 [%s] No resistance zone, using fixed SL: $%.4f (+%.2f%%)\n",
								result.Symbol, stopLoss, STOP_LOSS_PERCENT)
						}
					}

					if nearestSupport != nil {
						takeProfit = nearestSupport.ZoneBot
						if VERBOSE_MODE {
							fmt.Printf("   🎯 [%s] Using support zone TP: $%.4f (zone: $%.4f-$%.4f)\n",
								result.Symbol, takeProfit, nearestSupport.ZoneBot, nearestSupport.ZoneTop)
						}
					} else {
						takeProfit = currentPrice * (1 - TAKE_PROFIT_PERCENT/100)
						if VERBOSE_MODE {
							fmt.Printf("   🎯 [%s] No support zone, using fixed TP: $%.4f (-%.2f%%)\n",
								result.Symbol, takeProfit, TAKE_PROFIT_PERCENT)
						}
					}

					// CRITICAL FIX: Ensure SL is always ABOVE entry for SHORT
					if stopLoss <= entry {
						stopLoss = entry * (1 + STOP_LOSS_PERCENT/100)
						if VERBOSE_MODE {
							fmt.Printf("   ⚠️  [%s] WARNING: SL was at/below entry! Adjusted to $%.4f (+%.2f%%)\n",
								result.Symbol, stopLoss, STOP_LOSS_PERCENT)
						}
					}

					// CRITICAL FIX: Ensure TP is always BELOW entry for SHORT
					if takeProfit >= entry {
						takeProfit = entry * (1 - TAKE_PROFIT_PERCENT/100)
						if VERBOSE_MODE {
							fmt.Printf("   ⚠️  [%s] WARNING: TP was at/above entry! Adjusted to $%.4f (-%.2f%%)\n",
								result.Symbol, takeProfit, TAKE_PROFIT_PERCENT)
						}
					}

					risk := stopLoss - entry
					reward := entry - takeProfit
					rr := reward / risk

					if rr >= RISK_REWARD_RATIO && currentRSI > 70 {
						// ✅ FIXED: Use simple fixed allocation (realistic for 1x leverage)
						// Each trade gets equal share of initial balance
						positionSize := mp.StartingBalance / float64(mp.MaxPositions)

						// Optional: Log if risk-based sizing would have been larger (for analysis)
						riskAmount := mp.CurrentBalance * (MAX_RISK_PERCENT / 100)
						riskPercentPrice := (risk / entry) * 100
						riskBasedSize := riskAmount / (riskPercentPrice / 100)

						if riskBasedSize > positionSize && VERBOSE_MODE {
							fmt.Printf("   ⚠️  [%s] Risk-based size $%.0f capped to $%.0f (1x leverage)\n",
								result.Symbol, riskBasedSize, positionSize)
						}

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

// ==================== 3-TIER TRADE MANAGER CALLBACKS ====================

// handlePartialExit is called by the trade manager when Tier 2 triggers
func (mp *MultiPaperTradingEngine) handlePartialExit(symbol string, exitPercent, currentPrice float64) (float64, error) {
	trade, exists := mp.ActiveTrades[symbol]
	if !exists {
		return 0, fmt.Errorf("no active trade for %s", symbol)
	}

	// Calculate exit size
	exitSize := trade.Size * (exitPercent / 100.0)

	// Calculate profit from this partial exit
	var exitProfit float64
	if trade.Side == "SHORT" {
		exitProfit = (trade.EntryPrice - currentPrice) * (exitSize / trade.EntryPrice)
	} else { // LONG
		exitProfit = (currentPrice - trade.EntryPrice) * (exitSize / trade.EntryPrice)
	}

	// Update position size
	trade.Size -= exitSize
	mp.CurrentBalance += exitProfit

	if VERBOSE_MODE {
		fmt.Printf("💰 Partial Exit: %.0f%% of %s @ $%.4f | Profit: $%.4f | Remaining: $%.2f\n",
			exitPercent, symbol, currentPrice, exitProfit, trade.Size)
	}

	return exitProfit, nil
}

// handleStopUpdate is called by the trade manager when stops need to be moved
func (mp *MultiPaperTradingEngine) handleStopUpdate(symbol string, newStopLoss float64) error {
	trade, exists := mp.ActiveTrades[symbol]
	if !exists {
		return fmt.Errorf("no active trade for %s", symbol)
	}

	trade.StopLoss = newStopLoss
	return nil
}

func RunMultiPaperTrading() {
	// This is called from main when --multi-paper flag is used
	// Implementation will be added to binance_fetcher.go
}

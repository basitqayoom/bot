package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

type PaperTrade struct {
	ID            int
	Symbol        string
	Interval      string
	Side          string
	EntryPrice    float64
	EntryTime     time.Time
	StopLoss      float64
	TakeProfit    float64
	Size          float64
	Status        string
	ExitPrice     float64
	ExitTime      time.Time
	ProfitLoss    float64
	ProfitLossPct float64
	RiskReward    float64

	// Track price extremes during trade
	HighestPrice float64 // Highest price reached during trade
	LowestPrice  float64 // Lowest price reached during trade

	// Track maximum profit
	MaxProfit    float64 // Maximum profit in dollars
	MaxProfitPct float64 // Maximum profit percentage
}

type PaperTradingEngine struct {
	*TradingEngine
	StartingBalance float64
	CurrentBalance  float64
	Trades          []PaperTrade
	ActiveTrade     *PaperTrade
	TradeCounter    int
	WinCount        int
	LossCount       int
	TotalProfit     float64
	TotalLoss       float64
	Logger          *TradeLogger
}

func NewPaperTradingEngine(symbol, interval string, limit int, startingBalance float64) *PaperTradingEngine {
	// Initialize trade logger
	logger, err := NewTradeLogger(symbol)
	if err != nil {
		fmt.Printf("⚠️  Failed to create trade logger: %v\n", err)
		logger = nil
	}

	return &PaperTradingEngine{
		TradingEngine:   NewTradingEngine(symbol, interval, limit),
		StartingBalance: startingBalance,
		CurrentBalance:  startingBalance,
		Trades:          make([]PaperTrade, 0),
		TradeCounter:    0,
		Logger:          logger,
	}
}

func (p *PaperTradingEngine) OpenTrade(side string, entryPrice, stopLoss, takeProfit, size float64) {
	if p.ActiveTrade != nil {
		fmt.Println("⚠️  Already have an open trade. Close it first.")
		return
	}

	p.TradeCounter++
	trade := PaperTrade{
		ID:           p.TradeCounter,
		Symbol:       p.Symbol,
		Interval:     p.Interval,
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

	p.ActiveTrade = &trade

	if VERBOSE_MODE {
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║      PAPER TRADE OPENED                ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("\n📝 Trade #%d: %s %s\n", trade.ID, side, p.Symbol)
		fmt.Printf("💰 Entry:       $%.2f\n", entryPrice)
		fmt.Printf("🛑 Stop Loss:   $%.2f (%.2f%%)\n", stopLoss, (risk/entryPrice)*100)
		fmt.Printf("🎯 Take Profit: $%.2f (%.2f%%)\n", takeProfit, (reward/entryPrice)*100)
		fmt.Printf("📊 Size:        $%.2f\n", size)
		fmt.Printf("⚖️  Risk/Reward: %.2f:1\n", trade.RiskReward)
		fmt.Printf("⏰ Time:        %s\n", trade.EntryTime.Format("2006-01-02 15:04:05"))
		fmt.Println("════════════════════════════════════════")
	} else {
		fmt.Printf("\n🎯 [%s] %s OPENED @ $%.2f | SL: $%.2f | TP: $%.2f | Size: $%.2f\n",
			p.Symbol, side, entryPrice, stopLoss, takeProfit, size)
	}
}

func (p *PaperTradingEngine) CheckAndClosePosition(currentPrice float64) {
	if p.ActiveTrade == nil {
		return
	}

	trade := p.ActiveTrade

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
		p.CloseTrade(currentPrice, closeReason)
	}
}

func (p *PaperTradingEngine) CloseTrade(exitPrice float64, reason string) {
	if p.ActiveTrade == nil {
		return
	}

	trade := p.ActiveTrade
	trade.ExitPrice = exitPrice
	trade.ExitTime = time.Now()

	if trade.Side == "SHORT" {
		trade.ProfitLoss = (trade.EntryPrice - exitPrice) * (trade.Size / trade.EntryPrice)
	} else {
		trade.ProfitLoss = (exitPrice - trade.EntryPrice) * (trade.Size / trade.EntryPrice)
	}

	trade.ProfitLossPct = (trade.ProfitLoss / trade.Size) * 100

	if trade.ProfitLoss > 0 {
		p.WinCount++
		p.TotalProfit += trade.ProfitLoss
		if reason == "STOP_LOSS" {
			trade.Status = "CLOSED_SL_WIN"
		} else if reason == "TAKE_PROFIT" {
			trade.Status = "CLOSED_TP"
		} else {
			trade.Status = "CLOSED_WIN"
		}
	} else {
		p.LossCount++
		p.TotalLoss += trade.ProfitLoss
		if reason == "STOP_LOSS" {
			trade.Status = "CLOSED_SL"
		} else {
			trade.Status = "CLOSED_LOSS"
		}
	}

	p.CurrentBalance += trade.ProfitLoss
	p.Trades = append(p.Trades, *trade)

	// Log trade to CSV
	if p.Logger != nil {
		if err := p.Logger.LogTrade(trade); err != nil && VERBOSE_MODE {
			fmt.Printf("⚠️  Failed to log trade to CSV: %v\n", err)
		} else if VERBOSE_MODE {
			fmt.Printf("💾 Trade logged to CSV: %s\n", p.Logger.filename)
		}
	}

	duration := trade.ExitTime.Sub(trade.EntryTime)

	// Calculate "give back" (difference between max profit and actual profit)
	giveBack := trade.MaxProfit - trade.ProfitLoss
	giveBackPct := trade.MaxProfitPct - trade.ProfitLossPct

	if VERBOSE_MODE {
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║      PAPER TRADE CLOSED                ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("\n📝 Trade #%d: %s %s\n", trade.ID, trade.Side, p.Symbol)
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
	} else {
		// Quiet mode: concise output
		if trade.ProfitLoss > 0 {
			fmt.Printf("\n✅ [%s] %s CLOSED @ $%.2f | %s | P/L: +$%.2f (+%.2f%%)\n",
				p.Symbol, trade.Side, exitPrice, reason, trade.ProfitLoss, trade.ProfitLossPct)
		} else {
			fmt.Printf("\n❌ [%s] %s CLOSED @ $%.2f | %s | P/L: -$%.2f (%.2f%%)\n",
				p.Symbol, trade.Side, exitPrice, reason, -trade.ProfitLoss, trade.ProfitLossPct)
		}

		// Show give back in quiet mode if significant
		if giveBack > 0 && giveBack > 1.0 {
			fmt.Printf("   ⚠️  Max Profit: +$%.2f | Give Back: -$%.2f\n", trade.MaxProfit, giveBack)
		}
	}

	fmt.Printf("💵 Balance: $%.2f → $%.2f\n", p.StartingBalance, p.CurrentBalance)
	fmt.Println("════════════════════════════════════════")

	p.ActiveTrade = nil
}

func (p *PaperTradingEngine) PrintStats() {
	totalTrades := len(p.Trades)
	totalPL := p.CurrentBalance - p.StartingBalance
	totalPLPct := (totalPL / p.StartingBalance) * 100

	// Always show current portfolio balance
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║      PORTFOLIO SUMMARY                 ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n💰 Starting Balance: $%.2f\n", p.StartingBalance)
	fmt.Printf("💰 Current Balance:  $%.2f", p.CurrentBalance)

	if totalPL > 0 {
		fmt.Printf(" (+$%.2f, +%.2f%%) ✅\n", totalPL, totalPLPct)
	} else if totalPL < 0 {
		fmt.Printf(" (-$%.2f, %.2f%%) ❌\n", -totalPL, totalPLPct)
	} else {
		fmt.Printf("\n")
	}

	if totalTrades == 0 {
		fmt.Println("\n📊 No trades executed yet")
		fmt.Println("════════════════════════════════════════")
		return
	}

	winRate := (float64(p.WinCount) / float64(totalTrades)) * 100

	avgWin := 0.0
	if p.WinCount > 0 {
		avgWin = p.TotalProfit / float64(p.WinCount)
	}

	avgLoss := 0.0
	if p.LossCount > 0 {
		avgLoss = p.TotalLoss / float64(p.LossCount)
	}

	profitFactor := 0.0
	if p.TotalLoss != 0 {
		profitFactor = p.TotalProfit / -p.TotalLoss
	}

	// Trading statistics
	fmt.Printf("\n📊 Total Trades: %d\n", totalTrades)
	fmt.Printf("✅ Wins: %d (%.1f%%)\n", p.WinCount, winRate)
	fmt.Printf("❌ Losses: %d (%.1f%%)\n", p.LossCount, 100-winRate)

	fmt.Printf("\n📈 Average Win:  +$%.2f\n", avgWin)
	fmt.Printf("📉 Average Loss: -$%.2f\n", -avgLoss)
	fmt.Printf("⚖️  Profit Factor: %.2f\n", profitFactor)

	if len(p.Trades) > 0 {
		fmt.Println("\n📋 Recent Trades:")
		start := len(p.Trades) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(p.Trades); i++ {
			t := p.Trades[i]
			status := "❌"
			if t.ProfitLoss > 0 {
				status = "✅"
			}
			fmt.Printf("  #%d: %s $%.2f → $%.2f | P/L: $%.2f (%.2f%%) %s\n",
				t.ID, t.Side, t.EntryPrice, t.ExitPrice, t.ProfitLoss, t.ProfitLossPct, status)
		}
	}

	fmt.Println("════════════════════════════════════════")
}

func (p *PaperTradingEngine) RunPaperTrading() error {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║   PAPER TRADING BOT v1.0               ║")
	fmt.Println("║   SIMULATED TRADING - NO REAL MONEY    ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n🚀 Symbol: %s | Interval: %s\n", p.Symbol, p.Interval)
	fmt.Printf("💰 Starting Balance: $%.2f\n", p.StartingBalance)
	fmt.Printf("📊 Risk per trade: %.1f%%\n", MAX_RISK_PERCENT)
	fmt.Println()

	if ENABLE_LIVE_MODE {
		p.printCandleSchedule()
	}

	lastCheckTime := time.Now().UTC()
	analysisCount := 0

	for {
		if ENABLE_LIVE_MODE {
			if WAIT_FOR_CANDLE_CLOSE {
				// WaitForCandleClose() already waits internally with the countdown
				// It returns when the candle has closed
				p.WaitForCandleClose()
			}

			if !p.isCandleClosed(lastCheckTime) && WAIT_FOR_CANDLE_CLOSE {
				time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
				continue
			}
		}

		analysisCount++
		lastCheckTime = time.Now().UTC()
		lastCheckIST := getIST()

		if ENABLE_LIVE_MODE {
			fmt.Println("\n" + "═══════════════════════════════════════════════════════════")
			fmt.Printf("🔔 CANDLE CLOSED - Running Analysis #%d\n", analysisCount)
			fmt.Printf("⏰ %s IST (%s UTC)\n",
				lastCheckIST.Format("2006-01-02 15:04:05"),
				lastCheckTime.Format("2006-01-02 15:04:05"))
			fmt.Println("═══════════════════════════════════════════════════════════")
		}

		if err := p.FetchData(); err != nil {
			fmt.Printf("⚠️  Fetch error: %v\n", err)
			if !ENABLE_LIVE_MODE {
				return err
			}
			time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
			continue
		}

		p.CalculateIndicators()
		p.FindDivergences()
		p.IdentifySupportResistance()

		currentPrice := p.Candles[len(p.Candles)-1].Close
		currentRSI := p.RSI[len(p.RSI)-1]

		// Show current portfolio status
		totalPL := p.CurrentBalance - p.StartingBalance
		totalPLPct := (totalPL / p.StartingBalance) * 100
		fmt.Println("\n┌────────────────────────────────────────┐")
		if totalPL > 0 {
			fmt.Printf("│ 💼 PORTFOLIO: $%.2f (+%.2f%%) ✅    │\n", p.CurrentBalance, totalPLPct)
		} else if totalPL < 0 {
			fmt.Printf("│ 💼 PORTFOLIO: $%.2f (%.2f%%) ❌    │\n", p.CurrentBalance, totalPLPct)
		} else {
			fmt.Printf("│ 💼 PORTFOLIO: $%.2f                    │\n", p.CurrentBalance)
		}
		fmt.Println("└────────────────────────────────────────┘")

		if p.ActiveTrade != nil {
			p.CheckAndClosePosition(currentPrice)
		}

		if p.ActiveTrade == nil {
			recentDivergences := 0
			for _, div := range p.Divergences {
				divTime, _ := time.Parse("2006-01-02 15:04", div.EndTime)
				hoursSince := time.Since(divTime).Hours()
				if hoursSince < 72 {
					recentDivergences++
				}
			}

			if recentDivergences >= MIN_DIVERGENCES_FOR_SIGNAL && currentRSI > 70 {
				var nearestResistance *SRZone
				minDistanceUp := 1000000.0
				for i := range p.SRZones {
					if p.SRZones[i].Level > currentPrice {
						distance := p.SRZones[i].Level - currentPrice
						if distance < minDistanceUp {
							minDistanceUp = distance
							nearestResistance = &p.SRZones[i]
						}
					}
				}

				var nearestSupport *SRZone
				minDistanceDown := 1000000.0
				for i := range p.SRZones {
					if p.SRZones[i].Level < currentPrice {
						distance := currentPrice - p.SRZones[i].Level
						if distance < minDistanceDown {
							minDistanceDown = distance
							nearestSupport = &p.SRZones[i]
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

				if rr >= RISK_REWARD_RATIO {
					// ✅ FIXED: Use full balance for single symbol trading
					// (In single symbol mode, we only trade one pair at a time)
					positionSize := p.StartingBalance

					// Optional: Log risk-based calculation for comparison
					riskAmount := p.CurrentBalance * (MAX_RISK_PERCENT / 100)
					riskPercentPrice := (risk / entry) * 100
					riskBasedSize := riskAmount / (riskPercentPrice / 100)

					if riskBasedSize > positionSize && VERBOSE_MODE {
						fmt.Printf("   ⚠️  Risk-based size $%.0f capped to $%.0f (1x leverage)\n",
							riskBasedSize, positionSize)
					}

					fmt.Println("\n🎯 BEARISH SIGNAL DETECTED!")
					fmt.Printf("📊 RSI: %.2f (Overbought)\n", currentRSI)
					fmt.Printf("📈 Divergences: %d\n", recentDivergences)
					fmt.Printf("⚖️  R/R Ratio: %.2f:1 ✅\n", rr)

					p.OpenTrade("SHORT", entry, stopLoss, takeProfit, positionSize)
				} else {
					fmt.Println("\n⚠️  Signal detected but R/R ratio too low")
					fmt.Printf("   R/R: %.2f:1 (min: %.1f:1)\n", rr, RISK_REWARD_RATIO)
				}
			}
		}

		if p.ActiveTrade != nil {
			fmt.Println("\n📊 OPEN POSITION STATUS:")
			fmt.Printf("   Current Price: $%.2f\n", currentPrice)
			fmt.Printf("   Entry Price:   $%.2f\n", p.ActiveTrade.EntryPrice)

			unrealizedPL := (p.ActiveTrade.EntryPrice - currentPrice) * (p.ActiveTrade.Size / p.ActiveTrade.EntryPrice)
			unrealizedPct := (unrealizedPL / p.ActiveTrade.Size) * 100

			if unrealizedPL > 0 {
				fmt.Printf("   Unrealized P/L: +$%.2f (+%.2f%%) ✅\n", unrealizedPL, unrealizedPct)
			} else {
				fmt.Printf("   Unrealized P/L: -$%.2f (%.2f%%) ❌\n", -unrealizedPL, unrealizedPct)
			}

			slDistance := ((p.ActiveTrade.StopLoss - currentPrice) / currentPrice) * 100
			tpDistance := ((currentPrice - p.ActiveTrade.TakeProfit) / currentPrice) * 100
			fmt.Printf("   Distance to SL: %.2f%%\n", slDistance)
			fmt.Printf("   Distance to TP: %.2f%%\n", tpDistance)
		}

		p.PrintStats()

		if !ENABLE_LIVE_MODE {
			break
		}

		fmt.Printf("\n⏳ Next check in %d seconds...\n", CHECK_INTERVAL)
		time.Sleep(time.Duration(CHECK_INTERVAL) * time.Second)
	}

	return nil
}

func RunPaperTrading() {
	symbol := flag.String("symbol", DEFAULT_SYMBOL, "Trading symbol (e.g., BTCUSDT)")
	interval := flag.String("interval", DEFAULT_INTERVAL, "Candle interval (1m, 5m, 15m, 1h, 4h, 1d)")
	balance := flag.Float64("balance", 10000.0, "Starting balance in USD")
	flag.Parse()

	engine := NewPaperTradingEngine(*symbol, *interval, DEFAULT_LIMIT, *balance)

	// Print configuration at startup
	PrintBotConfig(*symbol, *interval, *balance, "PAPER TRADING")

	// Enable interactive commands with portfolio display on status
	StartInteractiveMode(func() {
		PrintBotConfig(*symbol, *interval, *balance, "PAPER TRADING")
	}, func() {
		// Show portfolio when 's' is pressed
		engine.PrintStats()
	})

	// Ensure logger is closed on exit
	defer func() {
		if engine.Logger != nil {
			engine.Logger.Close()
			fmt.Printf("\n📁 All trades saved to: %s\n", engine.Logger.filename)
		}
	}()

	if err := engine.RunPaperTrading(); err != nil {
		log.Fatalf("❌ Paper trading error: %v\n", err)
	}
}

# Quick Integration Guide: 3-Tier System

## 5-Minute Integration

### Step 1: Add TradeManager field to your engine

**File:** `multi_paper_trading.go`

```go
type MultiPaperTradingEngine struct {
    // ... existing fields ...
    TradeManager *trademanager.Manager  // ADD THIS LINE
}
```

### Step 2: Initialize in constructor

**File:** `multi_paper_trading.go` in `NewMultiPaperTradingEngine()`

```go
import "bot/internal/trademanager"  // Add import

func NewMultiPaperTradingEngine(...) *MultiPaperTradingEngine {
    // ... existing code ...
    
    engine := &MultiPaperTradingEngine{
        // ... existing fields ...
    }
    
    // Initialize trade manager
    tmConfig := trademanager.DefaultConfig()
    engine.TradeManager = trademanager.NewManager(tmConfig, VERBOSE_MODE)
    
    // Setup callbacks
    engine.TradeManager.SetCallbacks(
        engine.handlePartialExit,
        engine.handleStopUpdate,
        nil, // position close callback (optional)
    )
    
    return engine
}
```

### Step 3: Add callback methods

**File:** `multi_paper_trading.go` (add these new methods)

```go
func (mp *MultiPaperTradingEngine) handlePartialExit(symbol string, exitPercent, currentPrice float64) (float64, error) {
    trade, exists := mp.ActiveTrades[symbol]
    if !exists {
        return 0, fmt.Errorf("no active trade for %s", symbol)
    }
    
    exitSize := trade.Size * (exitPercent / 100.0)
    
    var exitProfit float64
    if trade.Side == "SHORT" {
        exitProfit = (trade.EntryPrice - currentPrice) * (exitSize / trade.EntryPrice)
    } else {
        exitProfit = (currentPrice - trade.EntryPrice) * (exitSize / trade.EntryPrice)
    }
    
    trade.Size -= exitSize
    mp.CurrentBalance += exitProfit
    
    fmt.Printf("💰 Partial Exit: %.0f%% of %s @ $%.4f | Profit: $%.4f\n", 
        exitPercent, symbol, currentPrice, exitProfit)
    
    return exitProfit, nil
}

func (mp *MultiPaperTradingEngine) handleStopUpdate(symbol string, newStopLoss float64) error {
    trade, exists := mp.ActiveTrades[symbol]
    if !exists {
        return fmt.Errorf("no active trade for %s", symbol)
    }
    
    trade.StopLoss = newStopLoss
    return nil
}
```

### Step 4: Hook into OpenTrade

**File:** `multi_paper_trading.go` in `OpenTrade()`

Add this at the END of the OpenTrade function:

```go
func (mp *MultiPaperTradingEngine) OpenTrade(...) {
    // ... all existing code stays the same ...
    
    // ADD THIS AT THE END:
    if mp.TradeManager != nil && mp.TradeManager.IsEnabled() {
        mp.TradeManager.AddPosition(
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
```

### Step 5: Hook into CheckAndClosePositions

**File:** `multi_paper_trading.go` in `CheckAndClosePositions()`

Add this BEFORE the shouldClose logic:

```go
func (mp *MultiPaperTradingEngine) CheckAndClosePositions(currentPrices map[string]float64) {
    mp.mutex.Lock()
    defer mp.mutex.Unlock()

    for symbol, trade := range mp.ActiveTrades {
        currentPrice, exists := currentPrices[symbol]
        if !exists {
            continue
        }

        // Track highest and lowest prices
        // ... existing tracking code ...

        // ADD THIS: Update trade manager
        if mp.TradeManager != nil && mp.TradeManager.IsEnabled() {
            if err := mp.TradeManager.UpdatePrice(symbol, currentPrice); err != nil {
                fmt.Printf("⚠️  Trade manager error for %s: %v\n", symbol, err)
            }
        }

        // ... rest of existing logic (shouldClose, etc.) ...
    }
}
```

### Step 6: Hook into closeTradeInternal

**File:** `multi_paper_trading.go` in `closeTradeInternal()`

Add this at the END of the function:

```go
func (mp *MultiPaperTradingEngine) closeTradeInternal(symbol string, exitPrice float64, reason string) {
    // ... all existing code stays the same ...
    
    // ADD THIS AT THE END:
    if mp.TradeManager != nil {
        mp.TradeManager.RemovePosition(symbol)
    }
}
```

## That's It!

The 3-Tier system is now integrated and will automatically:
- ✅ Move stops to breakeven at +0.5%
- ✅ Take 50% profit at +1.5%
- ✅ Lock in 60% of max profit after 5 minutes

## Testing

Run your existing bot:
```bash
go run *.go
```

Watch for new messages:
- `✅ Trade Manager: Added position ...`
- `🔒 Tier 1: Breakeven Lock at +X.X%`
- `💰 Tier 2: Partial Exit ...`
- `⏰ Tier 3: Time Lock ...`

## Enable/Disable

To disable temporarily (for A/B testing):
```go
mp.TradeManager.Disable()
```

To re-enable:
```go
mp.TradeManager.Enable()
```

## Configuration Options

Change anytime:
```go
// More aggressive
mp.TradeManager.SetConfig(trademanager.AggressiveConfig())

// More conservative
mp.TradeManager.SetConfig(trademanager.ConservativeConfig())

// Back to default
mp.TradeManager.SetConfig(trademanager.DefaultConfig())
```

## View Status

```go
// Print status of all managed positions
mp.TradeManager.PrintStatus()
```

## Troubleshooting

### "cannot find package"
```bash
go mod tidy
```

### "trade manager nil"
Make sure you initialized it in `NewMultiPaperTradingEngine()`

### "callbacks not working"
Verify `SetCallbacks()` is called with correct method names

---

**Next:** Run paper trading and compare results!

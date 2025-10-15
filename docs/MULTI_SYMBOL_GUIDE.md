# 🚀 Multi-Symbol Analysis Guide

This bot can now analyze **ALL Binance USDT pairs in parallel** using Go goroutines!

## 🎯 Features

- ✅ Fetches all trading pairs from Binance API
- ✅ Filters by 24h volume (top N symbols)
- ✅ Parallel processing using goroutines with semaphore (respects API limits)
- ✅ Shows symbols with trading signals
- ✅ Displays top overbought symbols (RSI > 60)
- ✅ Real-time progress tracking
- ✅ Can run in live mode for continuous monitoring

## 📊 Usage Examples

### 1. Quick Scan (Top 20 Symbols)
```bash
go run . --multi --top=20 --interval=4h
```
**Time:** ~8-10 seconds  
**Best for:** Quick market overview

### 2. Comprehensive Scan (Top 50 Symbols)
```bash
go run . --multi --top=50 --interval=4h
```
**Time:** ~15-20 seconds  
**Best for:** Finding best opportunities

### 3. Full Market Scan (ALL USDT Pairs - 500+ symbols)
```bash
go run . --all --interval=4h
```
**Time:** ~2-5 minutes  
**Best for:** Complete market analysis

### 4. Different Timeframes
```bash
# 1-hour timeframe (more signals)
go run . --multi --top=30 --interval=1h

# Daily timeframe (stronger signals)
go run . --multi --top=50 --interval=1d

# 15-minute timeframe (active trading)
go run . --multi --top=20 --interval=15m
```

### 5. Live Monitoring Mode
```bash
# First, enable live mode in engine.go:
# ENABLE_LIVE_MODE = true

# Then run:
go run . --multi --top=30 --interval=1h
```
This will scan 30 symbols every time a new candle closes!

## 📈 Sample Output

```
🔍 Fetching top 50 symbols by 24h volume...
✅ Found 50 symbols

╔════════════════════════════════════════╗
║   MULTI-SYMBOL PARALLEL ANALYSIS       ║
╚════════════════════════════════════════╝

🚀 Analyzing 50 symbols on 4h timeframe
⚡ Using 4 parallel workers

Progress:
[1/50] ✅   BTCUSDT - RSI: 65.23, Div: 0, S/R: 8 (1.12s)
[2/50] ✅   ETHUSDT - RSI: 68.45, Div: 1, S/R: 7 (1.05s)
[3/50] ✅ 🔔 SOLUSDT - RSI: 72.10, Div: 2, S/R: 6 (1.23s)
...
[50/50] ✅   DOGEUSDT - RSI: 52.30, Div: 0, S/R: 5 (0.98s)

⚡ Completed in 15.45 seconds
📊 Average per symbol: 0.31 seconds

════════════════════════════════════════
🎯 SYMBOLS WITH TRADE SIGNALS
════════════════════════════════════════

🔔 SOLUSDT
   📊 RSI: 72.10 (Overbought)
   📈 Divergences: 2
   🎯 S/R Zones: 6
   📉 Signal: SHORT

🔔 MATICUSDT
   📊 RSI: 71.45 (Overbought)
   📈 Divergences: 1
   🎯 S/R Zones: 8
   📉 Signal: SHORT

✅ Found signals in 2/50 symbols

════════════════════════════════════════
📊 TOP 10 OVERBOUGHT SYMBOLS (RSI > 60)
════════════════════════════════════════
  SOLUSDT: RSI 72.10 (Div: 2, S/R: 6)
  MATICUSDT: RSI 71.45 (Div: 1, S/R: 8)
  BTCUSDT: RSI 68.90 (Div: 0, S/R: 7)
  ETHUSDT: RSI 67.23 (Div: 0, S/R: 5)
  BNBUSDT: RSI 65.12 (Div: 1, S/R: 7)
  ADAUSDT: RSI 64.50 (Div: 0, S/R: 6)
  XRPUSDT: RSI 63.20 (Div: 0, S/R: 5)
  DOTUSDT: RSI 62.80 (Div: 0, S/R: 6)
  LINKUSDT: RSI 61.40 (Div: 0, S/R: 5)
  AVAXUSDT: RSI 60.90 (Div: 0, S/R: 4)
════════════════════════════════════════
```

## 🎯 Understanding the Output

### Progress Line Symbols
- `✅` - Analysis completed successfully
- `❌` - API error or data issue
- `🔔` - Trading signal detected for this symbol

### Signal Criteria
A SHORT signal is generated when:
1. ✅ RSI > 70 (Overbought)
2. ✅ At least 1 bearish divergence in last 72 hours
3. ✅ Support/resistance zones identified

### Top Overbought Symbols
- Shows symbols with RSI > 60
- Sorted by RSI (highest first)
- Includes divergence count and S/R zone count
- Useful for finding potential reversal candidates

## ⚙️ Configuration

Edit `engine.go` to adjust:

```go
// Number of parallel workers (more = faster, but respect API limits)
NUM_WORKERS = 4  // Default: 4 (recommended)

// Live mode (continuous monitoring)
ENABLE_LIVE_MODE = false  // Set to true for live monitoring

// Wait for candle close
WAIT_FOR_CANDLE_CLOSE = true  // Only analyze closed candles
```

## 🔧 Performance

### Execution Times (approximate)
- **5 symbols**: ~2-3 seconds
- **20 symbols**: ~6-8 seconds
- **50 symbols**: ~15-20 seconds
- **100 symbols**: ~30-40 seconds
- **500+ symbols (ALL)**: ~2-5 minutes

### Optimization
The bot uses:
- **Goroutines** for parallel processing
- **Semaphore** to limit concurrent API calls (respects Binance rate limits)
- **Worker pool** of 4 concurrent workers by default

## 📝 Tips

1. **Start small**: Test with `--top=5` first
2. **Use volume filter**: Top symbols by volume are more liquid
3. **Live mode**: Great for end-of-candle alerts
4. **Higher timeframes**: Fewer false signals (try 4h or 1d)
5. **Combine with paper trading**: Use multi-symbol to find signals, then paper trade specific pairs

## ⚠️ Important Notes

- **API Rate Limits**: Binance allows ~1200 requests/minute. The bot respects this with semaphores.
- **Network Speed**: Execution time depends on your internet connection
- **False Signals**: Always do your own analysis before trading
- **No Real Money**: This is for analysis only, not execution

## 🚀 Advanced: Combining Modes

### Find Signals, Then Paper Trade
```bash
# Step 1: Find symbols with signals
go run . --multi --top=50 --interval=4h

# Step 2: Paper trade the best signal
go run . --paper --symbol=SOLUSDT --interval=4h --balance=10000
```

### Live Multi-Symbol Monitoring
```bash
# In engine.go, set: ENABLE_LIVE_MODE = true

# Monitor top 30 symbols every hour
go run . --multi --top=30 --interval=1h
```

This will continuously scan 30 symbols and alert you when new signals appear!

---

**Pro Tip:** Use multi-symbol analysis to discover opportunities across the entire market, then focus on specific pairs for detailed analysis or paper trading! 🎯

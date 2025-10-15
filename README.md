# 🤖 Crypto Trading Bot (Go)

Fast cryptocurrency trading bot with technical analysis and parallel processing.

## ⚡ Features

- **Parallel Processing**: 3-4x faster using Go goroutines
- **Technical Indicators**: RSI (14), ATR (30)
- **Bearish Divergence Detection**: Price vs RSI divergences
- **Support/Resistance Zones**: TradingView Bjorgum algorithm
- **Trade Signals**: Entry, stop loss, take profit with R/R ratio
- **Live Mode**: Continuous monitoring on candle closes
- **IST Timezone**: Aligned with Indian markets (UTC+5:30)

## 🚀 Quick Start

### Snapshot Mode (Single Analysis)
```bash
go run . --symbol=BTCUSDT --interval=4h --limit=1000
```

### Paper Trading Mode (Simulated Trading)
```bash
go run . --paper --interval=1m --balance=10000
```
Features:
- ✅ Real-time countdown timer showing time until next candle close
- ✅ Automatic trade execution based on signals
- ✅ Position sizing with risk management (2% per trade)
- ✅ Automatic SL/TP monitoring
- ✅ Performance statistics (win rate, profit factor, P/L)
- ✅ No real money at risk!

### Multi-Symbol Analysis (Scan Multiple Coins)
```bash
# Top 50 symbols by volume (recommended)
go run . --multi --top=50 --interval=4h

# Top 20 symbols (faster)
go run . --multi --top=20 --interval=1h

# ALL USDT pairs (500+ symbols - takes 2-5 minutes)
go run . --all --interval=4h
```
Features:
- ✅ Parallel processing across all symbols using goroutines
- ✅ Automatically fetches top symbols by 24h volume
- ✅ Shows all symbols with trading signals
- ✅ Displays top 10 overbought symbols (RSI > 60)
- ✅ Real-time progress tracking
- ✅ Can be combined with live mode for continuous monitoring

### Multi-Symbol Paper Trading (Trade Multiple Coins) 🆕
```bash
# Paper trade top 50 symbols with max 5 simultaneous positions
go run . --multi-paper --top=50 --interval=4h --balance=10000 --max-pos=5

# Aggressive: Trade top 100 symbols with 10 positions
go run . --multi-paper --top=100 --interval=1h --balance=50000 --max-pos=10

# Conservative: Trade top 20 symbols with 3 positions
go run . --multi-paper --top=20 --interval=4h --balance=10000 --max-pos=3
```
Features:
- ✅ **Automatically trades multiple symbols** when signals appear
- ✅ **Portfolio-wide position management** (max simultaneous positions)
- ✅ **Real-time P/L tracking** across all positions
- ✅ **Auto SL/TP** monitoring for all open trades
- ✅ **Live portfolio balance** updates
- ✅ **Win rate & profit factor** statistics
- ✅ Scans all symbols each candle and opens best opportunities
- ✅ Perfect for diversified simulated trading!

### Live Mode (Continuous)
```go
// In engine.go, change:
ENABLE_LIVE_MODE = true
```
```bash
# Single symbol live monitoring
go run .

# Multi-symbol live monitoring (scans 50 coins every candle)
go run . --multi --top=50 --interval=1h
```

## ⚙️ Configuration

Edit constants in `engine.go`:

```go
// Trading
DEFAULT_SYMBOL   = "BTCUSDT"
DEFAULT_INTERVAL = "4h"
RSI_PERIOD       = 14
ATR_LENGTH       = 30

// Live Mode
ENABLE_LIVE_MODE      = false  // true for continuous
WAIT_FOR_CANDLE_CLOSE = true   // Only analyze closed candles

// Performance
ENABLE_PARALLEL_MODE = true   // 3-4x faster
NUM_WORKERS          = 4      // Concurrent workers
ENABLE_MULTI_SYMBOL  = false  // Multi-symbol analysis

// Display
VERBOSE_MODE = true  // Detailed logs
```

## 📊 Multi-Symbol Analysis

```go
// In engine.go:
ENABLE_MULTI_SYMBOL = true
```

Analyzes BTC, ETH, BNB, SOL, ADA concurrently.

## 🏗️ Architecture

```
┌─────────────────┐
│  Binance API    │  Fetch candle data
└────────┬────────┘
         │
┌────────▼────────┐
│  Indicators     │  RSI + ATR (parallel)
└────────┬────────┘
         │
┌────────▼────────┐
│  Analysis       │  Divergences + S/R (parallel)
└────────┬────────┘
         │
┌────────▼────────┐
│  Trade Signals  │  Entry/SL/TP
└─────────────────┘
```

## 📁 Files

- `binance_fetcher.go` - API client & main entry
- `divergence.go` - RSI & divergence detection
- `support_resistance.go` - ATR & S/R zones (TradingView algorithm)
- `engine.go` - Trading engine with parallel processing

## 🎯 Example Output

### Single Symbol Analysis
```
⚡ Starting parallel indicator calculations...
  ✅ [Thread 1] RSI completed
  ✅ [Thread 2] ATR completed

✅ All indicators calculated

⚡ Starting parallel analysis...
  ✅ [Thread 3] Found 2 divergences
  ✅ [Thread 4] Found 8 S/R zones

🎯 TRADE SIGNAL GENERATED
═══════════════════════════════════════════════════
Entry:      $67,245.00
Stop Loss:  $68,782.50 (↑ 2.29%)
Take Profit: $63,410.25 (↓ 5.70%)
R/R Ratio:  2.49:1  ✅ Good
Position Size: $800 (2.0% risk)
═══════════════════════════════════════════════════

⏱️  Execution time: 283ms (Parallel Mode)
```

### Multi-Symbol Analysis
```
╔════════════════════════════════════════╗
║   MULTI-SYMBOL PARALLEL ANALYSIS       ║
╚════════════════════════════════════════╝

🚀 Analyzing 50 symbols on 4h timeframe
⚡ Using 4 parallel workers

Progress:
[50/50] ✅ 🔔 BTCUSDT - RSI: 72.45, Div: 2, S/R: 8 (1.23s)

⚡ Completed in 15.45 seconds
📊 Average per symbol: 0.31 seconds

════════════════════════════════════════
🎯 SYMBOLS WITH TRADE SIGNALS
════════════════════════════════════════

🔔 BTCUSDT
   📊 RSI: 72.45 (Overbought)
   📈 Divergences: 2
   🎯 S/R Zones: 8
   📉 Signal: SHORT

🔔 ETHUSDT
   📊 RSI: 71.23 (Overbought)
   📈 Divergences: 1
   🎯 S/R Zones: 6
   📉 Signal: SHORT

✅ Found signals in 2/50 symbols

════════════════════════════════════════
📊 TOP 10 OVERBOUGHT SYMBOLS (RSI > 60)
════════════════════════════════════════
  BTCUSDT: RSI 72.45 (Div: 2, S/R: 8)
  ETHUSDT: RSI 71.23 (Div: 1, S/R: 6)
  SOLUSDT: RSI 68.90 (Div: 0, S/R: 7)
  BNBUSDT: RSI 65.12 (Div: 1, S/R: 5)
════════════════════════════════════════
```

### Multi-Symbol Paper Trading
```
╔════════════════════════════════════════╗
║   MULTI-SYMBOL PAPER TRADING v1.0      ║
║   SIMULATED TRADING - NO REAL MONEY    ║
╚════════════════════════════════════════╝

🚀 Symbols: 50 coins
⏰ Interval: 4h
💰 Starting Balance: $10,000.00
📊 Risk per trade: 2.0%
📈 Max Positions: 5

🔔 CANDLE CLOSED - Multi-Symbol Scan #1
⏰ 2025-10-15 10:25:00 IST

[50/50] ✅ 🔔 BTCUSDT - RSI: 72.45, Div: 2, S/R: 8 (1.23s)

🎯 SIGNAL: BTCUSDT (RSI: 72.45, Div: 2, R/R: 2.5:1)

╔════════════════════════════════════════╗
║   📝 NEW POSITION OPENED               ║
╚════════════════════════════════════════╝

🔔 Trade #1: SHORT BTCUSDT
💰 Entry:       $67,245.00
🛑 Stop Loss:   $68,782.50 (2.29%)
🎯 Take Profit: $63,410.25 (5.70%)
📊 Size:        $200.00
⚖️  Risk/Reward: 2.49:1
📈 Active Positions: 1/5

🎯 SIGNAL: ETHUSDT (RSI: 71.23, Div: 1, R/R: 2.2:1)

╔════════════════════════════════════════╗
║   📝 NEW POSITION OPENED               ║
╚════════════════════════════════════════╝

🔔 Trade #2: SHORT ETHUSDT
💰 Entry:       $3,456.00
🛑 Stop Loss:   $3,534.00 (2.26%)
🎯 Take Profit: $3,298.00 (4.57%)
📊 Size:        $200.00
⚖️  Risk/Reward: 2.03:1
📈 Active Positions: 2/5

╔════════════════════════════════════════╗
║      PORTFOLIO SUMMARY                 ║
╚════════════════════════════════════════╝

💰 Starting Balance: $10,000.00
💰 Current Balance:  $10,000.00
📊 Active Positions: 2/5
📈 Total Trades: 0 (W: 0, L: 0)

📋 Active Positions:
  BTCUSDT: SHORT @ $67,245.00 (5m ago)
  ETHUSDT: SHORT @ $3,456.00 (2m ago)
```

## 🕐 Candle Schedule (4h)

**IST Timezone** (Daily open: 5:30 AM IST = 00:00 UTC)

- 05:30 - 09:30 IST
- 09:30 - 13:30 IST  
- 13:30 - 17:30 IST
- 17:30 - 21:30 IST
- 21:30 - 01:30 IST
- 01:30 - 05:30 IST

## 🔧 Requirements

- Go 1.16+
- Internet connection (Binance API)

## 📝 Notes

- Only analyzes **closed candles** (no repainting)
- Multi-symbol analysis uses parallel processing (respects Binance rate limits)
- Live mode is lightweight (sleeps between candles)
- API rate limit: ~1200 requests/minute (multi-symbol mode uses semaphore to avoid limits)

## 🚀 Command Reference

```bash
# Single symbol snapshot
go run . --symbol=BTCUSDT --interval=4h

# Paper trading (single symbol, simulated trades)
go run . --paper --symbol=BTCUSDT --interval=1m --balance=10000

# Multi-symbol scan (top 50 by volume, no trading)
go run . --multi --top=50 --interval=4h

# Multi-symbol scan (top 20, faster)
go run . --multi --top=20 --interval=1h

# Scan ALL USDT pairs (500+ symbols, analysis only)
go run . --all --interval=4h

# Multi-symbol paper trading (auto-trade best signals) ⭐ NEW
go run . --multi-paper --top=50 --interval=4h --balance=10000 --max-pos=5

# Aggressive multi-trading (100 coins, 10 positions)
go run . --multi-paper --top=100 --interval=1h --balance=50000 --max-pos=10

# Multi-symbol live monitoring (requires ENABLE_LIVE_MODE=true)
go run . --multi --top=30 --interval=1h
```

## 📦 Project Files

- `binance_fetcher.go` - API client & main entry point
- `engine.go` - Trading engine with parallel processing
- `divergence.go` - RSI & divergence detection
- `support_resistance.go` - ATR & S/R zones (TradingView algorithm)
- `paper_trading.go` - Single-symbol simulated trading with P/L tracking
- `multi_symbol.go` - Multi-symbol parallel analysis (scan only)
- `multi_paper_trading.go` - Multi-symbol paper trading (auto-trade signals) 🆕

---

Built with Go ⚡ Powered by Binance API

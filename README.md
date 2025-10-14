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

### Live Mode (Continuous)
```go
// In engine.go, change:
ENABLE_LIVE_MODE = true
```
```bash
go run .
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
- Network latency affects API calls (not parallelized)
- Live mode is lightweight (sleeps between candles)

---

Built with Go ⚡ Powered by Binance API

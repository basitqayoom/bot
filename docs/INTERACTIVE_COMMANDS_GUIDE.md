# Interactive Commands Guide

## Overview

The bot now supports **interactive commands** that you can type while the bot is running. This allows you to check the configuration, status, and control the bot without stopping it.

## Available Commands

### 1. Configuration Display (`c` or `config`)

Shows the complete bot configuration including:
- Mode and symbol information
- Trading parameters (balance, risk, R/R ratio)
- Strategy settings (RSI, ATR, divergences)
- Support/Resistance zone configuration
- Risk management rules
- Timezone information with current time

**Usage:**
```bash
c       # Short form
config  # Long form
```

**Example Output (Single-Symbol):**
```
╔════════════════════════════════════════════════════════════════╗
║              BOT CONFIGURATION - PAPER TRADING                  ║
╚════════════════════════════════════════════════════════════════╝

📊 MODE & SYMBOL
  Mode: PAPER TRADING
  Symbol: BTCUSDT
  Interval: 4h
  Starting Balance: $10,000.00

⚙️  LIVE SETTINGS
  Live Mode: ENABLED (continuous monitoring)
  Parallel Mode: ENABLED (3-4x faster)
  Workers: 4

🎯 STRATEGY PARAMETERS
  RSI Period: 14
  RSI Overbought: > 70 (SHORT signals)
  RSI Oversold: < 30 (LONG signals)
  ATR Length: 30 candles
  Divergences: Enabled (Hidden & Regular)
```

**Example Output (Multi-Symbol):**
```
╔════════════════════════════════════════════════════════════════╗
║         BOT CONFIGURATION - MULTI-SYMBOL PAPER TRADING          ║
╚════════════════════════════════════════════════════════════════╝

📊 MODE & SYMBOLS
  Mode: MULTI-SYMBOL PAPER TRADING
  Symbols: 50 coins (BTCUSDT, ETHUSDT, BNBUSDT, ...)
  Interval: 4h
  Starting Balance: $10,000.00

🎯 MULTI-SYMBOL SETTINGS
  Max Positions: 5 simultaneous trades
  Scan Frequency: Every candle close
  Parallel Analysis: ENABLED (50 symbols in ~15s)
```

### 2. Status Check (`s` or `status`)

Shows current bot status, time information, **and portfolio summary** (when in paper trading mode).

**Usage:**
```bash
s       # Short form
status  # Long form
```

**Example Output (Single-Symbol Paper Trading):**
```
════════════════════════════════════════════════════════════════
⚡ BOT STATUS: RUNNING ✅
🕐 Current Time (IST): 2025-10-15 13:30:45
🕐 Current Time (UTC): 2025-10-15 08:00:45
════════════════════════════════════════════════════════════════

════════════════════════════════════════
📊 PAPER TRADING STATS
════════════════════════════════════════
💼 Portfolio Balance: $10,245.50
💰 Starting Balance: $10,000.00
📈 Total P/L: +$245.50 (+2.46%) ✅
📊 Active Trades: 1
📝 Total Trades: 5
✅ Wins: 4 | ❌ Losses: 1
🎯 Win Rate: 80.0%
💵 Profit Factor: 3.21
════════════════════════════════════════
```

**Example Output (Multi-Symbol Paper Trading):**
```
════════════════════════════════════════════════════════════════
⚡ BOT STATUS: RUNNING ✅
🕐 Current Time (IST): 2025-10-15 13:30:45
🕐 Current Time (UTC): 2025-10-15 08:00:45
════════════════════════════════════════════════════════════════

════════════════════════════════════════
💼 PORTFOLIO: $10,245.50 (+$245.50, +2.45%) ✅
📊 Active Positions: 3/5
📈 Total Trades: 12 (W: 8, L: 4)
✅ Win Rate: 66.7%
⚖️  Profit Factor: 2.15

📋 Active Positions:
  BTCUSDT: SHORT @ $42500.00 (45 ago)
  ETHUSDT: SHORT @ $2850.00 (23 ago)
  BNBUSDT: SHORT @ $315.00 (12 ago)
════════════════════════════════════════
```

### 3. Help (`h` or `help`)

Displays all available commands.

**Usage:**
```bash
h       # Short form
help    # Long form
```

**Example Output:**
```
════════════════════════════════════════════════════════════════
📖 INTERACTIVE COMMANDS HELP
════════════════════════════════════════════════════════════════
Available commands (type and press Enter):
  c, config  - Show full bot configuration
  s, status  - Show current status and time
  h, help    - Show this help message
  q, quit    - Exit the bot gracefully
════════════════════════════════════════════════════════════════
```

### 4. Quit (`q` or `quit`)

Gracefully exits the bot, closing all resources and saving trade logs.

**Usage:**
```bash
q       # Short form
quit    # Long form
```

## How It Works

### Single-Symbol Paper Trading

When you start single-symbol paper trading, the configuration is automatically displayed:

```bash
go run . --paper --symbol=BTCUSDT --interval=4h --balance=10000
```

**Output:**
1. Configuration display at startup
2. Bot starts monitoring
3. You can type commands at any time:
   - Type `c` + Enter → See full configuration
   - Type `s` + Enter → Check status
   - Type `q` + Enter → Exit bot

### Multi-Symbol Paper Trading

When you start multi-symbol paper trading, the configuration is automatically displayed:

```bash
go run . --multi-paper --top=50 --interval=4h --balance=10000 --max-pos=5
```

**Output:**
1. Multi-symbol configuration display at startup
2. Bot starts scanning all symbols
3. Interactive commands available:
   - Type `c` + Enter → See full configuration
   - Type `s` + Enter → Check status
   - Type `q` + Enter → Exit bot

## Background Operation

The interactive command listener runs in a **background goroutine**, which means:
- ✅ Commands are processed immediately
- ✅ Bot continues trading while you type
- ✅ No interruption to analysis or trade execution
- ✅ Safe concurrent access with mutex locks

## Use Cases

### 1. Verify Configuration During Long-Running Sessions

```bash
# Bot has been running for hours, want to double-check settings
> c
# Shows all parameters
```

### 2. Check Portfolio & Time During Trading

```bash
# Want to check current portfolio status and P/L
> s
# Shows current time + complete portfolio summary with active trades
```

### 3. Review Parameters After Market Changes

```bash
# Market volatility increased, want to verify risk settings
> c
# Shows risk management parameters (stop loss %, R/R ratio, etc.)
```

### 4. Graceful Shutdown

```bash
# Want to stop bot and save all trade logs properly
> q
# Bot exits cleanly, CSV logs are flushed and closed
```

## Technical Details

### Implementation

The interactive mode uses:
- `bufio.Scanner` to read stdin line-by-line
- Goroutine for non-blocking command processing
- Callback functions for configuration display
- Clean separation from main trading logic

### Thread Safety

- Configuration reads are safe (read-only data)
- No race conditions with trading logic
- Status checks use atomic operations
- Exit command triggers graceful shutdown

## Configuration Display Details

### Single-Symbol Configuration Includes:

1. **Mode & Symbol**: Current trading mode and symbol
2. **Live Settings**: Live mode, parallel mode, workers
3. **Strategy Parameters**: RSI, ATR, divergences
4. **Support/Resistance Zones**: Pivot detection, ATR multiplier
5. **Risk Management**: R/R ratio, max risk %, SL/TP percentages
6. **Timezone Info**: IST and UTC times

### Multi-Symbol Configuration Includes:

1. **Mode & Symbols**: Number of symbols and list preview
2. **Multi-Symbol Settings**: Max positions, scan frequency
3. **Strategy Parameters**: Same as single-symbol
4. **Support/Resistance Zones**: Same as single-symbol
5. **Risk Management**: Same as single-symbol
6. **Timezone Info**: IST and UTC times

## Examples

### Example 1: Checking Config During Paper Trading

```bash
# Start bot
$ go run . --paper --symbol=ETHUSDT --interval=1h --balance=5000

# Configuration displayed at startup...
# Bot is running...

# Type 'c' to see config again
> c

╔════════════════════════════════════════════════════════════════╗
║              BOT CONFIGURATION - PAPER TRADING                  ║
╚════════════════════════════════════════════════════════════════╝
...full configuration...

# Continue trading...
```

### Example 2: Status Check During Multi-Symbol Trading

```bash
# Start multi-symbol bot
$ go run . --multi-paper --top=50 --interval=4h --balance=10000

# Bot scanning multiple symbols...

# Check status
> s

════════════════════════════════════════════════════════════════
📊 BOT STATUS
════════════════════════════════════════════════════════════════
🤖 Status: RUNNING
⏰ Current Time (IST): 2025-10-15 13:30:45
🌍 Current Time (UTC): 2025-10-15 08:00:45
════════════════════════════════════════════════════════════════

# Continue scanning...
```

### Example 3: Graceful Exit

```bash
# Bot is running with active trades...

> q

💾 Closing trade logger...
📁 All trades saved to: trade_logs/trades_BTCUSDT.csv
👋 Bot stopped gracefully
```

## Notes

- Commands are **case-insensitive** (both `c` and `C` work)
- Commands must be followed by **Enter**
- Invalid commands are ignored with a warning message
- Configuration callback is set during bot initialization
- Works in both single-symbol and multi-symbol modes

## Best Practices

1. **Use `c` command regularly** to verify parameters during long sessions
2. **Use `s` command** to check time synchronization before important candle closes
3. **Use `q` command** instead of Ctrl+C to ensure proper CSV log saves
4. **Monitor the console** while typing commands (bot continues trading)

## Future Enhancements (Potential)

- [ ] Real-time portfolio summary command (`p` or `portfolio`)
- [ ] Trade history command (`t` or `trades`)
- [ ] Performance metrics command (`m` or `metrics`)
- [ ] Pause/Resume trading (`pause`, `resume`)
- [ ] Dynamic parameter adjustment (change risk, add symbols, etc.)

---

**Version:** 1.0  
**Last Updated:** October 2025  
**Related Guides:** `README.md`, `CSV_LOGGING_GUIDE.md`, `MULTI_SYMBOL_GUIDE.md`

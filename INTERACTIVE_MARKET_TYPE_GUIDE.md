# Interactive Market Type Display Guide

## Overview

The bot's interactive mode now displays detailed market type information when you press the `c` command. This helps you verify which market (Spot or Futures) you're currently trading on and which API endpoints are being used.

## Quick Reference

### Interactive Commands

| Command | Description |
|---------|-------------|
| `c` or `config` | Show bot configuration **including market type and base URL** |
| `s` or `status` | Show current status & portfolio |
| `h` or `help` | Show help message |
| `q` or `quit` | Quit bot gracefully |

## What's New in the 'c' Command

When you type `c` during bot execution, you'll now see:

### Market Configuration Section
```
🌐 MARKET CONFIGURATION:
   Market Type:       SPOT or FUTURES
   Base URL:          https://api.binance.com or https://fapi.binance.com
   Endpoint:          /api/v3/klines or /fapi/v1/klines
```

### For Multi-Symbol Mode
Multi-symbol paper trading shows additional endpoints:
```
🌐 MARKET CONFIGURATION:
   Market Type:       FUTURES
   Base URL:          https://fapi.binance.com
   Endpoint:          /fapi/v1/klines
   Exchange Info:     /fapi/v1/exchangeInfo
   24h Ticker:        /fapi/v1/ticker/24hr
```

## Usage Examples

### Example 1: Single Symbol Spot Market

**Start the bot:**
```bash
./bot --symbol BTCUSDT --interval 1m --limit 100
```

**During execution, type 'c' + Enter:**
```
╔════════════════════════════════════════════════════════════╗
║              BOT CONFIGURATION                             ║
╚════════════════════════════════════════════════════════════╝
🤖 Mode:              Single Symbol Analysis
📊 Symbol:            BTCUSDT
⏰ Interval:          1m
💰 Balance:           $10000.00

🌐 MARKET CONFIGURATION:
   Market Type:       SPOT
   Base URL:          https://api.binance.com
   Endpoint:          /api/v3/klines

⚙️  SYSTEM SETTINGS:
   Live Mode:         true
   Wait for Close:    true
   Parallel Mode:     true
   Workers:           8
   Verbose:           true

📈 STRATEGY PARAMETERS:
   RSI Period:        14
   ATR Length:        30
   Min Divergences:   1
   Swing Lookback:    2

🎯 S/R ZONE SETTINGS:
   Pivot Left:        20
   Pivot Right:       15
   ATR Multiplier:    0.5
   Max Zone %:        5.0%
   Align Zones:       true

💵 RISK MANAGEMENT:
   Risk/Reward:       1.5:1
   Max Risk:          1.0%
   Stop Loss:         0.4%
   Take Profit:       0.8%

🌍 TIMEZONE:
   Current Time (IST): 2025-10-15 16:47:30
   Current Time (UTC): 2025-10-15 11:17:30
   Offset:            UTC+5:30
════════════════════════════════════════════════════════════
```

### Example 2: Single Symbol Futures Market

**Start the bot:**
```bash
./bot --symbol ETHUSDT --interval 1m --limit 100 --futures
```

**During execution, type 'c' + Enter:**
```
╔════════════════════════════════════════════════════════════╗
║              BOT CONFIGURATION                             ║
╚════════════════════════════════════════════════════════════╝
🤖 Mode:              Single Symbol Analysis
📊 Symbol:            ETHUSDT
⏰ Interval:          1m
💰 Balance:           $10000.00

🌐 MARKET CONFIGURATION:
   Market Type:       FUTURES
   Base URL:          https://fapi.binance.com
   Endpoint:          /fapi/v1/klines

⚙️  SYSTEM SETTINGS:
   Live Mode:         true
   Wait for Close:    true
   Parallel Mode:     true
   Workers:           8
   Verbose:           true

... (rest of configuration)
```

### Example 3: Multi-Symbol Futures Paper Trading

**Start the bot:**
```bash
./bot --multi-paper --top 50 --futures --quiet
```

**During execution, type 'c' + Enter:**
```
╔════════════════════════════════════════════════════════════╗
║         MULTI-SYMBOL BOT CONFIGURATION                     ║
╚════════════════════════════════════════════════════════════╝
🤖 Mode:              Multi-Symbol Paper Trading
📊 Symbols:           50 pairs
⏰ Interval:          4h
💰 Starting Balance:  $10000.00
🎯 Max Positions:     5

🌐 MARKET CONFIGURATION:
   Market Type:       FUTURES
   Base URL:          https://fapi.binance.com
   Endpoint:          /fapi/v1/klines
   Exchange Info:     /fapi/v1/exchangeInfo
   24h Ticker:        /fapi/v1/ticker/24hr

⚙️  SYSTEM SETTINGS:
   Live Mode:         true
   Wait for Close:    true
   Parallel Mode:     true
   Workers:           8

📋 SYMBOLS LIST (First 10):
   1. BTCUSDT
   2. ETHUSDT
   3. BNBUSDT
   4. SOLUSDT
   5. XRPUSDT
   6. ADAUSDT
   7. DOGEUSDT
   8. MATICUSDT
   9. DOTUSDT
   10. LTCUSDT
   ... and 40 more

... (rest of configuration)
```

## Market Type Indicators

### Spot Market Indicators
- **Market Type**: `SPOT`
- **Base URL**: `https://api.binance.com`
- **Klines Endpoint**: `/api/v3/klines`
- **Exchange Info**: `/api/v3/exchangeInfo`
- **24h Ticker**: `/api/v3/ticker/24hr`

### Futures Market Indicators
- **Market Type**: `FUTURES`
- **Base URL**: `https://fapi.binance.com`
- **Klines Endpoint**: `/fapi/v1/klines`
- **Exchange Info**: `/fapi/v1/exchangeInfo`
- **24h Ticker**: `/fapi/v1/ticker/24hr`

## Why This Matters

### 1. **Verification**
- Instantly verify which market you're trading on
- Avoid confusion between spot and futures
- Confirm the correct API endpoints are being used

### 2. **Debugging**
- Troubleshoot API connection issues
- Verify configuration before placing trades
- Check if data is coming from the right source

### 3. **Safety**
- Futures and Spot have different risk profiles
- Knowing which market you're on helps prevent mistakes
- Easy to verify before making trading decisions

### 4. **Multi-Session Management**
- When running multiple bot instances
- Quickly identify which instance is on which market
- Better organization and tracking

## Practical Workflow

### Workflow 1: Quick Market Type Check
```bash
# Start bot
./bot --symbol BTCUSDT --interval 1m --futures

# During execution:
# 1. Type 'c' + Enter to check market type
# 2. Verify: Market Type: FUTURES
# 3. Verify: Base URL: https://fapi.binance.com
# 4. Continue trading or adjust as needed
```

### Workflow 2: Compare Spot vs Futures Setup
```bash
# Terminal 1: Spot market
./bot --multi-paper --top 20 --interval 1m

# Terminal 2: Futures market
./bot --multi-paper --top 20 --interval 1m --futures

# In each terminal:
# Type 'c' to see configuration
# Compare market types and endpoints
```

### Workflow 3: Debugging Connection Issues
```bash
# Start bot
./bot --symbol ETHUSDT --interval 5m --futures

# If experiencing issues:
# 1. Type 'c' + Enter
# 2. Check Base URL
# 3. Check Endpoint
# 4. Verify market type matches intention
# 5. Check if Binance API is accessible for that market
```

## Common Questions

### Q: How do I know if I'm on Spot or Futures?
**A:** Type `c` during bot execution. Look for the "Market Type" field in the output.

### Q: Can I switch markets without restarting?
**A:** No, the market type is set at startup via the `--futures` flag. You need to restart the bot to switch markets.

### Q: What if the wrong market is displayed?
**A:** Stop the bot (type `q`), then restart with or without the `--futures` flag:
- For Spot: `./bot --symbol BTCUSDT --interval 1m`
- For Futures: `./bot --symbol BTCUSDT --interval 1m --futures`

### Q: Does the market type affect my strategy?
**A:** The technical analysis strategy remains the same, but:
- Futures typically have higher volatility
- Different symbols may be available on each market
- Futures support both LONG and SHORT positions
- Risk management should be adjusted accordingly

## Integration with Other Features

### With Paper Trading
```bash
./bot --paper --balance 5000 --symbol BTCUSDT --futures
# Type 'c' to verify you're paper trading on futures
# Type 's' to see portfolio status
```

### With Multi-Symbol Analysis
```bash
./bot --multi --top 50 --futures
# Type 'c' to see market type and all 50 symbols
```

### With Quiet Mode
```bash
./bot --multi-paper --futures --quiet
# Type 'c' to see full config even in quiet mode
```

## Testing

Use the provided test script to see the market type display in action:

```bash
# Make script executable
chmod +x test_interactive_market.sh

# Run the test
./test_interactive_market.sh
```

This will demonstrate the 'c' command output for both Spot and Futures markets.

## Related Documentation

- **FUTURES_SPOT_GUIDE.md** - Complete guide on using Spot vs Futures markets
- **INTERACTIVE_COMMANDS_GUIDE.md** - Full interactive commands documentation
- **PORTFOLIO_STATUS_GUIDE.md** - Using the 's' command for portfolio tracking

## Summary

The enhanced 'c' command now provides:
- ✅ Clear market type indication (SPOT/FUTURES)
- ✅ Base URL for API calls
- ✅ Endpoint paths for data fetching
- ✅ Quick verification during runtime
- ✅ Better debugging capabilities
- ✅ Increased trading safety

Use it frequently to verify your bot's configuration and ensure you're trading on the intended market!

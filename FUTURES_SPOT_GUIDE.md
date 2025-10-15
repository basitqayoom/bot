# Binance Futures vs Spot Market Trading Guide

## Overview

The bot now supports both **Binance Spot** and **Binance Futures** markets. You can easily switch between them using the `--futures` flag.

## Quick Start

### Using Spot Market (Default)
```bash
# Single symbol on spot market
./bot --symbol BTCUSDT --interval 1m

# Multi-symbol paper trading on spot market
./bot --multi-paper --top 50 --balance 10000
```

### Using Futures Market
```bash
# Single symbol on futures market
./bot --symbol BTCUSDT --interval 1m --futures

# Multi-symbol paper trading on futures market
./bot --multi-paper --top 50 --balance 10000 --futures

# Analyze all USDT futures pairs
./bot --multi --all --futures
```

## Key Differences

### Spot Market
- **Endpoint**: `https://api.binance.com`
- **Trading Type**: Buy and hold actual cryptocurrencies
- **Leverage**: None (1x only)
- **Pairs**: BTCUSDT, ETHUSDT, etc.
- **Use Case**: Long-term investing, lower risk

### Futures Market
- **Endpoint**: `https://fapi.binance.com`
- **Trading Type**: Perpetual contracts (derivatives)
- **Leverage**: Up to 125x (bot uses paper trading values)
- **Pairs**: BTCUSDT, ETHUSDT (perpetual contracts)
- **Use Case**: Short-term trading, higher volatility

## Command Examples

### 1. Single Symbol Analysis

**Spot Market:**
```bash
./bot --symbol ETHUSDT --interval 4h --limit 1000
```

**Futures Market:**
```bash
./bot --symbol ETHUSDT --interval 4h --limit 1000 --futures
```

### 2. Paper Trading

**Spot Paper Trading:**
```bash
./bot --symbol BTCUSDT --interval 1m --paper --balance 5000
```

**Futures Paper Trading:**
```bash
./bot --symbol BTCUSDT --interval 1m --paper --balance 5000 --futures
```

### 3. Multi-Symbol Analysis

**Top 50 Spot Pairs by Volume:**
```bash
./bot --multi --top 50 --interval 1m
```

**Top 50 Futures Pairs by Volume:**
```bash
./bot --multi --top 50 --interval 1m --futures
```

### 4. Multi-Symbol Paper Trading

**Spot Multi-Paper Trading:**
```bash
./bot --multi-paper --top 30 --max-pos 5 --balance 10000
```

**Futures Multi-Paper Trading:**
```bash
./bot --multi-paper --top 30 --max-pos 5 --balance 10000 --futures
```

### 5. All USDT Pairs Analysis

**All Spot USDT Pairs:**
```bash
./bot --multi --all --interval 1m
```

**All Futures USDT Pairs:**
```bash
./bot --multi --all --interval 1m --futures
```

### 6. Quiet Mode (Minimal Output)

**Futures with Quiet Mode:**
```bash
./bot --multi-paper --top 50 --futures --quiet
```

## Flag Reference

| Flag | Description | Default |
|------|-------------|---------|
| `--futures` | Use Binance Futures market | false (uses Spot) |
| `--symbol` | Trading pair symbol | BTCUSDT |
| `--interval` | Timeframe (1m, 5m, 15m, 30m, 1h, 4h, 1d) | 4h |
| `--limit` | Number of candles to fetch | 1000 |
| `--paper` | Enable paper trading mode | false |
| `--balance` | Starting balance for paper trading | 10000.0 |
| `--multi` | Enable multi-symbol analysis | false |
| `--multi-paper` | Multi-symbol paper trading | false |
| `--top` | Number of top symbols by volume | 50 |
| `--all` | Analyze ALL USDT pairs | false |
| `--max-pos` | Max simultaneous positions | 5 |
| `--quiet` | Quiet mode (minimal output) | false |

## Implementation Details

### Market Type Detection

The bot automatically adjusts API endpoints based on the `--futures` flag:

**Spot Market URLs:**
- Exchange Info: `https://api.binance.com/api/v3/exchangeInfo`
- Klines: `https://api.binance.com/api/v3/klines`
- 24h Ticker: `https://api.binance.com/api/v3/ticker/24hr`

**Futures Market URLs:**
- Exchange Info: `https://fapi.binance.com/fapi/v1/exchangeInfo`
- Klines: `https://fapi.binance.com/fapi/v1/klines`
- 24h Ticker: `https://fapi.binance.com/fapi/v1/ticker/24hr`

### Global Variable

```go
// Set via --futures flag
var USE_FUTURES bool = false

// Helper functions
func GetBaseURL() string {
    if USE_FUTURES {
        return "https://fapi.binance.com"
    }
    return "https://api.binance.com"
}

func GetKlinesEndpoint() string {
    if USE_FUTURES {
        return "/fapi/v1/klines"
    }
    return "/api/v3/klines"
}
```

## Trading Log Files

Trade logs are stored in the same CSV file regardless of market type:
- **Location**: `trade_logs/trades_all_symbols.csv`
- **Note**: Consider the market type when analyzing your trading history

## Best Practices

### For Spot Trading
1. ✅ Suitable for long-term strategies
2. ✅ Lower risk, no liquidation
3. ✅ Good for learning the bot's behavior
4. ⚠️ Requires actual capital to execute

### For Futures Trading
1. ✅ Higher volatility = more opportunities
2. ✅ Can profit from both up and down markets (SHORT/LONG)
3. ⚠️ Higher risk due to leverage
4. ⚠️ Funding rates apply
5. ⚠️ Risk of liquidation with real trading

### Risk Management
- Start with **spot market** to understand the bot
- Test with **paper trading** before using real funds
- Use **--quiet** mode for cleaner logs
- Monitor **top volume pairs** for better liquidity
- Keep **--max-pos** low (5-10) to manage risk

## Troubleshooting

### "Unexpected status" errors
- Check if symbol exists on the selected market
- Some pairs are only available on Spot or only on Futures

### No signals detected
- Try different timeframes: `1m`, `5m`, `15m`, `1h`, `4h`
- Increase analysis period: `--limit 1000`
- Check if market is trending or ranging

### API Rate Limits
- Binance has rate limits (1200 requests/minute for Spot, 2400 for Futures)
- Use `--top N` with lower N values for faster analysis
- Avoid using `--all` too frequently

## Example Workflows

### Workflow 1: Quick Futures Scalping Test
```bash
# 1. Analyze top 30 futures pairs on 1m timeframe
./bot --multi --top 30 --interval 1m --futures

# 2. Run paper trading on best candidates
./bot --multi-paper --top 10 --interval 1m --futures --quiet
```

### Workflow 2: Spot Market Swing Trading
```bash
# 1. Analyze on 4h timeframe for swing opportunities
./bot --multi --top 50 --interval 4h

# 2. Paper trade specific pair
./bot --symbol ETHUSDT --interval 4h --paper --balance 5000
```

### Workflow 3: Compare Spot vs Futures Performance
```bash
# Run same strategy on both markets
./bot --multi-paper --top 20 --interval 1m --balance 10000          # Spot
./bot --multi-paper --top 20 --interval 1m --balance 10000 --futures # Futures

# Compare the trade logs
```

## Notes

- **Paper trading** simulates trades without real money
- **Real trading** requires API keys and actual capital (not implemented in this bot)
- Always test strategies thoroughly before considering real trading
- Past performance does not guarantee future results
- Cryptocurrency trading carries significant risk

## Support

For issues or questions:
1. Check the logs in `trade_logs/`
2. Review the technical documentation in other MD files
3. Ensure you're using the correct flags for your intended market

---

**Market Type Indicator**: The bot displays the market type on startup:
```
📊 Market Type: SPOT
```
or
```
📊 Market Type: FUTURES
```

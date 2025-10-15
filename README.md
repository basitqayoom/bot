# Crypto Trading Bot 🚀

A high-performance cryptocurrency trading bot with bearish divergence detection, support/resistance analysis, and multi-symbol paper trading.

## 📁 Project Structure

```
bot/
├── *.go                # Go source files
├── bot                 # Compiled executable
├── docs/              # All documentation
├── logs/              # Trade logs (CSV files)
├── config/            # Configuration files
├── scripts/           # Helper scripts
└── cmd/               # Additional commands
```

## 🚀 Quick Start

```bash
# Multi-symbol paper trading (100 coins, 1-minute)
./bot --multi --top 100 --interval 1m --max-pos 10

# Futures market
./bot --multi --top 100 --interval 1m --futures --max-pos 10

# Single symbol
./bot --symbol BTCUSDT --interval 1m
```

## 📚 Documentation

All documentation is in the `docs/` folder:
- Usage guides
- Implementation details
- Feature documentation
- Quick references

## 📊 Features

✅ Multi-symbol concurrent trading
✅ Max profit & give-back tracking
✅ Spot & Futures market support
✅ Real-time candle synchronization
✅ CSV trade logging
✅ Interactive commands

📝 **Trade Logs**: `logs/trade_logs/trades_all_symbols.csv`

---
**Happy Trading/Users/basitqayoomchowdhary/Desktop/personal/project/bot && mv trade_logs logs/ 2>/dev/null; mv test_*.sh scripts/ 2>/dev/null; echo "✅ Logs and scripts organized"* ��

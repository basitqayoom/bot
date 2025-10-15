# Portfolio Status Command - Quick Reference

## 🚀 New Feature: Portfolio Display on Status Check

When you type `s` (status) during paper trading, you now get:
1. ⏰ Current time (IST & UTC)
2. 💼 **Complete portfolio summary**
3. 📊 Active trades breakdown

---

## 📋 Example Workflow

### Start Paper Trading
```bash
go run . --paper --symbol=BTCUSDT --interval=4h --balance=10000
```

### While Bot is Running...

#### Check Status & Portfolio
```
> s    [press Enter]
```

**You'll see:**
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

📋 ACTIVE TRADES:
1. BTCUSDT SHORT @ $42500.00
   Entry: 2025-10-15 10:00:00
   Stop Loss: $43200.00
   Take Profit: $40500.00
   Position Size: $234.56
   Duration: 3h 30m
════════════════════════════════════════
```

---

## 📊 Multi-Symbol Example

### Start Multi-Symbol Paper Trading
```bash
go run . --multi-paper --top=50 --interval=4h --balance=10000 --max-pos=5
```

### Check Status
```
> s    [press Enter]
```

**You'll see:**
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
  BTCUSDT: SHORT @ $42500.00 (45 minutes ago)
  ETHUSDT: SHORT @ $2850.00 (23 minutes ago)
  BNBUSDT: SHORT @ $315.00 (12 minutes ago)
════════════════════════════════════════
```

---

## 🎯 What You Get

### Always Displayed:
- ✅ Bot status (running/stopped)
- ✅ Current time in IST
- ✅ Current time in UTC

### In Paper Trading Mode:
- ✅ Portfolio balance
- ✅ Total P/L ($ and %)
- ✅ Active trades count
- ✅ Win/Loss statistics
- ✅ Win rate percentage
- ✅ Profit factor
- ✅ List of active positions (if any)

---

## 💡 Use Cases

### 1. Quick Portfolio Check
```bash
# Bot running for hours, want to see P/L quickly
> s

# Shows portfolio without stopping the bot
```

### 2. Monitor Active Positions
```bash
# Have multiple positions open, want to review them
> s

# Shows all active trades with entry prices and duration
```

### 3. Check Performance Metrics
```bash
# Want to see win rate and profit factor
> s

# Displays full statistics including wins, losses, profit factor
```

### 4. Time Verification
```bash
# Want to verify current time before important candle close
> s

# Shows both IST and UTC time
```

---

## 🔄 Comparison: Before vs After

### Before (Old Status Command)
```
> s

⚡ BOT STATUS: RUNNING ✅
🕐 Current Time (IST): 2025-10-15 13:30:45
🕐 Current Time (UTC): 2025-10-15 08:00:45
```

### After (New Status Command with Portfolio)
```
> s

════════════════════════════════════════════════════════════════
⚡ BOT STATUS: RUNNING ✅
🕐 Current Time (IST): 2025-10-15 13:30:45
🕐 Current Time (UTC): 2025-10-15 08:00:45
════════════════════════════════════════════════════════════════

[COMPLETE PORTFOLIO SUMMARY HERE]
- Balance
- P/L
- Active trades
- Win rate
- Statistics
```

**Much more useful!** 🎉

---

## 🛠️ Technical Details

### Implementation
- Status callback function passed to `StartInteractiveMode()`
- Single-symbol: Calls `engine.PrintStats()`
- Multi-symbol: Calls `mp.PrintPortfolio()`
- Non-blocking (bot continues trading)
- Thread-safe with mutex locks

### Performance
- ⚡ Instant response (< 1ms)
- 🔒 Thread-safe concurrent access
- 🚀 No impact on trading logic

---

## 📚 Related Commands

| Command | What It Shows |
|---------|---------------|
| `s` | Status + **Portfolio** (this guide) |
| `c` | Full configuration |
| `h` | Help message |
| `q` | Quit gracefully |

---

## ✨ Pro Tips

1. **Check status regularly** - Type `s` every few hours to monitor P/L
2. **Before closing bot** - Type `s` to see final stats before quitting
3. **After opening trades** - Type `s` to verify positions were opened
4. **During high volatility** - Type `s` to check if stop losses were hit

---

**Version:** 1.1 (Enhanced with portfolio display)  
**Last Updated:** October 2025  
**Related:** `INTERACTIVE_COMMANDS_GUIDE.md`

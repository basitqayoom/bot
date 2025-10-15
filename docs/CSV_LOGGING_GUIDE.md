# 📊 CSV Trade Logging Guide

## Overview

All closed trades are automatically logged to CSV files for complete trade history and analysis.

## 📁 File Locations

### Single Symbol Paper Trading
```
trade_logs/trades_BTCUSDT.csv
trade_logs/trades_ETHUSDT.csv
```
- One file per symbol
- All trades for that symbol appended to the same file
- Persistent across multiple runs

### Multi-Symbol Paper Trading
```
trade_logs/trades_all_symbols.csv
```
- **Single file for ALL symbols**
- Complete history of every trade across all pairs
- All BTCUSDT, ETHUSDT, SOLUSDT, etc. trades in one place

## 📋 CSV Columns

| Column | Description | Example |
|--------|-------------|---------|
| `Trade_ID` | Unique trade identifier | 1, 2, 3... |
| `Symbol` | Trading pair | BTCUSDT, ETHUSDT |
| `Side` | Trade direction | SHORT, LONG |
| `Entry_Time` | Position open time | 2025-10-15 09:30:00 |
| `Entry_Price` | Entry price | 67500.00 |
| `Exit_Time` | Position close time | 2025-10-15 13:30:00 |
| `Exit_Price` | Exit price | 66800.00 |
| `Stop_Loss` | Stop loss price | 69000.00 |
| `Take_Profit` | Take profit price | 64500.00 |
| `Position_Size` | Position size in USD | 200.00 |
| `Status` | Close reason | CLOSED_TP, CLOSED_SL, CLOSED_WIN, CLOSED_LOSS |
| `Profit_Loss` | P/L in USD | 210.37, -28.57 |
| `Profit_Loss_Pct` | P/L percentage | 5.26, -3.57 |
| `Risk_Reward` | Risk/reward ratio | 2.00, 1.50 |
| `Duration_Minutes` | Trade duration | 240.00 (4 hours) |
| `Logged_At` | Timestamp logged | 2025-10-15 13:30:05 |

## 🚀 Usage Examples

### Single Symbol Paper Trading

```bash
# Start paper trading BTCUSDT
go run . --paper --symbol=BTCUSDT --interval=4h --balance=10000

# Trades will be saved to:
# trade_logs/trades_BTCUSDT.csv
```

**Output when trade closes:**
```
╔════════════════════════════════════════╗
║      PAPER TRADE CLOSED                ║
╚════════════════════════════════════════╝

📝 Trade #1: SHORT BTCUSDT
📍 Entry:  $67500.00 → Exit: $66800.00
📊 Reason: TAKE_PROFIT
⏱️  Duration: 4h 0m 0s
💰 P/L: +$210.37 (+5.26%) ✅
💵 Balance: $10000.00 → $10210.37
════════════════════════════════════════
💾 Trade logged to CSV: trade_logs/trades_BTCUSDT.csv
```

### Multi-Symbol Paper Trading

```bash
# Trade top 50 symbols with max 5 positions
go run . --multi-paper --top=50 --interval=4h --balance=10000 --max-pos=5

# All trades saved to:
# trade_logs/trades_all_symbols.csv
```

**On startup:**
```
📝 Appending to existing multi-symbol trade log: trade_logs/trades_all_symbols.csv

🚀 Symbols: 50 coins
⏰ Interval: 4h
💰 Starting Balance: $10000.00
📊 Risk per trade: 2.0%
📈 Max Positions: 5
```

**When trade closes:**
```
╔════════════════════════════════════════╗
║   ❌ POSITION CLOSED                   ║
╚════════════════════════════════════════╝

📝 Trade #3: SHORT ETHUSDT
📍 Entry:  $3500.00 → Exit: $3350.00
📊 Reason: TAKE_PROFIT
⏱️  Duration: 2h 15m 0s
💰 P/L: +$85.71 (+4.29%) ✅
💵 Portfolio: $10295.08 (+2.95%) ✅
📈 Active Positions: 2/5
════════════════════════════════════════
💾 Trade logged to CSV
```

## 📊 Example CSV Content

```csv
Trade_ID,Symbol,Side,Entry_Time,Entry_Price,Exit_Time,Exit_Price,Stop_Loss,Take_Profit,Position_Size,Status,Profit_Loss,Profit_Loss_Pct,Risk_Reward,Duration_Minutes,Logged_At
1,BTCUSDT,SHORT,2025-10-15 09:30:00,67500.00,2025-10-15 13:30:00,66800.00,69000.00,64500.00,200.00,CLOSED_TP,210.37,5.26,2.00,240.00,2025-10-15 13:30:05
2,ETHUSDT,SHORT,2025-10-15 09:30:00,3500.00,2025-10-15 11:45:00,3550.00,3600.00,3350.00,200.00,CLOSED_LOSS,-28.57,-3.57,2.00,135.00,2025-10-15 11:45:08
3,BTCUSDT,SHORT,2025-10-15 17:30:00,68200.00,2025-10-15 21:30:00,67100.00,69700.00,65200.00,200.00,CLOSED_TP,323.17,8.08,2.00,240.00,2025-10-15 21:30:12
4,SOLUSDT,SHORT,2025-10-15 17:30:00,145.50,2025-10-15 19:00:00,148.00,150.00,140.00,200.00,CLOSED_LOSS,-34.36,-4.30,2.00,90.00,2025-10-15 19:00:15
```

## 📈 Analyzing Your Trades

### Import into Excel/Google Sheets
1. Open Excel or Google Sheets
2. Import CSV file
3. Create pivot tables, charts, etc.

### Calculate Performance Metrics

**Win Rate:**
```
= COUNTIF(Status, "*WIN*") / COUNT(Trade_ID) * 100
```

**Average Win:**
```
= AVERAGEIF(Profit_Loss, ">0")
```

**Average Loss:**
```
= AVERAGEIF(Profit_Loss, "<0")
```

**Profit Factor:**
```
= SUMIF(Profit_Loss, ">0") / ABS(SUMIF(Profit_Loss, "<0"))
```

**Total P/L:**
```
= SUM(Profit_Loss)
```

### Python Analysis Example

```python
import pandas as pd
import matplotlib.pyplot as plt

# Load trades
df = pd.read_csv('trade_logs/trades_all_symbols.csv')

# Convert timestamps
df['Entry_Time'] = pd.to_datetime(df['Entry_Time'])
df['Exit_Time'] = pd.to_datetime(df['Exit_Time'])

# Performance metrics
total_trades = len(df)
win_rate = (df['Profit_Loss'] > 0).sum() / total_trades * 100
total_pl = df['Profit_Loss'].sum()
avg_win = df[df['Profit_Loss'] > 0]['Profit_Loss'].mean()
avg_loss = df[df['Profit_Loss'] < 0]['Profit_Loss'].mean()

print(f"Total Trades: {total_trades}")
print(f"Win Rate: {win_rate:.2f}%")
print(f"Total P/L: ${total_pl:.2f}")
print(f"Average Win: ${avg_win:.2f}")
print(f"Average Loss: ${avg_loss:.2f}")

# Plot cumulative P/L
df['Cumulative_PL'] = df['Profit_Loss'].cumsum()
plt.plot(df['Exit_Time'], df['Cumulative_PL'])
plt.title('Cumulative P/L Over Time')
plt.xlabel('Time')
plt.ylabel('Cumulative P/L ($)')
plt.grid(True)
plt.show()

# Best/worst trades
print("\nTop 5 Best Trades:")
print(df.nlargest(5, 'Profit_Loss')[['Symbol', 'Entry_Time', 'Profit_Loss']])

print("\nTop 5 Worst Trades:")
print(df.nsmallest(5, 'Profit_Loss')[['Symbol', 'Entry_Time', 'Profit_Loss']])
```

## 🔄 File Management

### Append vs Overwrite
- ✅ **Append mode** - New trades added to existing file
- ✅ **Persistent** - Never loses history
- ✅ **Safe** - Multiple runs don't overwrite data

### File Creation
- **First run**: Creates new file with headers
- **Subsequent runs**: Appends to existing file

### Backup Recommendation
```bash
# Backup trade logs periodically
cp -r trade_logs trade_logs_backup_$(date +%Y%m%d)
```

## 🎯 Trade Status Types

| Status | Meaning |
|--------|---------|
| `CLOSED_TP` | Take profit hit (win) |
| `CLOSED_SL` | Stop loss hit (loss) |
| `CLOSED_SL_WIN` | Stop loss hit but still profitable |
| `CLOSED_WIN` | Manual close with profit |
| `CLOSED_LOSS` | Manual close with loss |

## 📝 Notes

- CSV files are created in `trade_logs/` directory (auto-created)
- Files use UTF-8 encoding
- Timestamps in format: `YYYY-MM-DD HH:MM:SS`
- Prices and P/L formatted to 2 decimal places
- Duration in minutes (can be converted: 240 min = 4 hours)

## 🚨 Troubleshooting

**Logger fails to create:**
```
⚠️  Failed to create trade logger: permission denied
```
**Solution:** Check write permissions on current directory

**CSV not updating:**
- File is flushed after each trade
- Check if file is open in Excel (lock issue)

**Missing trades:**
- All closed trades are logged immediately
- Check terminal output for "💾 Trade logged to CSV"

---

**Happy Trading! 📊✅**

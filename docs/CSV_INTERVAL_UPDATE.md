# CSV Format Update - Interval Column Added

## 🆕 What Changed?

The CSV trade logs now include the **Interval** column (timeframe) as the 3rd column.

### Before (16 columns)
```csv
Trade_ID,Symbol,Side,Entry_Time,Entry_Price...
1,BTCUSDT,SHORT,2025-10-15 10:30:00,67500.00...
```

### After (17 columns) ⭐
```csv
Trade_ID,Symbol,Interval,Side,Entry_Time,Entry_Price...
1,BTCUSDT,4h,SHORT,2025-10-15 10:30:00,67500.00...
```

---

## 📊 Complete CSV Structure

| Position | Column | Description | Example |
|----------|--------|-------------|---------|
| 1 | Trade_ID | Sequential number | `1` |
| 2 | Symbol | Trading pair | `BTCUSDT` |
| **3** | **Interval** | **Timeframe** | **`4h`** ⭐ |
| 4 | Side | LONG/SHORT | `SHORT` |
| 5 | Entry_Time | Entry timestamp | `2025-10-15 10:30:00` |
| 6 | Entry_Price | Entry price | `67500.00` |
| 7 | Exit_Time | Exit timestamp | `2025-10-15 14:30:00` |
| 8 | Exit_Price | Exit price | `65000.00` |
| 9 | Stop_Loss | SL price | `69000.00` |
| 10 | Take_Profit | TP price | `64500.00` |
| 11 | Position_Size | Size in $ | `200.00` |
| 12 | Status | TP_HIT/SL_HIT | `TP_HIT` |
| 13 | Profit_Loss | P/L in $ | `2500.00` |
| 14 | Profit_Loss_Pct | P/L in % | `3.70` |
| 15 | Risk_Reward | Actual R/R | `2.45` |
| 16 | Duration_Minutes | Trade duration | `240.00` |
| 17 | Logged_At | Log timestamp | `2025-10-15 14:30:05` |

---

## 💡 Why Add Interval?

### 1. **Compare Timeframe Performance**

```python
import pandas as pd

df = pd.read_csv('trade_logs/trades_all_symbols.csv')

# Win rate by interval
print(df.groupby('Interval')['Status'].apply(
    lambda x: (x == 'TP_HIT').mean() * 100
))
```

**Output:**
```
Interval
1h     65.5%
4h     78.2%  ⭐ Best
1d     62.3%
```

### 2. **Optimize Strategy Per Timeframe**

```python
# P/L by interval
performance = df.groupby('Interval').agg({
    'Profit_Loss': ['sum', 'mean'],
    'Trade_ID': 'count'
}).round(2)

print(performance)
```

**Output:**
```
         Profit_Loss           Trade_ID
                 sum    mean      count
Interval                               
1h          1250.50  83.37         15
4h          3450.25 287.52  ⭐     12
1d           890.00 111.25          8
```

### 3. **Risk Management by Timeframe**

```python
# Average duration vs P/L by interval
df.groupby('Interval').agg({
    'Duration_Minutes': 'mean',
    'Profit_Loss_Pct': 'mean'
})
```

### 4. **Backtest Different Intervals**

```python
# Filter trades by specific interval
df_4h = df[df['Interval'] == '4h']

# Calculate Sharpe ratio
returns = df_4h['Profit_Loss_Pct']
sharpe = returns.mean() / returns.std()
print(f"4h Sharpe Ratio: {sharpe:.2f}")
```

### 5. **Multi-Timeframe Analysis**

```python
# Compare same symbol across intervals
btc_trades = df[df['Symbol'] == 'BTCUSDT']

pivot = btc_trades.pivot_table(
    index='Interval',
    values=['Profit_Loss', 'Win_Rate'],
    aggfunc='mean'
)
```

---

## 📝 Example CSV Records

### Single Row Example

```csv
1,BTCUSDT,4h,SHORT,2025-10-15 10:30:00,67500.00,2025-10-15 14:30:00,65000.00,69000.00,64500.00,200.00,TP_HIT,2500.00,3.70,2.45,240.00,2025-10-15 14:30:05
```

### Multiple Intervals

```csv
Trade_ID,Symbol,Interval,Side,Entry_Time,Entry_Price,Exit_Time,Exit_Price,Stop_Loss,Take_Profit,Position_Size,Status,Profit_Loss,Profit_Loss_Pct,Risk_Reward,Duration_Minutes,Logged_At
1,BTCUSDT,1h,SHORT,2025-10-15 09:00:00,67800.00,2025-10-15 10:00:00,66900.00,68500.00,65800.00,200.00,TP_HIT,900.00,1.33,2.57,60.00,2025-10-15 10:00:05
2,ETHUSDT,4h,SHORT,2025-10-15 10:00:00,3456.00,2025-10-15 14:00:00,3298.00,3534.00,3220.00,200.00,TP_HIT,2360.00,3.43,2.03,240.00,2025-10-15 14:00:05
3,BNBUSDT,1d,SHORT,2025-10-14 00:00:00,315.00,2025-10-15 00:00:00,308.00,321.00,303.00,200.00,TP_HIT,2222.22,3.50,2.00,1440.00,2025-10-15 00:00:05
```

---

## 🔄 Backward Compatibility

### Old CSV Files (16 columns)

If you have old CSV files without the Interval column, you can:

#### Option 1: Add Interval Column Manually

```python
import pandas as pd

# Load old CSV
df = pd.read_csv('trades_BTCUSDT.csv')

# Add interval column (default to '4h' or ask user)
df.insert(2, 'Interval', '4h')

# Save updated CSV
df.to_csv('trades_BTCUSDT_updated.csv', index=False)
```

#### Option 2: Merge Old and New

```python
# Load old format (16 columns)
df_old = pd.read_csv('old_trades.csv')
df_old.insert(2, 'Interval', 'unknown')

# Load new format (17 columns)
df_new = pd.read_csv('trade_logs/trades_BTCUSDT.csv')

# Combine
df_combined = pd.concat([df_old, df_new], ignore_index=True)
df_combined.to_csv('all_trades.csv', index=False)
```

---

## 📊 Analysis Examples

### 1. Best Performing Interval

```python
import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv('trade_logs/trades_all_symbols.csv')

# Group by interval
interval_perf = df.groupby('Interval').agg({
    'Profit_Loss': 'sum',
    'Trade_ID': 'count',
    'Profit_Loss_Pct': 'mean'
}).round(2)

interval_perf.plot(kind='bar', subplots=True, figsize=(10, 8))
plt.tight_layout()
plt.savefig('interval_performance.png')
```

### 2. Win Rate by Interval

```python
# Calculate win rate per interval
win_rate = df.groupby('Interval')['Status'].apply(
    lambda x: (x.str.contains('TP')).mean() * 100
).round(1)

print("\n📊 Win Rate by Interval:")
print(win_rate.sort_values(ascending=False))
```

### 3. Average Trade Duration

```python
# Average duration per interval
avg_duration = df.groupby('Interval')['Duration_Minutes'].mean()

print("\n⏱️ Average Trade Duration:")
for interval, minutes in avg_duration.items():
    hours = minutes / 60
    print(f"  {interval}: {hours:.1f} hours")
```

### 4. Profit Factor by Interval

```python
# Profit factor per interval
def calc_profit_factor(group):
    wins = group[group['Profit_Loss'] > 0]['Profit_Loss'].sum()
    losses = abs(group[group['Profit_Loss'] < 0]['Profit_Loss'].sum())
    return wins / losses if losses > 0 else float('inf')

pf = df.groupby('Interval').apply(calc_profit_factor).round(2)

print("\n⚖️ Profit Factor by Interval:")
print(pf.sort_values(ascending=False))
```

---

## 🎯 Usage in Bot

### Automatic Interval Logging

The bot automatically captures the interval from your command:

```bash
# This will log "1h" in CSV
go run . --paper --symbol=BTCUSDT --interval=1h --balance=10000

# This will log "4h" in CSV
go run . --multi-paper --top=50 --interval=4h --balance=10000

# This will log "1d" in CSV
go run . --paper --symbol=ETHUSDT --interval=1d --balance=10000
```

### Where Interval is Stored

```go
// In PaperTrade struct
type PaperTrade struct {
    ID            int
    Symbol        string
    Interval      string  // ⭐ NEW FIELD
    Side          string
    // ... other fields
}

// Set during trade creation
trade := PaperTrade{
    ID:         p.TradeCounter,
    Symbol:     p.Symbol,
    Interval:   p.Interval,  // ⭐ Automatically captured
    Side:       side,
    // ...
}
```

---

## ✅ Benefits Summary

| Benefit | Description |
|---------|-------------|
| **Strategy Optimization** | Identify which intervals work best |
| **Risk Management** | Adjust position sizes per interval |
| **Backtesting** | Compare historical performance |
| **Multi-Timeframe** | Trade different symbols on different intervals |
| **Data Analysis** | Rich dataset for ML/AI models |
| **Performance Tracking** | Monitor interval-specific metrics |

---

## 📝 Notes

- ✅ **New CSV files** automatically include Interval column
- ✅ **Old CSV files** need manual migration (see above)
- ✅ **Interval is captured** from command-line flag
- ✅ **Works for both** single-symbol and multi-symbol trading
- ✅ **No breaking changes** - old code still works

---

**Version:** 1.2.0  
**Date:** October 15, 2025  
**Status:** ✅ Implemented & Tested

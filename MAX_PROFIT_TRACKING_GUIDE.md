# Maximum Profit Tracking & Timestamped CSV Logs

## 📊 Overview

This implementation tracks the **maximum profit reached during each trade** and calculates "give back" - the difference between the maximum profit and the final profit. This helps identify trades that reached profit targets but reversed before closing.

## 🎯 New Features

### 1. **Price Extremes Tracking**
Each trade now tracks:
- `HighestPrice`: The highest price reached during the trade
- `LowestPrice`: The lowest price reached during the trade

### 2. **Maximum Profit Tracking**
- `MaxProfit`: Maximum profit in dollars reached during the trade
- `MaxProfitPct`: Maximum profit percentage reached during the trade

### 3. **"Give Back" Calculation**
- Shows how much profit was "given back" before the trade closed
- Formula: `Give Back = Max Profit - Final Profit`
- Example: Trade reached +$50 but closed at +$20 → Give Back = $30

### 4. **Timestamped CSV Files**
Each bot execution creates a **new CSV file** with a timestamp:
- Single symbol: `trades_BTCUSDT_2025-10-15_14-30-45.csv`
- Multi-symbol: `trades_all_symbols_2025-10-15_14-30-45.csv`

## 📈 How It Works

### During Trade Execution

On **every candle**, the bot:
1. Updates `HighestPrice` if current price is higher
2. Updates `LowestPrice` if current price is lower
3. Calculates current profit/loss
4. Updates `MaxProfit` if current profit exceeds previous maximum

### When Trade Closes

The bot displays and logs:
```
📈 Highest Price: $45,250.00
📉 Lowest Price:  $44,800.00

💰 Final P/L: +$120.50 (+1.20%) ✅
🎯 Max Profit: +$250.00 (+2.50%)
⚠️  Give Back:  -$129.50 (-1.30%) 📉
```

## 📄 CSV Output Format

### New Columns Added:
```csv
Trade_ID,Symbol,Interval,Side,Entry_Time,Entry_Price,Exit_Time,Exit_Price,
Stop_Loss,Take_Profit,Position_Size,Status,Profit_Loss,Profit_Loss_Pct,
Risk_Reward,Highest_Price,Lowest_Price,Max_Profit,Max_Profit_Pct,
Give_Back,Give_Back_Pct,Duration_Minutes,Logged_At
```

### Example Row:
```csv
1,BTCUSDT,1m,SHORT,2025-10-15 14:30:00,45000.00,2025-10-15 14:35:00,44900.00,
45500.00,44500.00,100.00,CLOSED_TP,+120.50,+1.20,1.50,
45250.00,44800.00,250.00,2.50,129.50,1.30,5.00,2025-10-15 14:35:00
```

## 🔍 Key Insights from Data

### 1. **Identify "Give Back" Trades**
Trades with high `Give_Back` values indicate:
- Strategy needs tighter exits
- Trailing stop loss could help
- Price reversed significantly after reaching profit

### 2. **Price Range Analysis**
- `Highest_Price - Lowest_Price` shows trade volatility
- Helps understand price action during trade duration
- Useful for adjusting stop loss distances

### 3. **Maximum Profit Analysis**
- Compare `Max_Profit` vs `Profit_Loss` across all trades
- Calculate average give back percentage
- Identify if most profits are being "given back"

## 📊 Example Scenarios

### Scenario 1: Perfect Take Profit Hit
```
Entry: $45,000
Max Profit: +$250 (+2.50%)
Final P/L: +$250 (+2.50%)
Give Back: $0 (0%)
✅ Take profit hit at maximum - perfect exit!
```

### Scenario 2: Significant Give Back
```
Entry: $45,000
Max Profit: +$500 (+5.00%)
Final P/L: +$50 (+0.50%)
Give Back: $450 (4.50%)
⚠️  Reached +5% but gave back 90% of profit!
```

### Scenario 3: Stop Loss Hit After Profit
```
Entry: $45,000
Max Profit: +$300 (+3.00%)
Final P/L: -$100 (-1.00%)
Give Back: $400 (4.00%)
❌ Was profitable but reversed to stop loss!
```

## 🚀 Running the Bot

### Single Symbol Trading
```bash
./bot --symbol BTCUSDT --interval 1m --live
```

Creates: `trade_logs/trades_BTCUSDT_2025-10-15_22-30-45.csv`

### Multi-Symbol Trading
```bash
./bot --multi --top 10 --interval 1m --live
```

Creates: `trade_logs/trades_all_symbols_2025-10-15_22-30-45.csv`

## 📁 File Organization

```
trade_logs/
├── trades_BTCUSDT_2025-10-15_14-30-00.csv
├── trades_BTCUSDT_2025-10-15_15-45-30.csv
├── trades_all_symbols_2025-10-15_14-30-00.csv
└── trades_all_symbols_2025-10-15_16-00-15.csv
```

Each execution creates a **separate file**, preserving all historical data.

## 🎯 Console Output

### Verbose Mode (`-v`)
```
╔════════════════════════════════════════╗
║      PAPER TRADE CLOSED                ║
╚════════════════════════════════════════╝

📝 Trade #5: SHORT BTCUSDT
📍 Entry:  $45,000.00 → Exit: $44,900.00
📊 Reason: TAKE_PROFIT
⏱️  Duration: 5m0s

📈 Highest Price: $45,250.00
📉 Lowest Price:  $44,800.00

💰 Final P/L: +$120.50 (+1.20%) ✅
🎯 Max Profit: +$250.00 (+2.50%)
⚠️  Give Back:  -$129.50 (-1.30%) 📉

💵 Balance: $1,000.00 → $1,120.50
════════════════════════════════════════
```

### Quiet Mode (default)
```
✅ [BTCUSDT] SHORT CLOSED @ $44,900.00 | TAKE_PROFIT | P/L: +$120.50 (+1.20%)
   ⚠️  Max Profit: +$250.00 | Give Back: -$129.50
```

## 📈 Analyzing Give Back Patterns

### Using CSV Data
```python
import pandas as pd

df = pd.read_csv('trade_logs/trades_all_symbols_2025-10-15_14-30-00.csv')

# Calculate average give back
avg_giveback = df['Give_Back'].mean()
avg_giveback_pct = df['Give_Back_Pct'].mean()

print(f"Average Give Back: ${avg_giveback:.2f} ({avg_giveback_pct:.2f}%)")

# Find trades with highest give back
high_giveback = df[df['Give_Back'] > 100].sort_values('Give_Back', ascending=False)
print(high_giveback[['Symbol', 'Max_Profit', 'Profit_Loss', 'Give_Back']])

# Calculate "efficiency ratio" (how much of max profit was captured)
df['Efficiency'] = (df['Profit_Loss'] / df['Max_Profit']) * 100
print(f"Average Efficiency: {df['Efficiency'].mean():.2f}%")
```

## 🛠️ Next Steps

Based on give back analysis, consider implementing:

1. **Trailing Stop Loss** - Lock in profits as price moves favorably
2. **Partial Exits** - Take 50% profit at +2%, let 50% run to target
3. **Time-Based Exits** - Close positions that stall after reaching profit
4. **Dynamic Take Profits** - Adjust targets based on volatility

## 🎓 Understanding the Data

### High Give Back Indicates:
- Market reversals after reaching profit
- Take profit targets might be too aggressive
- Need for profit protection mechanisms
- Potential for trailing stop implementation

### Low Give Back Indicates:
- Clean exits at or near maximum profit
- Well-placed take profit levels
- Good market timing
- Efficient strategy execution

## ✅ Summary

This implementation provides:
- ✅ Complete price tracking during trades
- ✅ Maximum profit monitoring
- ✅ "Give back" calculation and display
- ✅ Timestamped CSV files per execution
- ✅ Enhanced CSV columns with new metrics
- ✅ Detailed console output (verbose/quiet modes)
- ✅ Historical data preservation

Every trade now has complete lifecycle tracking from entry to exit!

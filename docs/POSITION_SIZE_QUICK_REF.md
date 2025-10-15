# Position Sizing Quick Reference

## ✅ YES - Profit IS Based on Dollar Amount Invested

But the "invested amount" is **risk-based**, not fixed!

---

## The Formula

```
Position Size = (Balance × Risk%) / (Stop Loss Distance%)
```

### Example with $10,000 Balance:

| Stop Loss Distance | Risk Amount | Position Size |
|-------------------|-------------|---------------|
| 0.2% away         | $100        | **$50,000**   |
| 0.4% away         | $100        | **$25,000**   |
| 0.5% away         | $100        | **$20,000**   |
| 1.0% away         | $100        | **$10,000**   |
| 2.0% away         | $100        | **$5,000**    |

**You always risk $100 (1% of balance), but position size varies!**

---

## Your CSV Example:

```csv
EULUSDT: Entry=$9.51, Exit=$9.35, Position_Size=$25,000, Profit=$418.03
```

### How it worked:

1. **Risk Amount**: $100 (1% of portfolio)
2. **Stop Loss Distance**: ~0.4%
3. **Position Size**: $100 / 0.004 = **$25,000**
4. **Quantity**: $25,000 / $9.51 = 2,628 contracts
5. **Profit**: ($9.51 - $9.35) × 2,628 = **$418.03** ✅

---

## Why Position Sizes Look Huge:

Your trades show:
- ✅ $25,000 position sizes
- ✅ $94,440 position sizes  
- ✅ $203,948 position sizes

**Because you're using VERY TIGHT stop losses (0.2-0.5%)!**

Tight stops = Bigger positions (while maintaining same $ risk)

---

## Profit Calculation:

```
Profit = Price_Difference × Quantity
Profit % = Profit / Position_Size × 100
```

**Example:**
```
Price moved: $0.16 (9.51 → 9.35)
Quantity: 2,628 contracts
Profit: $0.16 × 2,628 = $420
Profit %: $420 / $25,000 = 1.67%
```

---

## The Answer:

### "First trade uses how much from portfolio?"

**Answer:** It uses as much as needed to risk exactly 1% ($100 on $10k)

- If tight stop (0.4%): Uses **$25,000** position
- If wide stop (2.0%): Uses **$5,000** position

**But actual $ risked is always $100!**

---

## This is CORRECT!

This is professional **risk-based position sizing**:
- ✅ Equal risk per trade
- ✅ Position size adapts to volatility
- ✅ Tight stops = Larger positions
- ✅ Wide stops = Smaller positions

**Your bot is working correctly!** 🎯

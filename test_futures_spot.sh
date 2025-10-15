#!/bin/bash

# Test script to demonstrate Spot vs Futures market functionality
# This script shows how to use the --futures flag

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  Binance Spot vs Futures Market Test Script               ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Test 1: Spot Market (Default)
echo "📊 Test 1: Fetching top 5 symbols from SPOT market..."
echo "Command: ./bot --multi --top 5 --interval 1m"
echo "Expected: Should show 'Market Type: SPOT'"
echo ""
read -p "Press Enter to run test 1..."
./bot --multi --top 5 --interval 1m 2>&1 | head -20
echo ""
echo "─────────────────────────────────────────────────────────────"
echo ""

# Test 2: Futures Market
echo "📊 Test 2: Fetching top 5 symbols from FUTURES market..."
echo "Command: ./bot --multi --top 5 --interval 1m --futures"
echo "Expected: Should show 'Market Type: FUTURES'"
echo ""
read -p "Press Enter to run test 2..."
./bot --multi --top 5 --interval 1m --futures 2>&1 | head -20
echo ""
echo "─────────────────────────────────────────────────────────────"
echo ""

# Test 3: Compare symbol lists
echo "📊 Test 3: Getting symbol counts from both markets..."
echo ""
echo "Fetching SPOT symbols..."
SPOT_COUNT=$(./bot --multi --all 2>&1 | grep "Found" | grep -oE '[0-9]+' | head -1)
echo "✅ SPOT market has $SPOT_COUNT USDT pairs"
echo ""
echo "Fetching FUTURES symbols..."
FUTURES_COUNT=$(./bot --multi --all --futures 2>&1 | grep "Found" | grep -oE '[0-9]+' | head -1)
echo "✅ FUTURES market has $FUTURES_COUNT USDT pairs"
echo ""
echo "─────────────────────────────────────────────────────────────"
echo ""

echo "✅ All tests completed!"
echo ""
echo "Key Differences:"
echo "  • SPOT: Buy/sell actual cryptocurrencies"
echo "  • FUTURES: Trade perpetual contracts (derivatives)"
echo ""
echo "Usage Examples:"
echo "  # Single symbol on spot"
echo "  ./bot --symbol BTCUSDT --interval 1m"
echo ""
echo "  # Single symbol on futures"
echo "  ./bot --symbol BTCUSDT --interval 1m --futures"
echo ""
echo "  # Multi-symbol paper trading on futures"
echo "  ./bot --multi-paper --top 50 --futures --quiet"
echo ""

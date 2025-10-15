#!/bin/bash

# Test script to demonstrate the interactive 'c' command with market type display

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  Testing Interactive 'c' Command - Market Type Display    ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "This test will show how the 'c' command displays market type"
echo "and base URL configuration in the bot."
echo ""
echo "─────────────────────────────────────────────────────────────"
echo ""

# Test 1: SPOT market
echo "📊 Test 1: Running bot in SPOT market mode"
echo "Command: echo 'c' | ./bot --symbol BTCUSDT --interval 1m --limit 100"
echo ""
echo "Starting bot... (type 'c' to see config, 'q' to quit)"
echo ""
(sleep 2; echo "c"; sleep 2; echo "q") | ./bot --symbol BTCUSDT --interval 1m --limit 100 2>&1 | grep -A 50 "BOT CONFIGURATION"

echo ""
echo "─────────────────────────────────────────────────────────────"
echo ""

# Test 2: FUTURES market
echo "📊 Test 2: Running bot in FUTURES market mode"
echo "Command: echo 'c' | ./bot --symbol BTCUSDT --interval 1m --limit 100 --futures"
echo ""
echo "Starting bot... (type 'c' to see config, 'q' to quit)"
echo ""
(sleep 2; echo "c"; sleep 2; echo "q") | ./bot --symbol BTCUSDT --interval 1m --limit 100 --futures 2>&1 | grep -A 50 "BOT CONFIGURATION"

echo ""
echo "─────────────────────────────────────────────────────────────"
echo ""
echo "✅ Test completed!"
echo ""
echo "Key Points:"
echo "  • Press 'c' + Enter during bot execution to see configuration"
echo "  • Market Type shows: SPOT or FUTURES"
echo "  • Base URL shows the Binance API endpoint being used"
echo "  • Endpoints show the specific API paths for data fetching"
echo ""
echo "Available interactive commands:"
echo "  c, config  - Show bot configuration (including market type)"
echo "  s, status  - Show current status & portfolio"
echo "  h, help    - Show help message"
echo "  q, quit    - Quit bot gracefully"
echo ""

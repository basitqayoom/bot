#!/bin/bash

# Test Script for Interactive Configuration Feature
# This demonstrates all interactive commands

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║          INTERACTIVE COMMANDS TEST SCRIPT                       ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo "This script will help you test the interactive commands feature."
echo ""

# Function to display menu
show_menu() {
    echo "════════════════════════════════════════════════════════════════"
    echo "Choose a test mode:"
    echo "════════════════════════════════════════════════════════════════"
    echo "1. Single-Symbol Paper Trading (BTCUSDT)"
    echo "2. Multi-Symbol Paper Trading (Top 10 coins)"
    echo "3. Show what commands are available"
    echo "4. Exit"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
}

# Function to show commands
show_commands() {
    echo ""
    echo "════════════════════════════════════════════════════════════════"
    echo "📖 INTERACTIVE COMMANDS AVAILABLE WHILE BOT IS RUNNING"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    echo "Type these commands and press ENTER while the bot is running:"
    echo ""
    echo "  c  or  config   - Show full bot configuration"
    echo "  s  or  status   - Show current status and time"
    echo "  h  or  help     - Show help message"
    echo "  q  or  quit     - Exit bot gracefully"
    echo ""
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    echo "Example workflow:"
    echo "  1. Start the bot (choose option 1 or 2)"
    echo "  2. Wait for it to start running"
    echo "  3. Type 'c' and press ENTER to see configuration"
    echo "  4. Type 's' and press ENTER to see status"
    echo "  5. Type 'q' and press ENTER to exit cleanly"
    echo ""
    echo "════════════════════════════════════════════════════════════════"
    echo ""
}

# Main loop
while true; do
    show_menu
    read -p "Enter your choice (1-4): " choice
    echo ""
    
    case $choice in
        1)
            echo "🚀 Starting Single-Symbol Paper Trading..."
            echo ""
            echo "Configuration will be displayed automatically."
            echo "You can type 'c' anytime to see it again."
            echo ""
            echo "Press Ctrl+C to stop this test if needed."
            echo ""
            sleep 2
            go run . --paper --symbol=BTCUSDT --interval=4h --balance=10000
            ;;
        2)
            echo "🚀 Starting Multi-Symbol Paper Trading..."
            echo ""
            echo "Multi-symbol configuration will be displayed automatically."
            echo "You can type 'c' anytime to see it again."
            echo ""
            echo "Press Ctrl+C to stop this test if needed."
            echo ""
            sleep 2
            go run . --multi-paper --top=10 --interval=4h --balance=10000 --max-pos=3
            ;;
        3)
            show_commands
            ;;
        4)
            echo "👋 Exiting test script. Goodbye!"
            echo ""
            exit 0
            ;;
        *)
            echo "❌ Invalid choice. Please enter 1, 2, 3, or 4."
            echo ""
            ;;
    esac
done

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// ==================== INTERACTIVE CONFIGURATION DISPLAY ====================

// PrintBotConfig displays current bot configuration
func PrintBotConfig(symbol, interval string, balance float64, mode string) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              BOT CONFIGURATION                             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("🤖 Mode:              %s\n", mode)
	fmt.Printf("📊 Symbol:            %s\n", symbol)
	fmt.Printf("⏰ Interval:          %s\n", interval)
	fmt.Printf("💰 Balance:           $%.2f\n", balance)
	fmt.Printf("🔴 Live Mode:         %v\n", ENABLE_LIVE_MODE)
	fmt.Printf("⏳ Wait for Close:    %v\n", WAIT_FOR_CANDLE_CLOSE)
	fmt.Printf("⚡ Parallel Mode:     %v\n", ENABLE_PARALLEL_MODE)
	fmt.Printf("👷 Workers:           %d\n", NUM_WORKERS)
	fmt.Printf("📝 Verbose:           %v\n", VERBOSE_MODE)

	fmt.Println("\n📈 STRATEGY PARAMETERS:")
	fmt.Printf("   RSI Period:        %d\n", RSI_PERIOD)
	fmt.Printf("   ATR Length:        %d\n", ATR_LENGTH)
	fmt.Printf("   Min Divergences:   %d\n", MIN_DIVERGENCES_FOR_SIGNAL)
	fmt.Printf("   Swing Lookback:    %d\n", SWING_LOOKBACK)

	fmt.Println("\n🎯 S/R ZONE SETTINGS:")
	fmt.Printf("   Pivot Left:        %d\n", PIVOT_LEFT_LOOKBACK)
	fmt.Printf("   Pivot Right:       %d\n", PIVOT_RIGHT_LOOKBACK)
	fmt.Printf("   ATR Multiplier:    %.1f\n", ATR_MULTIPLIER)
	fmt.Printf("   Max Zone %%:        %.1f%%\n", MAX_ZONE_PERCENT)
	fmt.Printf("   Align Zones:       %v\n", ALIGN_ZONES)

	fmt.Println("\n💵 RISK MANAGEMENT:")
	fmt.Printf("   Risk/Reward:       %.1f:1\n", RISK_REWARD_RATIO)
	fmt.Printf("   Max Risk:          %.1f%%\n", MAX_RISK_PERCENT)
	fmt.Printf("   Stop Loss:         %.1f%%\n", STOP_LOSS_PERCENT)
	fmt.Printf("   Take Profit:       %.1f%%\n", TAKE_PROFIT_PERCENT)

	fmt.Println("\n🌍 TIMEZONE:")
	fmt.Printf("   Current Time (IST): %s\n", getIST().Format("2006-01-02 15:04:05"))
	fmt.Printf("   Current Time (UTC): %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Printf("   Offset:            UTC+5:30\n")

	fmt.Println("════════════════════════════════════════════════════════════")
}

// PrintMultiSymbolConfig displays multi-symbol configuration
func PrintMultiSymbolConfig(symbols []string, interval string, balance float64, maxPositions int, mode string) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         MULTI-SYMBOL BOT CONFIGURATION                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("🤖 Mode:              %s\n", mode)
	fmt.Printf("📊 Symbols:           %d pairs\n", len(symbols))
	fmt.Printf("⏰ Interval:          %s\n", interval)
	fmt.Printf("💰 Starting Balance:  $%.2f\n", balance)
	fmt.Printf("🎯 Max Positions:     %d\n", maxPositions)
	fmt.Printf("🔴 Live Mode:         %v\n", ENABLE_LIVE_MODE)
	fmt.Printf("⏳ Wait for Close:    %v\n", WAIT_FOR_CANDLE_CLOSE)
	fmt.Printf("⚡ Parallel Mode:     %v\n", ENABLE_PARALLEL_MODE)
	fmt.Printf("👷 Workers:           %d\n", NUM_WORKERS)

	fmt.Println("\n📋 SYMBOLS LIST (First 10):")
	displayCount := 10
	if len(symbols) < 10 {
		displayCount = len(symbols)
	}
	for i := 0; i < displayCount; i++ {
		fmt.Printf("   %d. %s\n", i+1, symbols[i])
	}
	if len(symbols) > 10 {
		fmt.Printf("   ... and %d more\n", len(symbols)-10)
	}

	fmt.Println("\n📈 STRATEGY PARAMETERS:")
	fmt.Printf("   RSI Period:        %d\n", RSI_PERIOD)
	fmt.Printf("   ATR Length:        %d\n", ATR_LENGTH)
	fmt.Printf("   Min Divergences:   %d\n", MIN_DIVERGENCES_FOR_SIGNAL)

	fmt.Println("\n🎯 S/R ZONE SETTINGS:")
	fmt.Printf("   Pivot Left:        %d\n", PIVOT_LEFT_LOOKBACK)
	fmt.Printf("   Pivot Right:       %d\n", PIVOT_RIGHT_LOOKBACK)
	fmt.Printf("   ATR Multiplier:    %.1f\n", ATR_MULTIPLIER)

	fmt.Println("\n💵 RISK MANAGEMENT:")
	fmt.Printf("   Risk/Reward:       %.1f:1\n", RISK_REWARD_RATIO)
	fmt.Printf("   Max Risk:          %.1f%% per trade\n", MAX_RISK_PERCENT)
	fmt.Printf("   Stop Loss:         %.1f%%\n", STOP_LOSS_PERCENT)
	fmt.Printf("   Take Profit:       %.1f%%\n", TAKE_PROFIT_PERCENT)

	fmt.Println("\n🌍 TIMEZONE:")
	fmt.Printf("   Current Time (IST): %s\n", getIST().Format("2006-01-02 15:04:05"))
	fmt.Printf("   Current Time (UTC): %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))

	fmt.Println("════════════════════════════════════════════════════════════")
}

// StartInteractiveMode listens for keyboard commands during runtime
// statusCallback is optional - if provided, it will be called when 's' is pressed
func StartInteractiveMode(configCallback func(), statusCallback ...func()) {
	fmt.Println("\n💡 INTERACTIVE MODE ENABLED")
	fmt.Println("   Type 'c' + Enter to show config")
	fmt.Println("   Type 's' + Enter to show status & portfolio")
	fmt.Println("   Type 'h' + Enter to show help")
	fmt.Println("   Type 'q' + Enter to quit")
	fmt.Println()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.ToLower(strings.TrimSpace(scanner.Text()))

			switch input {
			case "c", "config":
				configCallback()

			case "s", "status":
				fmt.Println("\n════════════════════════════════════════════════════════════")
				fmt.Println("⚡ BOT STATUS: RUNNING ✅")
				fmt.Printf("🕐 Current Time (IST): %s\n", getIST().Format("2006-01-02 15:04:05"))
				fmt.Printf("🕐 Current Time (UTC): %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
				fmt.Println("════════════════════════════════════════════════════════════")

				// Call status callback if provided (for portfolio display)
				if len(statusCallback) > 0 && statusCallback[0] != nil {
					statusCallback[0]()
				}
				fmt.Println()

			case "q", "quit", "exit":
				fmt.Println("\n👋 Shutting down bot...")
				os.Exit(0)

			case "h", "help":
				fmt.Println("\n📖 AVAILABLE COMMANDS:")
				fmt.Println("   c, config  - Show bot configuration")
				fmt.Println("   s, status  - Show current status & portfolio")
				fmt.Println("   h, help    - Show this help message")
				fmt.Println("   q, quit    - Quit bot gracefully")
				fmt.Println()

			default:
				if input != "" {
					fmt.Printf("❓ Unknown command: '%s' (type 'h' for help)\n", input)
				}
			}
		}
	}()
}

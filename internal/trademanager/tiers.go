package trademanager

import (
	"fmt"
)

// TierManager handles the 3-Tier trade management logic
type TierManager struct {
	config *Config
}

// NewTierManager creates a new tier manager with the given configuration
func NewTierManager(config *Config) *TierManager {
	if config == nil {
		config = DefaultConfig()
	}
	return &TierManager{
		config: config,
	}
}

// TierAction represents an action that should be taken
type TierAction struct {
	Type          string  // "MOVE_STOP", "PARTIAL_EXIT", "NONE"
	NewStopLoss   float64 // New stop loss price (for MOVE_STOP)
	ExitPercent   float64 // Percentage to exit (for PARTIAL_EXIT)
	Reason        string  // Human-readable reason
	TierActivated int     // Which tier triggered (1, 2, or 3)
}

// EvaluatePosition checks all tiers and returns the action to take
func (tm *TierManager) EvaluatePosition(pos *ManagedPosition) *TierAction {
	if !tm.config.Enabled {
		return &TierAction{Type: "NONE"}
	}

	// Check tiers in order: 1 -> 2 -> 3

	// Tier 1: Breakeven Lock
	if !pos.Tier1Activated {
		if action := tm.checkTier1(pos); action != nil {
			return action
		}
	}

	// Tier 2: Partial Exit (only if Tier 1 is already active)
	if pos.Tier1Activated && !pos.Tier2Activated {
		if action := tm.checkTier2(pos); action != nil {
			return action
		}
	}

	// Tier 3: Time-Based Lock (only if Tier 1 is active)
	if pos.Tier1Activated && !pos.Tier3Activated {
		if action := tm.checkTier3(pos); action != nil {
			return action
		}
	}

	// Tier 3: Continuous trailing if already activated
	if pos.Tier3Activated {
		if action := tm.updateTier3(pos); action != nil {
			return action
		}
	}

	return &TierAction{Type: "NONE"}
}

// checkTier1 evaluates Tier 1: Breakeven Lock
func (tm *TierManager) checkTier1(pos *ManagedPosition) *TierAction {
	profitPct := pos.GetCurrentProfitPct()

	if profitPct >= tm.config.Tier1BreakevenThreshold {
		return &TierAction{
			Type:          "MOVE_STOP",
			NewStopLoss:   pos.EntryPrice,
			Reason:        fmt.Sprintf("🔒 Tier 1: Breakeven Lock at +%.2f%%", profitPct),
			TierActivated: 1,
		}
	}

	return nil
}

// checkTier2 evaluates Tier 2: Partial Exit
func (tm *TierManager) checkTier2(pos *ManagedPosition) *TierAction {
	profitPct := pos.GetCurrentProfitPct()

	if profitPct >= tm.config.Tier2PartialExitThreshold {
		return &TierAction{
			Type:          "PARTIAL_EXIT",
			ExitPercent:   tm.config.Tier2PartialExitPercent,
			NewStopLoss:   pos.EntryPrice, // Keep at breakeven after partial exit
			Reason:        fmt.Sprintf("💰 Tier 2: Partial Exit %.0f%% at +%.2f%%", tm.config.Tier2PartialExitPercent, profitPct),
			TierActivated: 2,
		}
	}

	return nil
}

// checkTier3 evaluates Tier 3: Time-Based Lock (initial activation)
func (tm *TierManager) checkTier3(pos *ManagedPosition) *TierAction {
	profitPct := pos.GetCurrentProfitPct()
	timeInProfit := pos.TimeInProfit

	// Check if conditions are met for Tier 3 activation
	if profitPct >= tm.config.Tier3MinProfitThreshold &&
		timeInProfit >= float64(tm.config.Tier3TimeThreshold) {

		// Calculate lock price based on max profit reached
		lockPrice := tm.calculateTier3LockPrice(pos)

		return &TierAction{
			Type:        "MOVE_STOP",
			NewStopLoss: lockPrice,
			Reason: fmt.Sprintf("⏰ Tier 3: Time Lock (%.0fs in profit, locking %.0f%% of max %.2f%%)",
				timeInProfit,
				tm.config.Tier3ProfitLockPercent,
				pos.MaxProfitPct),
			TierActivated: 3,
		}
	}

	return nil
}

// updateTier3 continuously updates the Tier 3 trailing stop
func (tm *TierManager) updateTier3(pos *ManagedPosition) *TierAction {
	// Calculate new lock price based on current max profit
	newLockPrice := tm.calculateTier3LockPrice(pos)

	// Only tighten the stop, never widen it
	if (pos.Side == "SHORT" && newLockPrice < pos.StopLoss) ||
		(pos.Side == "LONG" && newLockPrice > pos.StopLoss) {

		return &TierAction{
			Type:        "MOVE_STOP",
			NewStopLoss: newLockPrice,
			Reason: fmt.Sprintf("⏰ Tier 3: Trail Update (locking %.0f%% of max %.2f%%)",
				tm.config.Tier3ProfitLockPercent,
				pos.MaxProfitPct),
			TierActivated: 3,
		}
	}

	return nil
}

// calculateTier3LockPrice calculates the stop loss price that locks in X% of max profit
func (tm *TierManager) calculateTier3LockPrice(pos *ManagedPosition) float64 {
	// We want to lock in X% of the max profit reached
	lockPercent := tm.config.Tier3ProfitLockPercent / 100.0

	if pos.Side == "SHORT" {
		// For SHORT: entry - (entry - lowest) * lockPercent
		maxMove := pos.EntryPrice - pos.LowestPrice
		lockMove := maxMove * lockPercent
		return pos.EntryPrice - lockMove
	} else {
		// For LONG: entry + (highest - entry) * lockPercent
		maxMove := pos.HighestPrice - pos.EntryPrice
		lockMove := maxMove * lockPercent
		return pos.EntryPrice + lockMove
	}
}

// GetConfig returns the current configuration
func (tm *TierManager) GetConfig() *Config {
	return tm.config
}

// SetConfig updates the configuration
func (tm *TierManager) SetConfig(config *Config) {
	tm.config = config
}

// PrintStatus prints the current status of a position with tier information
func (tm *TierManager) PrintStatus(pos *ManagedPosition) {
	profit, profitPct := pos.CalculateCurrentProfit()
	duration := pos.GetDuration()
	timeInProfit := pos.GetTimeInProfitDuration()

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Printf("║  📊 Position Status - %s         \n", pos.Symbol)
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("💵 Current P/L:    $%.2f (%.2f%%)\n", profit, profitPct)
	fmt.Printf("📈 Max Profit:     $%.2f (%.2f%%)\n", pos.MaxProfit, pos.MaxProfitPct)
	fmt.Printf("⏱️  Duration:       %.1f minutes\n", duration.Minutes())
	fmt.Printf("⏱️  Time in Profit: %.1f seconds\n", timeInProfit.Seconds())
	fmt.Println("\n🎯 Tier Status:")
	fmt.Printf("  Tier 1 (Breakeven): %s\n", tm.getTierStatus(pos.Tier1Activated))
	fmt.Printf("  Tier 2 (Partial):   %s", tm.getTierStatus(pos.Tier2Activated))
	if pos.Tier2Activated {
		fmt.Printf(" - Exited $%.2f (%.0f%%)", pos.Tier2ExitedProfit, (pos.Tier2ExitedSize/pos.OriginalSize)*100)
	}
	fmt.Println()
	fmt.Printf("  Tier 3 (Time Lock): %s\n", tm.getTierStatus(pos.Tier3Activated))
	fmt.Println("════════════════════════════════════════")
}

func (tm *TierManager) getTierStatus(activated bool) string {
	if activated {
		return "✅ Active"
	}
	return "⏳ Pending"
}

package trademanager

// Config holds the 3-Tier trade management configuration
type Config struct {
	// Tier 1: Breakeven Lock
	Tier1BreakevenThreshold float64 // % profit to trigger breakeven (default: 0.5)

	// Tier 2: Partial Exit
	Tier2PartialExitThreshold float64 // % profit to trigger partial exit (default: 1.5)
	Tier2PartialExitPercent   float64 // % of position to close (default: 50)

	// Tier 3: Time-Based Lock
	Tier3TimeThreshold      int     // Seconds in profit before tightening (default: 300 = 5 min)
	Tier3MinProfitThreshold float64 // Minimum profit % to activate time-based (default: 1.0)
	Tier3ProfitLockPercent  float64 // % of max profit to lock (default: 60)

	// General settings
	Enabled bool // Master switch to enable/disable 3-Tier system
}

// DefaultConfig returns the recommended default configuration
// OPTIMIZED FOR 0.4% SL / 0.8% TP SCALPING STRATEGY (1m timeframe)
func DefaultConfig() *Config {
	return &Config{
		Tier1BreakevenThreshold:   0.3,  // Move to breakeven at +0.3% profit (before 0.8% TP)
		Tier2PartialExitThreshold: 0.6,  // Take 50% off at +0.6% profit (BEFORE 0.8% TP!)
		Tier2PartialExitPercent:   50.0, // Close 50% of position
		Tier3TimeThreshold:        180,  // 3 minutes (appropriate for 1m scalping)
		Tier3MinProfitThreshold:   0.4,  // Must be at least +0.4% profit (matches SL)
		Tier3ProfitLockPercent:    60.0, // Lock 60% of max profit reached
		Enabled:                   true,
	}
}

// AggressiveConfig returns a more aggressive profit-locking configuration
func AggressiveConfig() *Config {
	return &Config{
		Tier1BreakevenThreshold:   0.3,  // Faster breakeven
		Tier2PartialExitThreshold: 1.0,  // Earlier partial exit
		Tier2PartialExitPercent:   60.0, // Take more off table
		Tier3TimeThreshold:        180,  // 3 minutes
		Tier3MinProfitThreshold:   0.7,
		Tier3ProfitLockPercent:    70.0, // Lock more profit
		Enabled:                   true,
	}
}

// ConservativeConfig returns a conservative configuration (lets winners run more)
func ConservativeConfig() *Config {
	return &Config{
		Tier1BreakevenThreshold:   0.7,  // Slower breakeven
		Tier2PartialExitThreshold: 2.0,  // Later partial exit
		Tier2PartialExitPercent:   40.0, // Take less off table
		Tier3TimeThreshold:        420,  // 7 minutes
		Tier3MinProfitThreshold:   1.5,
		Tier3ProfitLockPercent:    50.0, // Lock less (more room to run)
		Enabled:                   true,
	}
}

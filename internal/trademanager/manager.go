package trademanager

import (
	"fmt"
	"sync"
)

// Manager is the main trade management system that coordinates 3-Tier logic
type Manager struct {
	config          *Config
	tierManager     *TierManager
	positions       map[string]*ManagedPosition // symbol -> position
	mutex           sync.RWMutex
	verbose         bool
	partialExitCb   PartialExitCallback   // Callback for partial exits
	stopUpdateCb    StopUpdateCallback    // Callback for stop loss updates
	positionCloseCb PositionCloseCallback // Callback for position close
}

// Callbacks for integration with existing trading engine
type PartialExitCallback func(symbol string, exitPercent, currentPrice float64) (exitedProfit float64, err error)
type StopUpdateCallback func(symbol string, newStopLoss float64) error
type PositionCloseCallback func(symbol string, reason string) error

// NewManager creates a new trade manager
func NewManager(config *Config, verbose bool) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		config:      config,
		tierManager: NewTierManager(config),
		positions:   make(map[string]*ManagedPosition),
		verbose:     verbose,
	}
}

// SetCallbacks configures the callbacks for integration
func (m *Manager) SetCallbacks(
	partialExit PartialExitCallback,
	stopUpdate StopUpdateCallback,
	positionClose PositionCloseCallback,
) {
	m.partialExitCb = partialExit
	m.stopUpdateCb = stopUpdate
	m.positionCloseCb = positionClose
}

// AddPosition adds a new position to be managed
func (m *Manager) AddPosition(id int, symbol, side string, entryPrice, stopLoss, takeProfit, size float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pos := NewManagedPosition(id, symbol, side, entryPrice, stopLoss, takeProfit, size)
	m.positions[symbol] = pos

	if m.verbose {
		fmt.Printf("\n✅ Trade Manager: Added position %s (ID: %d)\n", symbol, id)
		fmt.Printf("   Entry: $%.2f | SL: $%.2f | TP: $%.2f\n", entryPrice, stopLoss, takeProfit)
		fmt.Printf("   3-Tier Protection: %s\n", m.getTierSummary())
	}
}

// AddPositionWithAdaptiveConfig adds a position with thresholds adapted to actual SL/TP distances
func (m *Manager) AddPositionWithAdaptiveConfig(id int, symbol, side string, entryPrice, stopLoss, takeProfit, size float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Calculate actual SL distance as percentage
	var slDistancePct float64
	if side == "SHORT" {
		slDistancePct = ((stopLoss - entryPrice) / entryPrice) * 100
	} else { // LONG
		slDistancePct = ((entryPrice - stopLoss) / entryPrice) * 100
	}

	// Adapt tier thresholds to the actual SL distance
	// Strategy: Tier 1 at 40% of SL distance, Tier 2 at 70% of SL distance
	adaptedConfig := &Config{
		Tier1BreakevenThreshold:   slDistancePct * 0.4,  // 40% to SL
		Tier2PartialExitThreshold: slDistancePct * 0.7,  // 70% to SL (before SL hits)
		Tier2PartialExitPercent:   m.config.Tier2PartialExitPercent,
		Tier3TimeThreshold:        m.config.Tier3TimeThreshold,
		Tier3MinProfitThreshold:   slDistancePct * 0.3,  // 30% to SL
		Tier3ProfitLockPercent:    m.config.Tier3ProfitLockPercent,
		Enabled:                   true,
	}

	// Create position with adapted config
	pos := NewManagedPosition(id, symbol, side, entryPrice, stopLoss, takeProfit, size)
	m.positions[symbol] = pos

	// Update tier manager with adapted config for this position
	m.tierManager.config = adaptedConfig

	if m.verbose {
		fmt.Printf("\n✅ Trade Manager: Added position %s (ID: %d) [ADAPTIVE MODE]\n", symbol, id)
		fmt.Printf("   Entry: $%.2f | SL: $%.2f (%.2f%%) | TP: $%.2f\n", 
			entryPrice, stopLoss, slDistancePct, takeProfit)
		fmt.Printf("   🔧 Adapted Tiers: %.2f%% BE | %.2f%% Partial | %ds Trailing\n",
			adaptedConfig.Tier1BreakevenThreshold,
			adaptedConfig.Tier2PartialExitThreshold,
			adaptedConfig.Tier3TimeThreshold)
	}
}

// UpdatePrice updates the price for a position and evaluates 3-Tier rules
func (m *Manager) UpdatePrice(symbol string, currentPrice float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pos, exists := m.positions[symbol]
	if !exists {
		return fmt.Errorf("no active position for %s", symbol)
	}

	// Update position price and metrics
	pos.UpdatePrice(currentPrice)

	// Evaluate 3-Tier rules
	action := m.tierManager.EvaluatePosition(pos)

	// Execute the action
	if action.Type != "NONE" {
		return m.executeAction(pos, action)
	}

	return nil
}

// executeAction performs the action returned by tier evaluation
func (m *Manager) executeAction(pos *ManagedPosition, action *TierAction) error {
	switch action.Type {
	case "MOVE_STOP":
		return m.executeMoveStop(pos, action)

	case "PARTIAL_EXIT":
		return m.executePartialExit(pos, action)

	default:
		return nil
	}
}

// executeMoveStop updates the stop loss
func (m *Manager) executeMoveStop(pos *ManagedPosition, action *TierAction) error {
	// Update internal state
	oldStopLoss := pos.StopLoss
	pos.StopLoss = action.NewStopLoss

	// Mark tier as activated
	switch action.TierActivated {
	case 1:
		pos.Tier1Activated = true
		pos.Tier1ActivationPrice = pos.CurrentPrice
		pos.Tier1ActivationTime = pos.EntryTime
	case 3:
		pos.Tier3Activated = true
		pos.Tier3ActivationPrice = pos.CurrentPrice
		pos.Tier3ActivationTime = pos.EntryTime
		pos.Tier3LockedProfit = pos.MaxProfit
	}

	// Call callback to update external system
	if m.stopUpdateCb != nil {
		if err := m.stopUpdateCb(pos.Symbol, action.NewStopLoss); err != nil {
			return fmt.Errorf("failed to update stop loss: %w", err)
		}
	}

	if m.verbose || action.TierActivated > 0 {
		fmt.Printf("\n%s\n", action.Reason)
		fmt.Printf("   Stop Loss: $%.4f → $%.4f\n", oldStopLoss, action.NewStopLoss)
	}

	return nil
}

// executePartialExit closes part of the position
func (m *Manager) executePartialExit(pos *ManagedPosition, action *TierAction) error {
	if m.partialExitCb == nil {
		return fmt.Errorf("partial exit callback not configured")
	}

	// Execute partial exit via callback
	exitedProfit, err := m.partialExitCb(pos.Symbol, action.ExitPercent, pos.CurrentPrice)
	if err != nil {
		return fmt.Errorf("failed to execute partial exit: %w", err)
	}

	// Update position state
	pos.ApplyPartialExit(action.ExitPercent, pos.CurrentPrice)

	// Also update stop to breakeven
	oldStopLoss := pos.StopLoss
	pos.StopLoss = action.NewStopLoss

	if m.stopUpdateCb != nil {
		if err := m.stopUpdateCb(pos.Symbol, action.NewStopLoss); err != nil {
			return fmt.Errorf("failed to update stop loss after partial: %w", err)
		}
	}

	if m.verbose {
		fmt.Printf("\n%s\n", action.Reason)
		fmt.Printf("   Closed %.0f%% (${%.2f}) | Profit: $%.4f\n",
			action.ExitPercent,
			pos.Tier2ExitedSize,
			exitedProfit)
		fmt.Printf("   Remaining: $%.2f | Stop: $%.4f → $%.4f\n",
			pos.RemainingSize,
			oldStopLoss,
			action.NewStopLoss)
	}

	return nil
}

// RemovePosition removes a position from management (when closed)
func (m *Manager) RemovePosition(symbol string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.positions, symbol)

	if m.verbose {
		fmt.Printf("✅ Trade Manager: Removed position %s\n", symbol)
	}
}

// GetPosition returns a managed position
func (m *Manager) GetPosition(symbol string) (*ManagedPosition, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	pos, exists := m.positions[symbol]
	return pos, exists
}

// GetAllPositions returns all managed positions
func (m *Manager) GetAllPositions() map[string]*ManagedPosition {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to avoid race conditions
	posCopy := make(map[string]*ManagedPosition)
	for k, v := range m.positions {
		posCopy[k] = v
	}
	return posCopy
}

// GetActivePositionCount returns the number of actively managed positions
func (m *Manager) GetActivePositionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.positions)
}

// PrintStatus prints status for all positions
func (m *Manager) PrintStatus() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.positions) == 0 {
		fmt.Println("\n📊 No positions under management")
		return
	}

	fmt.Printf("\n📊 Managing %d position(s):\n", len(m.positions))
	for _, pos := range m.positions {
		m.tierManager.PrintStatus(pos)
	}
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// SetConfig updates the configuration
func (m *Manager) SetConfig(config *Config) {
	m.config = config
	m.tierManager.SetConfig(config)
}

// Enable enables the 3-Tier system
func (m *Manager) Enable() {
	m.config.Enabled = true
	fmt.Println("✅ 3-Tier Trade Management: ENABLED")
}

// Disable disables the 3-Tier system
func (m *Manager) Disable() {
	m.config.Enabled = false
	fmt.Println("⏸️  3-Tier Trade Management: DISABLED")
}

// IsEnabled returns whether the system is enabled
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}

// getTierSummary returns a summary of tier thresholds
func (m *Manager) getTierSummary() string {
	return fmt.Sprintf("T1:%.1f%% T2:%.1f%% T3:%ds",
		m.config.Tier1BreakevenThreshold,
		m.config.Tier2PartialExitThreshold,
		m.config.Tier3TimeThreshold)
}

// SetVerbose enables/disables verbose logging
func (m *Manager) SetVerbose(verbose bool) {
	m.verbose = verbose
}

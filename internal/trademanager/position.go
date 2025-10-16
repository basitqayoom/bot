package trademanager

import "time"

// ManagedPosition wraps a position with 3-Tier state tracking
type ManagedPosition struct {
	// Original position data
	ID           int
	Symbol       string
	Side         string
	EntryPrice   float64
	EntryTime    time.Time
	StopLoss     float64
	TakeProfit   float64
	Size         float64
	OriginalSize float64 // Track original size for partial exits

	// Current state
	CurrentPrice  float64
	HighestPrice  float64
	LowestPrice   float64
	MaxProfit     float64
	MaxProfitPct  float64
	RemainingSize float64

	// Tier state tracking
	Tier1Activated       bool      // Breakeven lock activated
	Tier1ActivationTime  time.Time // When breakeven was activated
	Tier1ActivationPrice float64   // Price when breakeven activated

	Tier2Activated       bool      // Partial exit completed
	Tier2ActivationTime  time.Time // When partial exit happened
	Tier2ActivationPrice float64   // Price when partial exit happened
	Tier2ExitedSize      float64   // Amount exited in Tier 2
	Tier2ExitedProfit    float64   // Profit from Tier 2 exit

	Tier3Activated       bool      // Time-based lock activated
	Tier3ActivationTime  time.Time // When time lock activated
	Tier3ActivationPrice float64   // Price when time lock activated
	Tier3LockedProfit    float64   // Profit level locked by Tier 3

	// Time tracking for Tier 3
	FirstProfitableTime time.Time // When position first became profitable
	TimeInProfit        float64   // Seconds spent in profit
}

// NewManagedPosition creates a new managed position from basic parameters
func NewManagedPosition(id int, symbol, side string, entryPrice, stopLoss, takeProfit, size float64) *ManagedPosition {
	return &ManagedPosition{
		ID:            id,
		Symbol:        symbol,
		Side:          side,
		EntryPrice:    entryPrice,
		EntryTime:     time.Now(),
		StopLoss:      stopLoss,
		TakeProfit:    takeProfit,
		Size:          size,
		OriginalSize:  size,
		RemainingSize: size,
		CurrentPrice:  entryPrice,
		HighestPrice:  entryPrice,
		LowestPrice:   entryPrice,
	}
}

// UpdatePrice updates the current price and tracking metrics
func (p *ManagedPosition) UpdatePrice(price float64) {
	p.CurrentPrice = price

	// Track price extremes
	if price > p.HighestPrice {
		p.HighestPrice = price
	}
	if price < p.LowestPrice {
		p.LowestPrice = price
	}

	// Calculate current profit
	currentProfit, currentProfitPct := p.CalculateCurrentProfit()

	// Track max profit
	if currentProfit > p.MaxProfit {
		p.MaxProfit = currentProfit
		p.MaxProfitPct = currentProfitPct
	}

	// Track time in profit for Tier 3
	if currentProfit > 0 {
		if p.FirstProfitableTime.IsZero() {
			p.FirstProfitableTime = time.Now()
		}
		p.TimeInProfit = time.Since(p.FirstProfitableTime).Seconds()
	}
}

// CalculateCurrentProfit returns current profit in dollars and percentage
func (p *ManagedPosition) CalculateCurrentProfit() (float64, float64) {
	var profit float64

	if p.Side == "SHORT" {
		profit = (p.EntryPrice - p.CurrentPrice) * (p.RemainingSize / p.EntryPrice)
	} else { // LONG
		profit = (p.CurrentPrice - p.EntryPrice) * (p.RemainingSize / p.EntryPrice)
	}

	profitPct := (profit / p.RemainingSize) * 100

	return profit, profitPct
}

// GetCurrentProfitPct returns just the profit percentage
func (p *ManagedPosition) GetCurrentProfitPct() float64 {
	_, pct := p.CalculateCurrentProfit()
	return pct
}

// IsInProfit returns true if position is currently profitable
func (p *ManagedPosition) IsInProfit() bool {
	profit, _ := p.CalculateCurrentProfit()
	return profit > 0
}

// GetDuration returns how long the position has been open
func (p *ManagedPosition) GetDuration() time.Duration {
	return time.Since(p.EntryTime)
}

// GetTimeInProfit returns how long position has been in profit
func (p *ManagedPosition) GetTimeInProfitDuration() time.Duration {
	if p.FirstProfitableTime.IsZero() {
		return 0
	}
	return time.Since(p.FirstProfitableTime)
}

// ApplyPartialExit reduces position size and records the exit
func (p *ManagedPosition) ApplyPartialExit(exitPercent, exitPrice float64) float64 {
	exitSize := p.RemainingSize * (exitPercent / 100.0)

	// Calculate profit from this partial exit
	var exitProfit float64
	if p.Side == "SHORT" {
		exitProfit = (p.EntryPrice - exitPrice) * (exitSize / p.EntryPrice)
	} else {
		exitProfit = (exitPrice - p.EntryPrice) * (exitSize / p.EntryPrice)
	}

	// Update position state
	p.RemainingSize -= exitSize
	p.Tier2ExitedSize = exitSize
	p.Tier2ExitedProfit = exitProfit
	p.Tier2Activated = true
	p.Tier2ActivationTime = time.Now()
	p.Tier2ActivationPrice = exitPrice

	return exitProfit
}

// GetTotalProfit returns combined profit from all exits
func (p *ManagedPosition) GetTotalProfit() float64 {
	currentProfit, _ := p.CalculateCurrentProfit()
	return p.Tier2ExitedProfit + currentProfit
}

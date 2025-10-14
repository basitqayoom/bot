package main

import (
	"math"
	"time"
)

// ==================== ATR CALCULATION ====================

// calcATR computes Average True Range using Wilder's smoothing method
func calcATR(candles []Candle, period int) []float64 {
	if period <= 0 || len(candles) < period+1 {
		return make([]float64, len(candles))
	}

	atr := make([]float64, len(candles))
	for i := 0; i < len(atr); i++ {
		atr[i] = -1 // mark invalid
	}

	// Calculate initial True Range values
	trueRanges := make([]float64, len(candles))
	for i := 1; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)

		trueRanges[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	// Calculate initial ATR (simple average)
	var sum float64
	for i := 1; i <= period; i++ {
		sum += trueRanges[i]
	}
	atr[period] = sum / float64(period)

	// Calculate subsequent ATR using Wilder's smoothing
	for i := period + 1; i < len(candles); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + trueRanges[i]) / float64(period)
	}

	return atr
}

// ==================== PIVOT DETECTION ====================

// PivotPoint represents a swing high or low point
type PivotPoint struct {
	Index     int
	Price     float64
	Time      time.Time
	IsHigh    bool
	ATR       float64
	ZoneTop   float64
	ZoneBot   float64
	IsBullish bool // true for support, false for resistance
}

// findPivotHighs detects pivot highs using asymmetric lookback
// Uses lookLeft bars to the left and lookRight bars to the right
func findPivotHighs(candles []Candle, atr []float64, lookLeft, lookRight int, atrMult, maxPercent float64) []PivotPoint {
	var pivots []PivotPoint

	for i := lookLeft; i < len(candles)-lookRight; i++ {
		isPivot := true
		high := candles[i].High

		// Check left side
		for j := i - lookLeft; j < i; j++ {
			if candles[j].High >= high {
				isPivot = false
				break
			}
		}

		// Check right side
		if isPivot {
			for j := i + 1; j <= i+lookRight; j++ {
				if candles[j].High >= high {
					isPivot = false
					break
				}
			}
		}

		if isPivot && atr[i] > 0 {
			// Calculate zone width using ATR (with max percent limit)
			currentATR := atr[i]
			maxZoneWidth := high * (maxPercent / 100)
			band := math.Min(currentATR*atrMult, maxZoneWidth) / 2

			pivots = append(pivots, PivotPoint{
				Index:     i,
				Price:     high,
				Time:      candles[i].OpenTime,
				IsHigh:    true,
				ATR:       currentATR,
				ZoneTop:   high + band,
				ZoneBot:   high - band,
				IsBullish: false, // Starts as resistance
			})
		}
	}

	return pivots
}

// findPivotLows detects pivot lows using asymmetric lookback
func findPivotLows(candles []Candle, atr []float64, lookLeft, lookRight int, atrMult, maxPercent float64) []PivotPoint {
	var pivots []PivotPoint

	for i := lookLeft; i < len(candles)-lookRight; i++ {
		isPivot := true
		low := candles[i].Low

		// Check left side
		for j := i - lookLeft; j < i; j++ {
			if candles[j].Low <= low {
				isPivot = false
				break
			}
		}

		// Check right side
		if isPivot {
			for j := i + 1; j <= i+lookRight; j++ {
				if candles[j].Low <= low {
					isPivot = false
					break
				}
			}
		}

		if isPivot && atr[i] > 0 {
			// Calculate zone width using ATR (with max percent limit)
			currentATR := atr[i]
			maxZoneWidth := low * (maxPercent / 100)
			band := math.Min(currentATR*atrMult, maxZoneWidth) / 2

			pivots = append(pivots, PivotPoint{
				Index:     i,
				Price:     low,
				Time:      candles[i].OpenTime,
				IsHigh:    false,
				ATR:       currentATR,
				ZoneTop:   low + band,
				ZoneBot:   low - band,
				IsBullish: true, // Starts as support
			})
		}
	}

	return pivots
}

// ==================== ZONE MERGING (ALIGNMENT) ====================

// SRZone represents a support or resistance zone with merging capability
type SRZone struct {
	Level      float64
	ZoneTop    float64
	ZoneBot    float64
	Strength   int
	Type       string
	IsBullish  bool
	FirstTouch time.Time
	LastTouch  time.Time
	ZoneRange  float64
	PivotCount int
	AvgATR     float64
}

// zonesOverlap checks if two zones overlap
func zonesOverlap(zone1Top, zone1Bot, zone2Top, zone2Bot float64) bool {
	// Check all overlap conditions from the TradingView indicator
	return (zone2Top > zone1Bot && zone2Top < zone1Top) ||
		(zone2Bot < zone1Top && zone2Bot > zone1Bot) ||
		(zone2Top > zone1Top && zone2Bot < zone1Bot) ||
		(zone2Bot > zone1Bot && zone2Top < zone1Top)
}

// mergeZones combines overlapping zones into a single stronger zone
func mergeZones(zones []SRZone, alignZones bool) []SRZone {
	if !alignZones || len(zones) == 0 {
		return zones
	}

	merged := make([]SRZone, 0, len(zones))
	used := make([]bool, len(zones))

	for i := 0; i < len(zones); i++ {
		if used[i] {
			continue
		}

		currentZone := zones[i]
		mergedAny := false

		// Try to merge with subsequent zones
		for j := i + 1; j < len(zones); j++ {
			if used[j] {
				continue
			}

			if zonesOverlap(currentZone.ZoneTop, currentZone.ZoneBot, zones[j].ZoneTop, zones[j].ZoneBot) {
				// Merge zones - expand to encompass both
				newTop := math.Max(currentZone.ZoneTop, zones[j].ZoneTop)
				newBot := math.Min(currentZone.ZoneBot, zones[j].ZoneBot)
				newLevel := (newTop + newBot) / 2

				// Update zone with merged properties
				currentZone.ZoneTop = newTop
				currentZone.ZoneBot = newBot
				currentZone.Level = newLevel
				currentZone.ZoneRange = newTop - newBot
				currentZone.Strength += zones[j].Strength
				currentZone.PivotCount += zones[j].PivotCount

				// Weighted average of ATR
				totalStrength := float64(currentZone.Strength + zones[j].Strength)
				currentZone.AvgATR = (currentZone.AvgATR*float64(currentZone.Strength) +
					zones[j].AvgATR*float64(zones[j].Strength)) / totalStrength

				// Update time bounds
				if zones[j].FirstTouch.Before(currentZone.FirstTouch) {
					currentZone.FirstTouch = zones[j].FirstTouch
				}
				if zones[j].LastTouch.After(currentZone.LastTouch) {
					currentZone.LastTouch = zones[j].LastTouch
				}

				used[j] = true
				mergedAny = true
			}
		}

		if mergedAny || !used[i] {
			merged = append(merged, currentZone)
			used[i] = true
		}
	}

	return merged
}

// ==================== DYNAMIC ZONE COLOR UPDATE ====================

// updateZonePolarityByPrice updates whether zones are acting as support or resistance
// based on current price position (mimics the _color function)
func updateZonePolarityByPrice(zones []SRZone, currentPrice float64) []SRZone {
	for i := range zones {
		if currentPrice > zones[i].ZoneTop {
			// Price is above zone - it's now support
			zones[i].IsBullish = true
			zones[i].Type = "support"
		} else if currentPrice < zones[i].ZoneBot {
			// Price is below zone - it's now resistance
			zones[i].IsBullish = false
			zones[i].Type = "resistance"
		}
		// If price is within zone, keep existing polarity
	}
	return zones
}

// ==================== MAIN S/R DETECTION FUNCTION ====================

// findAdvancedSupportResistance implements the TradingView indicator logic
func findAdvancedSupportResistance(candles []Candle, config SRConfig) []SRZone {
	// Step 1: Calculate ATR
	atr := calcATR(candles, config.ATRLength)

	// Step 2: Find pivot highs (resistance zones)
	pivotHighs := findPivotHighs(candles, atr, config.LookLeft, config.LookRight,
		config.ATRMultiplier, config.MaxZonePercent)

	// Step 3: Find pivot lows (support zones)
	pivotLows := findPivotLows(candles, atr, config.LookLeft, config.LookRight,
		config.ATRMultiplier, config.MaxZonePercent)

	// Step 4: Convert pivots to zones
	var allZones []SRZone

	// Add resistance zones from pivot highs
	for _, pivot := range pivotHighs {
		zone := SRZone{
			Level:      pivot.Price,
			ZoneTop:    pivot.ZoneTop,
			ZoneBot:    pivot.ZoneBot,
			Strength:   1,
			Type:       "resistance",
			IsBullish:  false,
			FirstTouch: pivot.Time,
			LastTouch:  pivot.Time,
			ZoneRange:  pivot.ZoneTop - pivot.ZoneBot,
			PivotCount: 1,
			AvgATR:     pivot.ATR,
		}
		allZones = append(allZones, zone)
	}

	// Add support zones from pivot lows
	for _, pivot := range pivotLows {
		zone := SRZone{
			Level:      pivot.Price,
			ZoneTop:    pivot.ZoneTop,
			ZoneBot:    pivot.ZoneBot,
			Strength:   1,
			Type:       "support",
			IsBullish:  true,
			FirstTouch: pivot.Time,
			LastTouch:  pivot.Time,
			ZoneRange:  pivot.ZoneTop - pivot.ZoneBot,
			PivotCount: 1,
			AvgATR:     pivot.ATR,
		}
		allZones = append(allZones, zone)
	}

	// Step 5: Merge overlapping zones (zone alignment)
	mergedZones := mergeZones(allZones, config.AlignZones)

	// Step 6: Filter by minimum strength
	var significantZones []SRZone
	for _, zone := range mergedZones {
		if zone.Strength >= config.MinStrength {
			significantZones = append(significantZones, zone)
		}
	}

	// Step 7: Update zone polarity based on current price
	if len(candles) > 0 {
		currentPrice := candles[len(candles)-1].Close
		significantZones = updateZonePolarityByPrice(significantZones, currentPrice)
	}

	// Step 8: Limit number of zones (keep strongest/closest)
	if len(significantZones) > config.MaxZones && config.MaxZones > 0 {
		significantZones = filterTopZones(significantZones, candles[len(candles)-1].Close, config.MaxZones)
	}

	return significantZones
}

// filterTopZones keeps the most relevant zones (strongest and closest to current price)
func filterTopZones(zones []SRZone, currentPrice float64, maxZones int) []SRZone {
	// Sort by combination of strength and proximity
	type scoredZone struct {
		zone  SRZone
		score float64
	}

	scored := make([]scoredZone, len(zones))
	for i, zone := range zones {
		// Calculate distance from current price
		distance := math.Abs(zone.Level - currentPrice)
		distancePercent := (distance / currentPrice) * 100

		// Score: higher strength is better, closer distance is better
		// Normalize: strength weight = 50%, proximity weight = 50%
		strengthScore := float64(zone.Strength) * 10
		proximityScore := math.Max(0, 100-distancePercent)
		score := strengthScore + proximityScore

		scored[i] = scoredZone{zone: zone, score: score}
	}

	// Sort by score descending
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Take top N zones
	result := make([]SRZone, 0, maxZones)
	for i := 0; i < len(scored) && i < maxZones; i++ {
		result = append(result, scored[i].zone)
	}

	return result
}

// ==================== CONFIGURATION ====================

// SRConfig holds configuration for S/R detection matching TradingView indicator
type SRConfig struct {
	LookLeft       int     // Look left bars for pivot (default: 20)
	LookRight      int     // Look right bars for pivot (default: 15)
	ATRLength      int     // ATR period (default: 30)
	ATRMultiplier  float64 // Zone width = ATR * multiplier (default: 0.5)
	MaxZonePercent float64 // Max zone size as % of price (default: 5.0)
	AlignZones     bool    // Enable zone merging (default: true)
	MinStrength    int     // Minimum touches for significance (default: 1)
	MaxZones       int     // Maximum zones to return (default: 20)
}

// DefaultSRConfig returns configuration matching the TradingView indicator
func DefaultSRConfig() SRConfig {
	return SRConfig{
		LookLeft:       20,
		LookRight:      15,
		ATRLength:      30,
		ATRMultiplier:  0.5,
		MaxZonePercent: 5.0,
		AlignZones:     true,
		MinStrength:    1,
		MaxZones:       20,
	}
}

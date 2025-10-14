package main

import (
	"fmt"
	"time"
)

// calcRSI computes Wilder's RSI for the close prices. Returns slice aligned
// with closes (leading entries < period will be -1).
func calcRSI(closes []float64, period int) []float64 {
	if period <= 0 || len(closes) < period+1 {
		return make([]float64, len(closes))
	}
	rsi := make([]float64, len(closes))
	for i := 0; i < len(rsi); i++ {
		rsi[i] = -1 // mark invalid
	}
	var gainSum, lossSum float64
	for i := 1; i <= period; i++ {
		diff := closes[i] - closes[i-1]
		if diff > 0 {
			gainSum += diff
		} else {
			lossSum -= diff // diff negative
		}
	}
	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)
	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - 100/(1+rs)
	}
	for i := period + 1; i < len(closes); i++ {
		diff := closes[i] - closes[i-1]
		var gain, loss float64
		if diff > 0 {
			gain = diff
		} else {
			loss = -diff
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - 100/(1+rs)
		}
	}
	return rsi
}

// BearishDivergence represents a single bearish divergence with both swing points
type BearishDivergence struct {
	// First swing point (earlier)
	StartIdx   int
	StartTime  string
	StartPrice float64
	StartRSI   float64

	// Second swing point (later)
	EndIdx   int
	EndTime  string
	EndPrice float64
	EndRSI   float64
}

// findBearishDivergences identifies indices where price makes a higher high
// but RSI makes a lower high compared to previous swing high.
// swingLookback controls how many candles on each side define a swing high.
func findBearishDivergences(candles []Candle, rsi []float64, swingLookback int) []BearishDivergence {
	isSwingHigh := func(i int) bool {
		if i < swingLookback || i >= len(candles)-swingLookback {
			return false
		}
		h := candles[i].High
		for b := i - swingLookback; b <= i+swingLookback; b++ {
			if candles[b].High > h {
				return false
			}
		}
		return true
	}
	type swing struct {
		idx  int
		high float64
		rsi  float64
	}
	var swings []swing
	for i := range candles {
		if isSwingHigh(i) && rsi[i] > 0 {
			swings = append(swings, swing{i, candles[i].High, rsi[i]})
		}
	}
	var divergences []BearishDivergence
	for i := 1; i < len(swings); i++ {
		prev := swings[i-1]
		cur := swings[i]
		if cur.high > prev.high && cur.rsi < prev.rsi { // bearish divergence
			div := BearishDivergence{
				StartIdx:   prev.idx,
				StartTime:  candles[prev.idx].OpenTime.Format("2006-01-02 15:04"),
				StartPrice: prev.high,
				StartRSI:   prev.rsi,
				EndIdx:     cur.idx,
				EndTime:    candles[cur.idx].OpenTime.Format("2006-01-02 15:04"),
				EndPrice:   cur.high,
				EndRSI:     cur.rsi,
			}
			divergences = append(divergences, div)
		}
	}
	return divergences
}

// analyzeBearishDivergence computes RSI and prints bearish divergences.
func analyzeBearishDivergence(candles []Candle) {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	rsi := calcRSI(closes, 14)
	divs := findBearishDivergences(candles, rsi, 2)
	if len(divs) == 0 {
		fmt.Println("No bearish divergences found")
		return
	}
	fmt.Println("\n========== BEARISH DIVERGENCES ==========")
	fmt.Println("Draw lines on TradingView between these two points:")
	fmt.Println()
	for i, div := range divs {
		fmt.Printf("Divergence #%d:\n", i+1)
		fmt.Printf("  START POINT (Earlier Swing):\n")
		fmt.Printf("    Index: %d | Time: %s | Price: %.2f | RSI: %.2f\n",
			div.StartIdx, div.StartTime, div.StartPrice, div.StartRSI)
		fmt.Printf("  END POINT (Later Swing):\n")
		fmt.Printf("    Index: %d | Time: %s | Price: %.2f | RSI: %.2f\n",
			div.EndIdx, div.EndTime, div.EndPrice, div.EndRSI)
		fmt.Printf("  DIVERGENCE: Price %.2f → %.2f (↑ %.2f%%) but RSI %.2f → %.2f (↓ %.2f%%)\n\n",
			div.StartPrice, div.EndPrice,
			((div.EndPrice-div.StartPrice)/div.StartPrice)*100,
			div.StartRSI, div.EndRSI,
			((div.StartRSI-div.EndRSI)/div.StartRSI)*100)
	}
	fmt.Printf("Total divergences found: %d\n", len(divs))
	fmt.Println("=========================================")
	fmt.Println()
}

// SupportResistanceZone represents a price zone that acts as support or resistance
type SupportResistanceZone struct {
	Level      float64   // Central price level
	Strength   int       // Number of touches/tests
	Type       string    // "resistance" or "support"
	FirstTouch time.Time // When first identified
	LastTouch  time.Time // Most recent touch
	ZoneRange  float64   // Price range of the zone (tolerance)
}

// findSupportResistanceZones identifies key S/R levels from swing points and divergences
func findSupportResistanceZones(candles []Candle, divergences []BearishDivergence, tolerance float64) []SupportResistanceZone {
	if tolerance <= 0 {
		tolerance = 0.02 // 2% default tolerance
	}

	// Collect all swing highs and lows
	type pricePoint struct {
		price  float64
		time   time.Time
		isHigh bool
	}

	var points []pricePoint

	// Add swing highs from divergences (these are resistance)
	for _, div := range divergences {
		startTime, _ := time.Parse("2006-01-02 15:04", div.StartTime)
		endTime, _ := time.Parse("2006-01-02 15:04", div.EndTime)

		points = append(points, pricePoint{
			price:  div.StartPrice,
			time:   startTime,
			isHigh: true,
		})
		points = append(points, pricePoint{
			price:  div.EndPrice,
			time:   endTime,
			isHigh: true,
		})
	}

	// Add swing lows (support) - opposite of swing highs
	swingLookback := 2
	for i := swingLookback; i < len(candles)-swingLookback; i++ {
		isSwingLow := true
		low := candles[i].Low
		for b := i - swingLookback; b <= i+swingLookback; b++ {
			if b != i && candles[b].Low < low {
				isSwingLow = false
				break
			}
		}
		if isSwingLow {
			points = append(points, pricePoint{
				price:  low,
				time:   candles[i].OpenTime,
				isHigh: false,
			})
		}
	}

	// Cluster points into zones
	var zones []SupportResistanceZone

	for _, point := range points {
		foundZone := false

		// Try to add to existing zone
		for j := range zones {
			// Check if price is within tolerance of zone
			priceDiff := (point.price - zones[j].Level) / zones[j].Level
			if priceDiff < 0 {
				priceDiff = -priceDiff
			}

			if priceDiff <= tolerance {
				// Update zone
				zones[j].Strength++
				if point.time.After(zones[j].LastTouch) {
					zones[j].LastTouch = point.time
				}
				// Adjust level (weighted average)
				zones[j].Level = (zones[j].Level*float64(zones[j].Strength-1) + point.price) / float64(zones[j].Strength)
				foundZone = true
				break
			}
		}

		// Create new zone if not found
		if !foundZone {
			zoneType := "resistance"
			if !point.isHigh {
				zoneType = "support"
			}
			zones = append(zones, SupportResistanceZone{
				Level:      point.price,
				Strength:   1,
				Type:       zoneType,
				FirstTouch: point.time,
				LastTouch:  point.time,
				ZoneRange:  point.price * tolerance,
			})
		}
	}

	// Filter zones by strength (keep only significant ones)
	var significantZones []SupportResistanceZone
	for _, zone := range zones {
		if zone.Strength >= 2 { // At least 2 touches
			significantZones = append(significantZones, zone)
		}
	}

	return significantZones
}

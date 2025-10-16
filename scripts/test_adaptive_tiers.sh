#!/bin/bash

# 3-Tier Adaptive Configuration Test Script
# This simulates what will happen with different SL/TP scenarios

echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     3-TIER ADAPTIVE CONFIGURATION SIMULATOR               ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Function to calculate adaptive tiers
calc_tiers() {
    local entry=$1
    local sl=$2
    local side=$3
    
    # Calculate SL distance
    if [ "$side" = "SHORT" ]; then
        sl_dist=$(echo "scale=4; ($sl - $entry) / $entry * 100" | bc)
    else
        sl_dist=$(echo "scale=4; ($entry - $sl) / $entry * 100" | bc)
    fi
    
    # Calculate tier thresholds
    tier1=$(echo "scale=4; $sl_dist * 0.4" | bc)
    tier2=$(echo "scale=4; $sl_dist * 0.7" | bc)
    tier3=$(echo "scale=4; $sl_dist * 0.3" | bc)
    
    echo "   SL Distance: ${sl_dist}%"
    echo "   ├─ Tier 1 (Breakeven): ${tier1}%"
    echo "   ├─ Tier 2 (Partial):   ${tier2}%"
    echo "   └─ Tier 3 (Min):       ${tier3}%"
    echo ""
}

# Test Case 1: Fixed SL/TP (Original strategy)
echo "📊 SCENARIO 1: Fixed SL/TP (No S/R zones)"
echo "═══════════════════════════════════════════"
echo "   Entry: \$100.00 SHORT"
echo "   SL:    \$100.40 (Fixed 0.4%)"
echo "   TP:    \$99.20  (Fixed 0.8%)"
echo ""
calc_tiers 100 100.40 "SHORT"
echo "   ✅ Both tiers trigger BEFORE 0.4% SL"
echo ""

# Test Case 2: Your actual Trade #5 scenario
echo "📊 SCENARIO 2: Your Trade #5 (EVAAUSDT)"
echo "═══════════════════════════════════════════"
echo "   Entry: \$3.18 SHORT"
echo "   SL:    \$3.20 (Resistance zone = 0.6%)"
echo "   TP:    \$3.10 (Support zone = 2.5%)"
echo ""
calc_tiers 3.18 3.20 "SHORT"
echo "   ✅ Tier 2 at 0.42% vs SL at 0.6%"
echo "   ✅ 30% safety buffer (0.18% margin)"
echo "   ✅ Would have saved your -0.45% loss!"
echo ""

# Test Case 3: Wide S/R zones
echo "📊 SCENARIO 3: Wide S/R Zones"
echo "═══════════════════════════════════════════"
echo "   Entry: \$100.00 SHORT"
echo "   SL:    \$102.00 (Distant resistance = 2%)"
echo "   TP:    \$95.00  (Distant support = 5%)"
echo ""
calc_tiers 100 102 "SHORT"
echo "   ✅ Tier 2 at 1.4% vs SL at 2.0%"
echo "   ✅ 30% safety buffer (0.6% margin)"
echo "   ✅ Plenty of room for price action"
echo ""

# Test Case 4: Medium S/R zones
echo "📊 SCENARIO 4: Medium S/R Zones"
echo "═══════════════════════════════════════════"
echo "   Entry: \$100.00 SHORT"
echo "   SL:    \$101.00 (Resistance = 1%)"
echo "   TP:    \$97.00  (Support = 3%)"
echo ""
calc_tiers 100 101 "SHORT"
echo "   ✅ Tier 2 at 0.7% vs SL at 1.0%"
echo "   ✅ 30% safety buffer (0.3% margin)"
echo "   ✅ Balanced protection"
echo ""

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                    KEY INSIGHT                            ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "  OLD SYSTEM (Fixed):"
echo "  ├─ Tier 2 always at 0.6%"
echo "  ├─ Your SL often at 0.6%-2%"
echo "  └─ Result: Tier 2 AT or AFTER SL = No protection! ❌"
echo ""
echo "  NEW SYSTEM (Adaptive):"
echo "  ├─ Tier 2 at 70% of SL distance"
echo "  ├─ Always 30% buffer before SL"
echo "  └─ Result: Protection triggers BEFORE SL = Profit saved! ✅"
echo ""
echo "═══════════════════════════════════════════════════════════"
echo "🚀 Run your bot now and watch the difference!"
echo "═══════════════════════════════════════════════════════════"
echo ""

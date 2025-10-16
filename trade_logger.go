package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ==================== TRADE LOGGER ====================

type TradeLogger struct {
	filename string
	file     *os.File
	writer   *csv.Writer
}

// NewTradeLogger creates a logger for single-symbol paper trading
// Appends trades to a single file per symbol (e.g., trades_BTCUSDT.csv)
func NewTradeLogger(symbol string) (*TradeLogger, error) {
	logsDir := "trade_logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Single file per symbol - APPEND mode
	filename := filepath.Join(logsDir, fmt.Sprintf("trades_%s.csv", symbol))

	// Check if file exists to determine if we need to write headers
	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}

	// Open file in append mode (creates if doesn't exist)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Write headers only if file is new
	if !fileExists {
		headers := []string{
			"Trade_ID",
			"Symbol",
			"Interval",
			"Side",
			"Entry_Time",
			"Entry_Price",
			"Exit_Time",
			"Exit_Price",
			"Stop_Loss",
			"Take_Profit",
			"Position_Size",
			"Status",
			"Profit_Loss",
			"Profit_Loss_Pct",
			"Risk_Reward",
			"Highest_Price",
			"Lowest_Price",
			"Max_Profit",
			"Max_Profit_Pct",
			"Give_Back",
			"Give_Back_Pct",
			"Duration_Minutes",
			"Logged_At",
		}

		if err := writer.Write(headers); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write CSV headers: %w", err)
		}

		writer.Flush()
		fmt.Printf("📝 Created new trade log: %s\n", filename)
	} else {
		fmt.Printf("📝 Appending to existing trade log: %s\n", filename)
	}

	return &TradeLogger{
		filename: filename,
		file:     file,
		writer:   writer,
	}, nil
}

// NewMultiTradeLogger creates a logger for multi-symbol paper trading
// Appends all trades to a single trades_all_symbols.csv file
func NewMultiTradeLogger() (*TradeLogger, error) {
	logsDir := "./logs/trade_logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Single file for ALL multi-symbol trades - APPEND mode
	filename := filepath.Join(logsDir, "trades_all_symbols.csv")

	// Check if file exists to determine if we need to write headers
	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}

	// Open file in append mode (creates if doesn't exist)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Write headers only if file is new
	if !fileExists {
		headers := []string{
			"Trade_ID",
			"Symbol",
			"Interval",
			"Side",
			"Entry_Time",
			"Entry_Price",
			"Exit_Time",
			"Exit_Price",
			"Stop_Loss",
			"Take_Profit",
			"Position_Size",
			"Status",
			"Profit_Loss",
			"Profit_Loss_Pct",
			"Risk_Reward",
			"Highest_Price",
			"Lowest_Price",
			"Max_Profit",
			"Max_Profit_Pct",
			"Give_Back",
			"Give_Back_Pct",
			"Duration_Minutes",
			"Logged_At",
		}

		if err := writer.Write(headers); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write CSV headers: %w", err)
		}

		writer.Flush()
		fmt.Printf("📝 Created new multi-symbol trade log: %s\n", filename)
	} else {
		fmt.Printf("📝 Appending to existing multi-symbol trade log: %s\n", filename)
	}

	return &TradeLogger{
		filename: filename,
		file:     file,
		writer:   writer,
	}, nil
}

// LogTrade writes a completed trade to the CSV file
func (tl *TradeLogger) LogTrade(trade *PaperTrade) error {
	if tl == nil || tl.writer == nil {
		return fmt.Errorf("logger not initialized")
	}

	duration := trade.ExitTime.Sub(trade.EntryTime).Minutes()

	// Calculate "give back" (difference between max profit and actual profit)
	giveBack := trade.MaxProfit - trade.ProfitLoss
	giveBackPct := trade.MaxProfitPct - trade.ProfitLossPct

	record := []string{
		fmt.Sprintf("%d", trade.ID),
		trade.Symbol,
		trade.Interval,
		trade.Side,
		trade.EntryTime.Format("2006-01-02 15:04:05"),
		fmt.Sprintf("%.2f", trade.EntryPrice),
		trade.ExitTime.Format("2006-01-02 15:04:05"),
		fmt.Sprintf("%.2f", trade.ExitPrice),
		fmt.Sprintf("%.2f", trade.StopLoss),
		fmt.Sprintf("%.2f", trade.TakeProfit),
		fmt.Sprintf("%.2f", trade.Size),
		trade.Status,
		fmt.Sprintf("%.2f", trade.ProfitLoss),
		fmt.Sprintf("%.2f", trade.ProfitLossPct),
		fmt.Sprintf("%.2f", trade.RiskReward),
		fmt.Sprintf("%.2f", trade.HighestPrice),
		fmt.Sprintf("%.2f", trade.LowestPrice),
		fmt.Sprintf("%.2f", trade.MaxProfit),
		fmt.Sprintf("%.2f", trade.MaxProfitPct),
		fmt.Sprintf("%.2f", giveBack),
		fmt.Sprintf("%.2f", giveBackPct),
		fmt.Sprintf("%.2f", duration),
		time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := tl.writer.Write(record); err != nil {
		return fmt.Errorf("failed to write trade record: %w", err)
	}

	// Flush to ensure data is written immediately
	tl.writer.Flush()

	if err := tl.writer.Error(); err != nil {
		return fmt.Errorf("CSV writer error: %w", err)
	}

	return nil
}

// Close closes the CSV file
func (tl *TradeLogger) Close() error {
	if tl != nil && tl.file != nil {
		tl.writer.Flush()
		return tl.file.Close()
	}
	return nil
}

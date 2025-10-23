package backtesting

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// DataLoader loads historical data for backtesting
type DataLoader struct{}

// NewDataLoader creates a new data loader
func NewDataLoader() *DataLoader {
	return &DataLoader{}
}

// LoadFromCSV loads historical candle data from CSV file
// Expected CSV format: timestamp,open,high,low,close,volume
// timestamp can be in Unix timestamp (seconds or milliseconds) or RFC3339 format
func (dl *DataLoader) LoadFromCSV(filename string, symbol string) (*HistoricalData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header if exists
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Check if first row is header (contains non-numeric data)
	if _, err := strconv.ParseFloat(header[1], 64); err != nil {
		// First row is header, continue to next row
	} else {
		// First row is data, seek back
		file.Seek(0, 0)
		reader = csv.NewReader(file)
	}

	candles := make([]exchanges.Candle, 0)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		if len(record) < 6 {
			continue // Skip invalid records
		}

		candle, err := dl.parseCSVRecord(record, symbol)
		if err != nil {
			continue // Skip invalid records
		}

		candles = append(candles, candle)
	}

	// Sort candles by timestamp
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].Timestamp.Before(candles[j].Timestamp)
	})

	return &HistoricalData{
		Symbol:  symbol,
		Candles: candles,
	}, nil
}

// parseCSVRecord parses a single CSV record into a Candle
func (dl *DataLoader) parseCSVRecord(record []string, symbol string) (exchanges.Candle, error) {
	// Parse timestamp
	timestamp, err := dl.parseTimestamp(record[0])
	if err != nil {
		return exchanges.Candle{}, err
	}

	// Parse OHLCV
	open, err := decimal.NewFromString(record[1])
	if err != nil {
		return exchanges.Candle{}, fmt.Errorf("invalid open price: %w", err)
	}

	high, err := decimal.NewFromString(record[2])
	if err != nil {
		return exchanges.Candle{}, fmt.Errorf("invalid high price: %w", err)
	}

	low, err := decimal.NewFromString(record[3])
	if err != nil {
		return exchanges.Candle{}, fmt.Errorf("invalid low price: %w", err)
	}

	close, err := decimal.NewFromString(record[4])
	if err != nil {
		return exchanges.Candle{}, fmt.Errorf("invalid close price: %w", err)
	}

	volume, err := decimal.NewFromString(record[5])
	if err != nil {
		return exchanges.Candle{}, fmt.Errorf("invalid volume: %w", err)
	}

	return exchanges.Candle{
		Symbol:    symbol,
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}, nil
}

// parseTimestamp parses timestamp from string
// Supports Unix timestamp (seconds or milliseconds) and RFC3339 format
func (dl *DataLoader) parseTimestamp(s string) (time.Time, error) {
	// Try parsing as Unix timestamp
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Check if it's in milliseconds (13 digits) or seconds (10 digits)
		if ts > 10000000000 {
			// Milliseconds
			return time.Unix(ts/1000, (ts%1000)*1000000), nil
		}
		// Seconds
		return time.Unix(ts, 0), nil
	}

	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try parsing as common date formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", s)
}

// GenerateSampleData generates sample historical data for testing
func (dl *DataLoader) GenerateSampleData(symbol string, startTime time.Time, candles int, basePrice float64) *HistoricalData {
	data := &HistoricalData{
		Symbol:  symbol,
		Candles: make([]exchanges.Candle, 0, candles),
	}

	currentTime := startTime
	currentPrice := decimal.NewFromFloat(basePrice)

	for i := 0; i < candles; i++ {
		// Generate random price movement (simplified)
		change := decimal.NewFromFloat((float64(i%10) - 5) * 0.001) // ±0.5% movement
		open := currentPrice
		close := currentPrice.Add(currentPrice.Mul(change))

		high := decimal.Max(open, close).Mul(decimal.NewFromFloat(1.001))
		low := decimal.Min(open, close).Mul(decimal.NewFromFloat(0.999))
		volume := decimal.NewFromFloat(1000 + float64(i%500))

		candle := exchanges.Candle{
			Symbol:    symbol,
			Timestamp: currentTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}

		data.Candles = append(data.Candles, candle)

		currentTime = currentTime.Add(1 * time.Minute)
		currentPrice = close
	}

	return data
}

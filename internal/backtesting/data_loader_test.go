package backtesting

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/testutils"
	"github.com/shopspring/decimal"
)

func TestDataLoader_LoadFromCSV(t *testing.T) {
	loader := NewDataLoader()

	// Create temporary CSV file
	tempDir := t.TempDir()
	csvFile := filepath.Join(tempDir, "test_data.csv")

	// Create CSV content with header
	csvContent := `timestamp,open,high,low,close,volume
1640995200,50000,51000,49000,50500,100
1640995260,50500,51500,49500,51000,150
1640995320,51000,52000,50000,51500,200`

	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	testutils.AssertNoError(t, err, "Failed to create test CSV file")

	// Load data
	data, err := loader.LoadFromCSV(csvFile, "BTC-USD")
	testutils.AssertNoError(t, err, "Failed to load CSV data")
	testutils.AssertNotNil(t, data, "Data should not be nil")

	testutils.AssertEqual(t, "BTC-USD", data.Symbol, "Symbol should match")
	testutils.AssertEqual(t, 3, len(data.Candles), "Should have 3 candles")

	// Check first candle
	candle := data.Candles[0]
	testutils.AssertEqual(t, "BTC-USD", candle.Symbol, "Candle symbol should match")
	expectedOpen := decimal.NewFromFloat(50000)
	testutils.AssertTrue(t, candle.Open.Equal(expectedOpen), "Open price should match")
	expectedClose := decimal.NewFromFloat(50500)
	testutils.AssertTrue(t, candle.Close.Equal(expectedClose), "Close price should match")
}

func TestDataLoader_LoadFromCSV_NoHeader(t *testing.T) {
	loader := NewDataLoader()

	// Create temporary CSV file without header
	tempDir := t.TempDir()
	csvFile := filepath.Join(tempDir, "test_data_no_header.csv")

	// Create CSV content without header
	csvContent := `1640995200,50000,51000,49000,50500,100
1640995260,50500,51500,49500,51000,150`

	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	testutils.AssertNoError(t, err, "Failed to create test CSV file")

	// Load data
	data, err := loader.LoadFromCSV(csvFile, "BTC-USD")
	testutils.AssertNoError(t, err, "Failed to load CSV data")
	testutils.AssertEqual(t, 2, len(data.Candles), "Should have 2 candles")
}

func TestDataLoader_LoadFromCSV_InvalidFile(t *testing.T) {
	loader := NewDataLoader()

	_, err := loader.LoadFromCSV("nonexistent.csv", "BTC-USD")
	testutils.AssertError(t, err, "Should return error for nonexistent file")
}

func TestDataLoader_ParseTimestamp_UnixSeconds(t *testing.T) {
	loader := NewDataLoader()

	timestamp, err := loader.parseTimestamp("1640995200")
	testutils.AssertNoError(t, err, "Should parse Unix seconds")

	expected := time.Unix(1640995200, 0)
	testutils.AssertEqual(t, expected.Unix(), timestamp.Unix(), "Timestamp should match")
}

func TestDataLoader_ParseTimestamp_UnixMilliseconds(t *testing.T) {
	loader := NewDataLoader()

	timestamp, err := loader.parseTimestamp("1640995200000")
	testutils.AssertNoError(t, err, "Should parse Unix milliseconds")

	expected := time.Unix(1640995200, 0)
	testutils.AssertEqual(t, expected.Unix(), timestamp.Unix(), "Timestamp should match")
}

func TestDataLoader_ParseTimestamp_RFC3339(t *testing.T) {
	loader := NewDataLoader()

	timestamp, err := loader.parseTimestamp("2022-01-01T12:00:00Z")
	testutils.AssertNoError(t, err, "Should parse RFC3339")

	expected, _ := time.Parse(time.RFC3339, "2022-01-01T12:00:00Z")
	testutils.AssertEqual(t, expected.Unix(), timestamp.Unix(), "Timestamp should match")
}

func TestDataLoader_ParseTimestamp_Invalid(t *testing.T) {
	loader := NewDataLoader()

	_, err := loader.parseTimestamp("invalid")
	testutils.AssertError(t, err, "Should return error for invalid timestamp")
}

func TestDataLoader_GenerateSampleData(t *testing.T) {
	loader := NewDataLoader()

	startTime := time.Now()
	data := loader.GenerateSampleData("BTC-USD", startTime, 10, 50000)

	testutils.AssertNotNil(t, data, "Data should not be nil")
	testutils.AssertEqual(t, "BTC-USD", data.Symbol, "Symbol should match")
	testutils.AssertEqual(t, 10, len(data.Candles), "Should have 10 candles")

	// Check that candles are sequential
	for i := 1; i < len(data.Candles); i++ {
		prevTime := data.Candles[i-1].Timestamp
		currTime := data.Candles[i].Timestamp
		testutils.AssertTrue(t, currTime.After(prevTime), "Candles should be in chronological order")
	}
}

func TestDefaultBacktestConfig(t *testing.T) {
	config := DefaultBacktestConfig()

	testutils.AssertNotNil(t, config, "Config should not be nil")
	testutils.AssertTrue(t, config.InitialCapital.Equal(decimal.NewFromFloat(10000)), "Initial capital should be 10000")
	testutils.AssertTrue(t, config.CommissionRate.Equal(decimal.NewFromFloat(0.001)), "Commission rate should be 0.001")
	testutils.AssertFalse(t, config.UseFixedAmount, "Should not use fixed amount by default")
	testutils.AssertEqual(t, 1, config.MaxPositions, "Max positions should be 1")
	testutils.AssertFalse(t, config.AllowShort, "Should not allow short by default")
}

#!/bin/bash

# Constantine Backtesting Script
# Usage: ./scripts/run_backtest.sh [data_file]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                       ║${NC}"
echo -e "${BLUE}║        CONSTANTINE BACKTESTING SCRIPT                 ║${NC}"
echo -e "${BLUE}║                                                       ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if backtest binary exists
if [ ! -f "./bin/backtest" ]; then
    echo -e "${YELLOW}Building backtest binary...${NC}"
    go build -o bin/backtest ./cmd/backtest
    echo -e "${GREEN}✓ Build complete${NC}"
    echo ""
fi

# Default parameters
DATA_FILE="${1:-}"
SYMBOL="${SYMBOL:-BTC-USD}"
CAPITAL="${CAPITAL:-10000}"
COMMISSION="${COMMISSION:-0.001}"
SLIPPAGE="${SLIPPAGE:-0.0005}"
RISK="${RISK:-0.01}"

# Strategy parameters
SHORT_EMA="${SHORT_EMA:-9}"
LONG_EMA="${LONG_EMA:-21}"
RSI_PERIOD="${RSI_PERIOD:-14}"
RSI_OVERSOLD="${RSI_OVERSOLD:-30}"
RSI_OVERBOUGHT="${RSI_OVERBOUGHT:-70}"
TAKE_PROFIT="${TAKE_PROFIT:-0.5}"
STOP_LOSS="${STOP_LOSS:-0.25}"

# Output options
VERBOSE="${VERBOSE:-false}"

# Build command
CMD="./bin/backtest"

if [ -z "$DATA_FILE" ]; then
    echo -e "${YELLOW}No data file provided. Generating sample data...${NC}"
    CMD="$CMD --generate-sample --sample-candles=500"
else
    if [ ! -f "$DATA_FILE" ]; then
        echo -e "${RED}Error: Data file '$DATA_FILE' not found${NC}"
        exit 1
    fi
    CMD="$CMD --data=$DATA_FILE"
fi

CMD="$CMD --symbol=$SYMBOL"
CMD="$CMD --capital=$CAPITAL"
CMD="$CMD --commission=$COMMISSION"
CMD="$CMD --slippage=$SLIPPAGE"
CMD="$CMD --risk=$RISK"
CMD="$CMD --short-ema=$SHORT_EMA"
CMD="$CMD --long-ema=$LONG_EMA"
CMD="$CMD --rsi-period=$RSI_PERIOD"
CMD="$CMD --rsi-oversold=$RSI_OVERSOLD"
CMD="$CMD --rsi-overbought=$RSI_OVERBOUGHT"
CMD="$CMD --take-profit=$TAKE_PROFIT"
CMD="$CMD --stop-loss=$STOP_LOSS"

if [ "$VERBOSE" = "true" ]; then
    CMD="$CMD --verbose"
fi

echo -e "${GREEN}Running backtest with parameters:${NC}"
echo -e "  Symbol:         $SYMBOL"
echo -e "  Capital:        \$$CAPITAL"
echo -e "  Commission:     ${COMMISSION}%"
echo -e "  Risk/Trade:     ${RISK}%"
echo -e "  Short EMA:      $SHORT_EMA"
echo -e "  Long EMA:       $LONG_EMA"
echo -e "  RSI Period:     $RSI_PERIOD"
echo -e "  Take Profit:    ${TAKE_PROFIT}%"
echo -e "  Stop Loss:      ${STOP_LOSS}%"
echo ""

# Run backtest
eval $CMD

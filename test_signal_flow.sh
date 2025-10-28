#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Constantine Bot - Signal Generation Diagnostic${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"

echo -e "${YELLOW}This script will monitor the signal generation flow for 5 minutes.${NC}"
echo -e "${YELLOW}Watch for these key events:${NC}\n"

echo -e "  ${GREEN}✓${NC} Exchanges connecting (Connected to exchange)"
echo -e "  ${GREEN}✓${NC} Strategies launching (subscribing to market data)"
echo -e "  ${GREEN}✓${NC} Candles arriving (📊 candle received)"
echo -e "  ${GREEN}✓${NC} Ready for signals (ready_for_signals=true)"
echo -e "  ${GREEN}✓${NC} Signals being generated (integrated strategy signal)\n"

# Set log level to info to see candle messages
export LOG_LEVEL=info
export LOG_FORMAT=text

echo -e "${BLUE}Starting bot with signal monitoring...${NC}\n"

# Run bot and pipe through filters
./bin/bot 2>&1 | while IFS= read -r line; do
    # Highlight important events
    if echo "$line" | grep -q "exchange enabled\|Connected to"; then
        echo -e "${GREEN}[EXCHANGE]${NC} $line"
    elif echo "$line" | grep -q "subscribing to market data\|subscribed to"; then
        echo -e "${GREEN}[STRATEGY]${NC} $line"
    elif echo "$line" | grep -q "candle received"; then
        echo -e "${GREEN}[CANDLE]${NC} $line"
    elif echo "$line" | grep -q "ready_for_signals=true"; then
        echo -e "${GREEN}[READY]${NC} $line"
    elif echo "$line" | grep -q "strategy signal"; then
        echo -e "${GREEN}[SIGNAL]${NC} $line"
    elif echo "$line" | grep -q "execution"; then
        echo -e "${GREEN}[EXECUTION]${NC} $line"
    elif echo "$line" | grep -q "error\|Error\|ERROR\|failed"; then
        echo -e "${RED}[ERROR]${NC} $line"
    else
        echo "$line"
    fi
done

# TUI Enhancements: Dynamic Weights & Symbol Selection

## Overview

The Terminal User Interface (TUI) has been enhanced to provide real-time visibility into:
- **Dynamic Indicator Weights**: Adaptive weight adjustments based on market conditions
- **Intelligent Symbol Selection**: Selected trading opportunities with opportunity scores
- **Integrated Strategy Engine Status**: Real-time engine health and configuration

## New Features

### 1. Selected Symbols Display

The TUI now shows which symbols are actively selected for trading in the dashboard.

**Location**: Dashboard view (default view)
**View**: New "Selected Symbols" box showing:
- Symbol name with emoji indicator (📊)
- Opportunity score (0.0-1.0)
- Percentage representation (0-100%)
- Key metrics:
  - **Potential**: Estimated gain potential
  - **Sharpe Ratio**: Risk-adjusted return metric
- **Dynamic Weights** (when available):
  - EMA percentage
  - RSI percentage
  - Volume percentage
  - Bollinger Bands percentage
- **Last Refresh**: Time since last symbol selection update

**Color Coding**:
- Score > 60%: 🟢 Green (high opportunity)
- Score 30-60%: 🟡 Yellow (medium opportunity)
- Score < 30%: 🔴 Red (low opportunity)

### 2. Enhanced Summary Box

The Summary box now includes:
- **Selected**: Count and percentage of selected symbols
  - Example: "3/4 symbols (75%)"
  - Shows active trading focus

### 3. Detailed Symbols View

**Access**: Press `7` (new view) or existing ViewSymbols

Displays comprehensive analysis for each selected symbol:

```
📊 BTC-USD
  Opportunity Score: 0.7543 (75.4%)
  Gain Potential:    0.012345
  Risk Assessment:   0.000456
  Sharpe Ratio:      2.708965

  Dynamic Weights:
    EMA:            40.5%
    RSI:            25.3%
    Volume:         22.8%
    Bollinger Bands: 11.4%

  Current Signal: entry BUY (87.3%)

Symbol selection updated: 12s ago
Summary (4 configured, 3 selected)
```

### 4. Enhanced Settings View

**Access**: Press `6`

Now displays Integrated Strategy Engine configuration:

#### Features Enabled
```
✓ Dynamic Indicator Weights
✓ Intelligent Symbol Selection
✓ Real-time Signal Generation
✓ Automatic Risk Management
```

#### Symbol Selection
- Configured: Total trading symbols
- Selected: Active trading symbols
- Last Update: Time since refresh
- Selection Percentage: Current active ratio

#### Indicator Weights
- Base allocation percentages
- Information about dynamic adjustment
- Market condition factors

#### Risk Management
- Stop Loss: 0.40%
- Take Profit: 0.80%
- Minimum Signal: 0.50 (50%)

#### Refresh Interval
- Symbol Refresh: 30 seconds

### 5. Enhanced Header

The header now displays:
- Bot status (RUNNING/STOPPED)
- **Symbol Count**: Shows "Symbols: X/Y" (selected/configured)
- **Engine Status**: Shows "Engine: Active" or "Engine: Initializing"

**Example Header**:
```
⚡ SCALPING BOT  RUNNING  Symbols: 3/4  Engine: Active
```

## Data Flow

```
IntegratedStrategyEngine
    ↓ (Updates every 30 seconds)
    ├─ Recalculate symbol scores
    ├─ Update dynamic weights
    └─ Emit selected symbols
    
    ↓ (Captured by TUI)
    
TUI Model
    ├─ selectedSymbols map
    ├─ dynamicWeights map
    ├─ lastSymbolRefresh time
    └─ currentSignals map

    ↓ (Displayed via)
    
TUI Views
    ├─ Dashboard: Selected Symbols box
    ├─ Summary: Symbol selection percentage
    ├─ Symbols View (press 7): Detailed analysis
    └─ Settings View (press 6): Engine configuration
```

## Navigation

### Views (Press 1-6):

| Key | View | Shows |
|-----|------|-------|
| `1` | Dashboard | Summary, Selected Symbols, Active Signals, Messages |
| `2` | Order Book | Bid/Ask levels |
| `3` | Positions | Open positions across exchanges |
| `4` | Orders | Open orders |
| `5` | Exchanges | Exchange connection status |
| `6` | Settings | Engine config, features, risk parameters |

### Additional Keys:

| Key | Action |
|-----|--------|
| `s` | Start/Stop bot |
| `r` | Refresh data |
| `c` | Clear error |
| `q` | Quit |

## Model Updates

### New Fields in Model

```go
type Model struct {
    // ... existing fields ...
    
    // New fields
    integratedEngine     *strategy.IntegratedStrategyEngine
    selectedSymbols      map[string]strategy.RankedSymbol
    dynamicWeights       map[string]strategy.IndicatorWeights
    lastSymbolRefresh    time.Time
}
```

### New Methods

```go
// Update selected trading symbols
func (m *Model) UpdateSelectedSymbols(symbols map[string]strategy.RankedSymbol)

// Update dynamic weights for a symbol
func (m *Model) UpdateDynamicWeights(symbol string, weights strategy.IndicatorWeights)

// Get integrated engine reference
func (m *Model) GetIntegratedEngine() *strategy.IntegratedStrategyEngine

// Get selected symbols
func (m *Model) GetSelectedSymbols() map[string]strategy.RankedSymbol

// Get dynamic weights for symbol
func (m *Model) GetDynamicWeights(symbol string) (strategy.IndicatorWeights, bool)
```

## View Functions

### renderSelectedSymbols()

New function that renders selected symbols in the dashboard:
- Shows top candidates for trading
- Displays opportunity scores
- Shows key metrics (Potential, Sharpe Ratio)
- Displays dynamic weights when available
- Shows refresh time

### renderSymbols() (Enhanced)

Enhanced symbols view showing:
- All selected symbols with full analysis
- Detailed metric breakdown
- Dynamic weight visualization
- Current active signals
- Selection summary

### renderSettings() (Enhanced)

Now displays:
- Engine activation status
- Features enabled list
- Symbol selection statistics
- Weight configuration
- Risk management parameters
- Refresh intervals

### renderHeader() (Enhanced)

Improved header with:
- Engine status indicator
- Symbol count (selected/configured)
- Color-coded status

### renderSummary() (Enhanced)

Now includes:
- Symbol selection percentage
- Better portfolio visibility

## Real-time Updates

The TUI automatically updates every second to reflect:
- Current selected symbols
- Symbol scores and metrics
- Dynamic weight adjustments
- Active trading signals
- Engine refresh times

## Color Scheme

| Element | Color | Meaning |
|---------|-------|---------|
| Scores 60%+ | Green | High opportunity |
| Scores 30-60% | Muted | Medium opportunity |
| Scores <30% | Red | Low opportunity |
| EMA/RSI Weights | Green | Trend confidence |
| Volume | Muted | Supporting metric |
| BB | Muted | Volatility indicator |

## Example Workflow

1. **Bot Starts**
   - TUI loads with initialized model
   - Engine status shows as "Active"
   - Symbol count displays "0/4 symbols"

2. **First Symbol Refresh (30 seconds)**
   - Engine evaluates all configured symbols
   - Selects top performers (50%)
   - TUI displays "2/4 symbols (50%)"
   - Shows selected symbols with scores in dashboard

3. **Market Conditions Change**
   - Dynamic weights adjust automatically
   - TUI updates weight percentages
   - New signals may generate for different symbols
   - Symbol refresh shows updated time

4. **Viewing Details**
   - Press `6` to see Settings (engine config)
   - Press `7` for detailed symbol analysis
   - See real-time weights and metrics
   - Monitor signal generation

## Integration with Bot

### Callback Integration

The TUI receives updates through:

1. **Selected Symbols Update**
   ```go
   integratedEngine.GetSelectedSymbols()
   m.UpdateSelectedSymbols(symbols)
   ```

2. **Weight Display**
   ```go
   weights, ok := m.GetDynamicWeights(symbol)
   // Display in views
   ```

3. **Signal Handling**
   ```go
   integratedEngine.SetSignalCallback(func(signal *Signal) {
       m.UpdateSignal(signal.Symbol, signal)
   })
   ```

## Technical Details

### Data Refresh Cycle

```
Every 1 second (TUI tick):
├─ fetchData() called
├─ aggregator.RefreshData() → market data
├─ integratedEngine.GetSelectedSymbols() → updated scores
├─ m.UpdateSelectedSymbols() → refresh model
└─ View() renders with latest data

Every 30 seconds (Engine):
├─ Symbol selection re-evaluated
├─ Dynamic weights recalculated
├─ Model automatically updated via fetchData()
└─ TUI reflects changes
```

### Performance

- **TUI Updates**: 1-second refresh cycle
- **Engine Updates**: 30-second symbol refresh
- **Memory Usage**: Minimal (only current state)
- **CPU Usage**: Negligible for rendering

## Future Enhancements

1. **Graphical Visualization**
   - Chart symbol scores over time
   - Weight adaptation visualization
   - Signal confidence curves

2. **Real-time Weight Adjustment Feedback**
   - Show reason for weight changes
   - Market condition alerts
   - Volatility notifications

3. **Symbol Comparison**
   - Side-by-side metric comparison
   - Performance ranking
   - Correlation analysis

4. **Advanced Analytics**
   - Win rate by symbol
   - Average signal strength
   - Selection stability metrics

5. **Custom Thresholds**
   - Configurable minimum score
   - Selection percentage override
   - Weight adjustment intensity

## Troubleshooting

### Selected Symbols Not Showing

1. Check that trading symbols are configured
2. Verify engine is running (check header status)
3. Wait 30+ seconds for first symbol refresh
4. Check logs for symbol selection errors

### Weights Not Displaying

1. Verify symbols are selected
2. Check market data is available
3. Ensure sufficient price history (30+ candles)
4. Check for engine errors in messages

### Symbol Count Not Updating

1. Confirm integratedEngine is initialized
2. Check TUI Update() calls fetchData()
3. Verify GetSelectedSymbols() returns data
4. Monitor lastSymbolRefresh time

## Files Modified

```
internal/tui/
├── model.go      (+53, -2)   - Add engine, selectedSymbols, dynamicWeights
├── view.go       (+224, -20) - Add renderSelectedSymbols(), enhance other renders
├── update.go     (+6, -0)    - Update fetchData() for symbol selection

cmd/bot/
└── main.go       (+2, -1)    - Pass integratedEngine to TUI
```

## Summary

The TUI now provides complete visibility into the Integrated Strategy Engine's operations, displaying:
- ✅ Selected trading symbols
- ✅ Opportunity scores
- ✅ Dynamic indicator weights
- ✅ Engine status
- ✅ Symbol selection metrics
- ✅ Real-time updates

Users can now monitor symbol selection and weight adaptation in real-time while trading.

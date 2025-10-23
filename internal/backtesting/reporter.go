package backtesting

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Reporter generates reports from backtesting results
type Reporter struct{}

// NewReporter creates a new reporter
func NewReporter() *Reporter {
	return &Reporter{}
}

// GenerateReport generates a formatted text report
func (r *Reporter) GenerateReport(metrics *PerformanceMetrics) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════\n")
	sb.WriteString("           BACKTESTING PERFORMANCE REPORT\n")
	sb.WriteString("═══════════════════════════════════════════════════════\n\n")

	// Overall Performance
	sb.WriteString("📊 OVERALL PERFORMANCE\n")
	sb.WriteString("───────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total Return:         $%s (%.2f%%)\n",
		metrics.TotalReturn.StringFixed(2),
		metrics.TotalReturnPct.InexactFloat64()))
	sb.WriteString(fmt.Sprintf("Annualized Return:    %.2f%%\n",
		metrics.AnnualizedReturn.InexactFloat64()))
	sb.WriteString(fmt.Sprintf("Max Drawdown:         $%s (%.2f%%)\n",
		metrics.MaxDrawdown.StringFixed(2),
		metrics.MaxDrawdownPct.InexactFloat64()))
	if !metrics.SharpeRatio.IsZero() {
		sb.WriteString(fmt.Sprintf("Sharpe Ratio:         %.2f\n",
			metrics.SharpeRatio.InexactFloat64()))
	}
	sb.WriteString(fmt.Sprintf("Total Duration:       %s\n\n",
		formatDuration(metrics.TotalDuration)))

	// Trade Statistics
	sb.WriteString("📈 TRADE STATISTICS\n")
	sb.WriteString("───────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total Trades:         %d\n", metrics.TotalTrades))
	sb.WriteString(fmt.Sprintf("Winning Trades:       %d\n", metrics.WinningTrades))
	sb.WriteString(fmt.Sprintf("Losing Trades:        %d\n", metrics.LosingTrades))
	sb.WriteString(fmt.Sprintf("Win Rate:             %.2f%%\n",
		metrics.WinRate.InexactFloat64()))
	sb.WriteString(fmt.Sprintf("Avg Trade Duration:   %s\n\n",
		formatDuration(metrics.AvgTradeDuration)))

	// Profit/Loss Analysis
	sb.WriteString("💰 PROFIT/LOSS ANALYSIS\n")
	sb.WriteString("───────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total Profit:         $%s\n",
		metrics.TotalProfit.StringFixed(2)))
	sb.WriteString(fmt.Sprintf("Total Loss:           $%s\n",
		metrics.TotalLoss.StringFixed(2)))
	sb.WriteString(fmt.Sprintf("Profit Factor:        %.2f\n",
		metrics.ProfitFactor.InexactFloat64()))
	sb.WriteString(fmt.Sprintf("Avg Profit (Win):     $%s\n",
		metrics.AverageProfitWin.StringFixed(2)))
	sb.WriteString(fmt.Sprintf("Avg Loss (Lose):      $%s\n",
		metrics.AverageLossLose.StringFixed(2)))
	sb.WriteString(fmt.Sprintf("Largest Win:          $%s\n",
		metrics.LargestWin.StringFixed(2)))
	sb.WriteString(fmt.Sprintf("Largest Loss:         $%s\n\n",
		metrics.LargestLoss.StringFixed(2)))

	// Recent Trades
	if len(metrics.Trades) > 0 {
		sb.WriteString("📋 RECENT TRADES (Last 10)\n")
		sb.WriteString("───────────────────────────────────────────────────────\n")

		start := len(metrics.Trades) - 10
		if start < 0 {
			start = 0
		}

		for i := start; i < len(metrics.Trades); i++ {
			trade := metrics.Trades[i]
			symbol := "📈"
			if trade.PnL.LessThan(decimal.Zero) {
				symbol = "📉"
			}
			sb.WriteString(fmt.Sprintf("%s %s %s: Entry=$%s Exit=$%s PnL=$%s (%.2f%%) %s\n",
				symbol,
				trade.EntryTime.Format("01-02 15:04"),
				trade.Side,
				trade.EntryPrice.StringFixed(2),
				trade.ExitPrice.StringFixed(2),
				trade.PnL.StringFixed(2),
				trade.PnLPercent.InexactFloat64(),
				trade.ExitReason,
			))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("═══════════════════════════════════════════════════════\n")

	return sb.String()
}

// GenerateSummary generates a short summary
func (r *Reporter) GenerateSummary(metrics *PerformanceMetrics) string {
	return fmt.Sprintf(
		"Return: %.2f%% | Trades: %d | Win Rate: %.2f%% | Max DD: %.2f%% | Profit Factor: %.2f",
		metrics.TotalReturnPct.InexactFloat64(),
		metrics.TotalTrades,
		metrics.WinRate.InexactFloat64(),
		metrics.MaxDrawdownPct.InexactFloat64(),
		metrics.ProfitFactor.InexactFloat64(),
	)
}

// GenerateTradeLog generates a detailed trade log
func (r *Reporter) GenerateTradeLog(metrics *PerformanceMetrics) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("                                TRADE LOG\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n\n")

	for i, trade := range metrics.Trades {
		sb.WriteString(fmt.Sprintf("Trade #%d\n", i+1))
		sb.WriteString("───────────────────────────────────────────────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("Symbol:          %s\n", trade.Symbol))
		sb.WriteString(fmt.Sprintf("Side:            %s\n", trade.Side))
		sb.WriteString(fmt.Sprintf("Entry Time:      %s\n", trade.EntryTime.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("Exit Time:       %s\n", trade.ExitTime.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("Duration:        %s\n", formatDuration(trade.ExitTime.Sub(trade.EntryTime))))
		sb.WriteString(fmt.Sprintf("Entry Price:     $%s\n", trade.EntryPrice.StringFixed(2)))
		sb.WriteString(fmt.Sprintf("Exit Price:      $%s\n", trade.ExitPrice.StringFixed(2)))
		sb.WriteString(fmt.Sprintf("Amount:          %s\n", trade.Amount.StringFixed(4)))
		sb.WriteString(fmt.Sprintf("Stop Loss:       $%s\n", trade.StopLoss.StringFixed(2)))
		sb.WriteString(fmt.Sprintf("Take Profit:     $%s\n", trade.TakeProfit.StringFixed(2)))
		sb.WriteString(fmt.Sprintf("Exit Reason:     %s\n", trade.ExitReason))
		sb.WriteString(fmt.Sprintf("Commission:      $%s\n", trade.Commission.StringFixed(2)))

		pnlStatus := "PROFIT ✓"
		if trade.PnL.LessThan(decimal.Zero) {
			pnlStatus = "LOSS ✗"
		}
		sb.WriteString(fmt.Sprintf("P&L:             $%s (%.2f%%) [%s]\n",
			trade.PnL.StringFixed(2),
			trade.PnLPercent.InexactFloat64(),
			pnlStatus))
		sb.WriteString("\n")
	}

	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n")

	return sb.String()
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}

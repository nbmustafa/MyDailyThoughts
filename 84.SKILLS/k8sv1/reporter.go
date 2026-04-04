package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/your-org/k8sdiag/internal/types"
)

// Colour palettes
var (
	colorCritical = color.New(color.FgHiRed, color.Bold)
	colorWarning  = color.New(color.FgYellow, color.Bold)
	colorInfo     = color.New(color.FgCyan)
	colorOK       = color.New(color.FgGreen, color.Bold)
	colorHeader   = color.New(color.FgHiWhite, color.Bold, color.Underline)
	colorCategory = color.New(color.FgHiMagenta, color.Bold)
	colorDim      = color.New(color.Faint)
	colorTitle    = color.New(color.FgHiWhite, color.Bold)
	colorResource = color.New(color.FgHiBlue)
	colorSuggest  = color.New(color.FgGreen)
	colorBanner   = color.New(color.FgHiCyan, color.Bold)
)

// Reporter handles all output formatting
type Reporter struct {
	w      io.Writer
	noColor bool
	format  string
}

func New(format string, noColor bool) *Reporter {
	if noColor {
		color.NoColor = true
	}
	return &Reporter{
		w:       os.Stdout,
		noColor: noColor,
		format:  format,
	}
}

// Print renders the full diagnostic report
func (r *Reporter) Print(report types.DiagnosticReport) {
	switch r.format {
	case "json":
		r.printJSON(report)
	case "markdown":
		r.printMarkdown(report)
	default:
		r.printText(report)
	}
}

// ── Text (default) ────────────────────────────────────────────────────────────

func (r *Reporter) printText(report types.DiagnosticReport) {
	r.printBanner(report)
	r.printSummaryBox(report)

	// Group findings by category
	byCategory := map[types.Category][]types.Finding{}
	for _, result := range report.Results {
		if result.Error != nil {
			colorWarning.Fprintf(r.w, "  ⚠  [%s] check error: %v\n", result.Category, result.Error)
			continue
		}
		for _, f := range result.Findings {
			byCategory[f.Category] = append(byCategory[f.Category], f)
		}
	}

	// Sort categories
	cats := make([]types.Category, 0, len(byCategory))
	for cat := range byCategory {
		cats = append(cats, cat)
	}
	sort.Slice(cats, func(i, j int) bool { return string(cats[i]) < string(cats[j]) })

	for _, cat := range cats {
		findings := byCategory[cat]
		if len(findings) == 0 {
			continue
		}

		// Sort findings: CRITICAL → WARNING → INFO → OK
		sort.Slice(findings, func(i, j int) bool {
			return severityOrd(findings[i].Severity) < severityOrd(findings[j].Severity)
		})

		r.printCategoryHeader(string(cat), findings)
		for _, f := range findings {
			r.printFinding(f)
		}
	}

	r.printFooter(report)
}

func (r *Reporter) printBanner(report types.DiagnosticReport) {
	fmt.Fprintln(r.w)
	line := strings.Repeat("═", 70)
	colorBanner.Fprintln(r.w, line)

	scope := "cluster-wide"
	if report.Namespace != "" {
		scope = fmt.Sprintf("namespace: %s", report.Namespace)
	}

	colorBanner.Fprintf(r.w, "  ☸  K8S DIAGNOSTIC REPORT\n")
	colorBanner.Fprintf(r.w, "  Cluster : %s\n", report.ClusterName)
	colorBanner.Fprintf(r.w, "  Scope   : %s\n", scope)
	colorBanner.Fprintf(r.w, "  Time    : %s\n", report.StartTime.Format(time.RFC1123))
	colorBanner.Fprintln(r.w, line)
	fmt.Fprintln(r.w)
}

func (r *Reporter) printSummaryBox(report types.DiagnosticReport) {
	s := report.Summary
	colorHeader.Fprintln(r.w, "  SUMMARY")
	fmt.Fprintln(r.w)

	table := tablewriter.NewWriter(r.w)
	table.SetBorder(true)
	table.SetHeader([]string{"Severity", "Count"})
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER})

	table.Append([]string{colorCritical.Sprint("● CRITICAL"), fmt.Sprintf("%d", s.Critical)})
	table.Append([]string{colorWarning.Sprint("● WARNING"), fmt.Sprintf("%d", s.Warning)})
	table.Append([]string{colorInfo.Sprint("● INFO"), fmt.Sprintf("%d", s.Info)})
	table.Append([]string{colorOK.Sprint("● TOTAL"), fmt.Sprintf("%d", s.Total)})
	table.Render()
	fmt.Fprintln(r.w)
}

func (r *Reporter) printCategoryHeader(cat string, findings []types.Finding) {
	critCount := 0
	warnCount := 0
	for _, f := range findings {
		switch f.Severity {
		case types.SeverityCritical:
			critCount++
		case types.SeverityWarning:
			warnCount++
		}
	}

	fmt.Fprintln(r.w)
	badge := ""
	if critCount > 0 {
		badge = colorCritical.Sprintf(" [%d critical]", critCount)
	} else if warnCount > 0 {
		badge = colorWarning.Sprintf(" [%d warnings]", warnCount)
	}

	colorCategory.Fprintf(r.w, "  ▶ %s%s\n", strings.ToUpper(cat), badge)
	colorDim.Fprintln(r.w, "  "+strings.Repeat("─", 66))
}

func (r *Reporter) printFinding(f types.Finding) {
	icon, sevColor := severityStyle(f.Severity)

	fmt.Fprintf(r.w, "\n  %s ", icon)
	sevColor.Fprintf(r.w, "[%s]", f.Severity)
	fmt.Fprintf(r.w, " ")
	colorTitle.Fprintln(r.w, f.Title)

	if f.Resource != "" {
		fmt.Fprintf(r.w, "      ")
		colorDim.Fprintf(r.w, "Resource  : ")
		colorResource.Fprintln(r.w, f.Resource)
	}

	fmt.Fprintf(r.w, "      ")
	colorDim.Fprintf(r.w, "Detail    : ")
	fmt.Fprintln(r.w, f.Description)

	if f.Suggestion != "" {
		fmt.Fprintf(r.w, "      ")
		colorDim.Fprintf(r.w, "Suggestion: ")
		colorSuggest.Fprintln(r.w, f.Suggestion)
	}

	if !f.Timestamp.IsZero() {
		fmt.Fprintf(r.w, "      ")
		colorDim.Fprintf(r.w, "Last seen : %s\n", f.Timestamp.Format(time.RFC3339))
	}
}

func (r *Reporter) printFooter(report types.DiagnosticReport) {
	elapsed := report.EndTime.Sub(report.StartTime)
	fmt.Fprintln(r.w)
	line := strings.Repeat("─", 70)
	colorDim.Fprintln(r.w, "  "+line)
	colorDim.Fprintf(r.w, "  Diagnostic completed in %s\n\n", elapsed.Round(time.Millisecond))
}

// ── JSON output ───────────────────────────────────────────────────────────────

func (r *Reporter) printJSON(report types.DiagnosticReport) {
	enc := json.NewEncoder(r.w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(report)
}

// ── Markdown output ───────────────────────────────────────────────────────────

func (r *Reporter) printMarkdown(report types.DiagnosticReport) {
	scope := "Cluster-wide"
	if report.Namespace != "" {
		scope = fmt.Sprintf("Namespace: `%s`", report.Namespace)
	}

	fmt.Fprintf(r.w, "# K8s Diagnostic Report — %s\n\n", report.ClusterName)
	fmt.Fprintf(r.w, "**Scope**: %s  \n", scope)
	fmt.Fprintf(r.w, "**Time**: %s  \n\n", report.StartTime.Format(time.RFC1123))
	fmt.Fprintf(r.w, "## Summary\n\n")
	fmt.Fprintf(r.w, "| Severity | Count |\n|----------|-------|\n")
	s := report.Summary
	fmt.Fprintf(r.w, "| 🔴 CRITICAL | %d |\n| 🟡 WARNING | %d |\n| 🔵 INFO | %d |\n| Total | %d |\n\n",
		s.Critical, s.Warning, s.Info, s.Total)

	byCategory := map[types.Category][]types.Finding{}
	for _, result := range report.Results {
		for _, f := range result.Findings {
			byCategory[f.Category] = append(byCategory[f.Category], f)
		}
	}

	cats := make([]types.Category, 0, len(byCategory))
	for cat := range byCategory {
		cats = append(cats, cat)
	}
	sort.Slice(cats, func(i, j int) bool { return string(cats[i]) < string(cats[j]) })

	for _, cat := range cats {
		findings := byCategory[cat]
		if len(findings) == 0 {
			continue
		}
		fmt.Fprintf(r.w, "## %s\n\n", cat)
		for _, f := range findings {
			emoji := severityEmoji(f.Severity)
			fmt.Fprintf(r.w, "### %s %s\n\n", emoji, f.Title)
			fmt.Fprintf(r.w, "- **Resource**: `%s`\n", f.Resource)
			fmt.Fprintf(r.w, "- **Detail**: %s\n", f.Description)
			if f.Suggestion != "" {
				fmt.Fprintf(r.w, "- **Suggestion**: %s\n", f.Suggestion)
			}
			fmt.Fprintln(r.w)
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func severityOrd(s types.Severity) int {
	switch s {
	case types.SeverityCritical:
		return 0
	case types.SeverityWarning:
		return 1
	case types.SeverityInfo:
		return 2
	default:
		return 3
	}
}

func severityStyle(s types.Severity) (string, *color.Color) {
	switch s {
	case types.SeverityCritical:
		return "🔴", colorCritical
	case types.SeverityWarning:
		return "🟡", colorWarning
	case types.SeverityInfo:
		return "🔵", colorInfo
	default:
		return "✅", colorOK
	}
}

func severityEmoji(s types.Severity) string {
	switch s {
	case types.SeverityCritical:
		return "🔴"
	case types.SeverityWarning:
		return "🟡"
	case types.SeverityInfo:
		return "🔵"
	default:
		return "✅"
	}
}

// ComputeSummary tallies all findings in the report
func ComputeSummary(results []types.CheckResult) types.Summary {
	var s types.Summary
	for _, r := range results {
		for _, f := range r.Findings {
			s.Total++
			switch f.Severity {
			case types.SeverityCritical:
				s.Critical++
			case types.SeverityWarning:
				s.Warning++
			case types.SeverityInfo:
				s.Info++
			case types.SeverityOK:
				s.OK++
			}
		}
	}
	return s
}

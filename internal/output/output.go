package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yashbaddi/mxping/internal/verifier"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// Formatter handles rendering results to terminal or JSON.
type Formatter struct {
	writer   io.Writer
	asJSON   bool
	useColor bool
}

// NewFormatter creates a new output formatter.
func NewFormatter(w io.Writer, asJSON bool, disableColor bool) *Formatter {
	// Disable color if explicitly requested, NO_COLOR env is present, or not writing to stdout/file
	useColor := !disableColor && !asJSON && os.Getenv("NO_COLOR") == ""
	return &Formatter{
		writer:   w,
		asJSON:   asJSON,
		useColor: useColor,
	}
}

// PrintJSON writes the results as indented JSON.
func (f *Formatter) PrintJSON(v any) error {
	enc := json.NewEncoder(f.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintResult prints a detailed single email verification result.
func (f *Formatter) PrintResult(r verifier.Result) {
	if f.asJSON {
		_ = f.PrintJSON(r)
		return
	}

	badge := f.badge(r.Status)
	durationStr := r.Duration.Round(time.Millisecond).String()

	fmt.Fprintf(f.writer, "\n%s  %s%s%s %s(%s)%s\n",
		badge,
		f.bold(), r.Email, f.reset(),
		colorGray, durationStr, f.reset(),
	)
	fmt.Fprintf(f.writer, "   %sSummary:%s %s\n\n", f.bold(), f.reset(), r.SummaryMsg)

	// Syntax breakdown
	fmt.Fprintf(f.writer, "  %s[Syntax]%s\n", colorCyan, f.reset())
	if r.Syntax.IsValid {
		fmt.Fprintf(f.writer, "   • Status:      %sValid%s\n", colorGreen, f.reset())
		fmt.Fprintf(f.writer, "   • Local Part:  %s\n", r.Syntax.LocalPart)
		fmt.Fprintf(f.writer, "   • Domain:      %s\n", r.Syntax.Domain)
	} else {
		fmt.Fprintf(f.writer, "   • Status:      %sInvalid%s\n", colorRed, f.reset())
		fmt.Fprintf(f.writer, "   • Error:       %s\n", r.Syntax.Reason)
	}

	// DNS breakdown
	fmt.Fprintf(f.writer, "\n  %s[DNS / Mail Exchange]%s\n", colorCyan, f.reset())
	if r.DNS.HasMX || r.DNS.HasA {
		if r.DNS.HasMX {
			fmt.Fprintf(f.writer, "   • Records:     Found %d MX record(s)\n", len(r.DNS.MXRecords))
		} else {
			fmt.Fprintf(f.writer, "   • Records:     Fallback A/AAAA record\n")
		}
		for i, mx := range r.DNS.MXRecords {
			fmt.Fprintf(f.writer, "     %d. %s (priority %d)\n", i+1, mx.Host, mx.Pref)
		}
	} else if r.DNS.Reason != "" {
		fmt.Fprintf(f.writer, "   • Status:      %sNo MX/A records%s (%s)\n", colorRed, f.reset(), r.DNS.Reason)
	} else {
		fmt.Fprintf(f.writer, "   • Status:      %sNo records resolved%s\n", colorRed, f.reset())
	}

	// SMTP breakdown
	fmt.Fprintf(f.writer, "\n  %s[SMTP Handshake]%s\n", colorCyan, f.reset())
	if r.SMTP.Connected {
		fmt.Fprintf(f.writer, "   • Connected:   %sYes%s (%s)\n", colorGreen, f.reset(), r.SMTP.TargetMX)
		fmt.Fprintf(f.writer, "   • Response:    Code %d\n", r.SMTP.ResponseCode)
		if r.SMTP.ResponseMsg != "" {
			fmt.Fprintf(f.writer, "   • Message:     %s\n", strings.TrimSpace(r.SMTP.ResponseMsg))
		}
		if r.SMTP.IsCatchAll {
			fmt.Fprintf(f.writer, "   • Catch-All:   %sYes (server accepts all mailboxes)%s\n", colorYellow, f.reset())
		} else {
			fmt.Fprintf(f.writer, "   • Catch-All:   No\n")
		}
	} else if r.SMTP.PortBlocked {
		fmt.Fprintf(f.writer, "   • Status:      %sBlocked%s\n", colorRed, f.reset())
		fmt.Fprintf(f.writer, "   • Warning:     Outbound port 25 is blocked by network/ISP\n")
	} else if r.SMTP.Reason != "" {
		fmt.Fprintf(f.writer, "   • Connected:   %sNo%s\n", colorRed, f.reset())
		fmt.Fprintf(f.writer, "   • Error:       %s\n", r.SMTP.Reason)
	} else {
		fmt.Fprintf(f.writer, "   • Status:      Skipped\n")
	}

	fmt.Fprintln(f.writer)
}

// PrintBatchSummary prints a compact summary table for multiple verification results.
func (f *Formatter) PrintBatchSummary(results []verifier.Result) {
	if f.asJSON {
		_ = f.PrintJSON(results)
		return
	}

	fmt.Fprintf(f.writer, "\n%-35s %-15s %-12s %s\n", "EMAIL", "STATUS", "DURATION", "DETAILS")
	fmt.Fprintf(f.writer, "%s\n", strings.Repeat("─", 80))

	validCount := 0
	invalidCount := 0
	catchAllCount := 0
	otherCount := 0

	for _, r := range results {
		switch r.Status {
		case verifier.StatusValid:
			validCount++
		case verifier.StatusInvalid:
			invalidCount++
		case verifier.StatusCatchAll:
			catchAllCount++
		default:
			otherCount++
		}

		durationStr := r.Duration.Round(time.Millisecond).String()
		statusText := f.colorizeStatus(r.Status)
		emailTrunc := r.Email
		if len(emailTrunc) > 33 {
			emailTrunc = emailTrunc[:30] + "..."
		}

		details := r.SummaryMsg
		if len(details) > 30 {
			details = details[:27] + "..."
		}

		fmt.Fprintf(f.writer, "%-35s %-24s %-12s %s\n", emailTrunc, statusText, durationStr, details)
	}

	fmt.Fprintf(f.writer, "%s\n", strings.Repeat("─", 80))
	fmt.Fprintf(f.writer, "Total: %d | Valid: %s%d%s | Invalid: %s%d%s | Catch-All: %s%d%s | Other: %d\n\n",
		len(results),
		colorGreen, validCount, f.reset(),
		colorRed, invalidCount, f.reset(),
		colorYellow, catchAllCount, f.reset(),
		otherCount,
	)
}

func (f *Formatter) badge(status verifier.Status) string {
	if !f.useColor {
		return fmt.Sprintf("[%s]", strings.ToUpper(string(status)))
	}
	switch status {
	case verifier.StatusValid:
		return fmt.Sprintf("%s\033[42m\033[30m ✓ VALID %s", colorBold, colorReset)
	case verifier.StatusInvalid:
		return fmt.Sprintf("%s\033[41m\033[37m ✗ INVALID %s", colorBold, colorReset)
	case verifier.StatusCatchAll:
		return fmt.Sprintf("%s\033[43m\033[30m ? CATCH-ALL %s", colorBold, colorReset)
	case verifier.StatusGreylisted:
		return fmt.Sprintf("%s\033[45m\033[37m ⏳ GREYLISTED %s", colorBold, colorReset)
	case verifier.StatusBlocked:
		return fmt.Sprintf("%s\033[41m\033[37m ⊘ BLOCKED %s", colorBold, colorReset)
	default:
		return fmt.Sprintf("%s\033[47m\033[30m ❓ UNKNOWN %s", colorBold, colorReset)
	}
}

func (f *Formatter) colorizeStatus(status verifier.Status) string {
	if !f.useColor {
		return string(status)
	}
	switch status {
	case verifier.StatusValid:
		return fmt.Sprintf("%s%s%s", colorGreen, status, colorReset)
	case verifier.StatusInvalid:
		return fmt.Sprintf("%s%s%s", colorRed, status, colorReset)
	case verifier.StatusCatchAll:
		return fmt.Sprintf("%s%s%s", colorYellow, status, colorReset)
	case verifier.StatusGreylisted:
		return fmt.Sprintf("%s%s%s", colorPurple, status, colorReset)
	case verifier.StatusBlocked:
		return fmt.Sprintf("%s%s%s", colorRed, status, colorReset)
	default:
		return fmt.Sprintf("%s%s%s", colorGray, status, colorReset)
	}
}

func (f *Formatter) bold() string {
	if f.useColor {
		return colorBold
	}
	return ""
}

func (f *Formatter) reset() string {
	if f.useColor {
		return colorReset
	}
	return ""
}


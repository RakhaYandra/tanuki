// Package reporter renders human-readable CLI output.
package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	OK   = "✓"
	Warn = "!"
	Fail = "✗"
)

// ColorEnabled reports whether color output is appropriate.
// Respects NO_COLOR and dumb terminals.
func ColorEnabled() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}
	return true
}

// Reporter writes status lines to a writer.
type Reporter struct {
	Out io.Writer
}

// New returns a Reporter writing to stdout.
func New() *Reporter { return &Reporter{Out: os.Stdout} }

// Printf writes formatted output.
func (r *Reporter) Printf(format string, args ...any) {
	fmt.Fprintf(r.Out, format, args...)
}

// Line writes a symbol + message line.
func (r *Reporter) Line(symbol, msg string) {
	fmt.Fprintf(r.Out, "  %s %s\n", symbol, msg)
}

// Header writes a section header.
func (r *Reporter) Header(s string) {
	fmt.Fprintln(r.Out, strings.ToUpper(s))
}

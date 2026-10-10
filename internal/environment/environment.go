// Package environment defines shared diagnostic types for inspection commands.
package environment

// Status is a check outcome. Warn never blocks; Fail does.
type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// Check is one diagnostic result.
type Check struct {
	Group   string
	Name    string
	Status  Status
	Message string
	Hint    string
}

// Report is an ordered set of checks.
type Report struct {
	Checks []Check
}

// Counts tallies outcomes.
func (r Report) Counts() (pass, warn, fail int) {
	for _, c := range r.Checks {
		switch c.Status {
		case StatusPass:
			pass++
		case StatusWarn:
			warn++
		default:
			fail++
		}
	}
	return
}

// HasFail reports whether any check failed.
func (r Report) HasFail() bool {
	_, _, f := r.Counts()
	return f > 0
}

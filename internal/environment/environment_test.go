package environment

import "testing"

func TestCounts(t *testing.T) {
	var r Report
	if p, w, f := r.Counts(); p != 0 || w != 0 || f != 0 {
		t.Fatalf("empty = %d,%d,%d", p, w, f)
	}
	r = Report{Checks: []Check{
		{Status: StatusPass}, {Status: StatusWarn}, {Status: StatusFail},
	}}
	if p, w, f := r.Counts(); p != 1 || w != 1 || f != 1 {
		t.Fatalf("mixed = %d,%d,%d", p, w, f)
	}
	if !r.HasFail() {
		t.Fatal("expected HasFail")
	}
}

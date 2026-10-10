package reporter

import (
	"bytes"
	"strings"
	"testing"
)

func TestLine(t *testing.T) {
	var buf bytes.Buffer
	r := &Reporter{Out: &buf}
	r.Line(OK, "done")
	if got := buf.String(); !strings.Contains(got, "✓ done") {
		t.Fatalf("got %q", got)
	}
}

func TestColorRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if ColorEnabled() {
		t.Fatal("expected disabled with NO_COLOR")
	}
}

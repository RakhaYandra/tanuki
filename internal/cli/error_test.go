package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/config"
)

func writeCfg(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendCfg(t *testing.T, dir, extra string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(extra); err != nil {
		t.Fatal(err)
	}
}

func runCmdIn(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	root := NewRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	return buf.String(), root.Execute()
}

func TestRenderJSONKeys(t *testing.T) {
	c := &config.Config{Version: "1", Project: config.Project{Name: "p"}}
	out, err := render(c, "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"version"`) || !strings.Contains(out, `"project"`) {
		t.Fatalf("yaml keys missing: %s", out)
	}
	if strings.Contains(out, `"Version"`) || strings.Contains(out, `"Project"`) {
		t.Fatalf("Go names leaked: %s", out)
	}
}

func TestNoUsageOnRuntimeError(t *testing.T) {
	// validate on bad config must not print cobra Usage.
	dir := t.TempDir()
	writeCfg(t, dir, "version: \"1\"\nproject:\n  name: bad name!\n")
	out, err := runCmdIn(t, dir, "validate")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(out, "Usage:") {
		t.Fatalf("usage leaked: %q", out)
	}
}

func TestDriftWarn(t *testing.T) {
	dir := t.TempDir()
	writeCfg(t, dir, "version: \"1\"\nproject:\n  name: drift1\n")
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	out, err := runCmdIn(t, dir, "up", "--dry-run")
	if err != nil {
		t.Fatalf("dry: %v %s", err, out)
	}
	// Simulate a recorded fingerprint differing from current file.
	appendCfg(t, dir, "# drift\n")
	out, err = runCmdIn(t, dir, "status")
	if err != nil {
		t.Fatalf("status: %v %s", err, out)
	}
	// No fingerprint recorded yet (dry-run doesn't save) → no warn expected.
	if strings.Contains(out, "changed since last up") {
		t.Fatalf("unexpected drift warn: %q", out)
	}
}

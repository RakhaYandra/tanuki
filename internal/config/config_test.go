package config

import (
	"os"
	"path/filepath"
	"testing"
)

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseValid(t *testing.T) {
	for _, f := range []string{"../../testdata/valid/minimal.yaml", "../../testdata/valid/full.yaml"} {
		c, err := Parse(mustRead(t, f))
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if c.Project.Name != "pulse" {
			t.Fatalf("%s: name = %q", f, c.Project.Name)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, f := range []string{
		"../../testdata/invalid/unknown-field.yaml",
		"../../testdata/invalid/duplicate-port.yaml",
		"../../testdata/invalid/bad-version.yaml",
		"../../testdata/invalid/multiple.yaml",
	} {
		if _, err := Parse(mustRead(t, f)); err == nil {
			t.Fatalf("%s: expected error", f)
		}
	}
}

func TestParseFileUsesViper(t *testing.T) {
	c, err := ParseFile("../../testdata/valid/full.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if c.Services["postgres"].Port != 5432 {
		t.Fatalf("port = %d", c.Services["postgres"].Port)
	}
	if _, err := ParseFile("../../testdata/invalid/unknown-field.yaml"); err == nil {
		t.Fatal("expected strict error")
	}
}

func TestFindAscends(t *testing.T) {
	base := t.TempDir()
	nested := filepath.Join(base, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, FileName), []byte("version: \"1\"\nproject:\n  name: x1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Find(nested, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(base, FileName) {
		t.Fatalf("got %q", got)
	}
}

func TestFindMissing(t *testing.T) {
	if _, err := Find(t.TempDir(), ""); err == nil {
		t.Fatal("expected not found")
	}
}

func TestWriteExampleNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteExample(dir, "pulse", false); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteExample(dir, "pulse", false); err == nil {
		t.Fatal("expected overwrite guard")
	}
	if _, err := WriteExample(dir, "pulse", true); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseFile(filepath.Join(dir, FileName)); err != nil {
		t.Fatalf("generated file invalid: %v", err)
	}
}

func TestValidateDetails(t *testing.T) {
	c := &Config{Version: "1", Project: Project{Name: "Bad Name!"}}
	if err := Validate(c); err == nil {
		t.Fatal("expected name error")
	}
	c = &Config{Version: "1", Project: Project{Name: "ok-name"},
		Application: &Application{Command: []string{"go"}, DependsOn: []string{"ghost"}}}
	if err := Validate(c); err == nil {
		t.Fatal("expected depends_on error")
	}
}

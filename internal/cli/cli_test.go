package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/config"
)

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestVersion(t *testing.T) {
	out, err := runCmd(t, "version")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("empty version")
	}
}

func TestValidateOK(t *testing.T) {
	dir := t.TempDir()
	data, _ := os.ReadFile("../../testdata/valid/full.yaml")
	if err := os.WriteFile(filepath.Join(dir, config.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	if out, err := runCmd(t, "validate"); err != nil {
		t.Fatalf("validate: %v (%s)", err, out)
	}
}

func TestValidateBad(t *testing.T) {
	dir := t.TempDir()
	data, _ := os.ReadFile("../../testdata/invalid/duplicate-port.yaml")
	if err := os.WriteFile(filepath.Join(dir, config.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	if _, err := runCmd(t, "validate"); err == nil {
		t.Fatal("expected error")
	}
}

func TestInitCreatesFile(t *testing.T) {
	dir := t.TempDir()
	out, err := runCmd(t, "init", "--name", "pulse", "--dir", dir)
	if err != nil {
		t.Fatalf("init: %v (%s)", err, out)
	}
	if _, err := config.ParseFile(filepath.Join(dir, config.FileName)); err != nil {
		t.Fatalf("generated invalid: %v", err)
	}
	if _, err := runCmd(t, "init", "--name", "pulse", "--dir", dir); err == nil {
		t.Fatal("expected overwrite guard")
	}
}

func TestStatusStopped(t *testing.T) {
	dir := t.TempDir()
	min := "version: \"1\"\nproject:\n  name: stat1\n"
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(min), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	out, err := runCmd(t, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("PROJECT: stat1")) {
		t.Fatalf("missing project: %q", out)
	}
}

func TestPsEmpty(t *testing.T) {
	dir := t.TempDir()
	min := "version: \"1\"\nproject:\n  name: psempty1\n"
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(min), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	if out, err := runCmd(t, "ps"); err != nil {
		t.Fatalf("ps: %v (%s)", err, out)
	}
}

func TestLogsNoLog(t *testing.T) {
	dir := t.TempDir()
	min := "version: \"1\"\nproject:\n  name: logempty1\n"
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(min), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	if out, err := runCmd(t, "logs"); err != nil {
		t.Fatalf("logs: %v (%s)", err, out)
	}
}

func TestDoctorBadConfig(t *testing.T) {
	dir := t.TempDir()
	data, _ := os.ReadFile("../../testdata/invalid/duplicate-port.yaml")
	if err := os.WriteFile(filepath.Join(dir, config.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	out, err := runCmd(t, "doctor")
	if err == nil {
		t.Fatal("expected error")
	}
	if !bytes.Contains([]byte(out), []byte("TANUKI DOCTOR")) {
		t.Fatalf("missing header: %q", out)
	}
}

func TestUpDryRun(t *testing.T) {
	dir := t.TempDir()
	data, _ := os.ReadFile("../../testdata/valid/full.yaml")
	if err := os.WriteFile(filepath.Join(dir, config.FileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	out, err := runCmd(t, "up", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run: %v (%s)", err, out)
	}
	if !bytes.Contains([]byte(out), []byte("create")) {
		t.Fatalf("missing plan: %q", out)
	}
}

func TestDownNothing(t *testing.T) {
	dir := t.TempDir()
	min := "version: \"1\"\nproject:\n  name: nodown1\n"
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(min), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	if out, err := runCmd(t, "down"); err != nil {
		t.Fatalf("down: %v (%s)", err, out)
	}
}
func TestDoctorMinimalProject(t *testing.T) {
	dir := t.TempDir()
	min := "version: \"1\"\nproject:\n  name: doc1\n"
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(min), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfgFile = ""
	out, err := runCmd(t, "doctor")
	// Daemon may be down on CI; accept prereq fail but require grouped output.
	if !bytes.Contains([]byte(out), []byte("System")) || !bytes.Contains([]byte(out), []byte("Summary")) {
		t.Fatalf("missing groups: %q (err=%v)", out, err)
	}
}

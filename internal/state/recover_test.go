package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCorruptBackup(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	p := filepath.Join(dir, "tanuki", "c1.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Load("c1")
	if !errors.Is(err, ErrCorruptRecovered) {
		t.Fatalf("want corrupt error, got %v", err)
	}
	if s == nil || len(s.Containers) != 0 {
		t.Fatal("want fresh state")
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "tanuki"))
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "c1.json.corrupt.") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no backup: %v", entries)
	}
}

func TestFingerprint(t *testing.T) {
	f := filepath.Join(t.TempDir(), "tanuki.yaml")
	if err := os.WriteFile(f, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, err := FingerprintOf(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, _ := FingerprintOf(f)
	if a == b || len(a) != 64 {
		t.Fatal("hash broken")
	}
}

func TestPrune(t *testing.T) {
	s := &State{Containers: map[string]string{"a": "1", "b": "2"}, AppPID: 99}
	if n := s.Prune(map[string]bool{"a": true}, false); n != 2 {
		t.Fatalf("pruned %d", n)
	}
	if len(s.Containers) != 1 || s.AppPID != 0 {
		t.Fatalf("%+v", s)
	}
}

func TestLockExclusive(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	r1, err := AcquireLock("p")
	if err != nil {
		t.Fatal(err)
	}
	defer r1()
	if _, err := AcquireLock("p"); err == nil {
		t.Fatal("second lock must fail")
	}
	r1()
	if _, err := AcquireLock("p"); err != nil {
		t.Fatalf("relock must succeed: %v", err)
	}
}

package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	s, err := Load("p")
	if err != nil || len(s.Containers) != 0 {
		t.Fatalf("%v %+v", err, s)
	}
	s.Touch("postgres", "abc")
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := Load("p")
	if err != nil || got.Containers["postgres"] != "abc" {
		t.Fatalf("%v %+v", err, got)
	}
	info, _ := os.Stat(filepath.Join(Dir(), "p.json"))
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm %v", info.Mode())
	}
}

func TestMissingEmpty(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	s, err := Load("ghost")
	if err != nil || s == nil {
		t.Fatal(err)
	}
}

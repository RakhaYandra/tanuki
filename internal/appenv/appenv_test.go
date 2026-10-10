package appenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/config"
)

func TestMergeOverride(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("A=file\nB=file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env, warns, err := Load(dir, &config.Application{EnvFile: ".env"}, []string{"A=proc", "C=proc"})
	if err != nil || len(warns) != 0 {
		t.Fatalf("%v %v", err, warns)
	}
	m := map[string]string{}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	if m["A"] != "file" || m["C"] != "proc" || m["B"] != "file" {
		t.Fatalf("%v", m)
	}
}

func TestMissingFile(t *testing.T) {
	env, warns, _ := Load(t.TempDir(), &config.Application{EnvFile: ".env"}, []string{"A=1"})
	if len(warns) == 0 || len(env) != 1 {
		t.Fatalf("%v %v", warns, env)
	}
}

func TestEmptyWarnNoValues(t *testing.T) {
	dir := t.TempDir()
	secret := "zz-top-secret"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("K=\nS="+secret+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, warns, _ := Load(dir, &config.Application{EnvFile: ".env"}, nil)
	if len(warns) == 0 {
		t.Fatal("expected empty warn")
	}
	for _, w := range warns {
		if strings.Contains(w, secret) {
			t.Fatal("value leaked")
		}
	}
}

package envcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/RakhaYandra/tanuki/internal/environment"
)

func TestNoApp(t *testing.T) {
	cs := CheckEnvFile(t.TempDir(), nil)
	if cs[0].Status != environment.StatusPass {
		t.Fatalf("%+v", cs)
	}
}

func TestMissing(t *testing.T) {
	cs := CheckEnvFile(t.TempDir(), &config.Application{EnvFile: ".env"})
	if cs[0].Status != environment.StatusWarn {
		t.Fatalf("%+v", cs)
	}
}

func TestValid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=8080\n# c\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cs := CheckEnvFile(dir, &config.Application{EnvFile: ".env"})
	if cs[0].Status != environment.StatusPass {
		t.Fatalf("%+v", cs)
	}
}

func TestBadAndEmpty(t *testing.T) {
	dir := t.TempDir()
	secret := "SUPERSECRET123"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("BAD LINE HERE\nEMPTY=\nSECRET="+secret+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cs := CheckEnvFile(dir, &config.Application{EnvFile: ".env"})
	if cs[0].Status != environment.StatusWarn {
		t.Fatalf("%+v", cs)
	}
	if strings.Contains(cs[0].Message, secret) {
		t.Fatal("secret leaked")
	}
}

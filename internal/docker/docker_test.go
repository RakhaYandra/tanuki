package docker

import (
	"errors"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/environment"
	"github.com/RakhaYandra/tanuki/internal/runtime"
)

func fake(out string, err error) runtime.Runner {
	return func(n string, a ...string) (string, error) { return out, err }
}

func TestCLIok(t *testing.T) {
	c := CheckCLI(fake(`{"Client":{"Version":"29.7.2"}}`, nil))
	if c.Status != environment.StatusPass {
		t.Fatalf("%+v", c)
	}
}

func TestCLImissing(t *testing.T) {
	c := CheckCLI(fake("", errors.New("executable file not found in $PATH")))
	if c.Status != environment.StatusFail {
		t.Fatalf("%+v", c)
	}
}

func TestDaemonDown(t *testing.T) {
	c := CheckDaemon(fake("", errors.New("Cannot connect to the Docker daemon. Is the docker daemon running?")))
	if c.Status != environment.StatusFail {
		t.Fatalf("%+v", c)
	}
}

func TestDaemonPerm(t *testing.T) {
	c := CheckDaemon(fake("permission denied", errors.New("Got permission denied")))
	if c.Status != environment.StatusFail {
		t.Fatalf("%+v", c)
	}
	if c.Hint == "" {
		t.Fatal("expected hint")
	}
}

func TestDaemonOK(t *testing.T) {
	c := CheckDaemon(fake(`{"ServerVersion":"29.7.2"}`, nil))
	if c.Status != environment.StatusPass {
		t.Fatalf("%+v", c)
	}
}

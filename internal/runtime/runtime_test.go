package runtime

import (
	"errors"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/environment"
)

func okRunner(out string) Runner {
	return func(n string, a ...string) (string, error) { return out, nil }
}

func TestDetectGoOK(t *testing.T) {
	for _, req := range []string{"1.27", "1.27.1", "1.26"} {
		c := DetectGo(req, okRunner("go version go1.27.1 linux/amd64\n"))
		if c.Status != environment.StatusPass {
			t.Fatalf("req %s: %+v", req, c)
		}
	}
}

func TestDetectGoOld(t *testing.T) {
	c := DetectGo("1.27", okRunner("go version go1.21.0 linux/amd64\n"))
	if c.Status != environment.StatusFail {
		t.Fatalf("%+v", c)
	}
}

func TestDetectGoMissing(t *testing.T) {
	c := DetectGo("1.27", func(n string, a ...string) (string, error) {
		return "", errors.New("executable file not found in $PATH")
	})
	if c.Status != environment.StatusFail {
		t.Fatalf("%+v", c)
	}
}

func TestDetectGoGarbage(t *testing.T) {
	c := DetectGo("1.27", okRunner("go version devel xyz\n"))
	if c.Status != environment.StatusWarn {
		t.Fatalf("%+v", c)
	}
}

func TestDetectGoEmptyReq(t *testing.T) {
	c := DetectGo("", okRunner("whatever"))
	if c.Status != environment.StatusPass {
		t.Fatalf("%+v", c)
	}
}

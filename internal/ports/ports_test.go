package ports

import (
	"net"
	"testing"
	"time"

	"github.com/RakhaYandra/tanuki/internal/environment"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func TestFree(t *testing.T) {
	c := Check("127.0.0.1", freePort(t), 200*time.Millisecond)
	if c.Status != environment.StatusPass {
		t.Fatalf("%+v", c)
	}
}

func TestHeld(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port
	c := Check("127.0.0.1", port, 200*time.Millisecond)
	if c.Status != environment.StatusFail {
		t.Fatalf("%+v", c)
	}
}

func TestZeroSkipped(t *testing.T) {
	c := Check("127.0.0.1", 0, 200*time.Millisecond)
	if c.Status != environment.StatusPass {
		t.Fatalf("%+v", c)
	}
}

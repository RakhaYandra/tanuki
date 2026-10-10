package health

import (
	"net"
	"testing"
	"time"
)

func TestOpen(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port
	if err := WaitTCP("127.0.0.1", port, 3*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestClosed(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	start := time.Now()
	if err := WaitTCP("127.0.0.1", port, 1200*time.Millisecond); err == nil {
		t.Fatal("expected timeout")
	}
	if time.Since(start) < time.Second {
		t.Fatal("returned too fast, retries missing")
	}
}

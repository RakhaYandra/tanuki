// Package health executes readiness probes with bounded retries.
package health

import (
	"fmt"
	"net"
	"time"
)

// WaitTCP dials host:port until success or timeout. Polls every 500ms.
func WaitTCP(host string, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("service not ready after %s (port %d)", timeout, port)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// Package ports checks TCP port availability via dial. Advisory only:
// the final bind in `up` remains authoritative (TOCTOU).
package ports

import (
	"fmt"
	"net"
	"time"

	"github.com/RakhaYandra/tanuki/internal/environment"
)

// DefaultTimeout bounds each dial.
const DefaultTimeout = 500 * time.Millisecond

// InUse dials host:port; true means something answered.
func InUse(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Check returns Pass when free, Fail when occupied. Port 0 means unconfigured → Pass.
func Check(host string, port int, timeout time.Duration) environment.Check {
	c := environment.Check{Group: "Ports", Name: fmt.Sprintf("port %d", port)}
	if port == 0 {
		c.Status = environment.StatusPass
		c.Message = "no port configured, skipped"
		return c
	}
	if InUse(host, port, timeout) {
		c.Status = environment.StatusFail
		c.Message = fmt.Sprintf("port %d is already in use (advisory — final bind is authoritative)", port)
		c.Hint = "Inspect the process on port " + fmt.Sprint(port) + " or set another host port in tanuki.yaml"
		return c
	}
	c.Status = environment.StatusPass
	c.Message = fmt.Sprintf("port %d appears free", port)
	return c
}

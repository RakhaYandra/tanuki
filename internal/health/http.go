package health

import (
	"fmt"
	"net/http"
	"time"
)

// WaitHTTP polls url until 2xx or timeout.
func WaitHTTP(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(timeout)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("app not ready after %s (%s)", timeout, url)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

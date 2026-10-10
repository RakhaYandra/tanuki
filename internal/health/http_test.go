package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPok(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer s.Close()
	if err := WaitHTTP(s.URL, 3*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPfail(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer s.Close()
	if err := WaitHTTP(s.URL, 1200*time.Millisecond); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestHTTPdown(t *testing.T) {
	if err := WaitHTTP("http://127.0.0.1:1/health", time.Second); err == nil {
		t.Fatal("expected error")
	}
}

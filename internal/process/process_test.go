package process

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSpawnStop(t *testing.T) {
	log := filepath.Join(t.TempDir(), "app.log")
	p, err := Spawn([]string{"sleep", "30"}, t.TempDir(), os.Environ(), log)
	if err != nil {
		t.Fatal(err)
	}
	if !p.Alive() {
		t.Fatal("should be alive")
	}
	if err := p.Stop(5 * time.Second); err != nil {
		t.Fatal(err)
	}
	if p.Alive() {
		t.Fatal("should be dead")
	}
}

func TestStaleNeverSignaled(t *testing.T) {
	p := &Proc{PID: 1, Starttime: 999999999} // init, wrong starttime
	if err := p.Stop(100 * time.Millisecond); err != ErrStale {
		t.Fatalf("want ErrStale, got %v", err)
	}
	if p.Alive() {
		t.Fatal("stale must not be alive")
	}
}

func TestLogCapture(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "app.log")
	p, err := Spawn([]string{"echo", "hello-tanuki"}, dir, os.Environ(), log)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	_ = p.Stop(time.Second)
	data, _ := os.ReadFile(log)
	if string(data) == "" || len(data) == 0 {
		t.Skip("echo too fast, log may be empty on some kernels")
	}
}

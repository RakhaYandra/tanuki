// Package process spawns and stops detached application processes.
package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// ErrStale means the recorded PID no longer refers to our process.
var ErrStale = errors.New("stale process record (pid reused by another process)")

// Proc tracks a detached process.
type Proc struct {
	PID       int
	Starttime uint64
	Argv0     string
	LogPath   string
}

// Current reads /proc/<pid>/stat starttime and cmdline.
func Current(pid int) (starttime uint64, argv0 string, err error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, "", err
	}
	// comm may contain spaces/parens; fields after last ')' start with state.
	s := string(data)
	i := strings.LastIndex(s, ")")
	if i < 0 {
		return 0, "", errors.New("bad stat")
	}
	fields := strings.Fields(strings.TrimSpace(s[i+1:]))
	if len(fields) < 20 {
		return 0, "", errors.New("bad stat fields")
	}
	var st uint64
	if _, err := fmt.Sscan(fields[19], &st); err != nil {
		return 0, "", err
	}
	raw, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	argv0 = strings.Split(string(raw), "\x00")[0]
	return st, argv0, nil
}

// Spawn starts argv detached; stdout+stderr append to logPath.
func Spawn(argv []string, dir string, env []string, logPath string) (*Proc, error) {
	if len(argv) == 0 {
		return nil, errors.New("empty command")
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	c := exec.Command(argv[0], argv[1:]...)
	c.Dir = dir
	c.Env = env
	c.Stdout = f
	c.Stderr = f
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := c.Start(); err != nil {
		return nil, err
	}
	pid := c.Process.Pid
	_ = c.Process.Release() // detach: reparent to init, no zombie
	time.Sleep(100 * time.Millisecond)
	st, argv0, _ := Current(pid)
	return &Proc{PID: pid, Starttime: st, Argv0: argv0, LogPath: logPath}, nil
}

// Alive checks the process still runs and matches identity.
func (p *Proc) Alive() bool {
	if p == nil || p.PID <= 0 {
		return false
	}
	st, argv0, err := Current(p.PID)
	if err != nil {
		return false
	}
	if p.Starttime != 0 && st != p.Starttime {
		return false
	}
	if p.Argv0 != "" && argv0 != "" && argv0 != p.Argv0 {
		// argv0 basename compare: /tmp/x vs x after exec
		if !strings.HasSuffix(argv0, p.Argv0) && !strings.HasSuffix(p.Argv0, argv0) {
			return false
		}
	}
	// zombie check: state Z means dead but unreaped (ours are released, so Z = not ours)
	if s, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", p.PID)); err == nil {
		if i := strings.LastIndex(string(s), ")"); i > 0 {
			fields := strings.Fields(strings.TrimSpace(string(s)[i+1:]))
			if len(fields) > 0 && fields[0] == "Z" {
				return false
			}
		}
	}
	return true
}

// Stop sends SIGTERM, waits, then SIGKILL. Stale records are never signaled.
func (p *Proc) Stop(timeout time.Duration) error {
	if p == nil || p.PID <= 0 {
		return ErrStale
	}
	st, _, err := Current(p.PID)
	if err != nil {
		return ErrStale
	}
	if p.Starttime != 0 && st != p.Starttime {
		return ErrStale
	}
	proc, err := os.FindProcess(p.PID)
	if err != nil {
		return ErrStale
	}
	_ = proc.Signal(syscall.SIGTERM)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		if !(&Proc{PID: p.PID, Starttime: st}).Alive() {
			return nil
		}
	}
	_ = proc.Signal(syscall.SIGKILL)
	return nil
}

package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/RakhaYandra/tanuki/internal/appenv"
	"github.com/RakhaYandra/tanuki/internal/config"
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"github.com/RakhaYandra/tanuki/internal/health"
	"github.com/RakhaYandra/tanuki/internal/ports"
	"github.com/RakhaYandra/tanuki/internal/process"
	"github.com/RakhaYandra/tanuki/internal/runtime"
	"github.com/RakhaYandra/tanuki/internal/state"
)

// AppLogName is the log file inside the state dir.
func AppLogName(project string) string { return filepath.Join(state.Dir(), project+"-app.log") }

// SpawnFunc starts the app; replaceable in tests.
type SpawnFunc func(argv []string, dir string, env []string, log string) (*process.Proc, error)

// StartApp starts the app detached, or reuses a live recorded process.
// Returns proc, reused?, warnings, error.
func StartApp(cfg *config.Config, projectDir string, st *state.State, spawn SpawnFunc) (*process.Proc, bool, []string, error) {
	app := cfg.Application
	if app == nil {
		return nil, false, nil, nil
	}
	if st.AppPID > 0 {
		cur := &process.Proc{PID: st.AppPID, Starttime: st.AppStart}
		if cur.Alive() {
			return cur, true, nil, nil
		}
	}
	env, warns, _ := appenv.Load(projectDir, app, os.Environ())
	if app.Port != 0 && ports.InUse("127.0.0.1", app.Port, ports.DefaultTimeout) {
		return nil, false, warns, tanukierr.New(tanukierr.CodeConflict,
			fmt.Sprintf("port %d is already in use.\n\nTanuki did not stop or modify the existing process.", app.Port))
	}
	logPath := AppLogName(cfg.Project.Name)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return nil, false, warns, tanukierr.Wrap(tanukierr.CodeState, "cannot create log dir", err)
	}
	p, err := spawn(app.Command, projectDir, env, logPath)
	if err != nil {
		return nil, false, warns, tanukierr.Wrap(tanukierr.CodeStartup, "cannot start application", err)
	}
	// Early-exit detection: dead within 2s means bind/config failure.
	time.Sleep(2 * time.Second)
	if !p.Alive() {
		tail := tailFile(p.LogPath, 20)
		msg := "application exited immediately; check command and logs at " + p.LogPath
		if tail != "" {
			msg += "\n--- log tail ---\n" + tail
		}
		return nil, false, warns, tanukierr.New(tanukierr.CodeStartup, msg)
	}
	st.AppPID = p.PID
	st.AppStart = p.Starttime
	st.AppLog = p.LogPath
	if err := st.Save(); err != nil {
		return p, false, warns, err
	}
	return p, false, warns, nil
}

// WaitApp blocks until the app health check passes.
func WaitApp(cfg *config.Config) error {
	app := cfg.Application
	if app == nil || app.Health == nil {
		return nil
	}
	timeout := app.Health.TimeoutDuration()
	if app.Health.Type == "http" {
		url := fmt.Sprintf("http://127.0.0.1:%d%s", app.Port, app.Health.Path)
		return health.WaitHTTP(url, timeout)
	}
	return health.WaitTCP("127.0.0.1", app.Port, timeout)
}

// StopApp gracefully stops the recorded app process.
func StopApp(project string, run runtime.Runner) (bool, error) {
	st, err := state.Load(project)
	if err != nil && err != state.ErrCorruptRecovered {
		return false, err
	}
	if st.AppPID == 0 {
		return false, nil
	}
	p := &process.Proc{PID: st.AppPID, Starttime: st.AppStart}
	if err := p.Stop(10 * time.Second); err == process.ErrStale {
		st.ClearApp()
		_ = st.Save()
		return false, nil // stale record cleared, nothing signaled
	} else if err != nil {
		return false, tanukierr.Wrap(tanukierr.CodeShutdown, "cannot stop application", err)
	}
	st.ClearApp()
	_ = st.Save()
	return true, nil
}

func tailFile(path string, n int) string {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return ""
	}
	lines := splitLines(string(data))
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

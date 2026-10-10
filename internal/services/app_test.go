package services

import (
	"os"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/RakhaYandra/tanuki/internal/process"
	"github.com/RakhaYandra/tanuki/internal/state"
)

func appCfg(port int) *config.Config {
	return &config.Config{
		Version: "1", Project: config.Project{Name: "apptest"},
		Application: &config.Application{Name: "api", Command: []string{"sleep", "30"}, Port: port},
	}
}

func TestStartReuse(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	st, _ := state.Load("apptest")
	log := dir + "/app.log"
	p, err := process.Spawn([]string{"sleep", "30"}, dir, os.Environ(), log)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Stop(2000000000) }()
	st.AppPID = p.PID
	st.AppStart = p.Starttime
	if err := st.Save(); err != nil {
		t.Fatal(err)
	}
	calls := 0
	spawn := func(argv []string, d string, e []string, l string) (*process.Proc, error) {
		calls++
		return nil, nil
	}
	got, reused, _, err := StartApp(appCfg(0), dir, st, spawn)
	if err != nil || !reused || got == nil {
		t.Fatalf("%v %v %+v", err, reused, got)
	}
	if calls != 0 {
		t.Fatal("must not spawn when alive")
	}
}

func TestStartStaleRespawns(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	st, _ := state.Load("apptest")
	st.AppPID = 1
	st.AppStart = 999999999
	calls := 0
	spawn := func(argv []string, d string, e []string, l string) (*process.Proc, error) {
		calls++
		return process.Spawn([]string{"sleep", "30"}, d, e, l)
	}
	p, reused, _, err := StartApp(appCfg(0), dir, st, spawn)
	if err != nil || reused {
		t.Fatalf("%v %v", err, reused)
	}
	defer func() { _ = p.Stop(2000000000) }()
	if calls != 1 {
		t.Fatal("must spawn on stale")
	}
}

func TestStopAppStale(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	st, _ := state.Load("apptest")
	st.AppPID = 1
	st.AppStart = 999999999
	_ = st.Save()
	stopped, err := StopApp("apptest", nil)
	if err != nil || stopped {
		t.Fatalf("%v %v", err, stopped)
	}
}

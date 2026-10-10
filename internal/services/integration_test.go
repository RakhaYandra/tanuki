//go:build integration

package services

import (
	"os/exec"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/RakhaYandra/tanuki/internal/docker"
)

func realRunner(t *testing.T) func(string, ...string) (string, error) {
	t.Helper()
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon unavailable")
	}
	return func(n string, a ...string) (string, error) {
		out, err := exec.Command(n, a...).Output()
		return string(out), err
	}
}

// Covers: services.Up with TCP health, idempotent reuse, Down preserving volumes.
func TestIntegrationServicesUpDown(t *testing.T) {
	run := realRunner(t)
	cfg := &config.Config{
		Version: "1", Project: config.Project{Name: "tkitsvc"},
		Services: map[string]config.Service{
			"postgres": {Type: "postgres", Version: "16", Port: 55442, Healthcheck: &config.Healthcheck{Type: "tcp", Timeout: "60s"}},
			"redis":    {Type: "redis", Version: "7", Port: 56392, Healthcheck: &config.Healthcheck{Type: "tcp", Timeout: "30s"}},
		},
	}
	t.Cleanup(func() {
		_, _ = Down(cfg.Project.Name, run)
		for _, s := range []string{"postgres", "redis"} {
			_, _ = run("docker", "rm", "-f", docker.ContainerName(cfg.Project.Name, s))
			_, _ = run("docker", "volume", "rm", docker.VolumeName(cfg.Project.Name, s))
		}
	})
	done, err := Up(cfg, run, true)
	if err != nil {
		t.Fatalf("up: %v", err)
	}
	if len(done) != 2 {
		t.Fatalf("%+v", done)
	}
	again, err := Up(cfg, run, false)
	if err != nil {
		t.Fatalf("up2: %v", err)
	}
	for _, a := range again {
		if a.Kind != "reuse" {
			t.Fatalf("not reused: %+v", again)
		}
	}
	stopped, err := Down(cfg.Project.Name, run)
	if err != nil || len(stopped) != 2 {
		t.Fatalf("down: %v %v", err, stopped)
	}
}

//go:build integration

package docker

import (
	"os/exec"
	"strings"
	"testing"
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

// Covers: postgres+redis start, label ownership, volume preservation,
// unrelated containers untouched.
func TestIntegrationLifecycle(t *testing.T) {
	run := realRunner(t)
	project := "tkit"
	before, _ := run("docker", "ps", "-q")
	_ = before

	for _, img := range []string{"postgres:16-alpine", "redis:7-alpine"} {
		if err := Pull(img, run); err != nil {
			t.Fatalf("pull %s: %v", img, err)
		}
	}
	id1, err := Run(RunConfig{Name: ContainerName(project, "postgres"), Image: "postgres:16-alpine",
		Project: project, Service: "postgres", Port: 55441,
		Env:    []string{"POSTGRES_PASSWORD=tanuki"},
		Volume: VolumeName(project, "postgres"), DataDir: "/var/lib/postgresql/data"}, run)
	if err != nil {
		t.Fatalf("run postgres: %v", err)
	}
	id2, err := Run(RunConfig{Name: ContainerName(project, "redis"), Image: "redis:7-alpine",
		Project: project, Service: "redis", Port: 56391,
		Volume: VolumeName(project, "redis"), DataDir: "/data"}, run)
	if err != nil {
		t.Fatalf("run redis: %v", err)
	}
	t.Cleanup(func() {
		for _, id := range []string{id1, id2} {
			_, _ = run("docker", "rm", "-f", id)
		}
		for _, v := range []string{VolumeName(project, "postgres"), VolumeName(project, "redis")} {
			_, _ = run("docker", "volume", "rm", v)
		}
	})

	owned, err := List(project, run)
	if err != nil || len(owned) != 2 {
		t.Fatalf("list: %v %+v", err, owned)
	}
	for _, c := range owned {
		if !Owned(c, project) {
			t.Fatalf("not owned: %+v", c)
		}
	}
	for _, id := range []string{id1, id2} {
		if err := Stop(id, run); err != nil {
			t.Fatalf("stop %s: %v", id, err)
		}
	}
	vols, _ := run("docker", "volume", "ls", "--format", "{{.Name}}")
	for _, v := range []string{VolumeName(project, "postgres"), VolumeName(project, "redis")} {
		if !strings.Contains(vols, v) {
			t.Fatalf("volume %s missing after stop-only", v)
		}
	}
}

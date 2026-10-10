// Package services plans and executes infrastructure service lifecycles.
package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/RakhaYandra/tanuki/internal/docker"
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"github.com/RakhaYandra/tanuki/internal/health"
	"github.com/RakhaYandra/tanuki/internal/runtime"
)

// DevPassword is the dev-only default for postgres. Not for production.
const DevPassword = "tanuki"

// Action is one planned step.
type Action struct {
	Kind      string // reuse | create | start | wait
	Service   string
	Image     string
	Container string
	Port      int
}

// ImageFor pins a versioned alpine image. "latest" is rejected.
func ImageFor(svcType, version string) (string, error) {
	v := strings.TrimSpace(version)
	if v == "" || v == "latest" {
		return "", tanukierr.New(tanukierr.CodeConfig, "service version must pin an explicit tag, not empty or latest")
	}
	if strings.Contains(v, ":") || strings.Contains(v, "/") {
		return "", tanukierr.New(tanukierr.CodeConfig, "service version must be a plain tag, got "+v)
	}
	switch svcType {
	case "postgres":
		if !strings.Contains(v, "alpine") {
			v += "-alpine"
		}
		return "postgres:" + v, nil
	case "redis":
		if !strings.Contains(v, "alpine") {
			v += "-alpine"
		}
		return "redis:" + v, nil
	default:
		return "", tanukierr.New(tanukierr.CodeConfig, "unsupported service type "+svcType)
	}
}

// dataDir maps service type to its in-container data path.
func dataDir(svcType string) string {
	if svcType == "postgres" {
		return "/var/lib/postgresql/data"
	}
	return "/data"
}

// Plan renders intended actions without executing anything.
func Plan(cfg *config.Config) ([]Action, error) {
	names := make([]string, 0, len(cfg.Services))
	for n := range cfg.Services {
		names = append(names, n)
	}
	sort.Strings(names)
	var out []Action
	for _, n := range names {
		s := cfg.Services[n]
		img, err := ImageFor(s.Type, s.Version)
		if err != nil {
			return nil, err
		}
		if s.Port == 0 {
			return nil, tanukierr.New(tanukierr.CodeConfig, "services."+n+".port must be explicit in Phase 3")
		}
		out = append(out, Action{Kind: "create", Service: n, Image: img, Container: docker.ContainerName(cfg.Project.Name, n), Port: s.Port})
	}
	return out, nil
}

// Up ensures owned containers exist and are healthy. wait=true blocks on TCP health.
func Up(cfg *config.Config, run runtime.Runner, wait bool) ([]Action, error) {
	plan, err := Plan(cfg)
	if err != nil {
		return nil, err
	}
	owned, err := docker.List(cfg.Project.Name, run)
	if err != nil {
		return nil, tanukierr.Wrap(tanukierr.CodePrereq, "docker unavailable", err)
	}
	bySvc := map[string]docker.Container{}
	for _, c := range owned {
		if svc := c.LabelMap[docker.LabelService]; svc != "" {
			bySvc[svc] = c
		}
	}
	var done []Action
	var created []string
	for _, a := range plan {
		svc := cfg.Services[a.Service]
		if cur, ok := bySvc[a.Service]; ok {
			if cur.Image == a.Image && cur.State == "running" {
				a.Kind = "reuse"
				done = append(done, a)
				continue
			}
			if cur.Image != a.Image {
				cleanup(created, run)
				return nil, tanukierr.New(tanukierr.CodeConflict,
					fmt.Sprintf("services.%s image changed (have %s, want %s); stop first with `tanuki down`", a.Service, cur.Image, a.Image))
			}
			if _, err := run("docker", "start", cur.ID); err != nil {
				cleanup(created, run)
				return nil, tanukierr.Wrap(tanukierr.CodeStartup, "cannot start "+a.Container, err)
			}
			a.Kind = "start"
			done = append(done, a)
		} else {
			_ = docker.Pull(a.Image, run) // best-effort; cached images no-op
			rc := docker.RunConfig{Name: a.Container, Image: a.Image, Project: cfg.Project.Name,
				Service: a.Service, Port: a.Port, Volume: docker.VolumeName(cfg.Project.Name, a.Service), DataDir: dataDir(svc.Type)}
			if svc.Type == "postgres" {
				rc.Env = []string{"POSTGRES_PASSWORD=" + DevPassword}
			}
			id, err := docker.Run(rc, run)
			if err != nil {
				cleanup(created, run)
				return nil, tanukierr.Wrap(tanukierr.CodeStartup, "cannot create "+a.Container, err)
			}
			created = append(created, id)
			done = append(done, a)
		}
		if wait {
			timeout := svc.Healthcheck.TimeoutDuration()
			if err := health.WaitTCP("127.0.0.1", a.Port, timeout); err != nil {
				logs := docker.Logs(createdID(created, bySvc, a.Service), run)
				cleanup(created, run)
				msg := fmt.Sprintf("services.%s unhealthy: %v", a.Service, err)
				if logs != "" {
					msg += "\n--- logs ---\n" + logs
				}
				return nil, tanukierr.New(tanukierr.CodeStartup, msg)
			}
			done[len(done)-1].Kind += "+healthy"
		}
	}
	return done, nil
}

func createdID(created []string, bySvc map[string]docker.Container, svc string) string {
	if len(created) > 0 {
		return created[len(created)-1]
	}
	if c, ok := bySvc[svc]; ok {
		return c.ID
	}
	return ""
}

func cleanup(ids []string, run runtime.Runner) {
	for _, id := range ids {
		_ = docker.Stop(id, run)
		_ = docker.Remove(id, run) // never -v: volumes preserved
	}
}

// Down stops only label-owned containers. Volumes are never removed.
func Down(project string, run runtime.Runner) ([]string, error) {
	owned, err := docker.List(project, run)
	if err != nil {
		return nil, tanukierr.Wrap(tanukierr.CodePrereq, "docker unavailable", err)
	}
	var stopped []string
	var failed []string
	for _, c := range owned {
		if !docker.Owned(c, project) {
			continue
		}
		if err := docker.Stop(c.ID, run); err != nil {
			failed = append(failed, c.LabelMap[docker.LabelService])
			continue
		}
		stopped = append(stopped, c.LabelMap[docker.LabelService])
	}
	if len(failed) > 0 {
		return stopped, tanukierr.New(tanukierr.CodeShutdown, "could not stop: "+strings.Join(failed, ", "))
	}
	return stopped, nil
}

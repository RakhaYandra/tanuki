package services

import (
	"strings"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/config"
)

func testCfg() *config.Config {
	return &config.Config{
		Version: "1", Project: config.Project{Name: "p"},
		Services: map[string]config.Service{
			"postgres": {Type: "postgres", Version: "16", Port: 5432},
			"redis":    {Type: "redis", Version: "7", Port: 6379},
		},
	}
}

func TestImageFor(t *testing.T) {
	img, err := ImageFor("postgres", "16")
	if err != nil || img != "postgres:16-alpine" {
		t.Fatalf("%v %q", err, img)
	}
	if _, err := ImageFor("postgres", "latest"); err == nil {
		t.Fatal("latest must fail")
	}
}

func TestPlan(t *testing.T) {
	acts, err := Plan(testCfg())
	if err != nil {
		t.Fatal(err)
	}
	if len(acts) != 2 || acts[0].Service != "postgres" {
		t.Fatalf("%+v", acts)
	}
}

func TestUpReuse(t *testing.T) {
	var runs []string
	fake := func(n string, a ...string) (string, error) {
		runs = append(runs, n+" "+strings.Join(a, " "))
		if a[0] == "ps" {
			return "{\"ID\":\"a\",\"Name\":\"tanuki-p-postgres\",\"Image\":\"postgres:16-alpine\",\"State\":\"running\",\"Labels\":\"io.tanuki.managed=true,io.tanuki.project=p,io.tanuki.service=postgres\"}\n" +
				"{\"ID\":\"b\",\"Name\":\"tanuki-p-redis\",\"Image\":\"redis:7-alpine\",\"State\":\"running\",\"Labels\":\"io.tanuki.managed=true,io.tanuki.project=p,io.tanuki.service=redis\"}\n", nil
		}
		return "", nil
	}
	acts, err := Up(testCfg(), fake, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range acts {
		if a.Kind != "reuse" {
			t.Fatalf("%+v", acts)
		}
	}
	for _, r := range runs {
		if strings.Contains(r, "docker run") {
			t.Fatalf("must not create: %v", runs)
		}
	}
}

func TestDown(t *testing.T) {
	var stops []string
	fake := func(n string, a ...string) (string, error) {
		if a[0] == "ps" {
			return "{\"ID\":\"a\",\"Name\":\"x\",\"Image\":\"i\",\"State\":\"running\",\"Labels\":\"io.tanuki.managed=true,io.tanuki.project=p,io.tanuki.service=postgres\"}\n", nil
		}
		if a[0] == "stop" {
			stops = append(stops, a[1])
		}
		return "", nil
	}
	got, err := Down("p", fake)
	if err != nil || len(got) != 1 || len(stops) != 1 {
		t.Fatalf("%v %v %v", err, got, stops)
	}
}

func TestDownEmpty(t *testing.T) {
	fake := func(n string, a ...string) (string, error) { return "", nil }
	got, err := Down("p", fake)
	if err != nil || len(got) != 0 {
		t.Fatalf("%v %v", err, got)
	}
}

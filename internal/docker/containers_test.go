package docker

import (
	"strings"
	"testing"

	"github.com/RakhaYandra/tanuki/internal/runtime"
)

func TestListParse(t *testing.T) {
	fake := func(n string, a ...string) (string, error) {
		return "{\"ID\":\"abc\",\"Name\":\"tanuki-p-postgres\",\"Image\":\"postgres:16-alpine\",\"State\":\"running\",\"Labels\":\"\"}\n" +
			"{\"ID\":\"def\",\"Name\":\"tanuki-p-redis\",\"Image\":\"redis:7-alpine\",\"State\":\"running\"}\n", nil
	}
	cs, err := List("p", fake)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 2 || cs[0].ID != "abc" {
		t.Fatalf("%+v", cs)
	}
	// filter args must carry labels
	var gotArgs []string
	fake2 := func(n string, a ...string) (string, error) { gotArgs = a; return "", nil }
	_, _ = List("pulse", fake2)
	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, LabelProject+"=pulse") || !strings.Contains(joined, LabelManaged+"=true") {
		t.Fatalf("missing label filters: %v", gotArgs)
	}
	_ = runtime.DefaultRunner
}

func TestRunArgs(t *testing.T) {
	var got []string
	fake := func(n string, a ...string) (string, error) { got = a; return "id123\n", nil }
	id, err := Run(RunConfig{Name: "tanuki-p-postgres", Image: "postgres:16-alpine", Project: "p", Service: "postgres", Port: 5432, Env: []string{"POSTGRES_PASSWORD=***"}, Volume: "v", DataDir: "/data"}, fake)
	if err != nil || id != "id123" {
		t.Fatalf("%v %q", err, id)
	}
	j := strings.Join(got, " ")
	for _, want := range []string{LabelManaged + "=true", LabelProject + "=p", LabelService + "=postgres", "127.0.0.1:5432:5432", "-v", "POSTGRES_PASSWORD"} {
		if !strings.Contains(j, want) {
			t.Fatalf("missing %q in %v", want, got)
		}
	}
}

func TestStopEmpty(t *testing.T) {
	if err := Stop("", func(n string, a ...string) (string, error) { return "", nil }); err == nil {
		t.Fatal("expected error")
	}
}

func TestOwned(t *testing.T) {
	c := Container{}
	c.Labels = LabelManaged + "=true," + LabelProject + "=p"
	c.fillLabels()
	if !Owned(c, "p") || Owned(c, "q") {
		t.Fatal("ownership wrong")
	}
	empty := Container{}
	empty.fillLabels()
	if Owned(empty, "p") {
		t.Fatal("unlabeled must not be owned")
	}
}

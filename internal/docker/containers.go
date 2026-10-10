// Container lifecycle via Docker CLI structured JSON output.
package docker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/RakhaYandra/tanuki/internal/runtime"
)

const (
	LabelManaged = "io.tanuki.managed"
	LabelProject = "io.tanuki.project"
	LabelService = "io.tanuki.service"
)

// Container is a subset of docker ps fields. Labels arrives as
// "k=v,k2=v2" from --format json, parsed into LabelMap.
type Container struct {
	ID       string            `json:"ID"`
	Name     string            `json:"Names"`
	Image    string            `json:"Image"`
	State    string            `json:"State"`
	Labels   string            `json:"Labels"`
	LabelMap map[string]string `json:"-"`
}

// fillLabels parses the Labels string into LabelMap.
func (c *Container) fillLabels() {
	c.LabelMap = map[string]string{}
	for _, kv := range strings.Split(c.Labels, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(kv), "=")
		if ok && k != "" {
			c.LabelMap[k] = v
		}
	}
}

// RunConfig describes one container to create.
type RunConfig struct {
	Name    string
	Image   string
	Project string
	Service string
	Port    int
	Env     []string
	Volume  string
	DataDir string
}

// ContainerName returns the stable name for a project service.
func ContainerName(project, service string) string {
	return "tanuki-" + project + "-" + service
}

// VolumeName returns the named volume for a project service.
func VolumeName(project, service string) string {
	return "tanuki-" + project + "-" + service + "-data"
}

// List returns Tanuki-owned containers for a project (label filter).
func List(project string, run runtime.Runner) ([]Container, error) {
	out, err := run("docker", "ps", "-a",
		"--filter", "label="+LabelManaged+"=true",
		"--filter", "label="+LabelProject+"="+project,
		"--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	var res []Container
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var c Container
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			continue
		}
		c.fillLabels()
		res = append(res, c)
	}
	return res, nil
}

// Owned verifies labels before any mutation.
func Owned(c Container, project string) bool {
	return c.LabelMap[LabelManaged] == "true" && c.LabelMap[LabelProject] == project
}

// Run creates and starts a container. All args explicit, no shell.
func Run(cfg RunConfig, run runtime.Runner) (string, error) {
	if cfg.Name == "" || cfg.Image == "" {
		return "", errors.New("name and image required")
	}
	args := []string{"run", "-d", "--name", cfg.Name,
		"--label", LabelManaged + "=true",
		"--label", LabelProject + "=" + cfg.Project,
		"--label", LabelService + "=" + cfg.Service,
	}
	if cfg.Port != 0 {
		args = append(args, "-p", fmt.Sprintf("127.0.0.1:%d:%d", cfg.Port, cfg.Port))
	}
	if cfg.Volume != "" && cfg.DataDir != "" {
		args = append(args, "-v", cfg.Volume+":"+cfg.DataDir)
	}
	for _, e := range cfg.Env {
		args = append(args, "-e", e)
	}
	args = append(args, cfg.Image)
	out, err := run("docker", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Stop stops a container by id.
func Stop(id string, run runtime.Runner) error {
	if id == "" {
		return errors.New("id required")
	}
	_, err := run("docker", "stop", id)
	return err
}

// Remove removes a container without volumes.
func Remove(id string, run runtime.Runner) error {
	if id == "" {
		return errors.New("id required")
	}
	_, err := run("docker", "rm", id)
	return err
}

// Logs tails container logs for diagnostics.
func Logs(id string, run runtime.Runner) string {
	out, err := run("docker", "logs", "--tail", "50", id)
	if err != nil {
		return ""
	}
	return out
}

// Pull best-effort fetches an image; cached images make this a no-op.
func Pull(image string, run runtime.Runner) error {
	_, err := run("docker", "pull", image)
	return err
}

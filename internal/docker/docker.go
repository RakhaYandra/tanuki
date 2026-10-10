// Package docker checks Docker CLI and daemon via structured --format json output.
package docker

import (
	"encoding/json"
	"strings"

	"github.com/RakhaYandra/tanuki/internal/environment"
	"github.com/RakhaYandra/tanuki/internal/runtime"
)

type versionOut struct {
	Client struct {
		Version string `json:"Version"`
	} `json:"Client"`
	Server *struct {
		Version string `json:"Version"`
	} `json:"Server"`
}

// CheckCLI verifies the docker binary exists and reports its version.
func CheckCLI(run runtime.Runner) environment.Check {
	c := environment.Check{Group: "Dependencies", Name: "Docker CLI"}
	out, err := run("docker", "version", "--format", "{{json .}}")
	if err != nil {
		c.Status = environment.StatusFail
		c.Message = "Docker CLI not found or not executable"
		c.Hint = "Install Docker Engine: https://docs.docker.com/engine/install/"
		return c
	}
	var v versionOut
	if jerr := json.Unmarshal([]byte(out), &v); jerr != nil || v.Client.Version == "" {
		c.Status = environment.StatusWarn
		c.Message = "Docker CLI found but version unreadable"
		return c
	}
	c.Status = environment.StatusPass
	c.Message = "Docker CLI found (" + v.Client.Version + ")"
	return c
}

// CheckDaemon verifies the daemon answers. Distinct hints per failure class.
func CheckDaemon(run runtime.Runner) environment.Check {
	c := environment.Check{Group: "Dependencies", Name: "Docker daemon"}
	out, err := run("docker", "info", "--format", "{{json .}}")
	if err != nil {
		msg := err.Error() + " " + out
		c.Status = environment.StatusFail
		switch {
		case strings.Contains(msg, "permission denied"):
			c.Message = "Docker daemon unreachable: permission denied"
			c.Hint = "Add your user to the docker group or run with sufficient rights; never chmod the socket blindly"
		case strings.Contains(msg, "Cannot connect"), strings.Contains(msg, "connection refused"), strings.Contains(msg, "no such host"), strings.Contains(msg, "Is the docker daemon running"):
			c.Message = "Docker daemon not running or unreachable"
			c.Hint = "Start Docker (systemctl start docker) then retry"
		default:
			c.Message = "Docker daemon check failed"
			c.Hint = "Run `docker info` manually for details"
		}
		return c
	}
	c.Status = environment.StatusPass
	c.Message = "Docker daemon available"
	return c
}

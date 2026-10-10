package runtime

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/RakhaYandra/tanuki/internal/environment"
	"golang.org/x/mod/semver"
)

var goOutRe = regexp.MustCompile(`go(\d+\.\d+(?:\.\d+)?)`)

// Runner executes a binary without shell. Injectable for tests.
type Runner func(name string, args ...string) (string, error)

// DefaultRunner uses os/exec directly.
func DefaultRunner(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

// ParseGoVersion extracts major.minor and full version from `go version` output.
func ParseGoVersion(out string) (minor, full string, err error) {
	m := goOutRe.FindStringSubmatch(out)
	if m == nil {
		return "", "", errBadOutput
	}
	full = m[1]
	parts := strings.Split(full, ".")
	minor = strings.Join(parts[:2], ".")
	return minor, full, nil
}

type goErr string

func (e goErr) Error() string { return string(e) }

const errBadOutput = goErr("unrecognized go version output")

// DetectGo checks the installed Go toolchain against required ("1.27" or "1.27.1").
func DetectGo(required string, run Runner) environment.Check {
	base := environment.Check{Group: "Dependencies", Name: "Go"}
	if required == "" {
		base.Status = environment.StatusPass
		base.Message = "no Go requirement configured"
		return base
	}
	out, err := run("go", "version")
	if err != nil {
		base.Status = environment.StatusFail
		base.Message = "Go not found or not executable"
		base.Hint = "Install Go 1.27+ (mise install or apt install golang), ensure go is in PATH"
		return base
	}
	minor, full, perr := ParseGoVersion(out)
	if perr != nil {
		base.Status = environment.StatusWarn
		base.Message = "Go found but version output unrecognized"
		base.Hint = "Check `go version` manually; required " + required
		return base
	}
	norm := func(v string) string {
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		if c := strings.Count(v, "."); c == 1 {
			v += ".0"
		}
		return v
	}
	_ = minor
	if semver.Compare(norm(full), norm(required)) < 0 {
		base.Status = environment.StatusFail
		base.Message = "Go " + full + " does not satisfy required " + required
		base.Hint = "Install Go " + required + " or adjust runtime.go in tanuki.yaml"
		return base
	}
	base.Status = environment.StatusPass
	base.Message = "Go " + full + " satisfies required " + required
	return base
}

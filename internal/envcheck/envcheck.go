// Package envcheck validates the application env_file presence and syntax.
// Never returns secret values, only names and line numbers.
package envcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/RakhaYandra/tanuki/internal/environment"
	"github.com/subosito/gotenv"
)

// CheckEnvFile validates projectDir/app.EnvFile. Missing file is Warn
// (schema keeps env_file optional in Phase 2); malformed lines are Warn.
func CheckEnvFile(projectDir string, app *config.Application) []environment.Check {
	name := "env_file"
	if app == nil || strings.TrimSpace(app.EnvFile) == "" {
		return []environment.Check{{
			Group: "Environment", Name: name,
			Status:  environment.StatusPass,
			Message: "no env_file configured",
		}}
	}
	mk := func(s environment.Status, msg, hint string) []environment.Check {
		return []environment.Check{{Group: "Environment", Name: name, Status: s, Message: msg, Hint: hint}}
	}
	abs := filepath.Join(projectDir, app.EnvFile)
	data, err := os.ReadFile(abs)
	if err != nil {
		return mk(environment.StatusWarn, "env_file "+app.EnvFile+" not found or unreadable",
			"Create it from .env.example or remove env_file from tanuki.yaml")
	}
	var warns []string
	var empty []string
	for i, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		t = strings.TrimPrefix(t, "export ")
		if !strings.Contains(t, "=") {
			warns = append(warns, fmt.Sprintf("line %d: missing =", i+1))
			continue
		}
		k, v, _ := strings.Cut(t, "=")
		if strings.TrimSpace(k) == "" {
			warns = append(warns, fmt.Sprintf("line %d: empty key", i+1))
			continue
		}
		if strings.TrimSpace(v) == "" {
			empty = append(empty, strings.TrimSpace(k))
		}
	}
	// Full parse via gotenv to catch quoting errors; values discarded.
	if _, perr := gotenv.StrictParse(strings.NewReader(string(data))); perr != nil {
		warns = append(warns, "parse: "+firstLine(perr.Error()))
	}
	if len(warns) > 0 {
		return mk(environment.StatusWarn, "env_file "+app.EnvFile+" has "+fmt.Sprint(len(warns))+" issue(s): "+strings.Join(warns, "; "),
			"Fix listed lines; values are never shown here")
	}
	if len(empty) > 0 {
		return mk(environment.StatusWarn, "env_file "+app.EnvFile+" has empty value(s): "+strings.Join(empty, ", "),
			"Empty values are not substituted; set them or remove the keys")
	}
	return mk(environment.StatusPass, "env_file "+app.EnvFile+" readable", "")
}

func firstLine(s string) string {
	if i := strings.Index(s, "\n"); i >= 0 {
		return s[:i]
	}
	return s
}

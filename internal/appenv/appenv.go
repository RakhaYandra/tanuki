// Package appenv builds the application environment from process + .env file.
package appenv

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/subosito/gotenv"
)

// Load merges os.Environ with projectDir/app.EnvFile. File wins on clash.
// Returns merged env, warning strings (names only), error.
func Load(projectDir string, app *config.Application, base []string) ([]string, []string, error) {
	merged := map[string]string{}
	order := []string{}
	for _, kv := range base {
		k, v, _ := strings.Cut(kv, "=")
		if _, ok := merged[k]; !ok {
			order = append(order, k)
		}
		merged[k] = v
	}
	if app == nil || strings.TrimSpace(app.EnvFile) == "" {
		return joinEnv(order, merged), nil, nil
	}
	data, err := os.ReadFile(filepath.Join(projectDir, app.EnvFile))
	if err != nil {
		return joinEnv(order, merged), []string{"env_file " + app.EnvFile + " unreadable, using process env only"}, nil
	}
	parsed, err := gotenv.StrictParse(strings.NewReader(string(data)))
	if err != nil {
		return joinEnv(order, merged), []string{"env_file parse issue, using process env only"}, nil
	}
	var warns []string
	for k, v := range parsed {
		if v == "" {
			warns = append(warns, k+" is empty (not substituted with a value)")
		}
		if _, ok := merged[k]; !ok {
			order = append(order, k)
		}
		merged[k] = v
	}
	return joinEnv(order, merged), warns, nil
}

func joinEnv(order []string, m map[string]string) []string {
	out := make([]string, 0, len(order))
	for _, k := range order {
		out = append(out, k+"="+m[k])
	}
	return out
}

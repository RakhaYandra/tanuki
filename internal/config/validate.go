package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	nameRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9\-_]{1,62}$`)
	goVerRe = regexp.MustCompile(`^\d+\.\d+(\.\d+)?$`)
	maxSvcs = 16
)

// Parse parses YAML bytes strictly (unknown fields rejected).
func Parse(data []byte) (*Config, error) {
	var c Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return nil, tanukierr.Wrap(tanukierr.CodeConfig, "invalid YAML", err)
	}
	if err := Validate(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// ParseFile reads a file via viper (per ADR-0001) then strict-validates with yaml.v3.
func ParseFile(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, tanukierr.Wrap(tanukierr.CodeConfig, "cannot read config", err)
	}
	var c Config
	// ErrorUnused rejects unknown fields, matching strict-decoding rule.
	if err := v.UnmarshalExact(&c); err != nil {
		return nil, tanukierr.Wrap(tanukierr.CodeConfig, "invalid configuration", err)
	}
	if err := Validate(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Validate enforces the v0.1 schema rules.
func Validate(c *Config) error {
	var errs []string
	add := func(f string, args ...any) { errs = append(errs, fmt.Sprintf(f, args...)) }

	if c.Version != SupportedVersion {
		add("version must be %q, got %q", SupportedVersion, c.Version)
	}
	if !nameRe.MatchString(c.Project.Name) {
		add("project.name must match ^[a-z0-9][a-z0-9-_]{1,62}$, got %q", c.Project.Name)
	}
	for k, v := range c.Runtime {
		if k != "go" {
			add("runtime.%s: unsupported runtime %q (v0.1 supports only go)", k, k)
			continue
		}
		if !goVerRe.MatchString(v) {
			add("runtime.go must look like \"1.27\" or \"1.27.1\", got %q", v)
		}
	}
	if len(c.Services) > maxSvcs {
		add("too many services: %d (max %d)", len(c.Services), maxSvcs)
	}
	ports := map[int]string{}
	for name, s := range c.Services {
		if !nameRe.MatchString(name) {
			add("services.%s: invalid service name %q", name, name)
		}
		if s.Type != "postgres" && s.Type != "redis" {
			add("services.%s.type must be postgres or redis, got %q", name, s.Type)
		}
		if strings.TrimSpace(s.Version) == "" {
			add("services.%s.version is required", name)
		}
		if s.Port != 0 {
			if s.Port < 1 || s.Port > 65535 {
				add("services.%s.port must be 1-65535, got %d", name, s.Port)
			} else if owner, dup := ports[s.Port]; dup {
				add("duplicate port %d (services.%s and %s)", s.Port, name, owner)
			} else {
				ports[s.Port] = "services." + name
			}
		}
		if err := checkHealth("services."+name, s.Healthcheck); err != nil {
			add("%s", err)
		}
	}
	if c.Application != nil {
		a := c.Application
		if len(a.Command) == 0 {
			add("application.command must have at least one element")
		}
		for _, arg := range a.Command {
			if arg == "" {
				add("application.command must not contain empty strings")
				break
			}
		}
		if a.Port != 0 {
			if a.Port < 1 || a.Port > 65535 {
				add("application.port must be 1-65535, got %d", a.Port)
			} else if owner, dup := ports[a.Port]; dup {
				add("duplicate port %d (application and %s)", a.Port, owner)
			}
		}
		for _, d := range a.DependsOn {
			if _, ok := c.Services[d]; !ok {
				add("application.depends_on references unknown service %q", d)
			}
		}
		if a.EnvFile != "" {
			if filepath.IsAbs(a.EnvFile) || strings.Contains(a.EnvFile, "..") {
				add("application.env_file must be a relative path without .., got %q", a.EnvFile)
			}
		}
		if err := checkHealth("application", a.Health); err != nil {
			add("%s", err)
		}
	}
	if len(errs) > 0 {
		return tanukierr.New(tanukierr.CodeConfig, "invalid configuration:\n  - "+strings.Join(errs, "\n  - "))
	}
	return nil
}

func checkHealth(owner string, h *Healthcheck) error {
	if h == nil {
		return nil
	}
	if h.Type != "tcp" && h.Type != "http" {
		return fmt.Errorf("%s.healthcheck.type must be tcp or http, got %q", owner, h.Type)
	}
	if h.Type == "http" && !strings.HasPrefix(h.Path, "/") {
		return fmt.Errorf("%s.healthcheck.path must start with /, got %q", owner, h.Path)
	}
	if h.Timeout != "" {
		d, err := time.ParseDuration(h.Timeout)
		if err != nil {
			return fmt.Errorf("%s.healthcheck.timeout invalid: %q", owner, h.Timeout)
		}
		if d < time.Second || d > 5*time.Minute {
			return fmt.Errorf("%s.healthcheck.timeout must be 1s-5m, got %q", owner, h.Timeout)
		}
	}
	return nil
}

// WriteExample writes a starter config; refuses to overwrite unless force.
// Used by `tanuki init`.
func WriteExample(dir, project string, force bool) (string, error) {
	if !nameRe.MatchString(project) {
		return "", tanukierr.New(tanukierr.CodeConfig, fmt.Sprintf("invalid project name %q", project))
	}
	path := filepath.Join(dir, FileName)
	if _, err := os.Stat(path); err == nil && !force {
		return "", tanukierr.New(tanukierr.CodeConfig, "tanuki.yaml already exists (use --force to overwrite)")
	}
	body := fmt.Sprintf("version: %q\n\nproject:\n  name: %s\n\nruntime:\n  go: \"1.27\"\n\nservices:\n  postgres:\n    type: postgres\n    version: \"16\"\n    port: 5432\n    healthcheck:\n      type: tcp\n      timeout: 30s\n\n  redis:\n    type: redis\n    version: \"7\"\n    port: 6379\n    healthcheck:\n      type: tcp\n      timeout: 30s\n\napplication:\n  name: api\n  command:\n    - go\n    - run\n    - ./cmd/api\n  port: 8080\n  depends_on:\n    - postgres\n    - redis\n  healthcheck:\n    type: http\n    path: /health\n    timeout: 60s\n", SupportedVersion, project)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", tanukierr.Wrap(tanukierr.CodeState, "cannot write tanuki.yaml", err)
	}
	return path, nil
}

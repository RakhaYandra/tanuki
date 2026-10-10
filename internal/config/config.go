// Package config defines the tanuki.yaml schema, discovery, parsing, and validation.
package config

import "time"

// FileName is the canonical configuration file name.
const FileName = "tanuki.yaml"

// SupportedVersion is the only schema version accepted in v0.1.
const SupportedVersion = "1"

// Config is the typed representation of tanuki.yaml.
type Config struct {
	Version     string             `yaml:"version" mapstructure:"version"`
	Project     Project            `yaml:"project" mapstructure:"project"`
	Runtime     map[string]string  `yaml:"runtime,omitempty" mapstructure:"runtime"`
	Services    map[string]Service `yaml:"services,omitempty" mapstructure:"services"`
	Application *Application       `yaml:"application,omitempty" mapstructure:"application"`
}

// Project holds project identity.
type Project struct {
	Name string `yaml:"name" mapstructure:"name"`
}

// Service declares one infrastructure dependency.
type Service struct {
	Type        string       `yaml:"type" mapstructure:"type"`
	Version     string       `yaml:"version" mapstructure:"version"`
	Port        int          `yaml:"port,omitempty" mapstructure:"port"`
	Healthcheck *Healthcheck `yaml:"healthcheck,omitempty" mapstructure:"healthcheck"`
}

// Application declares the project process.
type Application struct {
	Name      string       `yaml:"name,omitempty" mapstructure:"name"`
	Command   []string     `yaml:"command" mapstructure:"command"`
	Port      int          `yaml:"port,omitempty" mapstructure:"port"`
	EnvFile   string       `yaml:"env_file,omitempty" mapstructure:"env_file"`
	DependsOn []string     `yaml:"depends_on,omitempty" mapstructure:"depends_on"`
	Health    *Healthcheck `yaml:"healthcheck,omitempty" mapstructure:"healthcheck"`
}

// Healthcheck supports tcp and http probes in v0.1.
type Healthcheck struct {
	Type    string `yaml:"type" mapstructure:"type"`
	Path    string `yaml:"path,omitempty" mapstructure:"path"`
	Timeout string `yaml:"timeout,omitempty" mapstructure:"timeout"`
}

// TimeoutDuration parses Healthcheck.Timeout, defaulting to 30s.
func (h *Healthcheck) TimeoutDuration() time.Duration {
	if h == nil || h.Timeout == "" {
		return 30 * time.Second
	}
	d, err := time.ParseDuration(h.Timeout)
	if err != nil {
		return -1
	}
	return d
}

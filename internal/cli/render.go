package cli

import (
	"encoding/json"

	"github.com/RakhaYandra/tanuki/internal/config"
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"gopkg.in/yaml.v3"
)

func render(c *config.Config, format string) (string, error) {
	// No secret fields in schema; env values live in .env, never rendered.
	switch format {
	case "", "yaml":
		b, err := yaml.Marshal(c)
		if err != nil {
			return "", tanukierr.Wrap(tanukierr.CodeGeneral, "cannot render", err)
		}
		return string(b), nil
	case "json":
		yb, err := yaml.Marshal(c)
		if err != nil {
			return "", tanukierr.Wrap(tanukierr.CodeGeneral, "cannot render", err)
		}
		var m map[string]any
		if err := yaml.Unmarshal(yb, &m); err != nil {
			return "", tanukierr.Wrap(tanukierr.CodeGeneral, "cannot render", err)
		}
		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return "", tanukierr.Wrap(tanukierr.CodeGeneral, "cannot render", err)
		}
		return string(b), nil
	default:
		return "", tanukierr.New(tanukierr.CodeConfig, "unsupported --format")
	}
}

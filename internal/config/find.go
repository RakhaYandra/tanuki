package config

import (
	"os"
	"path/filepath"

	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
)

// Find locates tanuki.yaml from startDir upward.
// If explicit is non-empty it is returned (must exist).
func Find(startDir, explicit string) (string, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", tanukierr.Wrap(tanukierr.CodeConfig, "invalid --config path", err)
		}
		if _, err := os.Stat(abs); err != nil {
			return "", tanukierr.Wrap(tanukierr.CodeConfig, "config file not found: "+abs, err)
		}
		return abs, nil
	}
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", tanukierr.Wrap(tanukierr.CodeConfig, "cannot resolve directory", err)
	}
	for {
		candidate := filepath.Join(dir, FileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", tanukierr.New(tanukierr.CodeConfig, "tanuki.yaml not found (searched up to filesystem root)")
		}
		dir = parent
	}
}

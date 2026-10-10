// Package state tracks Tanuki-managed containers in a local JSON file.
package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
)

// State is minimal local metadata. Reconstructible from docker labels + /proc.
type State struct {
	Project     string            `json:"project"`
	ConfigPath  string            `json:"config_path"`
	Fingerprint string            `json:"config_fingerprint,omitempty"`
	Containers  map[string]string `json:"containers"`
	AppPID      int               `json:"app_pid,omitempty"`
	AppStart    uint64            `json:"app_starttime,omitempty"`
	AppLog      string            `json:"app_log,omitempty"`
	Updated     int64             `json:"updated"`
}

// ErrCorruptRecovered signals Load backed up a corrupt file and returned fresh state.
var ErrCorruptRecovered = tanukierr.New(tanukierr.CodeState, "state file was corrupt; backed up, starting fresh")

// Dir returns the state directory per XDG conventions.
func Dir() string {
	if x := os.Getenv("XDG_STATE_HOME"); x != "" {
		return filepath.Join(x, "tanuki")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "tanuki")
}

func path(project string) string { return filepath.Join(Dir(), project+".json") }

// Load reads state; missing file returns empty state. Corrupt files are
// renamed to *.corrupt.<unix> and fresh state is returned with ErrCorruptRecovered.
func Load(project string) (*State, error) {
	s := &State{Project: project, Containers: map[string]string{}}
	data, err := os.ReadFile(path(project))
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, tanukierr.Wrap(tanukierr.CodeState, "cannot read state", err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		backup := fmt.Sprintf("%s.corrupt.%d", path(project), time.Now().Unix())
		_ = os.Rename(path(project), backup)
		return s, ErrCorruptRecovered
	}
	if s.Containers == nil {
		s.Containers = map[string]string{}
	}
	return s, nil
}

// Save writes state atomically with 0600 permissions.
func (s *State) Save() error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return tanukierr.Wrap(tanukierr.CodeState, "cannot create state dir", err)
	}
	s.Updated = time.Now().Unix()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return tanukierr.Wrap(tanukierr.CodeState, "cannot encode state", err)
	}
	tmp := path(s.Project) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return tanukierr.Wrap(tanukierr.CodeState, "cannot write state", err)
	}
	if err := os.Rename(tmp, path(s.Project)); err != nil {
		return tanukierr.Wrap(tanukierr.CodeState, "cannot commit state", err)
	}
	return nil
}

// Touch records a managed container id for a service.
func (s *State) Touch(service, id string) {
	if s.Containers == nil {
		s.Containers = map[string]string{}
	}
	s.Containers[service] = id
}

// ClearApp drops the application record after stop.
func (s *State) ClearApp() {
	s.AppPID = 0
	s.AppStart = 0
	s.AppLog = ""
}

// FingerprintOf hashes the config file to detect drift since up.
func FingerprintOf(configPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", tanukierr.Wrap(tanukierr.CodeState, "cannot hash config", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Prune drops container ids not in live and clears dead app records.
// Returns number of entries removed.
func (s *State) Prune(live map[string]bool, appAlive bool) int {
	n := 0
	for svc := range s.Containers {
		if !live[svc] {
			delete(s.Containers, svc)
			n++
		}
	}
	if !appAlive && s.AppPID != 0 {
		s.ClearApp()
		n++
	}
	return n
}

// AcquireLock takes an exclusive non-blocking lock for project operations.
// Returns a release func. Second holder gets an exit-7 error.
func AcquireLock(project string) (func(), error) {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return nil, tanukierr.Wrap(tanukierr.CodeState, "cannot create state dir", err)
	}
	f, err := os.OpenFile(filepath.Join(Dir(), ".lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, tanukierr.Wrap(tanukierr.CodeState, "cannot open lock", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, tanukierr.New(tanukierr.CodeState, "another tanuki operation is running for this machine; wait for it to finish")
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}

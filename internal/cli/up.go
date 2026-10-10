package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/RakhaYandra/tanuki/internal/docker"
	"github.com/RakhaYandra/tanuki/internal/environment"
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"github.com/RakhaYandra/tanuki/internal/ports"
	"github.com/RakhaYandra/tanuki/internal/process"
	"github.com/RakhaYandra/tanuki/internal/services"
	"github.com/RakhaYandra/tanuki/internal/state"
	"github.com/spf13/cobra"
)

func cliRunner(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

func loadConfig() (*config.Config, string, error) {
	cwd, _ := os.Getwd()
	path, err := config.Find(cwd, cfgFile)
	if err != nil {
		return nil, "", err
	}
	cfg, err := config.ParseFile(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

func newUpCmd() *cobra.Command {
	var dryRun, detach, waitFlag bool
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start project services and application",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig()
			if err != nil {
				return err
			}
			wait := waitFlag && !detach
			projectDir := filepath.Dir(path)
			hasApp := cfg.Application != nil
			if len(cfg.Services) == 0 && !hasApp {
				fmt.Fprintln(cmd.OutOrStdout(), "nothing configured; nothing to do")
				return nil
			}
			plan, err := services.Plan(cfg)
			if err != nil {
				return err
			}
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "plan for %s (%d services):\n", cfg.Project.Name, len(plan))
				for _, a := range plan {
					fmt.Fprintf(cmd.OutOrStdout(), "  create %s (%s) :%d\n", a.Container, a.Image, a.Port)
				}
				if hasApp {
					fmt.Fprintf(cmd.OutOrStdout(), "  start app %s: %s :%d (env names only, values never shown)\n",
						appName(cfg), strings.Join(cfg.Application.Command, " "), cfg.Application.Port)
				}
				return nil
			}
			// Preflight: daemon + port conflicts before any mutation.
			if len(plan) > 0 {
				if _, derr := cliRunner("docker", "info"); derr != nil {
					c := docker.CheckDaemon(cliRunner)
					if c.Status == environment.StatusFail {
						return tanukierr.New(tanukierr.CodePrereq, c.Message+". "+c.Hint)
					}
					return tanukierr.New(tanukierr.CodePrereq, "docker unavailable")
				}
			}
			for _, a := range plan {
				if ports.InUse("127.0.0.1", a.Port, ports.DefaultTimeout) {
					owned, _ := docker.List(cfg.Project.Name, cliRunner)
					mine := false
					for _, c := range owned {
						if c.LabelMap[docker.LabelService] == a.Service {
							mine = true
						}
					}
					if !mine {
						return tanukierr.New(tanukierr.CodeConflict,
							fmt.Sprintf("port %d is already in use.\n\nTanuki did not stop or modify the existing process.\n\nSuggested actions:\n  - Inspect the process using port %d.\n  - Configure another host port in tanuki.yaml.\n  - Run `tanuki validate` to check the configuration.", a.Port, a.Port))
					}
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "tanuki: starting %s\n", cfg.Project.Name)
			if detach {
				fmt.Fprintln(cmd.OutOrStdout(), "  ! --detach: health waits skipped, everything keeps running")
			} else if !wait {
				fmt.Fprintln(cmd.OutOrStdout(), "  ! --wait=false: health waits skipped")
			}
			release, err := state.AcquireLock(cfg.Project.Name)
			if err != nil {
				return err
			}
			defer release()
			done, err := services.Up(cfg, cliRunner, wait)
			if err != nil {
				return err
			}
			st, serr := state.Load(cfg.Project.Name)
			if serr == state.ErrCorruptRecovered {
				fmt.Fprintln(cmd.OutOrStdout(), "  ! state file was corrupt; backed up, starting fresh")
			}
			st.ConfigPath = path
			if fp, ferr := state.FingerprintOf(path); ferr == nil {
				st.Fingerprint = fp
			}
			owned, _ := docker.List(cfg.Project.Name, cliRunner)
			for _, c := range owned {
				st.Touch(c.LabelMap[docker.LabelService], c.ID)
			}
			_ = st.Save()
			total := len(done)
			if hasApp {
				total++
			}
			for i, a := range done {
				fmt.Fprintf(cmd.OutOrStdout(), "[%d/%d] %s (%s)  ✓ %s\n", i+1, total, a.Service, a.Image, a.Kind)
			}
			if hasApp {
				proc, reused, warns, err := services.StartApp(cfg, projectDir, st, process.Spawn)
				for _, w := range warns {
					fmt.Fprintf(cmd.OutOrStdout(), "  ! %s\n", w)
				}
				if err != nil {
					return err
				}
				if reused {
					fmt.Fprintf(cmd.OutOrStdout(), "[%d/%d] %s  ✓ already running (pid %d)\n", total, total, appName(cfg), proc.PID)
				} else if !wait {
					fmt.Fprintf(cmd.OutOrStdout(), "[%d/%d] %s  ✓ started (pid %d, log %s, health skipped)\n", total, total, appName(cfg), proc.PID, proc.LogPath)
				} else {
					if werr := services.WaitApp(cfg); werr != nil {
						fmt.Fprintln(cmd.OutOrStdout(), "\nServices and application are still running.")
						fmt.Fprintln(cmd.OutOrStdout(), "Run `tanuki logs api` for details, `tanuki down` to stop.")
						return tanukierr.Wrap(tanukierr.CodeStartup, "application unhealthy", werr)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "[%d/%d] %s  ✓ started healthy (pid %d)\n", total, total, appName(cfg), proc.PID)
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nEnvironment ready.")
			fmt.Fprintln(cmd.OutOrStdout(), "Run `tanuki status` to inspect. Run `tanuki down` to stop.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print plan without changing the system")
	cmd.Flags().BoolVar(&detach, "detach", false, "skip health waits, everything keeps running")
	cmd.Flags().BoolVar(&waitFlag, "wait", true, "wait for health checks (default true)")
	return cmd
}

func appName(cfg *config.Config) string {
	if cfg.Application != nil && cfg.Application.Name != "" {
		return cfg.Application.Name
	}
	return "api"
}
func newDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Stop application and project-managed services (volumes preserved)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}
			release, err := state.AcquireLock(cfg.Project.Name)
			if err != nil {
				return err
			}
			defer release()
			appStopped, err := services.StopApp(cfg.Project.Name, cliRunner)
			if err != nil {
				return err
			}
			if appStopped {
				fmt.Fprintln(cmd.OutOrStdout(), "stopped application")
			}
			stopped, err := services.Down(cfg.Project.Name, cliRunner)
			if err != nil {
				return err
			}
			if !appStopped && len(stopped) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "nothing managed by tanuki for project "+cfg.Project.Name)
				return nil
			}
			for _, s := range stopped {
				fmt.Fprintln(cmd.OutOrStdout(), "stopped "+s+" (volume preserved: "+docker.VolumeName(cfg.Project.Name, s)+")")
			}
			return nil
		},
	}
}

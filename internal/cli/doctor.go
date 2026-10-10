package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/RakhaYandra/tanuki/internal/config"
	"github.com/RakhaYandra/tanuki/internal/docker"
	"github.com/RakhaYandra/tanuki/internal/envcheck"
	"github.com/RakhaYandra/tanuki/internal/environment"
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"github.com/RakhaYandra/tanuki/internal/ports"
	inruntime "github.com/RakhaYandra/tanuki/internal/runtime"
	"github.com/spf13/cobra"
)

func execRunner(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check prerequisites without changing the system",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "TANUKI DOCTOR")
			var checks []environment.Check

			// System always runs, even without config.
			checks = append(checks, environment.Check{
				Group: "System", Name: "OS",
				Status: environment.StatusPass, Message: "Linux detected",
			})
			checks = append(checks, environment.Check{
				Group: "System", Name: "Arch",
				Status: environment.StatusPass, Message: "Architecture: " + runtime.GOARCH,
			})

			cwd, _ := os.Getwd()
			path, ferr := config.Find(cwd, cfgFile)
			var cfg *config.Config
			if ferr != nil {
				checks = append(checks, environment.Check{
					Group: "Configuration", Name: "tanuki.yaml",
					Status: environment.StatusFail, Message: ferr.Error(),
				})
				printReport(out, checks)
				return ferr
			}
			var perr error
			cfg, perr = config.ParseFile(path)
			if perr != nil {
				checks = append(checks, environment.Check{
					Group: "Configuration", Name: "tanuki.yaml",
					Status: environment.StatusFail, Message: perr.Error(),
				})
				printReport(out, checks)
				return perr
			}
			checks = append(checks, environment.Check{
				Group: "Configuration", Name: "tanuki.yaml",
				Status: environment.StatusPass, Message: "found and valid: " + path,
			})

			checks = append(checks,
				docker.CheckCLI(execRunner),
				docker.CheckDaemon(execRunner),
				inruntime.DetectGo(cfg.Runtime["go"], execRunner),
			)
			if _, err := exec.LookPath("psql"); err != nil {
				checks = append(checks, environment.Check{
					Group: "Dependencies", Name: "psql",
					Status: environment.StatusWarn, Message: "PostgreSQL CLI not installed",
					Hint: "Optional for doctor; install postgresql-client if you need psql",
				})
			} else {
				checks = append(checks, environment.Check{
					Group: "Dependencies", Name: "psql",
					Status: environment.StatusPass, Message: "PostgreSQL CLI found",
				})
			}
			seen := map[int]bool{}
			for _, s := range cfg.Services {
				if s.Port != 0 && !seen[s.Port] {
					seen[s.Port] = true
					checks = append(checks, ports.Check("127.0.0.1", s.Port, ports.DefaultTimeout))
				}
			}
			if cfg.Application != nil && cfg.Application.Port != 0 && !seen[cfg.Application.Port] {
				checks = append(checks, ports.Check("127.0.0.1", cfg.Application.Port, ports.DefaultTimeout))
			}
			projectDir := filepath.Dir(path)
			checks = append(checks, envcheck.CheckEnvFile(projectDir, cfg.Application)...)

			printReport(out, checks)

			rep := environment.Report{Checks: checks}
			if rep.HasFail() {
				for _, c := range checks {
					if c.Status == environment.StatusFail && c.Group == "Ports" {
						return tanukierr.New(tanukierr.CodeConflict, "doctor: port conflict detected")
					}
				}
				return tanukierr.New(tanukierr.CodePrereq, "doctor: prerequisite checks failed")
			}
			return nil
		},
	}
}

func printReport(out interface{ Write([]byte) (int, error) }, checks []environment.Check) {
	groups := []string{"System", "Dependencies", "Configuration", "Ports", "Environment"}
	seen := map[string][]environment.Check{}
	for _, c := range checks {
		seen[c.Group] = append(seen[c.Group], c)
	}
	for _, g := range groups {
		list, ok := seen[g]
		if !ok {
			continue
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, g)
		for _, c := range list {
			sym := "✓"
			if c.Status == environment.StatusWarn {
				sym = "!"
			} else if c.Status == environment.StatusFail {
				sym = "✗"
			}
			fmt.Fprintln(out, "  "+sym+" "+c.Name+": "+c.Message)
			if c.Hint != "" {
				fmt.Fprintln(out, "    hint: "+c.Hint)
			}
		}
	}
	rep := environment.Report{Checks: checks}
	p, w, f := rep.Counts()
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Summary")
	fmt.Fprintf(out, "  %d passed, %d warning, %d failed\n", p, w, f)
}

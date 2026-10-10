package cli

import (
	"errors"
	"fmt"
	"sort"

	"github.com/RakhaYandra/tanuki/internal/docker"
	"github.com/RakhaYandra/tanuki/internal/ports"
	"github.com/RakhaYandra/tanuki/internal/process"
	"github.com/RakhaYandra/tanuki/internal/state"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display current environment state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "PROJECT: %s\n", cfg.Project.Name)
			degraded := false

			st, serr := state.Load(cfg.Project.Name)
			if errors.Is(serr, state.ErrCorruptRecovered) {
				fmt.Fprintln(out, "  ! state file was corrupt; backed up, starting fresh")
			}
			if fp, ferr := state.FingerprintOf(path); ferr == nil && st.Fingerprint != "" && fp != st.Fingerprint {
				fmt.Fprintln(out, "  ! tanuki.yaml changed since last up; run `tanuki up` to reconcile")
			}

			fmt.Fprintln(out, "\nAPPLICATION")
			st, _ = state.Load(cfg.Project.Name)
			appName := appName(cfg)
			if cfg.Application == nil {
				fmt.Fprintln(out, "  (not configured)")
			} else if st.AppPID > 0 && (&process.Proc{PID: st.AppPID, Starttime: st.AppStart}).Alive() {
				fmt.Fprintf(out, "  %s  running (pid %d)\n", appName, st.AppPID)
				fmt.Fprintf(out, "  Port  %d\n", cfg.Application.Port)
			} else if st.AppPID > 0 {
				fmt.Fprintf(out, "  %s  stopped (stale record, run `tanuki up`)\n", appName)
				degraded = true
			} else {
				fmt.Fprintf(out, "  %s  stopped\n", appName)
				degraded = true
			}

			fmt.Fprintln(out, "\nSERVICES")
			names := make([]string, 0, len(cfg.Services))
			for n := range cfg.Services {
				names = append(names, n)
			}
			sort.Strings(names)
			owned, _ := docker.List(cfg.Project.Name, cliRunner)
			bySvc := map[string]docker.Container{}
			live := map[string]bool{}
			for _, c := range owned {
				svc := c.LabelMap[docker.LabelService]
				bySvc[svc] = c
				if c.State == "running" {
					live[svc] = true
				}
			}
			appAlive := st.AppPID > 0 && (&process.Proc{PID: st.AppPID, Starttime: st.AppStart}).Alive()
			if n := st.Prune(live, appAlive); n > 0 {
				_ = st.Save()
				fmt.Fprintf(out, "  ! state cleaned (%d stale entries removed)\n", n)
			}
			if len(names) == 0 {
				fmt.Fprintln(out, "  (none configured)")
			}
			for _, n := range names {
				s := cfg.Services[n]
				c, ok := bySvc[n]
				if !ok || c.State != "running" {
					fmt.Fprintf(out, "  %s  stopped\n", n)
					degraded = true
					continue
				}
				if s.Port != 0 && ports.InUse("127.0.0.1", s.Port, ports.DefaultTimeout) {
					fmt.Fprintf(out, "  %s  healthy\n", n)
					fmt.Fprintf(out, "  Port  %d\n", s.Port)
				} else {
					fmt.Fprintf(out, "  %s  running (not accepting connections)\n", n)
					degraded = true
				}
			}

			fmt.Fprintln(out, "\nRUNTIME")
			if g, ok := cfg.Runtime["go"]; ok {
				fmt.Fprintf(out, "  Go  required %s\n", g)
			} else {
				fmt.Fprintln(out, "  Go  no requirement")
			}
			fmt.Fprintln(out, "\nOverall status: "+overallStatus(degraded))
			return nil
		},
	}
}

func overallStatus(degraded bool) string {
	if degraded {
		return "DEGRADED"
	}
	return "READY"
}

func newPsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "List resources managed by Tanuki",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "NAME\tTYPE\tSTATE\tPORT")
			owned, _ := docker.List(cfg.Project.Name, cliRunner)
			for _, c := range owned {
				port := ""
				for svc, s := range cfg.Services {
					if c.LabelMap[docker.LabelService] == svc && s.Port != 0 {
						port = fmt.Sprint(s.Port)
					}
				}
				fmt.Fprintf(out, "%s\tcontainer\t%s\t%s\n", c.Name, c.State, port)
			}
			st, _ := state.Load(cfg.Project.Name)
			if cfg.Application != nil {
				astate := "stopped"
				if st.AppPID > 0 && (&process.Proc{PID: st.AppPID, Starttime: st.AppStart}).Alive() {
					astate = "running"
				}
				fmt.Fprintf(out, "%s\tprocess\t%s\t%d\n", appName(cfg), astate, cfg.Application.Port)
			}
			return nil
		},
	}
}

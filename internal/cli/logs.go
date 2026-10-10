package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/RakhaYandra/tanuki/internal/docker"
	"github.com/RakhaYandra/tanuki/internal/state"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "logs [postgres|redis|api]",
		Short: "Show logs from managed services and the application",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}
			target := "api"
			if len(args) == 1 {
				target = args[0]
			}
			out := cmd.OutOrStdout()
			// App target: own name, "api", "app", or "application".
			if target == "api" || target == "app" || target == "application" || (cfg.Application != nil && target == cfg.Application.Name) {
				st, _ := state.Load(cfg.Project.Name)
				if st.AppLog == "" {
					fmt.Fprintln(out, "no application log yet (run `tanuki up` first)")
					return nil
				}
				if follow {
					c := exec.Command("tail", "-f", st.AppLog)
					c.Stdout = out
					c.Stderr = out
					return c.Run()
				}
				data, err := os.ReadFile(st.AppLog)
				if err != nil {
					fmt.Fprintln(out, "cannot read application log: "+err.Error())
					return nil
				}
				fmt.Fprint(out, string(data))
				return nil
			}
			owned, err := docker.List(cfg.Project.Name, cliRunner)
			if err != nil {
				return err
			}
			for _, c := range owned {
				if c.LabelMap[docker.LabelService] == target {
					if follow {
						dc := exec.Command("docker", "logs", "-f", c.ID)
						dc.Stdout = out
						dc.Stderr = out
						return dc.Run()
					}
					logs := docker.Logs(c.ID, cliRunner)
					if logs == "" {
						fmt.Fprintln(out, "(no logs)")
						return nil
					}
					fmt.Fprintln(out, logs)
					return nil
				}
			}
			fmt.Fprintln(out, "unknown target "+target+"; use postgres, redis, or api")
			return nil
		},
	}
	cmd.Flags().BoolVar(&follow, "follow", false, "stream logs")
	return cmd
}

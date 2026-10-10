// Package cli wires cobra commands to config operations.
package cli

import (
	"github.com/spf13/cobra"
)

var (
	// Version is set via ldflags.
	Version = "dev"
	cfgFile string
)

// NewRoot builds the command tree. All commands implemented;
// errors print once via main.go (SilenceUsage/Errors set).
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "tanuki",
		Short:         "Development environment manager",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "explicit path to tanuki.yaml")
	root.AddCommand(newInitCmd(), newValidateCmd(), newVersionCmd(), newConfigCmd(), newDoctorCmd(), newUpCmd(), newDownCmd(), newLogsCmd(), newStatusCmd(), newPsCmd())
	for _, s := range stubCmds() {
		root.AddCommand(s)
	}
	return root
}

// Execute runs the root command.
func Execute() error {
	return NewRoot().Execute()
}

func stubCmds() []*cobra.Command {
	// All Phase 1-4 commands implemented; kept for future phases.
	return nil
}

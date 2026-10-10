package cli

import (
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"github.com/spf13/cobra"
)

func errNotImpl(name string) error {
	return tanukierr.New(tanukierr.CodeGeneral, "tanuki "+name+": not implemented yet")
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("tanuki %s\n", Version)
		},
	}
}

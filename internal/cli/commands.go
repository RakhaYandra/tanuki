package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/RakhaYandra/tanuki/internal/config"
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var name, dir string
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate tanuki.yaml with safe defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				base, err := os.Getwd()
				if err != nil {
					return tanukierr.Wrap(tanukierr.CodeGeneral, "cannot get cwd", err)
				}
				if dir != "" {
					base = dir
				}
				name = filepath.Base(base)
			}
			target := dir
			if target == "" {
				var err error
				target, err = os.Getwd()
				if err != nil {
					return tanukierr.Wrap(tanukierr.CodeGeneral, "cannot get cwd", err)
				}
			}
			path, err := config.WriteExample(target, name, force)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name (default: current directory)")
	cmd.Flags().StringVar(&dir, "dir", "", "target directory (default: cwd)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing tanuki.yaml")
	return cmd
}

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate tanuki.yaml without starting services",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			path, err := config.Find(cwd, cfgFile)
			if err != nil {
				return err
			}
			if _, err := config.ParseFile(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "configuration valid: %s\n", path)
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Print effective configuration (secrets redacted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if format != "" && format != "yaml" && format != "json" {
				return tanukierr.New(tanukierr.CodeConfig, "unsupported --format (yaml|json)")
			}
			cwd, _ := os.Getwd()
			path, err := config.Find(cwd, cfgFile)
			if err != nil {
				return err
			}
			c, err := config.ParseFile(path)
			if err != nil {
				return err
			}
			out, err := render(c, format)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "yaml", "output format: yaml|json")
	return cmd
}

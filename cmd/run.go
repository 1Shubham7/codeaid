package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a one-shot coding task (not yet implemented)",
}

func init() {
	rootCmd.AddCommand(runCmd)
}

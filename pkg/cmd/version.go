package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information
var (
	Version   string
	BuildDate string
	GitCommit string
)

// NewVersionCommand - Creates new Version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "prints the permify version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version:      %s\n", Version)
			fmt.Printf("Build Date:   %s\n", BuildDate)
			fmt.Printf("Git Commit: %s\n", GitCommit)

			return nil
		},
	}
}

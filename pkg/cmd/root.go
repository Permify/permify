package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCommand - Creates new root command
func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "permify",
		Short: "open-source authorization service & policy engine based on Google Zanzibar.",
		Long:  "open-source authorization service & policy engine based on Google Zanzibar.",
	}
}

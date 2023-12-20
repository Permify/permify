package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterAstFlags registers ast flags.
func RegisterAstFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.Bool("pretty", false, "If set to true, produces a human-readable output of the AST.")
	if err := viper.BindPFlag("pretty", flags.Lookup("pretty")); err != nil {
		panic(err)
	}
}

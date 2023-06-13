package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterValidationFlags registers validation flags.
func RegisterValidationFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.String("output-format", "", "output format. one of: verbose, json")
	if err := viper.BindPFlag("output-format", flags.Lookup("output-format")); err != nil {
		panic(err)
	}
}

package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterCoverageFlags registers coverage flags.
func RegisterCoverageFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.Int("coverage-relationships", 0, "the min coverage for relationships")
	if err := viper.BindPFlag("coverage-relationships", flags.Lookup("coverage-relationships")); err != nil {
		panic(err)
	}

	flags.Int("coverage-assertions", 0, "the min coverage for assertions")
	if err := viper.BindPFlag("coverage-assertions", flags.Lookup("coverage-assertions")); err != nil {
		panic(err)
	}
}

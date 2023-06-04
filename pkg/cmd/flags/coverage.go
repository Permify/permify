package flags

import (
	"github.com/Permify/permify/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RegisterCoverageFlags registers coverage flags.
func RegisterCoverageFlags(cmd *cobra.Command) {
	conf := config.DefaultConfig()
	flags := cmd.Flags()

	flags.Int("coverage-relationships", conf.Coverage.Relationships, "the min coverage for relationships")
	if err := viper.BindPFlag("coverage-relationships", flags.Lookup("coverage-relationships")); err != nil {
		panic(err)
	}

	flags.Int("coverage-assertions", conf.Coverage.Assertions, "the min coverage for assertions")
	if err := viper.BindPFlag("coverage-assertions", flags.Lookup("coverage-assertions")); err != nil {
		panic(err)
	}
}

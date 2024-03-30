package flags

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// RegisterCoverageFlags registers coverage flags.
func RegisterCoverageFlags(flags *pflag.FlagSet) {
	if err := viper.BindPFlag("coverage-relationships", flags.Lookup("coverage-relationships")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("coverage-attributes", flags.Lookup("coverage-attributes")); err != nil {
		panic(err)
	}

	if err := viper.BindPFlag("coverage-assertions", flags.Lookup("coverage-assertions")); err != nil {
		panic(err)
	}
}

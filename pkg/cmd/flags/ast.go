package flags

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// RegisterAstFlags registers ast flags.
func RegisterAstFlags(flags *pflag.FlagSet) {
	if err := viper.BindPFlag("pretty", flags.Lookup("pretty")); err != nil {
		panic(err)
	}
}

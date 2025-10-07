package flags // Command flags package

// Coverage flag registration utilities
import ( // Package imports
	"github.com/spf13/pflag" // Flag definitions
	"github.com/spf13/viper" // Configuration binding
) // End of imports

// RegisterCoverageFlags binds coverage-related flags to viper configuration
func RegisterCoverageFlags(flags *pflag.FlagSet) { // Register coverage flags
	if err := viper.BindPFlag("coverage-relationships", flags.Lookup("coverage-relationships")); err != nil { // Bind relationships flag
		panic(err) // Fatal error if binding fails
	} // Relationships flag bound
	// Bind attributes coverage flag
	if err := viper.BindPFlag("coverage-attributes", flags.Lookup("coverage-attributes")); err != nil {
		panic(err)
	}
	// Bind assertions coverage flag
	if err := viper.BindPFlag("coverage-assertions", flags.Lookup("coverage-assertions")); err != nil { // Bind assertions flag
		panic(err) // Fatal error if binding fails
	} // Assertions flag bound
} // End of RegisterCoverageFlags

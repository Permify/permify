package cmd // Command package
// Version command implementation
import ( // Package imports
	"fmt" // Formatting utilities
	// External dependencies
	"github.com/spf13/cobra" // Cobra CLI framework
	// Internal dependencies
	"github.com/Permify/permify/internal" // Internal version info
) // End of imports
// Version command implementation
// Version command function
// NewVersionCommand creates a new version command for displaying Permify version
// Returns a cobra command that displays the Permify version information
func NewVersionCommand() *cobra.Command { // Create version command
	return &cobra.Command{ // Return command configuration
		Use:   "version",                    // Command name
		Short: "prints the permify version", // Short description
		RunE: func(cmd *cobra.Command, args []string) error { // Command execution
			fmt.Printf("%s\n", internal.Version) // Print version
			return nil                           // No error
		}, // End of RunE
	} // End of command
} // End of NewVersionCommand

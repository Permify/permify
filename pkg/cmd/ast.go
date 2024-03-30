package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/pkg/cmd/flags"
	"github.com/Permify/permify/pkg/development/file"
)

// NewGenerateASTCommand creates a new cobra.Command to generate the Abstract Syntax Tree (AST) from a given file.
// The command expects exactly one argument, which is the URL or path to the file.
// It also allows for flags registration for further validation (like coverage flags).
func NewGenerateASTCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "ast <file>",
		Short: "Generates the AST from a specified file and prints it as JSON.",
		RunE:  runGenerateAST(),
		Args:  cobra.ExactArgs(1),
	}

	f := command.Flags()
	f.Bool("pretty", false, "If set to true, produces a human-readable output of the AST.")

	command.PreRun = func(cmd *cobra.Command, args []string) {
		flags.RegisterAstFlags(f)
	}

	return command
}

// runGenerateAST creates a closure that generates the AST and returns its JSON representation.
// Depending on the "pretty" flag, the output can be either pretty-printed or raw.
func runGenerateAST() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Fetch the value of the "pretty" flag.
		pretty := viper.GetBool("pretty")

		// Parse the provided URL.
		u, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		// Initialize a decoder for the provided URL.
		decoder, err := file.NewDecoderFromURL(u)
		if err != nil {
			return err
		}

		// Initialize an empty shape to store the decoded schema.
		s := &file.Shape{}

		// Decode the content from the URL into the shape.
		err = decoder.Decode(s)
		if err != nil {
			return err
		}

		// Convert the string definitions in the shape to a structured schema.
		def, err := schema.NewSchemaFromStringDefinitions(true, s.Schema)
		if err != nil {
			return err
		}

		// Serialize the schema definition into JSON.
		jsonData, err := protojson.Marshal(def)
		if err != nil {
			return err
		}

		// Print the JSON, either prettified or raw based on the "pretty" flag.
		if pretty {
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, jsonData, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(jsonData))
		}

		return nil
	}
}

package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Permify/permify/pkg/cmd/flags"
	cov "github.com/Permify/permify/pkg/development/coverage"
	"github.com/Permify/permify/pkg/development/file"
	"github.com/Permify/permify/pkg/schema"
)

// NewCoverageCommand - creates a new coverage command
func NewCoverageCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "coverage <file>",
		Short: "coverage analysis of authorization model and assertions",
		RunE:  coverage(),
		Args:  cobra.ExactArgs(1),
	}

	f := command.Flags()
	f.Int("coverage-relationships", 0, "the min coverage for relationships")
	f.Int("coverage-attributes", 0, "the min coverage for attributes")
	f.Int("coverage-assertions", 0, "the min coverage for assertions")

	// register flags for coverage
	command.PreRun = func(cmd *cobra.Command, args []string) {
		flags.RegisterCoverageFlags(f)
	}

	return command
}

// coverage - coverage analysis of authorization model and assertions
func coverage() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// parse the url from the first argument
		u, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		// get min coverage from viper
		coverageRelationships := viper.GetInt("coverage-relationships")
		coverageAttributes := viper.GetInt("coverage-attributes")
		coverageAssertions := viper.GetInt("coverage-assertions")

		// create a new decoder from the url
		decoder, err := file.NewDecoderFromURL(u)
		if err != nil {
			return err
		}

		// create a new shape
		s := &file.Shape{}

		// decode the schema from the decoder
		err = decoder.Decode(s)
		if err != nil {
			return err
		}

		loader := schema.NewSchemaLoader()
		loaded, err := loader.LoadSchema(s.Schema)
		if err != nil {
			return err
		}

		s.Schema = loaded

		color.Notice.Println("initiating validation... 🚀")
		validator := validate()
		err = validator(cmd, args)
		if err != nil {
			color.Danger.Println("failed to validate given file\n")
			color.Danger.Println("FAILED")
			return err
		}

		color.Notice.Println("initiating coverage analysis... 🚀")

		schemaCoverageInfo := cov.Run(*s)

		DisplayCoverageInfo(schemaCoverageInfo)

		if schemaCoverageInfo.TotalAssertionsCoverage < coverageAssertions {
			color.Danger.Printf("assertions coverage < %d%%\n", coverageAssertions)
			// print FAILED with color danger
			color.Danger.Println("FAILED")
			os.Exit(1)
		}

		if schemaCoverageInfo.TotalRelationshipsCoverage < coverageRelationships {
			color.Danger.Printf("relationships coverage < %d%%\n", coverageRelationships)
			// print FAILED with color danger
			color.Danger.Println("FAILED")
			os.Exit(1)
		}

		if schemaCoverageInfo.TotalAttributesCoverage < coverageAttributes {
			color.Danger.Printf("attributes coverage < %d%%\n", coverageAttributes)
			// print FAILED with color danger
			color.Danger.Println("FAILED")
			os.Exit(1)
		}
		return nil
	}
}

// DisplayCoverageInfo - Display the schema coverage information
func DisplayCoverageInfo(schemaCoverageInfo cov.SchemaCoverageInfo) {
	color.Notice.Println("schema coverage information:")

	for _, entityCoverageInfo := range schemaCoverageInfo.EntityCoverageInfo {
		color.Notice.Printf("entity: %s\n", entityCoverageInfo.EntityName)

		fmt.Printf("  uncovered relationships:\n")

		for _, value := range entityCoverageInfo.UncoveredRelationships {
			fmt.Printf("    - %v\n", value)
		}

		fmt.Printf("  uncovered assertions:\n")

		for key, value := range entityCoverageInfo.UncoveredAssertions {
			fmt.Printf("    %s:\n", key)
			for _, v := range value {
				fmt.Printf("    	%v\n", v)
			}
		}

		fmt.Printf("  coverage relationships percentage:")

		if entityCoverageInfo.CoverageRelationshipsPercent <= 50 {
			color.Danger.Printf(" %d%%\n", entityCoverageInfo.CoverageRelationshipsPercent)
		} else {
			color.Success.Printf(" %d%%\n", entityCoverageInfo.CoverageRelationshipsPercent)
		}

		fmt.Printf("  coverage attributes percentage:")

		if entityCoverageInfo.CoverageAttributesPercent <= 50 {
			color.Danger.Printf(" %d%%\n", entityCoverageInfo.CoverageAttributesPercent)
		} else {
			color.Success.Printf(" %d%%\n", entityCoverageInfo.CoverageAttributesPercent)
		}

		fmt.Printf("  coverage assertions percentage: \n")

		for key, value := range entityCoverageInfo.CoverageAssertionsPercent {
			fmt.Printf("    %s:", key)
			if value <= 50 {
				color.Danger.Printf(" %d%%\n", value)
			} else {
				color.Success.Printf(" %d%%\n", value)
			}
		}
	}
}

package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cobra"

	"github.com/Permify/permify/pkg/development"
	"github.com/Permify/permify/pkg/development/validation"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// NewValidateCommand -
func NewValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <file>",
		Short: "validate authorization model with assertions",
		RunE:  validate(),
		Args:  cobra.ExactArgs(1),
	}
}

// validate -
func validate() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		devContainer := development.NewContainer()

		u, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		decoder, err := validation.NewDecoderFromURL(u)
		if err != nil {
			return err
		}

		s := &validation.Shape{}
		err = decoder.Decode(s)
		if err != nil {
			return err
		}

		// Write schema -
		var version string
		version, err = devContainer.S.WriteSchema(ctx, s.Schema)
		if err != nil {
			return err
		}

		color.Success.Println("schema successfully created: ✓ ✅ ")

		// Write tuples -
		for _, t := range s.Tuples {
			tup, err := tuple.Tuple(t)
			if err != nil {
				return err
			}
			_, err = devContainer.R.WriteRelationships(ctx, []*base.Tuple{tup}, version)
			if err != nil {
				return err
			}
		}

		color.Success.Println("tuples successfully created: ✓ ✅ ")
		color.Success.Println("checking assertions...")

		// Check Assertions
		for i, assertion := range s.Assertions {
			for query, expected := range assertion {
				q, err := tuple.NewQueryFromString(query)
				if err != nil {
					return err
				}

				res, err := devContainer.P.CheckPermissions(ctx, q.Subject, q.Action, q.Entity, version, 20)
				if err != nil {
					return err
				}

				if res.Can == expected {
					fmt.Printf("%v. %s ? => ", i+1, query)
					if res.Can {
						color.Success.Println("expected: ✓ ✅ , actual: ✓ ✅ ")
					} else {
						color.Success.Println("expected: ✗ ❌ , actual: ✗ ❌ ")
					}
				} else {
					color.Danger.Printf("%v. %s ? => ", i+1, query)
					if res.Can {
						color.Danger.Println("expected: ✗ ❌ , actual: ✗ ✅ ")
					} else {
						color.Danger.Println("expected: ✓ ✅ , actual: ✓ ❌ ")
					}
					color.Danger.Println("FAILED.")
					os.Exit(1)
				}
			}
		}

		return nil
	}
}

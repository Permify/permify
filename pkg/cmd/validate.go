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
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// NewValidateCommand - Creates new validate command
func NewValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <file>",
		Short: "validate authorization model with assertions",
		RunE:  validate(),
		Args:  cobra.ExactArgs(1),
	}
}

// validate - permify validate command
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
		version, err = devContainer.S.WriteSchema(ctx, "noop", s.Schema)
		if err != nil {
			return err
		}

		color.Success.Println("schema successfully created: ✓ ✅ ")

		var tuples []*base.Tuple

		// Write tuples -
		for _, t := range s.Tuples {
			var tup *base.Tuple
			tup, err = tuple.Tuple(t)
			if err != nil {
				return err
			}
			tuples = append(tuples, tup)
		}

		_, err = devContainer.R.WriteRelationships(ctx, "noop", tuples, version)
		if err != nil {
			return err
		}

		color.Success.Println("tuples successfully created: ✓ ✅ ")
		color.Success.Println("checking assertions...")

		// Check Assertions
		for i, assertion := range s.Assertions {
			for query, expected := range assertion {
				exp := base.PermissionCheckResponse_RESULT_ALLOWED
				if !expected {
					exp = base.PermissionCheckResponse_RESULT_DENIED
				}

				q, err := tuple.NewQueryFromString(query)
				if err != nil {
					return err
				}

				res, err := devContainer.P.CheckPermissions(ctx, &base.PermissionCheckRequest{
					TenantId: "noop",
					Metadata: &base.PermissionCheckRequestMetadata{
						SchemaVersion: version,
						SnapToken:     token.NewNoopToken().Encode().String(),
						Depth:         100,
					},
					Entity:     q.Entity,
					Permission: q.Action,
					Subject:    q.Subject,
				})
				if err != nil {
					return err
				}

				if res.Can == exp {
					fmt.Printf("%v. %s ? => ", i+1, query)
					if res.Can == base.PermissionCheckResponse_RESULT_ALLOWED {
						color.Success.Println("expected: ✓ ✅ , actual: ✓ ✅ ")
					} else {
						color.Success.Println("expected: ✗ ❌ , actual: ✗ ❌ ")
					}
				} else {
					color.Danger.Printf("%v. %s ? => ", i+1, query)
					if res.Can == base.PermissionCheckResponse_RESULT_ALLOWED {
						color.Danger.Println("expected: ✗ ❌ , actual: ✓ ✅ ")
					} else {
						color.Danger.Println("expected: ✓ ✅ , actual: ✗ ❌ ")
					}
					color.Danger.Println("FAILED.")
					os.Exit(1)
				}
			}
		}

		return nil
	}
}

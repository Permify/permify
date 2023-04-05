package cmd

import (
	"context"
	"fmt"
	"net/url"
	`os`
	`strings`

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

// ErrList - error list
type ErrList struct {
	Errors []string
}

// Add - add error to error list
func (l *ErrList) Add(message string) {
	l.Errors = append(l.Errors, message)
}

// Print - print error list
func (l *ErrList) Print() {
	fmt.Println("")
	fmt.Println("fails:")
	for _, m := range l.Errors {
		color.Danger.Println(strings.ToLower("fail: " + strings.Replace(strings.Replace(m, "ERROR_CODE_", "", -1), "_", " ", -1)))
	}
	color.Danger.Println("FAILED")
}

// validate - permify validate command
func validate() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		var list = &ErrList{
			Errors: []string{},
		}

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

		color.Blue.Println("schema is creating... 🚀")

		// Write schema -
		var version string
		version, err = devContainer.S.WriteSchema(ctx, "t1", s.Schema)
		if err != nil {
			list.Add(err.Error())
			color.Danger.Printf("fail:       %s\n", strings.Replace(strings.Replace(err.Error(), "ERROR_CODE_", "", -1), "_", " ", -1))
			if len(list.Errors) != 0 {
				list.Print()
				os.Exit(1)
			}
		}

		if len(list.Errors) == 0 {
			color.Success.Println("success")
		}

		color.Blue.Println("relationships are creating... 🚀")

		// Write Relationships -
		for _, t := range s.Relationships {
			var tup *base.Tuple
			tup, err = tuple.Tuple(t)
			if err != nil {
				list.Add(err.Error())
			}

			_, err = devContainer.R.WriteRelationships(ctx, "t1", []*base.Tuple{
				tup,
			}, version)
			if err != nil {
				list.Add(fmt.Sprintf("fail: %s failed %s", t, err.Error()))
				color.Danger.Println(fmt.Sprintf("fail:       %s failed %s", t, strings.ToLower(strings.Replace(strings.Replace(err.Error(), "ERROR_CODE_", "", -1), "_", " ", -1))))
				continue
			}

			color.Success.Println(fmt.Sprintf("success:    %s ", t))
		}

		color.Blue.Println("checking assertions... 🚀")

		// Check Assertions
		for i, assertion := range s.Assertions {
			for query, expected := range assertion {
				exp := base.PermissionCheckResponse_RESULT_ALLOWED
				if !expected {
					exp = base.PermissionCheckResponse_RESULT_DENIED
				}

				q, err := tuple.NewQueryFromString(query)
				if err != nil {
					list.Add(err.Error())
				}

				res, err := devContainer.P.CheckPermissions(ctx, &base.PermissionCheckRequest{
					TenantId: "t1",
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
					list.Add(err.Error())
				}

				if res.Can == exp {
					color.Success.Print("success:    ")
					fmt.Printf("%v. %s ? passed \n", i+1, query)
				} else {
					color.Danger.Printf("fail:       %v. %s ? failed ", i+1, query)
					if res.Can == base.PermissionCheckResponse_RESULT_ALLOWED {
						color.Danger.Println("expected: DENIED actual: ALLOWED ")
						list.Add(fmt.Sprintf("fail: %s ? failed expected: DENIED actual: ALLOWED ", query))
					} else {
						color.Danger.Println("expected: ALLOWED actual: DENIED ")
						list.Add(fmt.Sprintf("fail: %s ? failed expected: ALLOWED actual: DENIED ", query))
					}
				}
			}
		}

		if len(list.Errors) != 0 {
			list.Print()
			os.Exit(1)
		}

		color.Blue.Println("schema successfully created")
		color.Blue.Println("relationships successfully created")
		color.Blue.Println("assertions successfully passed")

		color.Success.Println("SUCCESS")
		return nil
	}
}

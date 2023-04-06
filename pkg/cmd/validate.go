package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/gookit/color"
	"github.com/spf13/cobra"

	"github.com/Permify/permify/pkg/cmd/flags"
	"github.com/Permify/permify/pkg/development"
	"github.com/Permify/permify/pkg/development/validation"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// NewValidateCommand - creates a new validate command
func NewValidateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "validate <file>",
		Short: "validate authorization model with assertions",
		RunE:  validate(),
		Args:  cobra.ExactArgs(1),
	}

	// register flags for validation
	flags.RegisterValidationFlags(command)

	return command
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
		// print error message with color danger
		color.Danger.Println(strings.ToLower("fail: " + validationError(strings.Replace(strings.Replace(m, "ERROR_CODE_", "", -1), "_", " ", -1))))
	}
	// print FAILED with color danger
	color.Danger.Println("FAILED")
}

// validate returns a function that validates authorization model with assertions
func validate() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// set debug to false initially
		debug := false

		// get output format from viper
		format := viper.GetString("output-format")

		// if output format is not verbose or json, set it to verbose
		if format != "verbose" && format != "json" {
			format = "verbose"
		}

		// if output format is verbose, set debug to true
		if format == "verbose" {
			debug = true
		}

		// create an empty error list
		list := &ErrList{
			Errors: []string{},
		}

		// create a new context
		ctx := context.Background()

		// create a new development container
		devContainer := development.NewContainer()

		// parse the url from the first argument
		u, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		// create a new decoder from the url
		decoder, err := validation.NewDecoderFromURL(u)
		if err != nil {
			return err
		}

		// create a new shape
		s := &validation.Shape{}

		// decode the schema from the decoder
		err = decoder.Decode(s)
		if err != nil {
			return err
		}

		// if debug is true, print schema is creating with color blue
		if debug {
			color.Blue.Println("schema is creating... 🚀")
		}

		// write the schema
		var version string
		version, err = devContainer.S.WriteSchema(ctx, "t1", s.Schema)
		if err != nil {
			list.Add(err.Error())
			if debug {
				color.Danger.Printf("fail:       %s\n", validationError(err.Error()))
			}
			if len(list.Errors) != 0 {
				list.Print()
				os.Exit(1)
			}
		}

		// if there are no errors and debug is true, print success with color success
		if len(list.Errors) == 0 && debug {
			color.Success.Println("success")
		}

		// if debug is true, print relationships are creating with color blue
		if debug {
			color.Blue.Println("relationships are creating... 🚀")
		}

		// write relationships
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
				if debug {
					color.Danger.Println(fmt.Sprintf("fail:       %s failed %s", t, validationError(err.Error())))
				}
				continue
			}

			if debug {
				color.Success.Println(fmt.Sprintf("success:    %s ", t))
			}
		}

		// if debug is true, print checking assertions with color blue
		if debug {
			color.Blue.Println("checking assertions... 🚀")
		}

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
					if debug {
						color.Success.Print("success:    ")
						fmt.Printf("%v. %s ? passed \n", i+1, query)
					}
				} else {
					if debug {
						color.Danger.Printf("fail:       %v. %s ? failed ", i+1, query)
					}
					if res.Can == base.PermissionCheckResponse_RESULT_ALLOWED {
						if debug {
							color.Danger.Println("expected: DENIED actual: ALLOWED ")
						}
						list.Add(fmt.Sprintf("fail: %s ? failed expected: DENIED actual: ALLOWED ", query))
					} else {
						if debug {
							color.Danger.Println("expected: ALLOWED actual: DENIED ")
						}
						list.Add(fmt.Sprintf("fail: %s ? failed expected: ALLOWED actual: DENIED ", query))
					}
				}
			}
		}

		if len(list.Errors) != 0 {
			if debug {
				list.Print()
				os.Exit(1)
			}
			var b []byte
			b, err = json.Marshal(list.Errors)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			os.Exit(1)
		}

		if debug {
			color.Blue.Println("schema successfully created")
			color.Blue.Println("relationships successfully created")
			color.Blue.Println("assertions successfully passed")

			color.Success.Println("SUCCESS")
		}

		return nil
	}
}

// validationError - validation error
func validationError(message string) string {
	return strings.ToLower(strings.Replace(strings.Replace(message, "ERROR_CODE_", "", -1), "_", " ", -1))
}

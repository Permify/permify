package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/gookit/color"
	"github.com/rs/xid"
	"github.com/spf13/cobra"

	"github.com/Permify/permify/internal/storage"
	server_validation "github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/development"
	"github.com/Permify/permify/pkg/development/file"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
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
	color.Danger.Println("fails:")
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
		// create an empty error list
		list := &ErrList{
			Errors: []string{},
		}

		// create a new context
		ctx := context.Background()

		// create a new development container
		dev := development.NewContainer()

		// parse the url from the first argument
		u, err := url.Parse(args[0])
		if err != nil {
			return err
		}

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

		// if debug is true, print schema is creating with color blue
		color.Notice.Println("schema is creating... 🚀")

		sch, err := parser.NewParser(s.Schema).Parse()
		if err != nil {
			return err
		}

		_, _, err = compiler.NewCompiler(true, sch).Compile()
		if err != nil {
			return err
		}

		version := xid.New().String()

		cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
		for _, st := range sch.Statements {
			cnf = append(cnf, storage.SchemaDefinition{
				TenantID:             "t1",
				Version:              version,
				Name:                 st.GetName(),
				SerializedDefinition: []byte(st.String()),
			})
		}

		// write the schema
		err = dev.Container.SW.WriteSchema(ctx, cnf)
		if err != nil {
			list.Add(err.Error())
			color.Danger.Printf("fail: %s\n", validationError(err.Error()))
			if len(list.Errors) != 0 {
				list.Print()
				os.Exit(1)
			}
		}

		// if there are no errors and debug is true, print success with color success
		if len(list.Errors) == 0 {
			color.Success.Println("  success")
		}

		// if debug is true, print relationships are creating with color blue
		color.Notice.Println("relationships are creating... 🚀")

		// write relationships
		for _, t := range s.Relationships {
			var tup *base.Tuple
			tup, err = tuple.Tuple(t)
			if err != nil {
				list.Add(err.Error())
				continue
			}

			definition, _, err := dev.Container.SR.ReadEntityDefinition(ctx, "t1", tup.GetEntity().GetType(), version)
			if err != nil {
				return err
			}

			err = server_validation.ValidateTuple(definition, tup)
			if err != nil {
				return err
			}

			_, err = dev.Container.DW.Write(ctx, "t1", database.NewTupleCollection(tup), database.NewAttributeCollection())
			if err != nil {
				list.Add(fmt.Sprintf("%s failed %s", t, err.Error()))
				color.Danger.Println(fmt.Sprintf("fail: %s failed %s", t, validationError(err.Error())))
				continue
			}

			color.Success.Println(fmt.Sprintf("  success: %s ", t))
		}

		// write attributes
		for _, t := range s.Attributes {
			var attr *base.Attribute
			attr, err = attribute.Attribute(t)
			if err != nil {
				list.Add(err.Error())
				continue
			}

			definition, _, err := dev.Container.SR.ReadEntityDefinition(ctx, "t1", attr.GetEntity().GetType(), version)
			if err != nil {
				return err
			}

			err = server_validation.ValidateAttribute(definition, attr)
			if err != nil {
				return err
			}

			_, err = dev.Container.DW.Write(ctx, "t1", database.NewTupleCollection(), database.NewAttributeCollection(attr))
			if err != nil {
				list.Add(fmt.Sprintf("%s failed %s", t, err.Error()))
				color.Danger.Println(fmt.Sprintf("fail: %s failed %s", t, validationError(err.Error())))
				continue
			}

			color.Success.Println(fmt.Sprintf("  success: %s ", t))
		}

		// if debug is true, print checking assertions with color blue
		color.Notice.Println("checking scenarios... 🚀")

		// Check Assertions
		for sn, scenario := range s.Scenarios {
			color.Notice.Printf("%v.scenario: %s - %s\n", sn+1, scenario.Name, scenario.Description)
			color.Notice.Println("  checks:")

			for _, check := range scenario.Checks {
				entity, err := tuple.E(check.Entity)
				if err != nil {
					list.Add(err.Error())
					continue
				}

				ear, err := tuple.EAR(check.Subject)
				if err != nil {
					list.Add(err.Error())
					continue
				}

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				var cont *base.Context

				for permission, expected := range check.Assertions {
					exp := base.CheckResult_CHECK_RESULT_ALLOWED
					if !expected {
						exp = base.CheckResult_CHECK_RESULT_DENIED
					}

					res, err := dev.Container.Invoker.Check(ctx, &base.PermissionCheckRequest{
						TenantId: "t1",
						Metadata: &base.PermissionCheckRequestMetadata{
							SchemaVersion: version,
							SnapToken:     token.NewNoopToken().Encode().String(),
							Depth:         100,
						},
						Context:    cont,
						Entity:     entity,
						Permission: permission,
						Subject:    subject,
					})
					if err != nil {
						list.Add(err.Error())
						continue
					}

					query := tuple.SubjectToString(subject) + " " + permission + " " + tuple.EntityToString(entity)

					if res.Can == exp {
						color.Success.Print("    success:")
						fmt.Printf(" %s \n", query)
					} else {
						color.Danger.Printf("    fail: %s ->", query)
						if res.Can == base.CheckResult_CHECK_RESULT_ALLOWED {
							color.Danger.Println("  expected: DENIED actual: ALLOWED ")
							list.Add(fmt.Sprintf("%s -> expected: DENIED actual: ALLOWED ", query))
						} else {
							color.Danger.Println("  expected: ALLOWED actual: DENIED ")
							list.Add(fmt.Sprintf("%s -> expected: ALLOWED actual: DENIED ", query))
						}
					}
				}
			}

			color.Notice.Println("  entity_filters:")

			for _, filter := range scenario.EntityFilters {

				ear, err := tuple.EAR(filter.Subject)
				if err != nil {
					list.Add(err.Error())
					continue
				}

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				var cont *base.Context

				for permission, expected := range filter.Assertions {
					res, err := dev.Container.Invoker.LookupEntity(ctx, &base.PermissionLookupEntityRequest{
						TenantId: "t1",
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SchemaVersion: version,
							SnapToken:     token.NewNoopToken().Encode().String(),
							Depth:         100,
						},
						Context:    cont,
						EntityType: filter.EntityType,
						Permission: permission,
						Subject:    subject,
					})
					if err != nil {
						list.Add(err.Error())
						continue
					}

					query := tuple.SubjectToString(subject) + " " + permission + " " + filter.EntityType

					if isSameArray(res.GetEntityIds(), expected) {
						color.Success.Print("    success:")
						fmt.Printf(" %v\n", query)
					} else {
						color.Danger.Printf("    fail: %s -> expected: %+v actual: %+v\n", query, expected, res.GetEntityIds())
						list.Add(fmt.Sprintf("%s -> expected: %+v actual: %+v", query, expected, res.GetEntityIds()))
					}
				}
			}

			color.Notice.Println("  subject_filters:")

			for _, filter := range scenario.SubjectFilters {

				subjectReference := tuple.RelationReference(filter.SubjectReference)

				var entity *base.Entity
				entity, err = tuple.E(filter.Entity)
				if err != nil {
					list.Add(err.Error())
					continue
				}

				var cont *base.Context

				for permission, expected := range filter.Assertions {
					res, err := dev.Container.Invoker.LookupSubject(ctx, &base.PermissionLookupSubjectRequest{
						TenantId: "t1",
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SchemaVersion: version,
							SnapToken:     token.NewNoopToken().Encode().String(),
						},
						Context:          cont,
						SubjectReference: subjectReference,
						Permission:       permission,
						Entity:           entity,
					})
					if err != nil {
						list.Add(err.Error())
						continue
					}

					query := tuple.EntityToString(entity) + " " + permission + " " + filter.SubjectReference

					if isSameArray(res.GetSubjectIds(), expected) {
						color.Success.Print("    success:")
						fmt.Printf(" %v\n", query)
					} else {
						color.Danger.Printf("    fail: %s -> expected: %+v actual: %+v\n", query, expected, res.GetSubjectIds())
						list.Add(fmt.Sprintf("%s -> expected: %+v actual: %+v", query, expected, res.GetSubjectIds()))
					}
				}
			}
		}

		if len(list.Errors) != 0 {
			list.Print()
			os.Exit(1)
		}

		color.Notice.Println("schema successfully created")
		color.Notice.Println("relationships successfully created")
		color.Notice.Println("assertions successfully passed")
		color.Success.Println("SUCCESS")

		return nil
	}
}

// validationError - validation error
func validationError(message string) string {
	return strings.ToLower(strings.Replace(strings.Replace(message, "ERROR_CODE_", "", -1), "_", " ", -1))
}

// isSameArray - check if two arrays are the same
func isSameArray(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := make([]string, len(a))
	copy(sortedA, a)
	sort.Strings(sortedA)

	sortedB := make([]string, len(b))
	copy(sortedB, b)
	sort.Strings(sortedB)

	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}

	return true
}

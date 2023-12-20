package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/gookit/color"
	"github.com/rs/xid"
	"github.com/spf13/cobra"

	"github.com/Permify/permify/internal/storage"
	serverValidation "github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/development"
	"github.com/Permify/permify/pkg/development/file"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/schema"
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

		loader := schema.NewSchemaLoader()
		loaded, err := loader.LoadSchema(s.Schema)
		if err != nil {
			return err
		}

		sch, err := parser.NewParser(loaded).Parse()
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

		// Iterate over all relationships in the subject
		for _, t := range s.Relationships {
			// Convert each relationship to a Tuple
			var tup *base.Tuple
			tup, err = tuple.Tuple(t)
			// If an error occurs during the conversion, add the error message to the list and continue to the next iteration
			if err != nil {
				list.Add(err.Error())
				continue
			}

			// Retrieve the entity definition associated with the tuple's entity type
			definition, _, err := dev.Container.SR.ReadEntityDefinition(ctx, "t1", tup.GetEntity().GetType(), version)
			// If an error occurs while reading the entity definition, return the error
			if err != nil {
				return err
			}

			// Validate the tuple using the entity definition
			err = serverValidation.ValidateTuple(definition, tup)
			// If an error occurs during validation, return the error
			if err != nil {
				return err
			}

			// Write the validated tuple to the database
			_, err = dev.Container.DW.Write(ctx, "t1", database.NewTupleCollection(tup), database.NewAttributeCollection())
			// If an error occurs while writing to the database, add an error message to the list, log the error and continue to the next iteration
			if err != nil {
				list.Add(fmt.Sprintf("%s failed %s", t, err.Error()))
				color.Danger.Println(fmt.Sprintf("fail: %s failed %s", t, validationError(err.Error())))
				continue
			}

			// If the tuple was successfully written to the database, log a success message
			color.Success.Println(fmt.Sprintf("  success: %s ", t))
		}

		// if debug is true, print attributes are creating with color blue
		color.Notice.Println("attributes are creating... 🚀")

		// Iterate over all attributes in the subject
		for _, a := range s.Attributes {
			// Convert each attribute to an Attribute
			var attr *base.Attribute
			attr, err = attribute.Attribute(a)
			// If an error occurs during the conversion, add the error message to the list and continue to the next iteration
			if err != nil {
				list.Add(err.Error())
				continue
			}

			// Retrieve the entity definition associated with the attribute's entity type
			definition, _, err := dev.Container.SR.ReadEntityDefinition(ctx, "t1", attr.GetEntity().GetType(), version)
			// If an error occurs while reading the entity definition, return the error
			if err != nil {
				return err
			}

			// Validate the attribute using the entity definition
			err = serverValidation.ValidateAttribute(definition, attr)
			// If an error occurs during validation, return the error
			if err != nil {
				return err
			}

			// Write the validated attribute to the database
			_, err = dev.Container.DW.Write(ctx, "t1", database.NewTupleCollection(), database.NewAttributeCollection(attr))
			// If an error occurs while writing to the database, add an error message to the list, log the error and continue to the next iteration
			if err != nil {
				list.Add(fmt.Sprintf("%s failed %s", a, err.Error()))
				color.Danger.Println(fmt.Sprintf("fail: %s failed %s", a, validationError(err.Error())))
				continue
			}

			// If the attribute was successfully written to the database, log a success message
			color.Success.Println(fmt.Sprintf("  success: %s ", a))
		}

		// if debug is true, print checking assertions with color blue
		color.Notice.Println("checking scenarios... 🚀")

		// Check Assertions
		for sn, scenario := range s.Scenarios {
			color.Notice.Printf("%v.scenario: %s - %s\n", sn+1, scenario.Name, scenario.Description)

			// Start log output for checks
			color.Notice.Println("  checks:")

			// Iterate over all checks in the scenario
			for _, check := range scenario.Checks {
				// Extract entity from the check
				entity, err := tuple.E(check.Entity)
				if err != nil {
					list.Add(err.Error())
					continue
				}

				// Extract entity-attribute-relation from the check's subject
				ear, err := tuple.EAR(check.Subject)
				if err != nil {
					list.Add(err.Error())
					continue
				}

				// Define the subject based on the extracted entity-attribute-relation
				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				cont, err := Context(check.Context)
				if err != nil {
					list.Add(err.Error())
					continue
				}

				// Iterate over all assertions in the check
				for permission, expected := range check.Assertions {
					// Set expected result based on the assertion
					exp := base.CheckResult_CHECK_RESULT_ALLOWED
					if !expected {
						exp = base.CheckResult_CHECK_RESULT_DENIED
					}

					// Perform a permission check based on the context, entity, permission, and subject
					res, err := dev.Container.Invoker.Check(ctx, &base.PermissionCheckRequest{
						TenantId: "t1",
						Context:  cont,
						Metadata: &base.PermissionCheckRequestMetadata{
							SchemaVersion: version,
							SnapToken:     token.NewNoopToken().Encode().String(),
							Depth:         100,
						},
						Entity:     entity,
						Permission: permission,
						Subject:    subject,
					})
					if err != nil {
						list.Add(err.Error())
						continue
					}

					// Formulate the query string for log output
					query := tuple.SubjectToString(subject) + " " + permission + " " + tuple.EntityToString(entity)

					// If the check result matches the expected result, log a success message
					if res.Can == exp {
						color.Success.Print("    success:")
						fmt.Printf(" %s \n", query)
					} else {
						// If the check result does not match the expected result, log a failure message
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

			// Start of the entity filter processing.
			color.Notice.Println("  entity_filters:")

			// Iterate over each entity filter in the scenario.
			for _, filter := range scenario.EntityFilters {

				// Convert the subject from the filter into a base.Subject.
				ear, err := tuple.EAR(filter.Subject)
				if err != nil {
					// If an error occurs, add it to the list and continue to the next filter.
					list.Add(err.Error())
					continue
				}

				// Create a new base.Subject from the Entity-Attribute-Relation (EAR).
				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				// Convert the filter context into a base.Context.
				cont, err := Context(filter.Context)
				if err != nil {
					// If an error occurs, add it to the list and continue to the next filter.
					list.Add(err.Error())
					continue
				}

				// Iterate over each assertion in the filter.
				for permission, expected := range filter.Assertions {
					// Perform a permission lookup for the entity.
					res, err := dev.Container.Invoker.LookupEntity(ctx, &base.PermissionLookupEntityRequest{
						TenantId: "t1",
						Context:  cont,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SchemaVersion: version,
							SnapToken:     token.NewNoopToken().Encode().String(),
							Depth:         100,
						},
						EntityType: filter.EntityType,
						Permission: permission,
						Subject:    subject,
					})
					if err != nil {
						// If an error occurs, add it to the list and continue to the next assertion.
						list.Add(err.Error())
						continue
					}

					// Format the subject, permission, and entity type as a string for logging.
					query := tuple.SubjectToString(subject) + " " + permission + " " + filter.EntityType

					// Check if the actual result matches the expected result.
					if isSameArray(res.GetEntityIds(), expected) {
						// If the results match, log a success message.
						color.Success.Print("    success:")
						fmt.Printf(" %v\n", query)
					} else {
						// If the results don't match, log a failure message with the expected and actual results.
						color.Danger.Printf("    fail: %s -> expected: %+v actual: %+v\n", query, expected, res.GetEntityIds())
						list.Add(fmt.Sprintf("%s -> expected: %+v actual: %+v", query, expected, res.GetEntityIds()))
					}
				}
			}

			// Print a message indicating the start of the subject filter processing.
			color.Notice.Println("  subject_filters:")

			// Iterate over each subject filter in the scenario.
			for _, filter := range scenario.SubjectFilters {

				// Convert the subject reference from the filter into a relation reference.
				subjectReference := tuple.RelationReference(filter.SubjectReference)

				// Convert the entity from the filter into a base.Entity.
				var entity *base.Entity
				entity, err = tuple.E(filter.Entity)
				if err != nil {
					// If an error occurs, add it to the list and continue to the next filter.
					list.Add(err.Error())
					continue
				}

				// Convert the filter context into a base.Context.
				cont, err := Context(filter.Context)
				if err != nil {
					// If an error occurs, add it to the list and continue to the next filter.
					list.Add(err.Error())
					continue
				}

				// Iterate over each assertion in the filter.
				for permission, expected := range filter.Assertions {
					// Perform a permission lookup for the subject.
					res, err := dev.Container.Invoker.LookupSubject(ctx, &base.PermissionLookupSubjectRequest{
						TenantId: "t1",
						Context:  cont,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SchemaVersion: version,
							SnapToken:     token.NewNoopToken().Encode().String(),
							Depth:         100,
						},
						SubjectReference: subjectReference,
						Permission:       permission,
						Entity:           entity,
					})
					if err != nil {
						// If an error occurs, add it to the list and continue to the next assertion.
						list.Add(err.Error())
						continue
					}

					// Format the entity, permission, and subject reference as a string for logging.
					query := tuple.EntityToString(entity) + " " + permission + " " + filter.SubjectReference

					// Check if the actual result matches the expected result.
					if isSameArray(res.GetSubjectIds(), expected) {
						// If the results match, log a success message.
						color.Success.Print("    success:")
						fmt.Printf(" %v\n", query)
					} else {
						// If the results don't match, log a failure message with the expected and actual results.
						color.Danger.Printf("    fail: %s -> expected: %+v actual: %+v\n", query, expected, res.GetSubjectIds())
						list.Add(fmt.Sprintf("%s -> expected: %+v actual: %+v", query, expected, res.GetSubjectIds()))
					}
				}
			}
		}

		// If the error list is not empty, there were some errors during processing.
		if len(list.Errors) != 0 {
			// Print the errors collected during processing.
			list.Print()
			// Exit the program with a status of 1 to indicate an error.
			os.Exit(1)
		}

		// If there are no errors, print the success messages.
		color.Notice.Println("schema successfully created")
		color.Notice.Println("relationships successfully created")
		color.Notice.Println("assertions successfully passed")

		// Final success message to indicate everything completed successfully.
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

// Context is a function that takes a file context and returns a base context and an error.
func Context(fileContext file.Context) (cont *base.Context, err error) {
	// Initialize an empty base context to be populated from the file context.
	cont = &base.Context{
		Tuples:     []*base.Tuple{},
		Attributes: []*base.Attribute{},
		Data:       nil,
	}

	// Convert the file context's data to a Struct object.
	st, err := structpb.NewStruct(fileContext.Data)
	if err != nil {
		// If an error occurs, return it.
		return nil, err
	}

	// Assign the Struct object to the context's data field.
	cont.Data = st

	// Iterate over the file context's tuples.
	for _, t := range fileContext.Tuples {
		// Convert each tuple to a base tuple.
		tup, err := tuple.Tuple(t)
		if err != nil {
			// If an error occurs, return it.
			return nil, err
		}

		// Add the converted tuple to the context's tuples slice.
		cont.Tuples = append(cont.Tuples, tup)
	}

	// Iterate over the file context's attributes.
	for _, t := range fileContext.Attributes {
		// Convert each attribute to a base attribute.
		attr, err := attribute.Attribute(t)
		if err != nil {
			// If an error occurs, return it.
			return nil, err
		}

		// Add the converted attribute to the context's attributes slice.
		cont.Attributes = append(cont.Attributes, attr)
	}

	// If everything goes well, return the context and a nil error.
	return cont, nil
}

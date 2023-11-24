package bundle

import (
	"bytes"
	"html/template"

	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

func Operation(arguments map[string]string, operation *base.Operation) (tb database.TupleBundle, ab database.AttributeBundle, err error) {
	rWrites := operation.GetRelationshipsWrite()

	// Initialize a TupleCollection to store the processed tuples.
	var wtc database.TupleCollection

	// Iterate over each write operation.
	for _, w := range rWrites {
		// Parse the write operation string into a template.
		tmpl, err := template.New("template").Parse(w)
		if err != nil {
			return tb, ab, err
		}

		// Create a buffer to hold the template execution output.
		var buf bytes.Buffer

		// Execute the template with the provided arguments and store the output in the buffer.
		err = tmpl.Execute(&buf, arguments)
		if err != nil {
			return tb, ab, err
		}

		// Convert the write string into a tuple.
		t, err := tuple.Tuple(buf.String())
		if err != nil {
			return tb, ab, err
		}

		// Add the tuple to the write tuple collection.
		wtc.Add(t)
	}

	// Assign the processed write tuples to the tuple bundle.
	tb.Write = wtc

	// Retrieve the delete operations from the operation object.
	rDeletes := operation.GetRelationshipsDelete()

	// Initialize a TupleCollection to store the delete tuples.
	var dtc database.TupleCollection

	// Iterate over each delete operation.
	for _, w := range rDeletes {
		// Parse the write operation string into a template.
		tmpl, err := template.New("template").Parse(w)
		if err != nil {
			return tb, ab, err
		}

		// Create a buffer to hold the template execution output.
		var buf bytes.Buffer

		// Execute the template with the provided arguments and store the output in the buffer.
		err = tmpl.Execute(&buf, arguments)
		if err != nil {
			return tb, ab, err
		}

		// Convert the delete string into a tuple.
		t, err := tuple.Tuple(buf.String())
		if err != nil {
			return tb, ab, err
		}

		// Add the tuple to the delete tuple collection.
		dtc.Add(t)
	}

	// Assign the processed delete tuples to the tuple bundle.
	tb.Delete = dtc

	// Retrieve the write operations from the operation object.
	aWrites := operation.GetAttributesWrite()

	// Initialize an AttributeCollection to store the processed attributes.
	var wac database.AttributeCollection

	// Iterate over each write operation.
	for _, w := range aWrites {
		// Parse the write operation string into a template.
		tmpl, err := template.New("template").Parse(w)
		if err != nil {
			return tb, ab, err
		}

		// Create a buffer to hold the template execution output.
		var buf bytes.Buffer

		// Execute the template with the provided arguments and store the output in the buffer.
		err = tmpl.Execute(&buf, arguments)
		if err != nil {
			return tb, ab, err
		}

		// Convert the write string into an attribute.
		a, err := attribute.Attribute(buf.String())
		if err != nil {
			return tb, ab, err
		}

		// Add the attribute to the write attribute collection.
		wac.Add(a)
	}

	// Assign the processed write attributes to the attribute bundle.
	ab.Write = wac

	// Retrieve the delete operations from the operation object.
	aDeletes := operation.GetAttributesDelete()

	// Initialize an AttributeCollection to store the delete attributes.
	var dac database.AttributeCollection

	// Iterate over each delete operation.
	for _, w := range aDeletes {
		// Parse the write operation string into a template.
		tmpl, err := template.New("template").Parse(w)
		if err != nil {
			return tb, ab, err
		}

		// Create a buffer to hold the template execution output.
		var buf bytes.Buffer

		// Execute the template with the provided arguments and store the output in the buffer.
		err = tmpl.Execute(&buf, arguments)
		if err != nil {
			return tb, ab, err
		}

		// Convert the delete string into an attribute.
		a, err := attribute.Attribute(buf.String())
		if err != nil {
			return tb, ab, err
		}

		// Add the attribute to the delete attribute collection.
		dac.Add(a)
	}

	// Assign the processed delete attributes to the attribute bundle.
	ab.Delete = dac

	return tb, ab, nil
}

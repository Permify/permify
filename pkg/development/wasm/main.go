//go:build wasm
// +build wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/Permify/permify/pkg/development"
	"github.com/Permify/permify/pkg/development/graph"
)

// Requests for Permify Playground

var dev *development.Development

func run() js.Func {
	// Returns a new JavaScript function that wraps the Go function.
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create a new development container.
		dev = development.NewContainer()

		// Create an empty map to hold the JSON content.
		var t interface{}

		// Unmarshal the JSON received from the args into the map t.
		// args[0].String() is the JSON string.
		err := json.Unmarshal([]byte(args[0].String()), &t)
		// If there is an error in unmarshaling, return the error to JavaScript.
		if err != nil {
			return js.ValueOf([]interface{}{"invalid JSON"})
		}

		input, ok := t.(map[string]interface{})
		// If the JSON content is not a map, return an error to JavaScript.
		if !ok {
			return js.ValueOf([]interface{}{"invalid JSON"})
		}

		errors := dev.Run(context.Background(), input)

		if len(errors) == 0 {
			return js.ValueOf([]interface{}{})
		}

		vs := make([]interface{}, 0, len(errors))
		for _, r := range errors {
			result, err := json.Marshal(r)
			if err != nil {
				// If there's an error marshaling, we immediately return.
				// This is an important behavior decision.
				return js.ValueOf([]interface{}{err.Error()})
			}
			vs = append(vs, string(result))
		}

		return js.ValueOf(vs)
	})
}

// readSchemaGraph is a function that reads the Permify schema and converts it to a graph representation.
// It returns a JavaScript function that can be invoked from JavaScript code,
// enabling the interaction between JavaScript and Go code.
func visualize() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Call the ReadSchema method to read the schema.
		sch, err := dev.ReadSchema(context.Background())
		// If there's an error in reading the schema, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		// Convert the read schema to a graph representation.
		r, err := graph.NewBuilder(sch).SchemaToGraph()
		// If there's an error in converting, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		// Marshal the nodes and edges of the graph into JSON format.
		result, err := json.Marshal(struct {
			Nodes []*graph.Node `json:"nodes"`
			Edges []*graph.Edge `json:"edges"`
		}{Nodes: r.Nodes(), Edges: r.Edges()})
		// If there's an error in marshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}

		// Marshal the read schema into JSON format.
		s, err := protojson.Marshal(sch)
		// If there's an error in marshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}

		// Return the marshalled graph and no error.
		return js.ValueOf([]interface{}{string(result), string(s), nil})
	})
}

func main() {
	ch := make(chan struct{}, 0)
	dev = development.NewContainer()
	js.Global().Set("run", run())
	js.Global().Set("visualize", visualize())
	<-ch
}

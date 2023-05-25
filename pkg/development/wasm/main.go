//go:build wasm
// +build wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/development"
	"github.com/Permify/permify/pkg/development/graph"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// Requests for Permify Playground

var dev *development.Development

// check is a function that checks permissions for a specific request.
// It returns a JavaScript function that can be called from JavaScript code,
// allowing you to interact with your Go code from a JavaScript environment.
func check() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create an instance of the PermissionCheckRequest message.
		params := &v1.PermissionCheckRequest{}
		// Unmarshal the input JSON string into the params struct.
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		// If there's an error in unmarshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{false, err.Error()})
		}

		var result *v1.PermissionCheckResponse
		// Call the Check method to check the permissions.
		result, err = dev.Check(context.Background(), params.Subject, params.Permission, params.Entity)
		// If there's an error in checking the permissions, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{false, err.Error()})
		}

		// If the permission check response indicates that the action is allowed,
		// return true and no error.
		if result.GetCan() == v1.PermissionCheckResponse_RESULT_ALLOWED {
			return js.ValueOf([]interface{}{true, nil})
		}

		// If the permission check response doesn't indicate that the action is allowed,
		// return false and no error.
		return js.ValueOf([]interface{}{false, nil})
	})
}

// lookupEntity is a function that looks up an entity for a specific request.
// It returns a JavaScript function that can be called from JavaScript code,
// allowing you to interact with your Go code from a JavaScript environment.
func lookupEntity() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create an instance of the PermissionLookupEntityRequest message.
		params := &v1.PermissionLookupEntityRequest{}
		// Unmarshal the input JSON string into the params struct.
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		// If there's an error in unmarshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{[]interface{}{}, err.Error()})
		}

		var result *v1.PermissionLookupEntityResponse
		// Call the LookupEntity method to get the entities.
		result, err = dev.LookupEntity(context.Background(), params.Subject, params.Permission, params.EntityType)
		// If there's an error in looking up the entity, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{[]interface{}{}, err.Error()})
		}

		// Prepare a slice to store the entity IDs.
		ids := make([]interface{}, len(result.GetEntityIds()))
		// Loop through the result and add each entity ID to the ids slice.
		for i, v := range result.GetEntityIds() {
			ids[i] = v
		}

		// Return the list of entity IDs and no error.
		return js.ValueOf([]interface{}{ids, nil})
	})
}

// lookupSubject is a function that looks up a subject for a specific request.
// It returns a JavaScript function that can be invoked from JavaScript code,
// enabling the interaction with the Go code from a JavaScript environment.
func lookupSubject() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create an instance of the PermissionLookupSubjectRequest message.
		params := &v1.PermissionLookupSubjectRequest{}
		// Unmarshal the input JSON string into the params struct.
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		// If there's an error in unmarshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{[]interface{}{}, err.Error()})
		}

		var result *v1.PermissionLookupSubjectResponse
		// Call the LookupSubject method to get the subjects.
		result, err = dev.LookupSubject(context.Background(), params.Entity, params.Permission, params.SubjectReference)
		// If there's an error in looking up the subject, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{[]interface{}{}, err.Error()})
		}

		// Prepare a slice to store the subject IDs.
		ids := make([]interface{}, len(result.GetSubjectIds()))
		// Loop through the result and add each subject ID to the ids slice.
		for i, v := range result.GetSubjectIds() {
			ids[i] = v
		}

		// Return the list of subject IDs and no error.
		return js.ValueOf([]interface{}{ids, nil})
	})
}

// writeSchema is a function that writes a schema based on a provided JSON string.
// It returns a JavaScript function that can be invoked from JavaScript code,
// allowing for interaction between JavaScript and Go code.
func writeSchema() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Call the WriteSchema method with the provided JSON string.
		err := dev.WriteSchema(context.Background(), string(args[0].String()))
		// If there's an error in writing the schema, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		// If the schema is written successfully, return nil error.
		return js.ValueOf([]interface{}{nil})
	})
}

// writeTuple is a function that writes relation tuples for a specific request.
// It returns a JavaScript function that can be called from JavaScript code,
// allowing the interaction with the Go code from a JavaScript environment.
func writeTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create an instance of the Tuple message.
		t := &v1.Tuple{}
		// Unmarshal the input JSON string into the tuple struct.
		err := protojson.Unmarshal([]byte(string(args[0].String())), t)
		// If there's an error in unmarshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}

		// Call the WriteTuple method with the unmarshalled tuple.
		err = dev.WriteTuple(context.Background(), []*v1.Tuple{t})
		// If there's an error in writing the tuple, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}

		// If the tuple is written successfully, return nil error.
		return js.ValueOf([]interface{}{nil})
	})
}

// deleteTuple is a function that deletes a relation tuple based on a provided JSON string.
// It returns a JavaScript function that can be invoked from JavaScript code,
// enabling the interaction between JavaScript and Go code.
func deleteTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create an instance of the Tuple message.
		t := &v1.Tuple{}
		// Unmarshal the input JSON string into the tuple struct.
		err := protojson.Unmarshal([]byte(string(args[0].String())), t)
		// If there's an error in unmarshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}

		// Call the DeleteTuple method with the constructed TupleFilter.
		_, err = dev.DeleteTuple(context.Background(), &v1.TupleFilter{
			Entity: &v1.EntityFilter{
				Type: t.GetEntity().GetType(),
				Ids:  []string{t.GetEntity().GetId()},
			},
			Relation: t.GetRelation(),
			Subject: &v1.SubjectFilter{
				Type:     t.GetSubject().GetType(),
				Ids:      []string{t.GetSubject().GetId()},
				Relation: t.GetSubject().GetRelation(),
			},
		})
		// If there's an error in deleting the tuple, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}

		// If the tuple is deleted successfully, return nil error.
		return js.ValueOf([]interface{}{nil})
	})
}

// readSchema is a function that reads the Permify schema.
// It returns a JavaScript function that can be invoked from JavaScript code,
// enabling interaction between JavaScript and Go code.
func readSchema() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Call the ReadSchema method to read the schema.
		sch, err := dev.ReadSchema(context.Background())
		// If there's an error in reading the schema, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		// Marshal the read schema into JSON format.
		result, err := protojson.Marshal(sch)
		// If there's an error in marshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		// Return the marshalled schema and no error.
		return js.ValueOf([]interface{}{string(result), nil})
	})
}

// readTuple is a function that reads and filters relation tuples based on a provided JSON string.
// It returns a JavaScript function that can be invoked from JavaScript code,
// enabling interaction between JavaScript and Go code.
func readTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Create an instance of the TupleFilter message.
		filter := &v1.TupleFilter{}
		// Unmarshal the input JSON string into the TupleFilter struct.
		err := protojson.Unmarshal([]byte(string(args[0].String())), filter)
		// If there's an error in unmarshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}

		// Call the ReadTuple method with the unmarshalled filter.
		var collection *database.TupleCollection
		collection, _, err = dev.ReadTuple(context.Background(), filter)
		// If there's an error in reading the tuple, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}

		// Create a Tuples instance with the returned collection.
		t := &v1.Tuples{
			Tuples: collection.GetTuples(),
		}
		// Marshal the tuples into JSON format.
		var result []byte
		result, err = protojson.Marshal(t)
		// If there's an error in marshalling, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}

		// Return the marshalled tuples and no error.
		return js.ValueOf([]interface{}{string(result), nil})
	})
}

// readSchemaGraph is a function that reads the Permify schema and converts it to a graph representation.
// It returns a JavaScript function that can be invoked from JavaScript code,
// enabling the interaction between JavaScript and Go code.
func readSchemaGraph() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Call the ReadSchema method to read the schema.
		sch, err := dev.ReadSchema(context.Background())
		// If there's an error in reading the schema, return the error.
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		// Convert the read schema to a graph representation.
		r, err := graph.SchemaToGraph(sch)
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
		// Return the marshalled graph and no error.
		return js.ValueOf([]interface{}{string(result), nil})
	})
}

func main() {
	ch := make(chan struct{}, 0)
	dev = development.NewContainer()
	js.Global().Set("check", check())
	js.Global().Set("lookupEntity", lookupEntity())
	js.Global().Set("lookupSubject", lookupSubject())
	js.Global().Set("writeSchema", writeSchema())
	js.Global().Set("writeTuple", writeTuple())
	js.Global().Set("readSchema", readSchema())
	js.Global().Set("readTuple", readTuple())
	js.Global().Set("deleteTuple", deleteTuple())
	js.Global().Set("readSchemaGraph", readSchemaGraph())
	<-ch
}

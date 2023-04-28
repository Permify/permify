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

// check - Permission check request
func check() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		params := &v1.PermissionCheckRequest{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		if err != nil {
			return js.ValueOf([]interface{}{false, err.Error()})
		}
		var result *v1.PermissionCheckResponse
		result, err = dev.Check(context.Background(), params.Subject, params.Permission, params.Entity)
		if err != nil {
			return js.ValueOf([]interface{}{false, err.Error()})
		}
		if result.GetCan() == v1.PermissionCheckResponse_RESULT_ALLOWED {
			return js.ValueOf([]interface{}{true, nil})
		}
		return js.ValueOf([]interface{}{false, nil})
	})
}

// lookupEntity -
func lookupEntity() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		params := &v1.PermissionLookupEntityRequest{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		if err != nil {
			return js.ValueOf([]interface{}{[]string{}, err.Error()})
		}
		var result *v1.PermissionLookupEntityResponse
		result, err = dev.LookupEntity(context.Background(), params.Subject, params.Permission, params.EntityType)
		if err != nil {
			return js.ValueOf([]interface{}{[]string{}, err.Error()})
		}
		ids := make([]interface{}, len(result.GetEntityIds()))
		for i, v := range result.GetEntityIds() {
			ids[i] = v
		}
		return js.ValueOf([]interface{}{ids, nil})
	})
}

// writeSchema - Writes schema
func writeSchema() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		err := dev.WriteSchema(context.Background(), string(args[0].String()))
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		return js.ValueOf([]interface{}{nil})
	})
}

// writeTuple - Writes relation tuples
func writeTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		t := &v1.Tuple{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), t)
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		err = dev.WriteTuple(context.Background(), []*v1.Tuple{t})
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		return js.ValueOf([]interface{}{nil})
	})
}

// deleteTuple - Delete relation tuple
func deleteTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		t := &v1.Tuple{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), t)
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
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
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		return js.ValueOf([]interface{}{nil})
	})
}

// readSchema - Read Permify Schema
func readSchema() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		sch, err := dev.ReadSchema(context.Background())
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		result, err := protojson.Marshal(sch)
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		return js.ValueOf([]interface{}{string(result), nil})
	})
}

// readTuple - Read, filter relation tuples
func readTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		filter := &v1.TupleFilter{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), filter)
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		var collection *database.TupleCollection
		collection, _, err = dev.ReadTuple(context.Background(), filter)
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		var result []byte
		t := &v1.Tuples{
			Tuples: collection.GetTuples(),
		}
		result, err = protojson.Marshal(t)
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		return js.ValueOf([]interface{}{string(result), nil})
	})
}

// readSchemaGraph - read schema graph
func readSchemaGraph() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		sch, err := dev.ReadSchema(context.Background())
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		r, err := graph.SchemaToGraph(sch)
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		result, err := json.Marshal(struct {
			Nodes []*graph.Node `json:"nodes"`
			Edges []*graph.Edge `json:"edges"`
		}{Nodes: r.Nodes(), Edges: r.Edges()})
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		return js.ValueOf([]interface{}{string(result), nil})
	})
}

func main() {
	ch := make(chan struct{}, 0)
	dev = development.NewContainer()
	js.Global().Set("check", check())
	js.Global().Set("lookupEntity", lookupEntity())
	js.Global().Set("writeSchema", writeSchema())
	js.Global().Set("writeTuple", writeTuple())
	js.Global().Set("readSchema", readSchema())
	js.Global().Set("readTuple", readTuple())
	js.Global().Set("deleteTuple", deleteTuple())
	js.Global().Set("readSchemaGraph", readSchemaGraph())
	<-ch
}

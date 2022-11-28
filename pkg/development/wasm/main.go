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

var dev *development.Container

// check -
func check() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		params := &v1.PermissionCheckRequest{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		if err != nil {
			return js.ValueOf([]interface{}{false, err.Error()})
		}
		var result *v1.PermissionCheckResponse
		result, err = development.Check(context.Background(), dev.P, params.Subject, params.Permission, params.Entity, string(args[1].String()), "")
		if err != nil {
			return js.ValueOf([]interface{}{false, err.Error()})
		}
		return js.ValueOf([]interface{}{result.GetCan(), nil})
	})
}

// lookupEntity -
func lookupEntity() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		params := &v1.PermissionLookupEntityRequest{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		if err != nil {
			return js.ValueOf([]interface{}{false, err.Error()})
		}
		var result *v1.PermissionLookupEntityResponse
		result, err = development.LookupEntity(context.Background(), dev.P, params.Subject, params.Permission, params.EntityType, string(args[1].String()), "")
		if err != nil {
			return js.ValueOf([]interface{}{[]string{}, err.Error()})
		}
		return js.ValueOf([]interface{}{result.GetEntityIds(), nil})
	})
}

// writeSchema -
func writeSchema() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		version, err := development.WriteSchema(context.Background(), dev.S, string(args[0].String()))
		if err != nil {
			return js.ValueOf([]interface{}{"", err.Error()})
		}
		return js.ValueOf([]interface{}{version, nil})
	})
}

// writeTuple -
func writeTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		t := &v1.Tuple{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), t)
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		_, err = development.WriteTuple(context.Background(), dev.R, []*v1.Tuple{t}, string(args[1].String()))
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		return js.ValueOf([]interface{}{nil})
	})
}

// deleteTuple -
func deleteTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		t := &v1.Tuple{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), t)
		if err != nil {
			return js.ValueOf([]interface{}{err.Error()})
		}
		_, err = development.DeleteTuple(context.Background(), dev.R, &v1.TupleFilter{
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

// readSchema -
func readSchema() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		sch, err := development.ReadSchema(context.Background(), dev.S, string(args[0].String()))
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

// readTuple -
func readTuple() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		params := &v1.TupleFilter{}
		err := protojson.Unmarshal([]byte(string(args[0].String())), params)
		if err != nil {
			return js.ValueOf([]interface{}{nil, err.Error()})
		}
		var collection database.ITupleCollection
		collection, err = development.ReadTuple(context.Background(), dev.R, params, "")
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

// readSchemaGraph -
func readSchemaGraph() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		sch, err := development.ReadSchema(context.Background(), dev.S, string(args[0].String()))
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

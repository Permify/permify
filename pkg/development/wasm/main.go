//go:build wasm
// +build wasm

package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/pkg/development"
)

var dev *development.Development

// check -
func check() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var err error
		params := &development.CheckQuery{}
		err = json.Unmarshal([]byte(string(args[0].String())), params)
		if err != nil {
			return err.Error()
		}
		var result commands.CheckResponse
		result, err = development.Check(context.Background(), dev.P, params.Subject, params.Action, params.Entity)
		if err != nil {
			return err.Error()
		}
		return result.Can
	})
}

// writeSchema -
func writeSchema() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var err error
		var version string
		version, err = development.WriteSchema(context.Background(), dev.M, string(args[0].String()))
		if err != nil {
			return err.Error()
		}
		return version
	})
}

func main() {
	ch := make(chan struct{}, 0)
	dev = development.NewDevelopment()

	js.Global().Set("check", check())
	js.Global().Set("writeSchema", writeSchema())
	<-ch
}

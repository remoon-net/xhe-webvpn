package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"
)

var ErrCanceledThroughJsSignal = errors.New("canceled through js signal")

var DOMException = js.Global().Get("DOMException")

func signal2ctx(signal js.Value) context.Context {
	ctx := context.Background()
	if signal.IsUndefined() {
		return ctx
	}
	ctx, cause := context.WithCancelCause(ctx)
	signal.Call("addEventListener", "abort", js.FuncOf(func(this js.Value, args []js.Value) any {
		reason := this.Get("reason")
		if reason.InstanceOf(DOMException) && reason.Get("name").String() == "AbortError" {
			cause(nil)
			return js.Undefined()
		}
		msg := reason.Call("toString").String()
		err := fmt.Errorf("%w. %s", ErrCanceledThroughJsSignal, msg)
		cause(err)
		return js.Undefined()
	}))
	return ctx
}

func getConfig[T any](v js.Value) (c T, err error) {
	s := js.Global().Get("JSON").Call("stringify", v).String()
	err = json.Unmarshal([]byte(s), &c)
	return
}

//go:build js && wasm

package main

import (
	"syscall/js"

	"tlog.app/go/errors"

	"nikand.dev/go/timeline/go/timeline"
)

func main() {
	js.Global().Set("processTimeline", js.FuncOf(do))

	select {}
}

func do(this js.Value, args []js.Value) any {
	file := args[0]
	l := file.Get("length").Int()

	b := make([]byte, l)
	js.CopyBytesToGo(b, file)

	points, err := timeline.Parse(b, nil)
	if err != nil {
		return errors.Wrap(err, "parse timeline")
	}

	flat := js.Global().Get("Float32Array").New(2 * len(points))

	for i, p := range points {
		flat.SetIndex(2*i+0, p.X)
		flat.SetIndex(2*i+1, p.Y)
	}

	return flat
}

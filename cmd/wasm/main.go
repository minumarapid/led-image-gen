//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"syscall/js"

	"led-image-gen/processor"
)

func main() {
	js.Global().Set("generateLEDImage", js.FuncOf(generateLEDImage))

	<-make(chan struct{})
}

func generateLEDImage(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return "Error: insufficient arguments"
	}

	configJsonStr := args[1].String()

	cfg := processor.DefaultConfig()
	cfg.MaxWorkers = 1
	if err := json.Unmarshal([]byte(configJsonStr), cfg); err != nil {
		return "Error: invalid config JSON"
	}

	jsFileData := args[0]
	goFileData := make([]byte, jsFileData.Get("length").Int())
	js.CopyBytesToGo(goFileData, jsFileData)
	src, _, err := image.Decode(bytes.NewReader(goFileData))

	if err != nil {
		return "Error: invalid image"
	}

	result, err := processor.GenerateLEDImage(src, cfg)
	if err != nil {
		return "Error: failed to process image"
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, result); err != nil {
		return "Error: failed to encode image"
	}
	jsArray := js.Global().Get("Uint8Array").New(buf.Len())
	js.CopyBytesToJS(jsArray, buf.Bytes())

	return jsArray
}

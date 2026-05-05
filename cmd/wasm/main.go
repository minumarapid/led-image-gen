//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
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
		errObj := js.Global().Get("Error").New("insufficient arguments")
		return errObj
	}

	var configJsonStr string
	if args[1].Type() == js.TypeObject {
		configJsonStr = js.Global().Get("JSON").Call("stringify", args[1]).String()
	} else {
		configJsonStr = args[1].String()
	}

	cfg := processor.DefaultConfig()
	if err := json.Unmarshal([]byte(configJsonStr), cfg); err != nil {
		errObj := js.Global().Get("Error").New(fmt.Sprintf("Invalid config: %v", err))
		return errObj
	}
	cfg.MaxWorkers = 1

	jsFileData := args[0]
	uint8ArrayCtor := js.Global().Get("Uint8Array")
	uint8Arr := jsFileData
	if !jsFileData.InstanceOf(uint8ArrayCtor) {
		uint8Arr = uint8ArrayCtor.New(jsFileData)
	}
	goFileData := make([]byte, uint8Arr.Get("length").Int())
	js.CopyBytesToGo(goFileData, uint8Arr)

	src, _, err := image.Decode(bytes.NewReader(goFileData))
	if err != nil {
		errObj := js.Global().Get("Error").New(fmt.Sprintf("Failed to decode image: %v", err))
		return errObj
	}

	result, err := processor.GenerateLEDImage(src, cfg)
	if err != nil {
		errObj := js.Global().Get("Error").New(fmt.Sprintf("Failed to process image: %v", err))
		return errObj
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, result); err != nil {
		errObj := js.Global().Get("Error").New(fmt.Sprintf("Failed to encode image: %v", err))
		return errObj
	}
	jsArray := js.Global().Get("Uint8Array").New(buf.Len())
	js.CopyBytesToJS(jsArray, buf.Bytes())

	return jsArray
}

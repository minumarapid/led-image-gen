# WASM bridge

This package exposes a single global function for browsers:

- `generateLEDImage(imageBytes, config)`

`imageBytes` must be a `Uint8Array` or `ArrayBuffer`. `config` can be a JSON string or a plain object. The function returns a `Uint8Array` with PNG bytes or an `Error` object.

## Config keys

```json
{
  "border": 10,
  "ledSize": 4,
  "ledGap": 2,
  "ledGamma": 1.0,
  "ledExposure": 1.0,
  "ledShape": "circle",
  "enableGlow": true,
  "glowRange": 3.0,
  "glowStrength": 1.75,
  "glowGamma": 1.0,
  "glowExposure": 1.0,
  "offLightColor": {
    "R": 50,
    "G": 50,
    "B": 50,
    "A": 255
  }
}
```

## Build notes

```bash
# Copy wasm_exec.js once (Go provides it)
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/public/

# Build wasm
GOOS=js GOARCH=wasm go build -o web/public/led-image-gen.wasm ./cmd/wasm
```

## JS usage sketch

```js
// Load wasm (wasm_exec.js is required)
const go = new Go();
const resp = await fetch("/led-image-gen.wasm");
const bytes = await resp.arrayBuffer();
const { instance } = await WebAssembly.instantiate(bytes, go.importObject);
await go.run(instance);

// Call the Go function
const pngBytes = generateLEDImage(imageBytes, { ledSize: 6, ledShape: "square" });
if (pngBytes instanceof Error) throw pngBytes;
```


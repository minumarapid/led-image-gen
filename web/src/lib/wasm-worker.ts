declare global {
  interface Window {
    Go: any;
    generateLEDImage: (imageBytes: Uint8Array, config: any) => Uint8Array | Error;
  }

  interface WorkerGlobalScope {
    Go: any;
  }
}

// Worker内で `self.generateLEDImage` を叩く時のエラー回避用
declare var Go: any;
declare var generateLEDImage: (imageBytes: Uint8Array, config: any) => Uint8Array | Error;

importScripts('/wasm_exec.js');

{
  const go = new Go();
  let isWasmReady = false;

  WebAssembly.instantiateStreaming(fetch("/led-gen.wasm"), go.importObject)
    .then((result) => {
      go.run(result.instance);
      isWasmReady = true;
      self.postMessage({type: "READY"});
    })
    .catch((error) => {
      console.error("🚨 [WASM ロード失敗]:", error);
      self.postMessage({ type: "INIT_ERROR", error: error.message });
    });

  self.onmessage = async (e) => {
    const {type, payload} = e.data;
    if (type === "PROCESS_IMAGE") {
      try {
        if (!isWasmReady) {
          self.postMessage({type: "ERROR", error: "WASM is not ready yet", requestId: payload?.requestId});
          return;
        } else {
          const {imageBytes, config, requestId} = payload;
          const result = generateLEDImage(imageBytes, config);
          if (result instanceof Error) {
            throw result;
          } else {
            self.postMessage({type: "SUCCESS", result, requestId}, [result.buffer]);
          }
        }
      } catch (error) {
        self.postMessage({type: "ERROR", error: error instanceof Error ? error.message : String(error), requestId: payload?.requestId});
      }
    }
  }
}
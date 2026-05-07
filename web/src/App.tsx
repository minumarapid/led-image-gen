import {createEffect, createSignal, onCleanup, onMount, Show, untrack} from 'solid-js'
import { cancelLEDImage, defaultConfig, generateLEDImage, initWasmWorker} from "./lib/wasm-bridge.ts";
import {AppSymbol} from "./components/loading.tsx";
import {Dropzone, file, setFile} from "./components/dropzone.tsx";

function App() {
  const [wasmReady, setWasmReady] = createSignal(false);
  const [isProcessing, setIsProcessing] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [outputUrl, setOutputUrl] = createSignal<string | null>(null);

  const clearSelection = () => {
    cancelLEDImage();
    const currentUrl = outputUrl();
    if (currentUrl) URL.revokeObjectURL(currentUrl);
    setOutputUrl(null);
    setError(null);
    setIsProcessing(false);
    setFile(null);
  }

  onMount(() => {
    initWasmWorker(() => {
      console.log("WASM Worker is ready");
      setWasmReady(true);
    },
    (error) => {
      console.error("Failed to initialize WASM Worker:", error);
      setError(error);
    });
  });

  createEffect(() => {
    const selectedFile = file();
    const ready = wasmReady();
    if (!selectedFile || !ready) return;

    const previousUrl = untrack(() => outputUrl());
    if (previousUrl) URL.revokeObjectURL(previousUrl);
    setOutputUrl(null);
    setIsProcessing(true);
    setError(null);

    selectedFile.arrayBuffer()
      .then((buffer) => generateLEDImage(new Uint8Array(buffer), defaultConfig))
      .then((pngBytes) => {
        const pngCopy = new Uint8Array(pngBytes.byteLength);
        pngCopy.set(pngBytes);
        const blob = new Blob([pngCopy.buffer], { type: "image/png" });
        setOutputUrl(URL.createObjectURL(blob));
      })
      .catch((err) => {
        if (err === "LED image generation was canceled" || (err instanceof Error && err.message === "LED image generation was canceled")) {
          return;
        }
        setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        setIsProcessing(false);
      });
  });

  onCleanup(() => {
    const url = outputUrl();
    if (url) URL.revokeObjectURL(url);
  });

  return (
    <div class={"w-screen h-screen"}>
      <Show when={file()} fallback={
        <Dropzone class={"fixed top-0"} />
      }>
        <div class={"w-full h-full flex items-center justify-center"}>
          <Show when={outputUrl()} fallback={<div class={"text-sm text-gray-400"}>処理結果を生成中...</div>}>
            <div class={"relative"}>
              <img src={outputUrl() ?? ""} alt={"LED"} class={"max-w-full max-h-full"} />
              <button
                type={"button"}
                class={"absolute top-2 right-2 rounded-md bg-black/60 px-3 py-1 text-xs text-white"}
                onClick={clearSelection}
              >
                クリア
              </button>
            </div>
          </Show>
        </div>
      </Show>
      <Show when={!wasmReady() || isProcessing()}>
        <AppSymbol isLoading={true} class={"w-12 fixed bottom-6 right-6"} />
      </Show>
      <Show when={error()}>
        <div class={"fixed bottom-6 left-6 text-sm text-red-400"}>{error()}</div>
      </Show>
    </div>
  )
}

export default App

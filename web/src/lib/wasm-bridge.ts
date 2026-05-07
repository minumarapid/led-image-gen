import type {goColorRgba, ledConfig} from "../type/config.ts";

let worker: Worker | null = null;
let currentRequestId = 0;
let inFlight: {
  id: number;
  resolve: (value: Uint8Array) => void;
  reject: (reason?: string) => void;
} | null = null;

export const initWasmWorker = (onReady: () => void, onError?: (arg0: string) => void) => {
  if (worker) return;

  worker = new Worker(new URL("./wasm-worker.ts", import.meta.url), { type: "classic" });

  worker.onmessage = (e) => {
    if (e.data.type === "READY") {
      onReady();
    } else if (e.data.type === "INIT_ERROR") {
      if (onError) onError(e.data.error);
    } else if (e.data.type === "SUCCESS") {
      if (inFlight && inFlight.id === e.data.requestId) {
        inFlight.resolve(e.data.result as Uint8Array);
        inFlight = null;
      }
    } else if (e.data.type === "ERROR") {
      if (inFlight && inFlight.id === e.data.requestId) {
        inFlight.reject(e.data.error);
        inFlight = null;
      }
    }
  };
};

export const cancelLEDImage = () => {
  if (inFlight) {
    inFlight.reject("LED image generation was canceled");
    inFlight = null;
  }
  currentRequestId += 1;
};

export const defaultConfig: ledConfig = {
  border: 10,
  ledSize: 4,
  ledGap: 2,
  ledGamma: 1.0,
  ledExposure: 1.0,
  ledShape: "circle",
  maxWorkers: 1,
  enableGlow: true,
  glowRange: 3,
  glowStrength: 1.75,
  glowGamma: 1.0,
  glowExposure: 1.0,
  offLightColor: "#323232",
}

export const generateLEDImage = async (imageBytes: Uint8Array, config: ledConfig) => {
  if (!worker) throw new Error("WASM Worker is not initialized");
  if (inFlight) throw new Error("LED image generation is already running");

  const wasmConfig = {
    ...config,
    offLightColor: hexToRgba(config.offLightColor),
  };

  const requestId = currentRequestId + 1;
  currentRequestId = requestId;

  const imageBuffer = imageBytes.buffer.slice(
    imageBytes.byteOffset,
    imageBytes.byteOffset + imageBytes.byteLength
  );
  const imageCopy = new Uint8Array(imageBuffer);

  return new Promise<Uint8Array>((resolve, reject) => {
    inFlight = { id: requestId, resolve, reject };
    worker?.postMessage(
      {
        type: "PROCESS_IMAGE",
        payload: { requestId, imageBytes: imageCopy, config: wasmConfig },
      },
      [imageBuffer]
    );
  });
}

const hexToRgba = (hex: string): goColorRgba  => {
  const cleanHex = hex.startsWith("#") ? hex.slice(1) : hex;
  const r = parseInt(cleanHex.slice(0, 2), 16) || 0;
  const g = parseInt(cleanHex.slice(2, 4), 16) || 0;
  const b = parseInt(cleanHex.slice(4, 6), 16) || 0;
  return { R: r, G: g, B: b, A: 255 };
}
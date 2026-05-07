type ledShape = "circle" | "square";

export type goColorRgba = {
  R: number;
  G: number;
  B: number;
  A: number;
}

export interface ledConfig {
  border: number;
  ledSize: number;
  ledGap: number;
  ledGamma: number;
  ledExposure: number;
  ledShape: ledShape;
  maxWorkers: number;
  enableGlow: boolean;
  glowRange: number;
  glowStrength: number;
  glowGamma: number;
  glowExposure: number;
  offLightColor: string;
}

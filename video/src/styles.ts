// ─── Color Palette — Black & Red (Godshell Official) ─────────────
export const COLORS = {
  bg: "#080808",
  bgDeep: "#060606",
  bgSubtle: "#0A0A0A",
  surface: "#0F0000",
  surfaceHover: "#1A0000",
  border: "#2A0000",
  borderSubtle: "#1A0000",
  text: "#CC4444",
  textSecondary: "#884444",
  textDim: "#440000",
  accent: "#FF2222",
  accentMuted: "#CC0000",
  accentSubtle: "#3A0000",
  accentGlow: "rgba(255,34,34,0.15)",
  red: "#FF2222",
  green: "#00FF88",
  orange: "#FF6622",
  yellow: "#FFAA00",
  cyan: "#FF4444",
} as const;

// ─── Typography ──────────────────────────────────────────────────
import { interFamily, jetbrainsFamily } from "./fonts";

export const FONT = {
  mono: jetbrainsFamily,
  sans: interFamily,
} as const;

// ─── Spacing ─────────────────────────────────────────────────────
export const SPACING = {
  letterbox: 60,
  pagePad: 120,
} as const;

// ─── Video Config ────────────────────────────────────────────────
export const VIDEO = {
  width: 1920,
  height: 1080,
  fps: 30,
} as const;

// ─── Scene Durations (in seconds) ───────────────────────────────
export const DURATIONS = {
  coldOpen: 4,
  problem: 8,
  brand: 8,
  demo1: 16,
  demo2: 22,
  demo3: 14,
  moat: 10,
  cta: 8,
} as const;

// ─── Helpers ─────────────────────────────────────────────────────
export const sec = (s: number, fps: number) => Math.round(s * fps);

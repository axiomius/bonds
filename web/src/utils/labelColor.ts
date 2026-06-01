import type { CSSProperties } from "react";

const MIN_READABLE_CONTRAST = 4.5;

const NAMED_TAG_BACKGROUNDS: Record<string, string> = {
  default: "#fafafa",
  blue: "#e6f4ff",
  cyan: "#e6fffb",
  error: "#fff1f0",
  geekblue: "#f0f5ff",
  gold: "#fffbe6",
  green: "#f6ffed",
  lime: "#fcffe6",
  magenta: "#fff0f6",
  orange: "#fff7e6",
  processing: "#e6f4ff",
  purple: "#f9f0ff",
  red: "#fff1f0",
  success: "#f6ffed",
  volcano: "#fff2e8",
  warning: "#fff7e6",
  yellow: "#feffe6",
};

const TAILWIND_LABEL_COLORS: Record<string, string> = {
  "bg-blue-200": "#bfdbfe",
  "bg-green-200": "#bbf7d0",
  "bg-orange-200": "#fed7aa",
  "bg-purple-200": "#e9d5ff",
  "bg-red-200": "#fecaca",
  "bg-yellow-200": "#fef08a",
  "bg-zinc-200": "#e4e4e7",
  "text-blue-700": "#1d4ed8",
  "text-green-700": "#15803d",
  "text-orange-700": "#c2410c",
  "text-purple-700": "#7e22ce",
  "text-red-700": "#b91c1c",
  "text-yellow-700": "#a16207",
  "text-zinc-700": "#3f3f46",
};

interface RgbColor {
  r: number;
  g: number;
  b: number;
}

export interface ReadableLabelTagColors {
  color?: string;
  style: CSSProperties;
}

export function getReadableLabelTagColors(bgColor?: string, textColor?: string): ReadableLabelTagColors {
  const backgroundColor = resolveColor(bgColor);
  if (!backgroundColor) {
    return { color: "default", style: {} };
  }

  return {
    style: {
      backgroundColor,
      borderColor: getReadableBorderColor(backgroundColor),
      color: getReadableTextColor(backgroundColor, resolveColor(textColor)),
    },
  };
}

function resolveColor(color?: string): string | null {
  if (!color) return null;

  const value = color.trim().toLowerCase();
  if (value === "") return null;

  const hexColor = normalizeHexColor(value);
  if (hexColor) return hexColor;

  return NAMED_TAG_BACKGROUNDS[value] ?? TAILWIND_LABEL_COLORS[value] ?? null;
}

function normalizeHexColor(value: string): string | null {
  if (/^#[0-9a-f]{3}$/.test(value)) {
    const r = value.charAt(1);
    const g = value.charAt(2);
    const b = value.charAt(3);
    return `#${r}${r}${g}${g}${b}${b}`;
  }

  if (/^#[0-9a-f]{6}$/.test(value)) {
    return value;
  }

  return null;
}

function getReadableTextColor(backgroundColor: string, requestedTextColor: string | null): string {
  if (requestedTextColor && getContrastRatio(backgroundColor, requestedTextColor) >= MIN_READABLE_CONTRAST) {
    return requestedTextColor;
  }

  const blackContrast = getContrastRatio(backgroundColor, "#000000");
  const whiteContrast = getContrastRatio(backgroundColor, "#ffffff");
  return blackContrast >= whiteContrast ? "#000000" : "#ffffff";
}

function getReadableBorderColor(backgroundColor: string): string {
  return getRelativeLuminance(hexToRgb(backgroundColor)) > 0.85 ? "rgba(0, 0, 0, 0.18)" : backgroundColor;
}

function getContrastRatio(colorA: string, colorB: string): number {
  const luminanceA = getRelativeLuminance(hexToRgb(colorA));
  const luminanceB = getRelativeLuminance(hexToRgb(colorB));
  const lighter = Math.max(luminanceA, luminanceB);
  const darker = Math.min(luminanceA, luminanceB);
  return (lighter + 0.05) / (darker + 0.05);
}

function getRelativeLuminance(color: RgbColor): number {
  const channels = [color.r, color.g, color.b].map((channel) => {
    const normalized = channel / 255;
    return normalized <= 0.03928 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
  });

  return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
}

function hexToRgb(hexColor: string): RgbColor {
  return {
    r: Number.parseInt(hexColor.slice(1, 3), 16),
    g: Number.parseInt(hexColor.slice(3, 5), 16),
    b: Number.parseInt(hexColor.slice(5, 7), 16),
  };
}

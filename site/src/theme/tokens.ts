/**
 * The single source of truth for planton.dev's color and radius tokens.
 *
 * planton.dev is a dark-only marketing surface on Planton's monochrome design
 * language (the planton.ai marketing shade `#0a0a0a`, not the console `#0d1117`).
 * Color is reserved for one signal — a status green — and provider brand logos.
 *
 * These hex values are consumed by BOTH styling systems, with no duplication of
 * intent:
 *   - MUI reads them directly (createTheme needs concrete values; it cannot
 *     derive light/dark/contrast from `var(--x)` strings).
 *   - Tailwind reads them via the CSS custom properties declared in globals.css,
 *     which mirror the exact values below. globals.css is the second emitter of
 *     this one source — keep the two in sync (they are small and rarely change).
 */

/** Neutral ramp, dark → light. The whole UI is built from these plus `status`. */
export const mono = {
  /** Page canvas. */
  950: "#0a0a0a",
  /** Elevated surface (cards, terminal/app chrome bars). */
  900: "#111111",
  /** Popovers, raised panels. */
  850: "#141414",
  /** Subtle surface (tabs, inputs, hover). */
  800: "#1a1a1a",
  /** Borders / dividers. */
  700: "#242424",
  /** Strong border, muted iconography. */
  600: "#333333",
  /** Decorative faint text (large tracked labels only — meets AA at large sizes). */
  500: "#6f6f6f",
  /** Secondary text (meets WCAG-AA on the canvas). */
  400: "#a1a1a1",
  /** Bright secondary / hover text. */
  300: "#c7c7c7",
  /** Primary text and near-white CTA surface. */
  100: "#ededed",
  /** Pure white — used sparingly for peak emphasis. */
  0: "#ffffff",
} as const;

/** The one semantic accent: deployment/health status. Desaturated for the palette. */
export const status = {
  success: "#3fb950",
} as const;

/**
 * Semantic aliases mapped onto the ramp. Components reference semantics, never
 * raw ramp stops, so the palette can shift in one place. These names mirror the
 * shadcn token contract the `ui/*` primitives already depend on.
 */
export const semantic = {
  background: mono[950],
  foreground: mono[100],

  card: mono[900],
  cardForeground: mono[100],

  popover: mono[850],
  popoverForeground: mono[100],

  /** Near-white primary CTA with inverted (dark) text. */
  primary: mono[100],
  primaryForeground: mono[950],

  secondary: mono[800],
  secondaryForeground: mono[100],

  muted: mono[800],
  mutedForeground: mono[400],

  accent: mono[800],
  accentForeground: mono[100],

  destructive: "#cf4a3c",
  destructiveForeground: mono[100],

  border: mono[700],
  input: mono[700],
  /** Focus ring — light enough to be clearly visible on the dark canvas. */
  ring: mono[400],
} as const;

/** Corner radius scale (rem). */
export const radius = {
  sm: "0.375rem",
  md: "0.5rem",
  lg: "0.75rem",
  xl: "1rem",
} as const;

export const tokens = { mono, status, semantic, radius } as const;

export type MonoStop = keyof typeof mono;
export default tokens;

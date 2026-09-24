/**
 * MUI-shaped views of the palette in `tokens.ts`. Nothing here introduces a
 * color; it arranges the tokens into the keys MUI's palette wants.
 */
import { tokens } from './tokens';

/**
 * The grey ramp in MUI's 50-900 convention (lower key = lighter). The
 * console also consumes this theme and reaches for a few non-standard keys
 * its shared components were written against; those map onto the nearest
 * surface so they resolve to a visible color here.
 */
export const websiteGrey = {
  50: '#f5f5f5',
  100: tokens.text.primary,
  200: '#d4d4d4',
  300: tokens.text.secondary,
  400: '#888888',
  500: tokens.text.muted,
  600: tokens.text.faint,
  700: tokens.edge.hover,
  800: tokens.edge.default,
  900: tokens.surface.raised,

  // Console-compatible keys (the console's 0-100 ramp, steps of 10).
  20: tokens.surface.panel,
  70: tokens.surface.raised,
  80: tokens.surface.raised,
} as const;

export const websiteColors = {
  background: {
    default: tokens.surface.canvas,
    paper: tokens.surface.panel,
    tertiary: tokens.surface.raised,
  },
  text: {
    primary: tokens.text.secondary,
    secondary: tokens.text.secondary,
    disabled: tokens.text.muted,
  },
  border: {
    default: tokens.edge.default,
    hover: tokens.edge.hover,
  },
  cta: {
    background: tokens.cta.background,
    text: tokens.cta.text,
  },
} as const;

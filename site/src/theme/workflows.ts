import { tokens } from '../../packages/website-shell/src/theme/tokens';

/** Focused dark illustrations on the light landing page. Blue encodes resource
 * outputs or delivery handoffs; it is not a health or compliance verdict. */
export const workflowDarkTokens = {
  ...tokens,
  surface: { ...tokens.surface, canvas: tokens.surface.card, panel: tokens.surface.card },
  edge: { ...tokens.edge, default: tokens.edge.hover },
  semantic: { ...tokens.semantic, flow: '#659dff' },
} as const;

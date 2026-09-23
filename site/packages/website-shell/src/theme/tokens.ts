/**
 * The planton.ai palette, defined once.
 *
 * Every color the website paints comes from this object. The MUI theme
 * (`websiteTheme.ts`) and the site's Tailwind config both read it, so a hex
 * value appears in exactly one file and a component never types one. The
 * design law behind the values is `public/branding/design-system.md`:
 * monochrome chrome, hierarchy by luminance, color only where it carries
 * meaning.
 *
 * Naming is by role, never by shade: a component asks for the canvas or a
 * card, not for "#151515", so the palette can move without touching the
 * component. Two card surfaces exist on purpose: `card` is the marketing
 * card (borders, not fills, separate it from the canvas), `raised` is the
 * lighter surface for inputs, pills, and tooltips that must read as sitting
 * on top of a card.
 */
export const tokens = {
  surface: {
    /** The page background. */
    canvas: '#0a0a0a',
    /** Header, footer, and panel backgrounds one step up from the canvas. */
    panel: '#111111',
    /** Marketing cards. */
    card: '#151515',
    /** A card under the pointer. */
    cardHover: '#1f1f1f',
    /** Inputs, pills, tooltips, code blocks: the surface that sits on a card. */
    raised: '#1a1a1a',
  },
  text: {
    /** Headings and anything that must be read first. */
    primary: '#ededed',
    /** Body copy inside cards and long-form marketing paragraphs. */
    body: '#b0b0b0',
    /** Supporting copy, subtitles, and content-page headings. */
    secondary: '#a0a0a0',
    /** Captions and labels that recede. */
    muted: '#666666',
    /** Fine print under a call to action; the quietest text on the page. */
    faint: '#555555',
  },
  edge: {
    /** Card and section borders. */
    default: '#2a2a2a',
    /** A border under the pointer. */
    hover: '#3a3a3a',
  },
  semantic: {
    /** Success, "included", a check. Never decorative. */
    ok: '#10b981',
    /** Failure, "not included", an X. Never decorative. */
    danger: '#ef4444',
    /** A warning or a partial. Never decorative. */
    warn: '#f59e0b',
  },
  cta: {
    /** Primary buttons are true white on black; the Tailwind `white` alias is not this. */
    background: '#ffffff',
    text: '#000000',
  },
} as const;

/**
 * The palette flattened into Tailwind color names. The site's
 * `tailwind.config.ts` spreads this into `theme.extend.colors`, so
 * `bg-canvas`, `bg-card`, `text-fg-secondary`, `border-edge`, `text-ok`
 * are the classes components use.
 */
export const tailwindColors = {
  canvas: tokens.surface.canvas,
  panel: tokens.surface.panel,
  card: { DEFAULT: tokens.surface.card, hover: tokens.surface.cardHover },
  raised: tokens.surface.raised,
  fg: {
    DEFAULT: tokens.text.primary,
    body: tokens.text.body,
    secondary: tokens.text.secondary,
    muted: tokens.text.muted,
    faint: tokens.text.faint,
  },
  edge: { DEFAULT: tokens.edge.default, hover: tokens.edge.hover },
  ok: tokens.semantic.ok,
  danger: tokens.semantic.danger,
  warn: tokens.semantic.warn,
  cta: { DEFAULT: tokens.cta.background, text: tokens.cta.text },
} as const;

/** Opt-in route palette. Without CSS variables, every consumer keeps the existing dark pixels. */
export const scopedTokens = Object.fromEntries(Object.entries(tokens).map(([group, roles]) => [group,
  Object.fromEntries(Object.entries(roles).map(([role, value]) => [role, `var(--website-${group}-${role}, ${value})`])),
])) as { [G in keyof typeof tokens]: { [R in keyof typeof tokens[G]]: string } };

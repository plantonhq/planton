
/** Light is opt-in at the route boundary; all existing dark consumers keep their defaults. */
export const homepageLightTokens = {
  surface: { canvas: '#f6f6f3', panel: '#eeeeeb', raised: '#e3e3e0', card: '#eeeeeb', cardHover: '#e3e3e0' },
  text: { primary: '#171717', body: '#454545', secondary: '#595959', muted: '#616161', faint: '#616161' },
  edge: { default: '#c7c7c4', hover: '#999996' },
  semantic: { ok: '#167044', danger: '#b42318', warn: '#875200' },
  cta: { background: '#171717', text: '#ffffff' },
} as const;
const p = homepageLightTokens;
const css: Record<string, string> = {};
for (const [group, roles] of Object.entries(p)) for (const [role, value] of Object.entries(roles)) css[`--website-${group}-${role}`] = value;
const colors = { canvas:p.surface.canvas, panel:p.surface.panel, raised:p.surface.raised, card:p.surface.card, 'card-hover':p.surface.cardHover, fg:p.text.primary, 'fg-body':p.text.body, 'fg-secondary':p.text.secondary, 'fg-muted':p.text.muted, 'fg-faint':p.text.faint, edge:p.edge.default, 'edge-hover':p.edge.hover, cta:p.cta.background, 'cta-text':p.cta.text, ok:p.semantic.ok, danger:p.semantic.danger, warn:p.semantic.warn };
for (const [name, hex] of Object.entries(colors)) css[`--marketing-${name}`] = [1,3,5].map(i=>parseInt(hex.slice(i,i+2),16)).join(' ');
css['--website-header-background'] = p.surface.canvas;
css['--website-header-border'] = p.edge.default;
export const lightVariables = css;

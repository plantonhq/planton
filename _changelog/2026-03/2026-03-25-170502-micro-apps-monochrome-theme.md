# Micro-Apps Monochrome Theme

**Date**: March 25, 2026
**Type**: Design
**Components**: Invest Decks, Tour, Demo, Meets, Blog, Tutorials, Hackathon, Legal, Common Components, ACME (deleted)

## Summary

Extended the black-and-white theme redesign to every micro-app on planton.ai — invest, tour, demo, meets, blog, tutorials, hackathon, legal, and common components. Deleted the ACME fictional customer section entirely. 254 files changed across 9 phases, eliminating all pink/violet/indigo/cyan/amber/blue/orange gradients and accent colors from the marketing site and micro-apps.

## Problem Statement / Motivation

The landing page and main marketing site were redesigned to a monochrome black-and-white theme (changelog: `2026-03-25-144804`), but all micro-apps still used the old colorful aesthetic — heavy pink-to-violet gradients, colored roadmap borders, tinted slide backgrounds, blue badges, and purple CTA buttons. This created a jarring visual disconnect when navigating between the main site and any micro-app.

### Pain Points

- Invest decks used a pink/violet/cyan gradient system with `#7c3aed` → `#0ea5e9` accent colors
- Tour had full-screen `from-purple-900 via-purple-600 to-pink-500` backgrounds
- Demo components used `#110D1F` dark-purple backgrounds and violet gradient buttons
- Meets decks mirrored the invest color system with pink/violet/emerald/blue accents
- Blog and tutorials used blue badges, links, and active states
- Hackathon used a pervasive purple/pink/rose/indigo gradient system
- Legal investor-updates had emerald/pink/cyan badge colors
- ACME section (fictional customer demo) was no longer needed
- Common components (sidebar, Mermaid diagrams, menus) had blue decorative accents

## Solution / What's New

### Phase 0: ACME Deletion

Removed the entire ACME fictional customer section — `src/app/(acme)/` (layout, CSS, 8 page routes) and `src/components/acme/` (9 component files). Will be replaced with a more robust customer showcase in the future.

### Phase 1: Foundation — CSS and Shared Tokens

Rewrote the central style hubs that downstream components consume:

- **invest.css**: `.highlight-gradient` from pink-to-violet gradient text to solid `#ededed`; roadmap borders to brightness-based hierarchy (white/gray/dark); slide backgrounds to neutral `#0a0a0a`; progress bar fill to solid white
- **meets.css**: All 315 lines neutralized — gradient text, flow arrows, quote boxes, stat values, demo badges, presenter notes, slide backgrounds. Comparison before/after borders preserved as semantic (red/green)
- **demo.css**: `.btn-gradient` from violet-to-purple gradient to solid white
- **invest/v2/shared.tsx**: Complete design token rewrite — `Card`, `Badge`, `Metric`, `RoadmapItem`, `Callout`, `TeamMember`, `CustomerCard`, `FundsItem` components all use monochrome palette
- **invest/explainer/shared.tsx**: Complete design token rewrite — `SectionTitle`, `GradientText`, `Section`, `Card`, `Callout`, `Badge`, `Metric`, `Step`, `List`, `Table` components neutralized

### Phase 2: Tour (2 files)

TourPage and CalloutCard: purple/pink backgrounds → `#0a0a0a`; gradient buttons → `bg-[#111]`; colored icon wells → neutral; progress bar → white fill; callout markers from purple borders/ping to neutral gray. Fixed typo `from purple-600` (missing hyphen).

### Phase 3: Blog + Tutorials + Legal (8 files)

Blog: blue category pills → neutral gray; blue "Read more" links → white. Tutorials: blue sort chips, sidebar active states, badges, CTAs → white/gray. Tutorial list row orange links neutralized. Legal investor-updates: emerald/pink/cyan badges and gradient backgrounds → neutral dark palette.

### Phase 4: Hackathon (7 files)

All 7 section files stripped of purple/pink/rose/indigo/blue/cyan gradient system. Replaced with brightness-based white/gray equivalents.

### Phase 5: Invest (54 files)

After Phase 1 rewrote the shared tokens, remaining inline Tailwind classes were neutralized: v2 deck (InvestorDeckV2 + 14 slides), v1 deck (InvestorDeckPage + 10 slides with indigo `#1e1b4b` gradient), explainer pages (IfYouAre, AndYouGet, calculator, controls, layouts), standalone pages (landing, opportunity, process, carta-walkthrough). SVG fills (`color="#ec4899"`) replaced with `#ededed`. YC orange (`#f26625`) replaced with `#a0a0a0`.

### Phase 6: Meets (28 files)

MeetsDeck, primitives, navigation, presenter-notes, and all 21 SEP slide files stripped of pink/violet/emerald/blue/amber/cyan classes and tinted hex backgrounds.

### Phase 7: Demo (58 files)

Fully monochrome including interactive console simulation panels: `#110D1F` dark-purple backgrounds → `#0a0a0a`; InfraJobLogger dark-blue panels (`#242F5E`, `#242C4B`) → neutral darks; SVG strokes in LegoCatalog and InfraVisualization from blue/violet/cyan to white/gray; all form, deployment, infrastructure, and log viewer components neutralized.

### Phase 8: Common Components (3 files)

Selective monochrome preserving semantic colors: content-sidebar blue active states → white/gray; MermaidDiagram blue primary color → `#ededed` (red error styling preserved); ActionsMenu blue hover → white/10. Green success states in CodeBlock, CopyButton, MarkdownViewDialog preserved.

## Implementation Details

### Monochrome Palette (consistent across all micro-apps)

| Token | Value | Usage |
|---|---|---|
| `bgPrimary` | `#0a0a0a` | Page backgrounds, slide backgrounds |
| `bgSecondary` | `#111111` | Secondary panels |
| `bgTertiary` | `#1a1a1a` | Cards, tertiary surfaces |
| `textPrimary` | `#ededed` | Headings, primary text (via Tailwind `white` override) |
| `textSecondary` | `#a0a0a0` | Descriptions, secondary text |
| `textMuted` | `#666666` | Muted labels |
| `border` | `#2a2a2a` | Default borders |
| `borderHover` | `#3a3a3a` | Hover borders |
| CTA Buttons | `bg-[#fff]` | Pure white for action elements (landing), `bg-[#111]` for dark-on-white (tour modals) |
| Semantic green | `#10b981` | Success indicators, check icons |
| Semantic red | `#ef4444` | Error indicators, X icons |
| Semantic amber | `#f59e0b` | Warning indicators |

### What Was Preserved

- Semantic green/red/amber for functional status indicators (comparison tables, check/X icons, warnings)
- hljs code highlighting colors (industry standard)
- Mermaid diagram red error styling
- CodeBlock/CopyButton green success states
- Provider/brand logos (original colors with opacity treatment)

### What Was NOT Touched

- Markdown documentation files (`.md` with color examples in prose)
- Cursor rules (`.mdc` files mentioning gradient examples) — should be updated separately
- `globals.css` hljs syntax highlighting

## Benefits

- **Complete brand consistency**: Every micro-app now matches the main site and Planton console
- **Zero visual jarring**: Navigating between landing page, invest decks, demo, tour, blog feels seamless
- **Simplified design system**: Single monochrome palette serves all micro-apps
- **Reduced cognitive load**: No more color-coding that required learning (which purple means what?)
- **Brightness-based hierarchy**: Visual priority communicated through luminance, not hue

## Impact

- **254 files** changed (including ACME deletion)
- **9 phases** executed with build verification after each
- **Zero build errors** throughout
- All micro-apps (invest, tour, demo, meets, blog, tutorials, hackathon, legal) affected
- ACME section completely removed

## Related Work

- Extends the monochrome theme redesign (changelog: `2026-03-25-144804`)
- Extends the white softening (changelog: `2026-03-25-151117`)
- Follows the "Planton" to "Planton" rebrand (changelog: `2026-03-25-095619`)

---

**Status**: Live
**Timeline**: Single session, 9 phases

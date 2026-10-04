# Monochrome Code Block and Chrome Alignment

**Date**: March 31, 2026
**Type**: Design
**Components**: Design System, Documentation, UI Components

## Summary

Aligned all code blocks, tables, Mermaid diagrams, callouts, and search UI across the docs/blog/branding surfaces to the established monochrome palette. Replaced Tailwind `gray-*` defaults with explicit palette tokens, introduced a muted-desaturated syntax highlighting theme, and redesigned the callout (blockquote) component for better visual hierarchy.

## Problem Statement / Motivation

The monochrome design system established a precise palette (`#0a0a0a`, `#111`, `#1a1a1a`, `#ededed`, `#a0a0a0`, `#666`, `#2a2a2a`) but code blocks and supporting elements still used two kinds of off-palette values:

### Pain Points

- **Tailwind `gray-*` defaults everywhere.** `gray-700` (`#374151`), `gray-800` (`#1f2937`), `gray-900` (`#111827`) scattered across CodeBlock, tables, blockquotes, Mermaid, and search modal. Close in spirit but not the documented ramp.
- **Full-spectrum VS Code Dark+ syntax colors** in hljs — bright blues, greens, oranges, reds clashing with the "hierarchy by luminance, not hue" principle.
- **Dual-layer code blocks.** `INLINE_CODE_CLASSES` was applied to `<code>` elements inside `<pre>`, creating a visible inner rectangle (`#2a2a2a`) inside the code block container (`#1a1a1a`).
- **Callouts blended with code blocks.** Blockquotes used `bg-[#1a1a1a]` (same as code blocks), italic text (hurts readability with inline code), and a subtle `border-white/30` border.
- **Mermaid diagrams** used Tailwind gray-700/800 family instead of the palette, plus light-theme fills for ER and requirement diagrams.
- **Search modal** used `rgb(17, 24, 39)` instead of `#111111`.

## Solution / What's New

### Centralized Token System

Extended `src/theme/docs.ts` from 10 tokens to 23 tokens. Every docs component now imports class strings from this single source of truth. New tokens cover:

- `CODE_BLOCK_CLASSES`, `CODE_BLOCK_COPY_CLASSES`, `CODE_BLOCK_COPY_ACTIVE_CLASSES`
- `TABLE_CLASSES`, `TABLE_HEAD_CLASSES`, `TABLE_ROW_CLASSES`, `TABLE_HEADER_CLASSES`, `TABLE_CELL_CLASSES`, `TABLE_WRAPPER_CLASSES`
- `MERMAID_CONTAINER_CLASSES`
- `HR_CLASSES`, `PARAGRAPH_CLASSES`, `LIST_CLASSES`
- `BLOCKQUOTE_WARNING_CLASSES`

### Muted Syntax Highlighting

Replaced the VS Code Dark+ hljs theme with a muted monochrome palette at ~25% saturation:

| Token | Before | After |
|-------|--------|-------|
| Keywords | `#569cd6` (bright blue) | `#8b9bb5` (cool gray-blue) |
| Functions | `#dcdcaa` (bright yellow) | `#c5c0a0` (warm muted) |
| Comments | `#6a9955` (bright green) | `#666666` (neutral, matches textMuted) |
| Strings | `#ce9178` (bright orange) | `#b5a08f` (warm gray) |
| Types | `#4ec9b0` (bright teal) | `#85ada3` (muted teal) |
| Variables | `#9cdcfe` (bright cyan) | `#9cb5c5` (cool gray) |

Each hue family is preserved so tokens remain distinguishable, but the overall impression is tinted grays. Matches how Cursor, Vercel, and Linear handle code in monochrome contexts.

### Callout Redesign

Blockquotes now render as structured callouts:

| Property | Before | After |
|----------|--------|-------|
| Background | `#1a1a1a` (same as code blocks) | `#111111` (bgSecondary, distinct surface) |
| Left border | `border-l-4 border-white/30` (thick, dim) | `border-l-2 border-[#3a3a3a]` (thin, bright) |
| Text style | Italic | Regular (italic hurts code readability) |
| Type awareness | None | Detects **Warning:**/**Caution:** for semantic red border |

### Mermaid Palette Alignment

All Mermaid theme variables now use the documented palette:

- Background: `#1f2937` → `#1a1a1a`
- Node fills: `#374151` → `#2a2a2a`
- Borders: `#4b5563` → `#3a3a3a`
- Lines: `#9ca3af` → `#666666`
- Text: `#f3f4f6` → `#ededed`
- Font: `system-ui` → `Inter`
- ER `fill: 'honeydew'` → `#2a2a2a` (dark)
- Requirement `rect_fill: '#f9f9f9'` → `#2a2a2a` (dark)

Git/pie chart rainbow colors retained — functional for branch/segment differentiation.

### Search Modal Alignment

- Dialog background: `rgb(17, 24, 39)` → `#111111`
- All `gray-*` classes → palette hex values
- Footer: `border-gray-700 bg-gray-900/50` → `border-[#2a2a2a] bg-[#0a0a0a]/50`

## Implementation Details

### Files Changed (planton.ai)

| File | Changes |
|------|---------|
| `src/theme/docs.ts` | 13 new tokens, 4 updated tokens |
| `src/app/globals.css` | Muted hljs theme, `pre code` reset, removed ghost copy indicator, fixed excerpt gradient |
| `src/components/common/CodeBlock.tsx` | Imports centralized tokens |
| `src/lib/MDXRenderer.tsx` | All renderers use tokens; smart callout type detection |
| `src/components/common/MermaidDiagram.tsx` | Full palette alignment, Inter font, dark ER/requirement |
| `src/app/(root)/docs/components/SearchModal.tsx` | Palette-aligned colors |
| `src/app/(root)/docs/components/SearchTrigger.tsx` | Palette-aligned colors |
| `public/branding/design-system.md` | New syntax highlighting section, callout section, updated token table |

### Files Changed (planton monorepo)

| File | Changes |
|------|---------|
| `_team/agents/_knowledge-base/context/product/ux/planton-console-design-system.md` | New "Syntax Highlighting Alignment" follow-up section with flagged issues |

### Key Technical Decision: CSS `pre code` Reset

The `code` renderer in MDXRenderer applies `INLINE_CODE_CLASSES` to all `<code>` elements. For fenced code blocks (with or without a language specifier), the `<code>` sits inside `CodeBlock`'s `<pre>`, creating a dual-layer background. Rather than complex React-level detection (fragile with react-markdown v9), a CSS `pre code` override with `!important` resets background, padding, and border-radius. This is the standard pattern used by Tailwind Typography.

## Benefits

- **Single source of truth.** All 23 docs tokens in one file — future palette changes are a single-file edit.
- **Zero off-palette values.** Every `gray-*` default in the docs rendering path replaced with explicit hex from the documented ramp.
- **Readable code blocks.** Muted syntax colors maintain token differentiation without breaking the monochrome aesthetic.
- **Clear visual hierarchy.** Three distinct surface levels (canvas `#0a0a0a`, callouts `#111`, code blocks `#1a1a1a`) are visually separable.
- **Callouts work with inline code.** Removing italic eliminates the visual clash between italic prose and upright monospace.

## Impact

- **Docs, blog, tutorials, branding, and changelog** pages all benefit — they share `MDXRenderer`.
- **Console is unaffected** — it uses a separate rendering pipeline (`react-syntax-highlighter`). Follow-up effort is documented in the console design system.

## Related Work

- [Black and White Theme Redesign](_changelog/2026-03/2026-03-25-144804-black-and-white-theme-redesign.md) — established the monochrome palette
- [Docs Monochrome and Pricing Cleanup](_changelog/2026-03/2026-03-26-070350-docs-monochrome-and-pricing-cleanup.md) — created `src/theme/docs.ts`
- [Branding Design System Page](_changelog/2026-03/2026-03-29-164308-branding-design-system-page.md) — created `/branding/design-system`

---

**Status**: Live
**Timeline**: Single session

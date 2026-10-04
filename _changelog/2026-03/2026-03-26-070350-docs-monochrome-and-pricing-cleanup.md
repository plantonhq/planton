# Docs Monochrome Centralization & Pricing Page Cleanup

**Date**: March 26, 2026
**Type**: Design
**Components**: Documentation, Pricing Page, Design System

## Summary

Centralized the docs design system into a single theme file (`src/theme/docs.ts`), stripped all sidebar icons, converted every remaining blue/green/colored element to monochrome, and cleaned up the pricing page -- fixing the Most Popular card layout, neutralizing gold/purple/blue gradients, and aligning the CTA section with the standard site-wide pattern.

## Problem Statement / Motivation

After the four monochrome passes (changelogs `144804`, `151117`, `170502`, `193704`), the docs section still operated with its own independent design system. Blue links, blue tag pills, blue blockquote borders, green inline code, a blue-bordered search modal, and an extensive emoji icon system in the sidebar all remained. Colors were hardcoded as raw Tailwind classes across 6+ files with no central source of truth.

The pricing page had separate issues: the Most Popular banner was rendered inside the card, pushing its content down and misaligning all four cards. Gold gradients, blue sliders, purple FAQ backgrounds, and colorful decorative SVGs clashed with the monochrome theme.

### Pain Points

- Docs links used `text-blue-400` while the rest of the site used monochrome
- ~170 emoji icon mappings in `fileSystem.ts` cluttered the sidebar
- Tag pills, blockquotes, and inline code all used blue/green colors
- Search modal had a blue border accent
- Pricing cards were vertically misaligned due to the Most Popular banner
- Calculator slider was bright blue (`#0099FF`)
- FAQ section had purple gradient blobs and decorative SVGs
- Pricing CTA was a bespoke one-off component unlike every other page

## Solution / What's New

### Centralized Docs Theme (`src/theme/docs.ts`)

Created a single file exporting all docs-specific style constants as Tailwind class strings. Every docs component imports from here instead of hardcoding colors. Constants include:

- `LINK_CLASSES` -- monochrome links with subtle underline decoration
- `TAG_CLASSES` -- neutral pills replacing blue variants
- `BLOCKQUOTE_CLASSES` -- `border-white/30` replacing `border-blue-500`
- `INLINE_CODE_CLASSES` -- white text replacing green
- `NEXT_ARTICLE_BUTTON_CLASSES` -- white CTA replacing blue
- `SEARCH_DIALOG_BORDER` -- neutral border replacing blue accent
- `SIDEBAR_BADGE_COLORS` / `SIDEBAR_ACTIVE_CLASSES` -- centralized sidebar tokens

### Docs Monochrome Conversions

**MDXRenderer.tsx** (shared by docs, blog, tutorials):
- Links: `text-blue-400` → `text-white/80` with `decoration-white/30` underline
- Tags: `bg-blue-900 text-blue-200` → `bg-white/10 text-white/70`
- Blockquotes: `border-blue-500` → `border-white/30`
- Inline code: `text-green-400` → `text-white`
- "Read next" button: `bg-blue-600` → `bg-[#fff] text-black`

**SearchModal.tsx**: Blue border `rgba(59, 130, 246, 0.25)` → neutral `rgba(255, 255, 255, 0.1)`

**DocsLayout.tsx**: Border colors aligned to monochrome palette (`border-[#2a2a2a]`)

### Sidebar Icon Removal

Removed the entire icon system from the docs sidebar:

- **DocsSidebar.tsx**: Removed `renderIcon()`, `FolderIcon`, `FileIcon` imports. Sidebar is now text-only with indentation for hierarchy.
- **fileSystem.ts**: Removed `iconMap` (~170 emoji entries), `categoryIcons` (~20 entries), `resolveIcon()`, and `getDefaultIcon()` functions. Stopped populating `icon` field on `DocItem`.
- Kept expand/collapse chevrons (navigation controls) and semantic badges (Popular, Deprecated).

### Pricing Card Layout Fix

Restructured `PriceCard` so the Most Popular badge sits as an absolutely positioned pill on the card's top border edge -- half inside, half outside. All four cards now share identical internal structure with `items-stretch` on the grid for equal heights and aligned CTA buttons.

Price and period text changed from `<sub>` to baseline-aligned `<span>`, fixing vertical misalignment of "/month" and "/seat/month" across cards.

### Pricing Page Monochrome

- **Slider**: `#0099FF` (blue) → `#ededed`
- **Card border gradient**: gold → monochrome white-to-gray
- **Tab selected state**: amber → `rgba(255, 255, 255, 0.1)`
- **Show/Hide Calculation**: `text-[#0099FF]` → monochrome underlined link
- **Price card gradient**: gold `#FDA935` → `bg-[#111] border border-[#2a2a2a]`
- **Contact Sales card**: same gold → same neutral treatment
- **Decorative SVGs**: All four 3D dice images removed
- **FAQ section**: Purple gradient blob, `burger-shape.svg`, gradient container → `bg-[#111] border border-[#2a2a2a]`
- **Shared tabs**: Purple gradient selected states → neutral `rgba(255, 255, 255, 0.1)`

### Pricing CTA Standardization

Replaced the bespoke `ReadyToTry` component (left-aligned dark strip with ghost button) with the standard CTA pattern used across all feature pages: centered `Card` with subtitle, heading, description, and `PrimaryButton` + `SecondaryButton`.

## Implementation Details

### Design Token Architecture

The `src/theme/docs.ts` file follows the same pattern as the landing page's `shared.tsx` `colors` object but scoped to docs. It exports Tailwind class strings (not CSS variables) because the codebase already uses Tailwind utility classes everywhere. This gives TypeScript-level visibility into what's being used and makes future palette changes a single-file edit.

### Semantic Colors Preserved

- HeadingWithAnchor check icon (`#10b981` green) -- "link copied" indicator
- Sidebar badges: Popular (green), Deprecated (red) -- functional status
- hljs syntax highlighting -- industry standard
- Mermaid diagram colors -- data visualization

### MDXRenderer Impact

`MDXRenderer.tsx` is shared by docs, blog, and tutorials. The monochrome changes cascade to all three consumers, which is consistent since blog and tutorials were already monochrome-converted in previous passes.

## Benefits

- **Single source of truth**: All docs styling lives in `src/theme/docs.ts`
- **No more scattered colors**: Components import tokens instead of hardcoding
- **Cleaner sidebar**: Text-only navigation with indentation, matching cursor.com's pattern
- **Aligned pricing cards**: All four cards at the same vertical position with consistent internal layout
- **Consistent CTA**: Pricing page CTA now matches every other page on the site
- **~200 lines of dead code removed**: Icon maps, resolve functions, category mappings

## Impact

- **13 files** modified, 1 new file created
- **Zero build errors** -- clean `next build` exit code 0
- Docs, pricing, and shared tabs all affected
- ~200 lines of emoji icon infrastructure removed from `fileSystem.ts`

## Related Work

- Follows the monochrome polish pass (changelog: `2026-03-25-193704`)
- Follows the micro-apps monochrome theme (changelog: `2026-03-25-170502`)
- Follows the original black-and-white redesign (changelog: `2026-03-25-144804`)

---

**Status**: Live
**Timeline**: Single session

# Monochrome Canvas Unification and Palette Alignment

**Date**: April 1, 2026
**Type**: Design
**Components**: Design System, Documentation, Blog System, Tutorials, Pricing Page, Solutions Pages, Navigation, UI Components

## Summary

Unified the entire planton.ai website to the design system's canonical `#0a0a0a` canvas background, eliminated all blue-tinted Tailwind gray classes in favor of the neutral monochrome palette, and elevated navigation dropdowns to a distinct surface level for proper figure-ground separation on dark backgrounds.

## Problem Statement / Motivation

The website had accumulated three categories of visual inconsistency:

1. **Canvas fragmentation**: The `<body>` tag used Tailwind's `bg-black` (`#000000`) while the design system specified `#0a0a0a` as the page canvas. Pages using the `<Section>` component (homepage, features) painted `#0a0a0a` over the body, masking the mismatch. Pages that didn't (docs, blog, changelog, pricing, tutorials) exposed the pure black body, creating a "shinier" appearance compared to the "paper-like" feel of other pages.

2. **Blue-tinted grays**: Multiple components used Tailwind's default `gray-XXX` scale (`gray-300` through `gray-800`), which carries a cool blue tint from Tailwind's slate-adjacent default palette. On the pure-neutral `#0a0a0a` canvas, this blue cast was perceptible and broke the monochrome design language.

3. **Flat overlay hierarchy**: Navigation dropdowns used `#111` (Panel level) with `#2a2a2a` borders -- the same visual treatment as pricing cards and FAQ containers on the page beneath them. On the pricing page, the dropdown was visually indistinguishable from page content.

### Pain Points

- Docs, blog, changelog, and pricing pages appeared "shinier" than the homepage
- Sidebars had a visible darker band from `bg-black/95` (`rgba(0,0,0,0.95)`) instead of sharing the canvas background
- Blog post cards used `bg-gray-800` (`#1f2937`) -- a blue-tinted gray that felt out of place in the monochrome theme
- Navigation dropdown on pricing page blended completely into the page content
- Footer used `#000` instead of the design system's canvas color
- Multiple components mixed Tailwind named grays with the neutral palette hex values

## Solution / What's New

### Phase 1: Canvas Unification

Changed the `<body>` background from `bg-black` (`#000`) to `bg-[#0a0a0a]` and swept every page-level component that set its own `bg-black` or `bg-[#010101]` canvas. This covered 20+ files across pricing, solutions, CLI, features, self-service DevOps, and Plantora AI sections.

### Phase 2: Unified Canvas Sidebars

Removed `bg-black/95` from all desktop sidebars (docs, blog, tutorials, branding). Sidebars now inherit the `#0a0a0a` canvas through the flex parent, with only `border-[#2a2a2a]` providing visual separation -- exactly as the design system's "unified canvas" principle dictates. MUI Drawer overlays (mobile) use solid `bg-[#0a0a0a]`.

### Phase 3: Footer Canvas Alignment

Changed both `WebsiteFooter` (`bgcolor: '#000'`) and the legacy `footer.tsx` (`bg-black`) to `#0a0a0a`, matching the design system surface table which places the footer in the Canvas layer.

### Phase 4: BlogPostCard Monochrome Alignment

Replaced the blue-tinted `bg-gray-800` / `border-gray-700` / `shadow-lg` card with `bg-[#1a1a1a]` / `border-[#2a2a2a]` / `hover:border-[#3a3a3a]` -- no shadows, per the "borders not backgrounds" principle. Text colors moved from `text-gray-300`/`text-gray-400` to the neutral palette equivalents.

### Phase 5: Blue-Tinted Gray Purge

Systematically replaced every `gray-300` through `gray-800` Tailwind class across 15+ files with neutral palette equivalents: `border-[#2a2a2a]`, `border-[#3a3a3a]`, `bg-[#111]`, `hover:bg-white/5`, `text-[#a0a0a0]`, `text-[#666]`, `text-white`.

### Phase 6: Sidebar Active State Alignment

Changed sidebar active item styling from `bg-white/20` (too bright) to `bg-white/10` (matching `SIDEBAR_ACTIVE_CLASSES` in `docs.ts`). Fixed `TutorialListRow` to use neutral palette colors and removed incorrect `dark:` prefixes.

### Phase 7: Navigation Dropdown Elevation

Bumped the MegaMenu dropdown panel from Panel level (`#111` + `#2a2a2a` border) to Card/Overlay level (`#1a1a1a` + `#3a3a3a` border). Added a subtle `0 0 0 1px rgba(255,255,255,0.06)` ring shadow for luminance-based elevation on dark backgrounds. This ensures the dropdown is visually distinct from any page content it floats over.

## Implementation Details

### Files Changed

**Canvas fixes** (20+ files):
- `src/app/layout.tsx` -- body `bg-black` to `bg-[#0a0a0a]`
- `src/components/pricing/plans-for-every-stage.tsx`, `ready-to-crunch.tsx`, `faqs.tsx`
- `src/components/pricing-calculator/calculation.tsx`
- `src/components/solutions/` -- 7 files
- `src/components/apps/cli/` -- 7 files
- `src/components/self-service-dev-ops/` -- 3 files
- `src/components/features/all/dev-ops-workflows.tsx`
- `src/components/plantora-ai/explore-and-manage.tsx`
- `src/components/common/content-layout.tsx`

**Sidebar fixes** (8 files):
- `DocsLayout.tsx`, `RightSidebar.tsx`, `content-layout.tsx`, `content-sidebar.tsx`, `content-details-sidebar.tsx`, `TutorialsPageClient.tsx`, `BrandingContentLayout.tsx`

**Footer fixes** (2 files):
- `WebsiteFooter.tsx`, `footer.tsx`

**Neutral palette sweep** (15+ files):
- `content-sidebar.tsx`, `TutorialsPageClient.tsx`, `TutorialsSidebar.tsx`, `TutorialListRow.tsx`, `BlogPostCard.tsx`, `AuthorSection.tsx`, `TableOfContents.tsx`, `ActionsMenu.tsx`, `blog/page.tsx`, `changelog/page.tsx`, `changelog/[slug]/page.tsx`

**Dropdown elevation** (1 file):
- `packages/website-shell/src/components/header/MegaMenu.tsx`

**Design system doc** (1 file):
- `public/branding/design-system.md` -- updated surface table, corrected body class reference

### Key Design Decisions

- **Sidebars share canvas, not a darker shade**: The design system's "unified canvas" principle means sidebar, header, and content area all use `#0a0a0a`. Borders provide the visual separation. This follows the GitHub Dark Default pattern.
- **Navigation dropdowns are overlays, not panels**: Overlays float above all page content and need a higher surface level than inline panels. Dropdowns now use Card level (`#1a1a1a`) with hover-level borders (`#3a3a3a`).
- **Ring shadow for dark-mode elevation**: Traditional drop shadows are invisible on dark backgrounds. A `1px` white-at-6% ring provides a luminance-based elevation cue that works regardless of the underlying surface.
- **No Tailwind named grays**: All `gray-300` through `gray-800` classes were replaced with explicit neutral hex values from the design system palette. Tailwind's default gray scale has a cool blue tint that breaks the monochrome language.

## Benefits

- Every page on planton.ai now shares the same `#0a0a0a` canvas -- no more "shinier" vs "paper-like" differences
- Sidebars blend seamlessly with the page canvas per the unified canvas principle
- Blog post cards, sort menus, and action menus all use the correct neutral palette
- Navigation dropdowns are clearly distinguishable from page content on every page
- The design system document accurately reflects the implemented surface hierarchy
- Zero blue-tinted Tailwind grays remain in the sidebar, blog, tutorials, and changelog components

## Impact

- **Visual consistency**: All 126 generated pages now share a unified dark canvas
- **Design system compliance**: Every surface color traces back to the documented palette
- **Maintainability**: Future contributors have clear guidance on which hex values to use
- **No layout changes**: Only background colors, borders, and text colors were modified -- no spacing, sizing, or interactive behavior changes

## Related Work

- [Black and White Theme Redesign](2026-03/2026-03-25-144804-black-and-white-theme-redesign.md) -- established the monochrome palette
- [Monochrome Polish and Contrast Fixes](2026-03/2026-03-25-193704-monochrome-polish-and-contrast-fixes.md) -- initial contrast pass
- [Monochrome Code Block Alignment](2026-03/2026-03-31-173025-monochrome-code-block-alignment.md) -- aligned code blocks to the three-level surface stack
- [Website Shell Component Extraction](2026-04/2026-04-01-175418-website-shell-component-extraction.md) -- extracted shared header/footer into `@plantonhq/website-shell`

---

**Status**: Live
**Timeline**: Single session

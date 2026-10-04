# Docs Site Mobile Responsiveness

**Date**: February 14, 2026
**Type**: Enhancement
**Components**: Documentation, Responsive Design, UI Components

## Summary

Made the Planton documentation site mobile-responsive by fixing layout, typography, sticky positioning, and overflow issues across 6 component files. Eliminated the double-hamburger problem on mobile by replacing the docs header with a distinct inline navigation trigger. Page-level horizontal scroll partially addressed — deeper investigation needed in follow-up session.

## Problem Statement / Motivation

The documentation site had zero mobile adaptation in the content area. While a sidebar Drawer existed for mobile navigation, every other element was hardcoded for desktop widths.

### Pain Points

- **Content padding `px-12` (48px/side)** consumed most of the mobile viewport, leaving only ~279px for text on a 375px phone
- **Two identical hamburger icons** stacked vertically — one for the main site nav, another for the docs sidebar — caused user confusion
- **Fixed typography sizes** (`text-4xl` titles, `prose-lg` body) caused overflow on narrow screens
- **Broken sticky positioning** — the docs header used `sticky top-0` but the main site header was `fixed top-0 z-20`, causing the docs header to scroll behind the main header
- **No background on docs header** — content bled through the sticky header on scroll
- **Search bar fixed at `w-64`** (256px) consumed most of the mobile header
- **Heading anchor links** at `absolute -left-6` overflowed outside the viewport on mobile
- **No page-level overflow protection** — any wide element caused horizontal scroll

## Solution / What's New

### Layout Container (`DocsLayout.tsx`)

- Responsive content padding: `px-4 sm:px-6 md:px-8 lg:px-12`
- Docs header hidden on mobile (`hidden md:sticky md:block`), replaced with an inline "Documentation menu" button using a `FormatListBulleted` icon — visually distinct from the main site hamburger
- Sticky positioning corrected: docs header at `top: 70px`, sidebars at `top: 127px`
- Header background: `bg-gray-900` prevents content bleed-through
- `overflow-x-hidden` on content area, `min-w-0` on flex child to prevent overflow
- Named constants (`SITE_HEADER_HEIGHT`, `BELOW_BOTH_HEADERS`) for maintainability

### Typography (`MDXRenderer.tsx`)

- Page title: `text-2xl sm:text-3xl md:text-4xl`
- Prose container: `prose md:prose-lg` (base 16px on mobile, 18px on desktop)
- All headings (h1-h6): responsive sizes via `sm:` and `md:` breakpoints
- Metadata row: `flex-wrap` + `gap-y-2` for wrapping on narrow screens
- Tags: `flex-wrap` for natural wrapping
- Inline code: `break-words` to prevent overflow
- Links: `break-words` for long URLs
- Tables: edge-to-edge scroll on mobile with `-mx-4 px-4 sm:mx-0 sm:px-0`

### Heading Anchors (`HeadingWithAnchor.tsx`)

- Anchor link hidden on mobile (`hidden md:inline-block`) — prevents `-left-6` overflow
- Scroll margin updated from `scroll-mt-24` to `scroll-mt-36` (144px) for both headers

### Copy Button (`CopyButton.tsx`)

- Mobile: compact icon-only `IconButton`
- Desktop: unchanged full button with "Copy page" text

### CSS (`globals.css`)

- `overflow-x: clip` on `html` element to prevent page-level horizontal scroll
- `pre { max-width: 100% }` for code blocks
- Smaller code font on mobile (13px below 640px)
- Inline code: `word-break: break-word` + `overflow-wrap: break-word`
- Mermaid diagrams: `max-width: 100%; overflow-x: auto`

## Implementation Details

### Double Hamburger Resolution

The main site header (from `header.tsx`) renders a `DensityMedium` hamburger on mobile. The docs header previously rendered a second `MenuIcon` hamburger. Both looked identical, confusing users.

Solution: On mobile, the docs header is completely hidden. A compact button with a `FormatListBulleted` (list) icon and "Documentation menu" text appears at the top of the content area. This is visually distinct from the hamburger and clearly communicates its purpose.

### Sticky Positioning Fix

The root cause was that the main site header uses `fixed top-0 z-20` with a 70px height (`pt-[70px]` in MainLayout). The docs header used `sticky top-0 z-10`, causing it to scroll behind the main header. The sidebars used `sticky top-16`, also incorrect.

Fix: Defined constants for the header offsets and used `style={{ top: SITE_HEADER_HEIGHT }}` for the docs header and `style={{ top: BELOW_BOTH_HEADERS, height: calc(100vh - ...) }}` for sidebars.

### Overflow Protection Strategy

An early attempt to add `overflow-x-hidden` to the root DocsLayout container broke all `position: sticky` elements (known CSS behavior — overflow on a parent breaks sticky children). The fix was:
1. `overflow-x-hidden` only on the content area flex child
2. `overflow-x: clip` on the `html` element in globals.css (`clip` doesn't create a scroll container, so sticky is preserved)

## Known Limitations

- **Horizontal scroll not fully resolved**: The `overflow-x: clip` on `html` may not catch all overflow sources. The user reports horizontal scroll persists on real mobile devices. A deeper investigation in a follow-up session is needed to identify the specific element(s) causing the overflow.
- **Search bar hidden on mobile**: Search is not yet implemented, so hiding it is appropriate. When search is built, a mobile search experience should be added.
- **Table of Contents not accessible on mobile**: The right sidebar (TOC) is hidden on mobile with no alternative access point. A future enhancement could add a mobile TOC.

## Files Changed

| File | Changes |
|------|---------|
| `src/app/(root)/docs/components/DocsLayout.tsx` | Layout, sticky positioning, mobile nav trigger, overflow protection |
| `src/lib/MDXRenderer.tsx` | Responsive typography, wrapping, table scroll |
| `src/components/docs/HeadingWithAnchor.tsx` | Anchor positioning, scroll margin |
| `src/app/(root)/docs/components/DocsPageActions/CopyButton.tsx` | Mobile compact button |
| `src/app/globals.css` | Overflow clip, code block sizing, inline code wrapping |

## Benefits

- Documentation is now readable on mobile devices (previously unusable)
- Single clear navigation pattern on mobile (no double hamburger confusion)
- Sticky headers and sidebars properly positioned below the main site header
- Content uses full available width on mobile instead of being squeezed by 96px padding
- Typography scales naturally across viewport sizes

## Impact

All 50 documentation pages across 8 sections are affected. Mobile users can now read documentation, navigate sections, and interact with code blocks. Desktop layout is preserved with no regressions.

## Related Work

- Planton Docs Overhaul project (`_projects/20260212.02.planton-docs-overhaul/`)
- Sessions 1-16 created the documentation content; this session makes it mobile-accessible

---

**Status**: 🔵 In Progress (horizontal scroll investigation pending)
**Timeline**: 1 session, follow-up needed

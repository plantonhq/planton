# Docs Search: Pagefind Integration with Trigger + Modal Pattern

**Date**: February 14, 2026
**Type**: Feature
**Components**: Documentation, Build System, UI Components

## Summary

Added full-text search to planton.ai/docs using Pagefind, a static search engine that indexes built HTML at build time and loads a lightweight JS bundle at runtime. The UI follows the trigger-button-plus-centered-modal pattern used by Tailwind CSS, Vercel, and Stripe documentation sites, adapted to planton.ai's blue/gray theme.

## Problem Statement / Motivation

The docs site had a stub search bar in the header (`SearchBar.tsx`, 53 lines) that rendered an MUI TextField with a `TODO: Implement search functionality` comment. With 50 documentation pages across 8 sections, users had no way to find content other than browsing the sidebar.

### Pain Points

- No search functionality at all across 50 documentation pages
- Users had to manually browse the sidebar to find content
- The stub search bar was visually present but non-functional (confusing UX)
- Mobile users had no access to search (the docs header was hidden on mobile)

## Solution / What's New

Ported the search architecture from openmcf.org/docs (which was built and refined in a previous session) and adapted it to planton.ai's theme and layout constraints.

### Architecture

Three focused components replace the single stub:

- **SearchBar.tsx** (~80 lines) — Orchestrator that owns the modal open/close state, the global keyboard shortcut (Cmd+K / Ctrl+K / `/`), and exposes an imperative `onOpenRef` for the mobile trigger
- **SearchTrigger.tsx** (~55 lines) — A native `<button>` styled as a search input field with placeholder text and keyboard shortcut badge, rendered in the desktop docs header
- **SearchModal.tsx** (~300 lines) — MUI Dialog with Pagefind integration, debounced search, keyboard navigation (arrow keys + Enter), grouped results with page section headers, and footer keyboard hints

### Build Pipeline

Pagefind runs automatically after `next build`:

```
next build && pagefind --site out --output-path out/_pagefind --exclude-selectors 'pre,code'
```

The `data-pagefind-body` attribute on the `<article>` element in `MDXRenderer.tsx` scopes indexing to documentation content only — sidebar, header, footer, and non-docs pages are excluded automatically.

### Mobile Access

A "Search" button with icon appears next to the "Documentation menu" button on mobile screens, opening the same modal dialog. The keyboard shortcut (Cmd+K) also works on desktop regardless of which element has focus.

## Implementation Details

### Theme Adaptation (OpenMCF to planton.ai)

Key color mappings from OpenMCF's purple theme to planton.ai's blue/gray theme:

- Modal background: `rgb(17, 24, 39)` (gray-900) instead of `rgb(15, 23, 42)` (slate-900)
- Border accent: `blue-500/25` instead of `purple-500/25`
- Active result: `bg-blue-500/15`, `text-blue-300` instead of purple
- Search highlights (`<mark>` from Pagefind): `bg-blue-500/30 text-blue-300`
- Trigger border: `border-gray-600` with `hover:border-blue-500/50`
- Loading spinner: `border-blue-500/30 border-t-blue-500`

### MUI Compatibility

OpenMCF uses MUI 7; planton.ai uses MUI 6. The Dialog API (`open`, `onClose`, `maxWidth`, `sx` targeting `MuiDialog-container`/`MuiDialog-paper`/`MuiBackdrop-root`) is identical between versions — no adaptation needed.

### Mobile Integration via Ref

The `SearchBar` component accepts an `onOpenRef` prop (a `MutableRefObject`) that DocsLayout uses to open the search modal from the mobile trigger button. This avoids lifting state management into the layout while keeping the keyboard shortcut ownership inside `SearchBar`.

### Pagefind Scoping

Adding `data-pagefind-body` to the `<article>` element means:
- Only pages with MDXRenderer (docs pages) are indexed
- Non-docs pages (landing, pricing, features, blog, etc.) are automatically excluded
- Within docs pages, only the article content is indexed — not the sidebar, header, or ToC

### Build Performance

Pagefind indexed 55 pages (2,217 words) in 0.244 seconds. Negligible impact on the ~43-second total build time.

## Benefits

- **Users can now search** across all 50 documentation pages with sub-second results
- **Keyboard-first workflow**: Cmd+K opens search, arrow keys navigate, Enter selects, Esc closes
- **Zero runtime cost** until first search interaction (Pagefind JS loaded lazily)
- **Mobile accessible**: Search button visible on small screens alongside the docs menu
- **No external service**: Pagefind is fully static — no Algolia, no server, no API keys
- **Graceful dev mode**: Helpful notice explains that search requires a production build

## Impact

- **Users**: Can now find documentation content by keyword instead of manual browsing
- **Developers**: Three focused components (~435 lines total) replace one non-functional stub (53 lines)
- **Build pipeline**: One additional step (`pagefind`) adds ~0.25 seconds to build time
- **Bundle**: Pagefind's runtime JS (~50KB) is loaded only on first search interaction

## Related Work

- OpenMCF docs search modal refactor (2026-02-14) — the reference implementation
- Docs mobile responsiveness (2026-02-14) — established the dual-header layout and mobile patterns
- Docs mobile UX and page actions (2026-02-14) — the mobile navigation bar that search integrates with

---

**Status**: Production Ready
**Build**: Verified — 55 pages indexed, 2,217 words, zero errors

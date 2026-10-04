# Changelog Detail Page: Back Navigation and Page Actions

**Date**: February 15, 2026
**Type**: Enhancement
**Components**: Changelog, UI Components, Design System

## Summary

Added back-navigation and page actions (Copy as Markdown, View as Markdown, Open Raw) to the changelog detail page. As a prerequisite, relocated the `DocsPageActions` component from a docs-scoped path to a shared location (`src/components/common/PageActions/`), fixing an existing architectural smell where `MDXRenderer.tsx` (a shared library) was importing from a page-specific location.

## Problem Statement / Motivation

The changelog detail page (`/changelog/[slug]`) serves as the shareable permalink for individual changelog entries. Users arrive from external sources -- Slack links, the console dashboard widget, email shares. Two gaps existed:

### Pain Points

- **No way back to the list**: Users landing on a detail page had no in-page navigation to return to the full changelog timeline at `/changelog`. They had to use the browser back button or the site header navigation.
- **No copy/raw access**: The docs pages had "Copy as Markdown", "View as Markdown", and "Open Raw" actions, but the changelog detail page didn't -- despite changelog entries being equally useful as shareable markdown content.
- **Architectural smell**: The page actions component (`DocsPageActions`) was scoped inside `src/app/(root)/docs/components/` but was imported by `src/lib/MDXRenderer.tsx`, a shared renderer. Reusing it on the changelog page would compound this cross-cutting dependency.

## Solution / What's New

### Back Navigation Link

A minimal `ArrowLeft` + "Changelog" link at the top of the detail page, above the article. Positioned above the content (not in a footer) because the detail page is a permalink target -- users need the escape hatch immediately.

### Page Actions

The same "Copy page" dropdown from docs pages now appears next to the changelog entry title. Users can copy the raw markdown (with frontmatter), view it in a dialog, or open the raw `.md` file in a new tab.

### PageActions Component Relocation

Moved the 4-file `DocsPageActions` component from `src/app/(root)/docs/components/DocsPageActions/` to `src/components/common/PageActions/` and renamed it to `PageActions`. This is now a properly shared component alongside `MermaidDiagram` and `CodeBlock`.

## Implementation Details

### Files Created

- `src/components/common/PageActions/index.tsx` -- Orchestrator (renamed from `DocsPageActions` to `PageActions`)
- `src/components/common/PageActions/ActionsMenu.tsx` -- MUI dropdown menu
- `src/components/common/PageActions/CopyButton.tsx` -- Trigger button with "Copy page" text
- `src/components/common/PageActions/MarkdownViewDialog.tsx` -- Modal showing raw markdown

### Files Modified

- `src/components/common/index.ts` -- Added `PageActions` to barrel exports
- `src/lib/MDXRenderer.tsx` -- Updated import from docs-scoped path to shared `@/components/common/PageActions`
- `src/app/(root)/changelog/[slug]/page.tsx` -- Added back-nav link, added `PageActions` to title row

### Files Deleted

- `src/app/(root)/docs/components/DocsPageActions/index.tsx`
- `src/app/(root)/docs/components/DocsPageActions/ActionsMenu.tsx`
- `src/app/(root)/docs/components/DocsPageActions/CopyButton.tsx`
- `src/app/(root)/docs/components/DocsPageActions/MarkdownViewDialog.tsx`

### Key Design Decisions

- **Top-placed back link (not footer)**: The investor-updates page uses a footer back link, but for changelog the primary browsing experience is the list page with inline expand. The detail page is secondary -- users need the escape hatch at the top.
- **ArrowLeft icon (not Unicode arrow)**: Consistent with how `lucide-react` icons are used across the codebase (already imported in ChangelogTimeline and InvestorDeckV2).
- **Raw path points to static `.md` file**: Changelog entries live in `public/changelog/`, so "Open Raw" correctly opens `/changelog/slug.md` -- the actual markdown file served by Next.js static export.

## Benefits

- **User experience**: Users can navigate back to the changelog list from any detail page without relying on browser history
- **Content reuse**: Changelog entries can be copied as markdown for Slack posts, documentation, or issue descriptions
- **Architecture**: The `PageActions` component is now properly shared -- any future content type (blog, tutorials) can use it without importing from a page-scoped path
- **Consistency**: The changelog detail page now matches the docs pages in functionality

## Impact

- **Changelog readers**: Can navigate between list and detail views naturally
- **Content consumers**: Can copy/view raw markdown for any changelog entry
- **Future development**: Any markdown content page can reuse `PageActions` via `import { PageActions } from '@/components/common'`

## Related Work

- **T02+T03 (Changelog Pages)**: This enhances the detail page built in the previous session (commit `849e901`)
- **DocsPageActions original implementation**: The component was originally built for the docs system; this change makes it a platform-wide utility

---

**Status**: ✅ Live
**Timeline**: Single session

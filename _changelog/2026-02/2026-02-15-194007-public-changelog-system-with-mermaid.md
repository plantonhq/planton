# Public Changelog System with Mermaid Diagram Rendering

**Date**: February 15, 2026
**Type**: Feature
**Components**: Changelog, Build System, UI Components, Navigation, Content Management

## Summary

Built a complete public changelog system for planton.ai featuring an inline expand/collapse timeline (inspired by the Fastlane pattern), category filtering, text search, Mermaid diagram rendering, a shareable detail page, and a `recent.json` build-time artifact for the console dashboard widget. Also wired the existing `MermaidDiagram` component into the site-wide `MDXRenderer` so docs, blog, and tutorials now render Mermaid diagrams.

## Problem Statement / Motivation

Planton ships improvements continuously, but platform users had no public-facing visibility into what changed. Internal engineering changelogs captured implementation details, but nothing was distilled into a user-facing format. A public changelog was needed to provide transparency, build trust, and let users know when features they care about land.

### Pain Points

- No public changelog -- users relied on word-of-mouth or commit messages
- Internal changelogs were too technical and detailed for external consumption
- The console dashboard had a "Recent from the Changelog" section with no data source
- Mermaid diagrams in markdown content rendered as plain code blocks instead of diagrams

## Solution / What's New

### Changelog Timeline (`/changelog`)

A single-page timeline where entries expand inline. The primary reading experience happens on the list page -- no need to navigate to a separate detail page for each entry.

- **Click to expand**: Full markdown content appears in-place below the entry header
- **Click to collapse**: Content disappears, returning to the compact view
- **Single expansion**: Only one entry expanded at a time
- **Hover action icons**: Copy shareable link, open in new tab, chevron indicator
- **Content blends with background**: No card/box wrappers, subtle hover states

### Text Search + Category Filters

Two complementary filtering mechanisms above the timeline:

- **Search input**: Filters entries client-side by title, excerpt, and tags (debounced 300ms)
- **Category tabs**: All / Features / Improvements / Fixes / Breaking
- Both applied together via `useMemo`

### Detail Page (`/changelog/[slug]`)

Simple centered layout for shareable links. Category badge, date, title, tags, and the full markdown body -- no sidebars, no 3-column layout.

### Site-wide Mermaid Rendering

Wired the existing `MermaidDiagram` component into `MDXRenderer.tsx`. Now all content types (docs, blog, tutorials, changelog) render Mermaid diagrams as interactive SVGs instead of code blocks.

### `recent.json` for Console Dashboard

A pre-build Node.js script reads all changelog markdown files, parses frontmatter, selects the 10 most recent entries, and writes `public/changelog/recent.json`. This file is served via GitHub Pages CDN and consumed by the console dashboard widget (T05).

## Implementation Details

### Architecture

```
src/lib/changelog.ts              Data layer (read, parse, sort)
src/lib/constants.tsx              +CHANGELOG_DIRECTORY
src/lib/types-client.ts            +ChangelogEntry, ChangelogCategory
src/lib/MDXRenderer.tsx            +Mermaid detection in pre component

src/components/changelog/
  ChangelogTimeline.tsx            Client: search + filters + expand
  ChangelogCategoryBadge.tsx       Colored badge (green/blue/amber/red)
  ChangelogMarkdownBody.tsx        Markdown renderer with Mermaid
  index.ts                         Barrel exports

src/app/(root)/changelog/
  page.tsx                         List page (server component)
  [slug]/page.tsx                  Detail page with generateStaticParams

scripts/generate-recent-changelog.mjs    Pre-build script
```

### Key Technical Decisions

- **Content loading follows the `tutorials.ts` pattern**: `getAllChangelogEntries()`, `getChangelogContentBySlug()`, `getNextChangelogEntry()` -- all in one file, no functions added to `mdx.ts`
- **Mermaid detection in `pre` component**: Inspects the HAST node for `language-mermaid` className, extracts code text, renders `MermaidDiagram` instead of `CodeBlock`
- **Date normalisation**: `gray-matter` parses YAML dates as JS Date objects; both `changelog.ts` and the pre-build script normalise to `YYYY-MM-DD`
- **`recent.json` is gitignored**: Generated build artifact, not committed

### Files Created

- `src/lib/changelog.ts`
- `src/components/changelog/ChangelogTimeline.tsx`
- `src/components/changelog/ChangelogCategoryBadge.tsx`
- `src/components/changelog/ChangelogMarkdownBody.tsx`
- `src/components/changelog/index.ts`
- `src/app/(root)/changelog/page.tsx`
- `src/app/(root)/changelog/[slug]/page.tsx`
- `scripts/generate-recent-changelog.mjs`
- `public/changelog/2026-02-15-introducing-the-planton-changelog.md` (seed content)

### Files Modified

- `src/lib/constants.tsx` -- Added `CHANGELOG_DIRECTORY`
- `src/lib/types-client.ts` -- Added `ChangelogEntry`, `ChangelogCategory`
- `src/lib/MDXRenderer.tsx` -- Added Mermaid rendering in `pre` component
- `src/components/layout/header/header.tsx` -- Added "Changelog" to `menuResources` and `menuExplorer`
- `package.json` -- Added pre-build step for `recent.json`
- `.gitignore` -- Added `public/changelog/recent.json`

## Benefits

- **Users** can now see what's changed across the Planton platform in a user-friendly format
- **Inline expand** keeps the browsing experience fast -- no page navigation required
- **Mermaid diagrams** now render across the entire site (docs, blog, tutorials, changelog)
- **Console dashboard** has a data source (`recent.json`) for the "Recent from the Changelog" widget
- **Shareable links** via the detail page with OpenGraph metadata
- **Zero backend changes** -- everything is static, CDN-cached, and built at deploy time

## Impact

- **New pages**: `/changelog` (list) + `/changelog/[slug]` (detail)
- **Navigation**: "Changelog" added to Resources and Explorer menus
- **Build pipeline**: One additional pre-build step (~0.8s) for `recent.json` generation
- **Site-wide**: Mermaid diagrams now render in all content types

## Related Work

- This is T02 + T03 of the `20260215.03.public-changelog-system` project
- T04 (Changelog Curator Cursor rule) and T05 (Console dashboard widget) are next
- The `MermaidDiagram` component already existed in `src/components/common/`; this work wired it into the rendering pipeline

---

**Status**: ✅ Live (pending merge to main)
**Build**: Verified -- clean build, 112 pages, Pagefind indexed 57 pages

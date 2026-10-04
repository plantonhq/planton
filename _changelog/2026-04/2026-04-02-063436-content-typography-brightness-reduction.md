# Content Typography Brightness Reduction

**Date**: April 2, 2026
**Type**: Design
**Components**: Design System, Documentation, Tutorials, Blog, Changelog, Legal Pages

## Summary

Reduced text brightness across all content pages (docs, tutorials, blog) from `#ededed` to `#b0b0b0` for headings, section titles, table headers, sidebar labels, and page titles. Body text remains `#a0a0a0`. The heading-to-body gap shrank from 77 units (48% brighter) to 16 units (10% brighter), creating a calmer reading experience in dense documentation while preserving scanability through size and weight.

## Problem Statement / Motivation

After the monochrome canvas unification (April 1), every heading level (h1–h6) in markdown content used `text-white` (`#ededed`) with `font-bold` (700) against `#a0a0a0` body text. This produced a 77-unit luminance gap on a 0–255 scale — 4x larger than GitHub dark mode's heading-to-body gap. In documentation pages with 15–20+ section headings, every heading shouted with equal intensity, creating a strobe-like visual pattern that caused reading fatigue.

### Pain Points

- In-body headings at `#ededed` were indistinguishable in brightness from the page title — no typographic hierarchy beyond font size
- `font-bold` (700) on all six heading levels amplified perceived brightness through increased glyph surface area
- Table headers, sidebar section titles, TOC heading, and contributor names all used the same bright `#ededed`, making every label compete for attention
- The design system's own weight scale prescribed Semibold (600) for section titles but the implementation used Bold (700) everywhere
- The `ChangelogMarkdownBody.tsx` and `LegalContent.tsx` renderers still had blue-tinted Tailwind `gray-*` classes from the incomplete monochrome purge

## Solution / What's New

### Phase 1: Heading Token Centralization

Added six `HEADING_H*_CLASSES` constants to `src/theme/docs.ts`, centralizing heading color, weight, size, and spacing in one file. All three markdown renderers (MDXRenderer, ChangelogMarkdownBody, LegalContent) now import these tokens instead of hardcoding inline class strings.

### Phase 2: Color Reduction — #ededed → #d4d4d4 → #b0b0b0

Initial change to `#d4d4d4` (32% brighter than body) was still too aggressive. After visual comparison with Cursor's editor and iterative feedback, settled on `#b0b0b0` (`secondary.80` in the design system palette) — 10% brighter than body text. Combined with `font-semibold` (600) instead of `font-bold` (700), this produces a comfortable reading density where size and weight carry the structural hierarchy.

### Phase 3: Content Chrome Alignment

Extended the `#b0b0b0` treatment beyond markdown headings to all content-area chrome:

- **Markdown table headers** (`TABLE_HEADER_CLASSES` in `docs.ts`)
- **Page titles** (frontmatter h1 in `MDXRenderer.tsx`)
- **Tutorial list page** title, entry titles, sort button, author names (`TutorialsPageClient.tsx`, `TutorialListRow.tsx`)
- **Sidebar section titles** (`TutorialsSidebar.tsx`, `content-sidebar.tsx`, `DocsLayout.tsx`)
- **TOC heading** "On this page" (`TableOfContents.tsx`)
- **Contributors heading** and author names (`AuthorSection.tsx`)
- **Bold text** in markdown body (`<strong>` mapping in `MDXRenderer.tsx`)

Interactive states (hover highlights, active/selected indicators) remain at `text-white` for feedback contrast.

### Phase 4: Duplicate Heading Removal

Removed the redundant "Documentation" header from `DocsSidebar.tsx`. The sticky docs header already displays "Planton Documentation" — the sidebar repeated "Documentation" immediately below it, wasting vertical space.

### Phase 5: Duplicate Title Stripping

Added logic to `MDXRenderer.tsx` to strip the leading `# Title` from markdown body when it exactly matches the frontmatter `title`. Many docs pages rendered the title twice — once from frontmatter metadata and once from the first markdown heading.

### Phase 6: Monochrome Consistency Pass

- **ChangelogMarkdownBody.tsx**: Replaced all blue-tinted Tailwind `gray-*` classes with neutral palette values (`#a0a0a0`, `#2a2a2a`, `#3a3a3a`, `#111`). Imported heading tokens from `docs.ts`.
- **LegalContent.tsx**: Same gray purge, plus replaced `text-blue-400` links with the monochrome `LINK_CLASSES` pattern. Imported heading tokens from `docs.ts`.

## Implementation Details

### Files Modified

**Design tokens** (1 file):
- `src/theme/docs.ts` — Added 6 heading constants, changed table header color

**Primary markdown renderer** (1 file):
- `src/lib/MDXRenderer.tsx` — Imported heading tokens, added `<strong>` mapping, changed page title color, added duplicate h1 stripping

**Consistency pass** (2 files):
- `src/components/changelog/ChangelogMarkdownBody.tsx` — Heading tokens + full gray purge
- `src/components/legal/LegalContent.tsx` — Heading tokens + gray purge + blue link removal

**Tutorial system** (3 files):
- `src/components/tutorials/TutorialsPageClient.tsx` — Page heading, mobile drawer title, sort button, empty state
- `src/components/tutorials/TutorialListRow.tsx` — Title link, author link
- `src/components/tutorials/TutorialsSidebar.tsx` — Section header

**Docs system** (2 files):
- `src/app/(root)/docs/components/DocsLayout.tsx` — Sticky header, mobile drawer header
- `src/app/(root)/docs/components/DocsSidebar.tsx` — Removed redundant "Documentation" header

**Shared components** (3 files):
- `src/components/common/content-sidebar.tsx` — Section title, inactive record titles
- `src/components/blog/TableOfContents.tsx` — "On this page" heading
- `src/components/blog/AuthorSection.tsx` — "Contributors" heading, author names

**Documentation** (2 files):
- `public/branding/design-system.md` — Updated heading hierarchy table and content chrome documentation
- `../../planton/_team/agents/_knowledge-base/context/product/ux/web/why-different-color-shades-for-planton-website-and-console.md` — Added content pages as a third brightness tier in the dwell-time analysis

### Key Design Decisions

- **`#b0b0b0` not `#d4d4d4`**: The first iteration (`#d4d4d4`, 32% brighter than body) was visually too close to the original `#ededed`. User feedback confirmed that headings should be "only 10-20% brighter than the text under the heading." `#b0b0b0` (10% brighter) was the right landing point — still above body text, close to GitHub's heading-to-body gap.
- **Uniform color for all heading levels**: GitHub, Vercel, Stripe, Notion, and Linear all use a single heading color with size as the sole level differentiator. Adding a color gradient per level would add complexity for no meaningful UX gain.
- **`font-semibold` (600) not `font-bold` (700)**: The design system weight scale already prescribed Semibold for section titles. The implementation was diverging from its own spec.
- **Interactive states stay white**: Hover highlights and active selection indicators remain at `text-white` because they serve state feedback, not content hierarchy.
- **Content pages as a third brightness tier**: Documentation and tutorials are reading surfaces with 5-15 minute dwell times — between marketing pages (10-30 seconds) and the console (30 min to hours). The dwell-time gradient in the UX documentation now reflects this: marketing (15:1) → content (9:1) → console (10.2:1) → editor (7.8:1).

## Benefits

- Heading-dense documentation pages no longer create visual fatigue from strobe-like brightness alternation
- Clear three-tier typographic hierarchy: marketing headings (`#ededed`) → content headings (`#b0b0b0`) → body text (`#a0a0a0`)
- All heading color/weight is centralized in `src/theme/docs.ts` — future tuning is a single-file edit
- Zero blue-tinted Tailwind `gray-*` classes remain in the changelog or legal renderers
- No duplicate page titles on docs pages
- No redundant "Documentation" sidebar header
- WCAG AAA accessibility maintained: `#b0b0b0` on `#0a0a0a` = ~9:1 contrast ratio

## Impact

- All 55+ docs pages, 3 tutorial pages, and 1 blog page now use the calmer typography
- Changelog and legal pages aligned to the same neutral palette
- Design system documentation updated to reflect the new convention
- UX rationale document updated with the content-page brightness tier

## Related Work

- [Monochrome Canvas Unification](2026-04-01-190727-monochrome-canvas-unification.md) — established the neutral palette this change extends
- [Website Shell Component Extraction](2026-04-01-175418-website-shell-component-extraction.md) — extracted shared header/footer into MUI-only package

---

**Status**: Live
**Timeline**: Single session

# Branding Design System Page

**Date**: March 29, 2026
**Type**: Feature
**Components**: Design System, Build System, Content Management

## Summary

Added a public-facing design system reference page at `/branding/design-system` that synthesizes the platform-wide monochrome visual language across the marketing website and web console. The page renders markdown through the existing `MDXRenderer` pipeline and also serves raw markdown at `/branding/design-system.md` for LLM agent consumption.

## Problem Statement / Motivation

The Planton platform recently completed a comprehensive monochrome redesign across both the marketing website (15 changelogs, March 25-29) and the web console (12 changelogs, March 28-29). The visual language decisions — token pipelines, semantic color budgets, surface hierarchy, typography choices — were spread across changelog files and an internal-only console design system document.

### Pain Points

- No single public reference for the platform-wide visual language
- Internal console design system document not accessible to external contributors or partners
- LLM agents working on the codebase had no machine-readable design system reference
- Website and console design decisions documented separately with no unified view

## Solution / What's New

### Architecture: Static File + Page Route

Follows the established planton.ai content pattern — markdown in `public/` for raw serving, a Next.js page route for rendered HTML:

- **`public/branding/design-system.md`** — Markdown source, automatically served at `/branding/design-system.md` as a static asset
- **`src/app/(root)/branding/design-system/page.tsx`** — Reads the file at build time, renders via `MDXRenderer`, exports as static HTML

### Layout

A new `BrandingContentLayout` component provides a focused single-document layout:
- Content area with full `MDXRenderer` rendering (tables, code blocks, mermaid, syntax highlighting)
- Sticky right-side Table of Contents sidebar (reuses existing `TableOfContents` component)
- No left sidebar — this is a standalone reference, not a multi-page section
- TOC collapses on smaller screens

### Content

The design system document covers:
- Design philosophy (monochrome, dark-first, data-dense)
- Dual-surface model — why website (`#0a0a0a`, Tailwind) and console (`#0d1117`, MUI) differ
- Complete color system with hex values for both surfaces
- Typography (Inter, weight scale, heading hierarchy)
- Surface hierarchy (canvas/subtle/muted/deep)
- Button hierarchy, icon rules, dark/light mode
- Console token pipeline (5-layer architecture summary)
- Website theme tokens
- Rules for contributors and AI agents

## Implementation Details

### Files Created

| File | Purpose |
|------|---------|
| `public/branding/design-system.md` | Markdown source (also raw .md endpoint) |
| `src/app/(root)/branding/design-system/page.tsx` | Page route with `generateMetadata`, `MDXRenderer`, `rawPath` wiring |
| `src/components/branding/BrandingContentLayout.tsx` | Content + sticky TOC layout |

### Files Modified

| File | Change |
|------|--------|
| `src/lib/constants.tsx` | Added `BRANDING_DIRECTORY` alongside existing content directory constants |

### Key Decisions

- **Direct page route, not catch-all**: A `[[...slug]]` route would add filesystem scanning complexity for one page. Refactoring to catch-all later is straightforward if `/branding/` grows.
- **`public/` for raw .md**: Files in `public/` are copied verbatim to `out/`, giving free raw markdown serving without an API route.
- **Reused `MDXRenderer`**: Same renderer as docs/blog — consistent rendering of tables, code blocks, mermaid diagrams, PageActions (copy markdown, view raw).

## Benefits

- **Single source of truth**: One document covering both website and console visual language
- **LLM-consumable**: Raw markdown at `/branding/design-system.md` for agent workflows
- **Zero new dependencies**: Reuses existing `MDXRenderer`, `TableOfContents`, `PageActions`, and `gray-matter`
- **Build-verified**: Static export succeeds with 125 pages; `/branding/design-system` appears in the route manifest

## Impact

- **Contributors and AI agents**: Canonical reference for the monochrome design system, accessible from both the rendered page and raw markdown
- **Build output**: `out/branding/design-system.html` (rendered) and `out/branding/design-system.md` (raw) both present in static export

## Related Work

- Console design system document: `planton/_team/agents/_knowledge-base/context/product/ux/planton-console-design-system.md`
- Website monochrome redesign: 15 changelogs from March 25-29
- Console monochrome theme project: `planton/_projects/20260328.02.console-monochrome-theme/`
- Cloudflare routing update: `/branding` added to website-origin exemption list in `cloudflare-ruleset.planton-ai-origin-routing.yaml` (planton monorepo)

---

**Status**: ✅ Live (pending deployment)
**Timeline**: Single session

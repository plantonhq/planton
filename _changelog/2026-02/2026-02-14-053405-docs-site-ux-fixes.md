# Docs Site UX Fixes: Sidebar, URLs, Copy, Styling

**Date**: February 14, 2026
**Type**: Enhancement
**Components**: Documentation Site, Build Scripts, Sidebar Navigation, MDX Renderer

## Summary

Seven UX issues in the Planton documentation site were fixed in a single session: broken sidebar labels, ugly URL slugs, white flash on page navigation, sidebar scroll position loss, broken icon fallbacks, missing code block copy button, and overly bright inline code styling. The build pipeline, sidebar component, page layout, and markdown renderer were all updated.

## Problem Statement / Motivation

After deploying 136 catalog pages and 40 hand-written docs pages, several usability issues surfaced during manual site review:

### Pain Points

- Sidebar labels like "Route53dnsrecord", "S3objectset", "Ecsservice" were illegible — the `generateTitle()` function failed for all-lowercase compound words with no word boundaries
- URLs like `/docs/catalog/aws/awsroute53dnsrecord` were redundant (provider prefix repeated) and not human-readable
- Every page navigation caused a visible white flash because `DocsLayout` re-mounted, re-fetching `docs-structure.json` each time
- Clicking a sidebar item below the viewport scrolled the sidebar back to the top, losing the user's position
- Missing component icons showed browser-default broken image placeholders (green/teal squares)
- Code blocks had no copy-to-clipboard functionality
- Inline code used `text-purple-300` on `bg-purple-900/30` — described as "too glittery"

## Solution / What's New

### Title Extraction from Content Headings

Instead of relying on the flawed `generateTitle()` algorithm, the build script now extracts titles from the first `# ` heading in each catalog-page.md file. These headings are hand-written and always correct (e.g., `# AWS Route53 DNS Record`). The provider prefix is stripped for sidebar labels ("Route53 DNS Record"), and a kebab-case slug is derived for URLs (`route53-dns-record`).

### Next.js Layout Architecture

Created `app/docs/layout.tsx` to hold the header, left sidebar, and mobile drawer. This layout persists across page navigations — the sidebar never re-mounts, never re-fetches its JSON, and never loses scroll position. A `loading.tsx` skeleton prevents any flash during page transitions.

### Sidebar Scroll Preservation

The `expandedPaths` state now merges new ancestors into the existing set rather than replacing it. A scroll-to-active mechanism (`data-active` attribute + `scrollIntoView`) ensures the clicked item stays visible.

### Icon Fallback with Letter Badge

Missing icons now display a styled letter badge (first letter of the component title, 20x20px, `bg-slate-700` rounded square) instead of hiding entirely — maintaining alignment with icons that do load.

### Code Block Copy Button

New `CodeBlock` component wraps `<pre>` elements with a copy-to-clipboard button that appears on hover. Uses `navigator.clipboard.writeText()` with a 2-second checkmark confirmation.

### Inline Code Restyling

Changed from `bg-purple-900/30 text-purple-300` to `bg-slate-800/60 text-sky-300` — soft blue on neutral dark slate, similar to VS Code's inline code treatment.

## Implementation Details

### Build Script Changes (`site/scripts/copy-component-docs.ts`)

- Added `extractTitleFromContent()` — regex for first `^# (.+)$` heading
- Added `stripProviderPrefix()` — removes "AWS ", "GCP ", "Azure ", etc.
- Added `generateSlug()` — title to kebab-case (`"Route53 DNS Record"` → `"route53-dns-record"`)
- Added `yamlEscape()` — escapes double quotes in YAML frontmatter values (fixed a build failure from a legacy DigitalOcean doc heading containing `"1-Click"`)
- Output filenames changed from `{component}.md` to `{slug}.md`
- Provider index links updated to use slug-based URLs
- `componentName` preserved in frontmatter for icon path resolution

### Structure Generation (`generate-docs-structure.ts`, `fileSystem.ts`)

- Added `componentName` to `DocItem` interface
- Propagated from frontmatter through to the sidebar JSON

### Layout Restructure

- Created `app/docs/layout.tsx` — client component with header, sidebar, mobile drawer
- Refactored `page.tsx` — returns content + right sidebar as fragment (flex children of layout)
- Created `loading.tsx` — dark-themed skeleton matching site colors

### Sidebar Component (`DocsSidebar.tsx`)

- Structure fetch moved to mount-only (ref guard prevents re-fetch)
- `expandedPaths` useEffect changed from `new Set()` to `new Set(prev)`
- Added `data-active` attribute + `scrollIntoView({ behavior: 'smooth', block: 'nearest' })`
- Icon resolution now uses `componentName` from item data (survives slug changes)
- Added `onError` handler on `Image` — fallback to letter badge on load failure

### MDX Renderer (`MDXRenderer.tsx`)

- `pre` override replaced with `CodeBlock` component
- Inline `code` styling changed to `bg-slate-800/60 text-sky-300`

## Benefits

- **Correct sidebar labels** for all 190 catalog pages — extracted from hand-written headings, not generated
- **Clean URLs** — `/docs/catalog/aws/alb` instead of `/docs/catalog/aws/awsalb`
- **No navigation flash** — sidebar persists, skeleton loading for content
- **Scroll position preserved** — sidebar stays put when navigating between pages
- **No broken icons** — letter badge fallback maintains visual consistency
- **Copy code in one click** — hover any code block to copy
- **Professional inline code styling** — readable without being visually dominant

## Impact

- **All 190 catalog page URLs changed** — old URLs will 404 (acceptable for pre-1.0 docs)
- **Build pipeline updated** — `copy-component-docs.ts` now generates slug-based filenames
- **No breaking changes to docs content** — only the site infrastructure and rendering changed

## Files Changed

**New files (3):**
- `site/src/app/docs/layout.tsx`
- `site/src/app/docs/[[...slug]]/loading.tsx`
- `site/src/app/docs/components/CodeBlock.tsx`

**Modified files (6):**
- `site/scripts/copy-component-docs.ts` — title extraction, slug generation, YAML escaping
- `site/scripts/generate-docs-structure.ts` — componentName propagation
- `site/src/app/docs/utils/fileSystem.ts` — componentName in DocItem interface
- `site/src/app/docs/[[...slug]]/page.tsx` — layout refactor
- `site/src/app/docs/components/DocsSidebar.tsx` — scroll, icons, componentName
- `site/src/app/docs/components/MDXRenderer.tsx` — CodeBlock, inline code styling

**Generated files (~190):**
- All catalog pages regenerated with new slug-based filenames and corrected titles

---

**Status**: Production Ready
**Timeline**: Single session (~2 hours)

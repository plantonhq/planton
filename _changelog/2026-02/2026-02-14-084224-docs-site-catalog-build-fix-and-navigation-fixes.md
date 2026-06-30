# Docs Site: Catalog Build Deduplication and Navigation Fixes

**Date**: February 14, 2026
**Type**: Bug Fix
**Components**: Build System, Docs Site, User Experience

## Summary

Fixed two categories of bugs in the Planton docs site: (1) a build script bug that caused duplicate sidebar entries for Auth0, OpenFGA, and Scaleway providers, and (2) four navigation issues where internal links in catalog pages navigated to the home page instead of their target, "Read next article" skipped section index pages, and showed sidebar labels instead of full page titles.

## Problem Statement / Motivation

### Catalog Build Duplicates

The `copy-component-docs.ts` build script hardcoded a list of 11 provider directories to clear between builds. Three providers added later — `auth0`, `openfga`, `scaleway` — were missing from this list. Their output directories were never cleaned, so stale files from older builds persisted alongside new ones. This caused visible duplicates in the sidebar: "Authorization Model" and "Authorizationmodel" both appeared for OpenFGA, and "Store" appeared twice.

### Navigation Issues

1. **Catalog list links navigated to home page** — Component links on provider index pages (e.g., `/docs/catalog/aws`) used plain `<a>` tags instead of Next.js `<Link>`. In a static-exported app served by `serve`, these caused full browser navigations that bypassed the client-side router.
2. **"Read next article" navigated to home page** — Same root cause: the `NextArticle` component used a plain `<a>` tag.
3. **"Read next article" skipped section index pages** — The `getNextDocItem` function flattened only leaf files, so section transitions jumped past index pages.
4. **"Read next article" showed sidebar labels** — Titles came from frontmatter (sidebar labels like "ALB") instead of the page's `#` heading (e.g., "AWS ALB").

### Pain Points

- Users clicking any link in the catalog component list were redirected to the home page
- "Read next article" at the bottom of every catalog page was broken
- Section transitions in "Read next article" skipped the next section's overview
- Catalog page titles in "Read next article" were truncated (provider prefix stripped)

## Solution / What's New

### Build Fix: Dynamic Provider Directory Clearing

Replaced the hardcoded `providerDirs` array with a dynamic scan of the output directory. Every subdirectory in `site/public/docs/catalog/` is now cleared before regeneration, regardless of which providers exist.

### Navigation Fix: Next.js Link for Internal Links

Replaced plain `<a>` tags with Next.js `<Link>` for all internal links in the markdown renderer and the `NextArticle` component. External links continue to use `<a>` with `target="_blank"`.

### Section Transitions: Include Directory Index Pages

Modified `getNextDocItem` to include directories with `hasIndex: true` in the flattened reading order, positioned before their children. This produces the natural reading order: last page of Section A, then Section B index, then first child of Section B.

### Page Titles: Extract from Content Heading

Added a `pageTitle` field to `DocItem`, populated during structure building by extracting the first `#` heading from each markdown file. The "Read next article" block now shows `pageTitle` (falling back to `title` for pages without a heading).

## Implementation Details

### Files Changed

| File | Change |
|------|--------|
| `site/scripts/copy-component-docs.ts` | Replaced hardcoded 11-provider list with dynamic `fs.readdirSync` scan of catalog output directory |
| `site/src/app/docs/components/MDXRenderer.tsx` | Added `Link` import; replaced `<a>` with `<Link>` for internal links in both the markdown `a` component and `NextArticle` |
| `site/src/app/docs/utils/fileSystem.ts` | Added `pageTitle` to `DocItem`; extract `#` heading during `buildStructure`; include `hasIndex` directories in `getNextDocItem` flattening |
| `site/src/app/docs/[[...slug]]/page.tsx` | Updated `nextArticle` prop to prefer `pageTitle` over `title` |

### Key Code Change: Dynamic Provider Clearing

```typescript
// Before: hardcoded, missing auth0/openfga/scaleway
const providerDirs = [
    'aws', 'gcp', 'azure', 'kubernetes',
    'cloudflare', 'civo', 'digitalocean',
    'atlas', 'confluent', 'openstack', 'snowflake'
];

// After: dynamic, future-proof
const providerDirs = fs.existsSync(siteDocsRoot)
    ? fs.readdirSync(siteDocsRoot).filter(item =>
        fs.statSync(path.join(siteDocsRoot, item)).isDirectory()
      )
    : [];
```

### Key Code Change: Internal Link Routing

```typescript
// Before: plain <a> tag for all links
a: ({ href, children }) => (
  <a href={href} className="...">{children}</a>
)

// After: Next.js Link for internal, <a> for external
a: ({ href, children }) => {
  const isExternal = href?.startsWith('http');
  if (!isExternal && href) {
    return <Link href={href} className="...">{children}</Link>;
  }
  return <a href={href} target="_blank" rel="noopener noreferrer" className="...">{children}</a>;
}
```

## Benefits

- **Zero duplicate sidebar entries** — All 14 providers now properly cleared between builds
- **Working catalog navigation** — All links in catalog pages navigate correctly via client-side routing
- **Natural reading order** — Section transitions include index pages
- **Full page titles** — "Read next article" shows "AWS ALB" not "ALB"
- **Future-proof** — New providers are automatically included in build cleanup without code changes

## Impact

- **End users**: All catalog page links now work; "Read next article" navigates correctly and shows meaningful titles
- **Developers**: No manual maintenance needed when adding new providers — the build script handles cleanup automatically
- **Build**: 213 components across 14 providers, zero skipped, zero duplicates

## Related Work

- Previous docs site UX fixes session (2026-02-14): sidebar labels, URL slugs, white flash, scroll position, broken icons, code copy, inline code styling
- Catalog page coverage completion (2026-02-14): all 14 providers at 100% catalog page coverage

---

**Status**: Production Ready
**Timeline**: Single session

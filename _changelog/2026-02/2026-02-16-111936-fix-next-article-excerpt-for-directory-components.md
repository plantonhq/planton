# Fix Next Article Excerpt for Directory-Based Component Pages

**Date**: February 16, 2026
**Type**: Bug Fix
**Components**: Documentation Site, Next Article Navigation

## Summary

Fixed missing excerpt text in the "Next article" navigation card at the bottom of documentation pages. Directory-based component pages (introduced during the T09 preset documentation migration) had their excerpt hardcoded to an empty string, causing the card to display only the title with no summary preview.

## Problem Statement / Motivation

The "Next article" card at the bottom of each docs page provides a preview of what comes next in the reading order -- title, summary excerpt, and a link. After the T09 migration from flat files to directories, every component page's excerpt was empty, degrading the reading experience by removing the contextual preview that helps users decide whether to continue reading.

### Pain Points

- **No summary preview**: The "Next article" card showed only the article title with no description, unlike the production site which shows a rich multi-line excerpt
- **Reduced navigation quality**: Users lost the at-a-glance preview that helps them decide whether the next page is relevant to their task
- **Same root cause as sidebar icons**: Part of the broader regression from the flat-file-to-directory migration in T09

## Solution / What's New

Both structure builders (`generate-docs-structure.ts` and `fileSystem.ts`) already read the full content of each directory's `index.md` file to extract frontmatter and page titles. The fix captures that content and passes it through the existing `generateExcerptFromContent()` function instead of discarding it.

### Implementation

**`site/src/app/docs/utils/fileSystem.ts`**:
- Added `dirExcerpt` variable alongside existing `dirPageTitle`
- Populated it with `generateExcerptFromContent(fileContent)` inside the index file parsing loop
- Replaced `excerpt: ''` with `excerpt: dirExcerpt`

**`site/scripts/generate-docs-structure.ts`**:
- Same pattern: captured the raw file content (previously inlined into `matter()`) into a variable
- Generated excerpt using the locally defined `generateExcerptFromContent()` function
- Replaced `excerpt: ''` with `excerpt: dirExcerpt`

## Benefits

- **Rich previews restored**: All 267 component pages now show meaningful excerpt text in the "Next article" card
- **Parity with production**: The localhost development site now matches the production site's reading experience
- **Zero overhead**: The content was already being read from disk; the only addition is the excerpt generation call

## Impact

- **Users**: Documentation readers see a helpful summary in the "Next article" card, improving navigation and content discovery
- **Build output**: `docs-structure.json` grew from ~114 KB to ~251 KB due to the added excerpt text -- a reasonable trade-off for the UX improvement

## Related Work

- Sidebar icon fix (`_changelog/2026-02/2026-02-16-110355-fix-sidebar-component-icons-after-directory-migration.md`) -- companion fix for the same T09 migration regression
- T09 Preset Documentation Pages (`_changelog/2026-02/2026-02-15-213453-preset-documentation-pages.md`) -- the session that introduced the directory migration

---

**Status**: Production Ready
**Timeline**: Single session fix

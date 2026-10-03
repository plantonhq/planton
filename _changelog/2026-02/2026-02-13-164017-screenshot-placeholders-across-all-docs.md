# Screenshot Placeholders Across All Documentation Pages

**Date**: February 13, 2026
**Type**: Enhancement
**Components**: Documentation

## Summary

Added standardized screenshot placeholders to 51 of 53 documentation pages, completing Phase 10 of the Planton docs overhaul and finishing all Week 1 structural work. Every page that references a web console view, wizard, form, or dashboard now has invisible HTML comment markers ready for screenshot collection.

## Problem Statement / Motivation

Documentation pages frequently describe UI workflows — creating connections, browsing the Infra Catalog, monitoring pipeline progress, managing team members — but contain no visual aids. When a developer reads "Navigate to Connections and click the AWS card," they have no reference image to confirm they are in the right place. Screenshots at key points in these workflows reduce cognitive load and build confidence.

### Pain Points

- 49 of 53 pages had zero screenshot placeholders before this session
- 4 pages used inconsistent blockquote-style placeholders (`> **Screenshot Placeholder**:`) instead of the standard HTML comment format
- No systematic mapping between documentation content and web console routes existed

## Solution / What's New

### Standardized Placeholder Format

Every placeholder follows the convention defined in `coding-guidelines/docs-format-conventions.md`:

```markdown
<!-- SCREENSHOT: {brief context}
  Page: {web console URL path}
  Action: {what state the UI should be in}
  Focus: {what to highlight or crop to}
  Alt: {accessibility text for the future image}
-->
```

### Coverage

| Section | Pages | Placeholders |
|---------|-------|--------------|
| Root | 2 of 3 | 3 |
| Platform | 5 of 5 | 22 |
| Connect | 8 of 8 | 12 |
| Infra Hub | 13 of 13 | 22 |
| Service Hub | 12 of 12 | 16 |
| Secrets and Config | 4 of 4 | 6 |
| Cloud Ops | 3 of 3 | 6 |
| Runner | 2 of 3 | 2 |
| Teams and Access | 2 of 2 | 5 |
| **Total** | **51 of 53** | **89** |

Two pages intentionally excluded:
- `cli.md` — CLI installation and reference only; no console UI to capture
- `runner/security-model.md` — Conceptual security guarantees; no UI to capture

### Format Standardization

Converted 9 blockquote-style placeholders in `platform/resource-hierarchy.md` (1) and `platform/platform-tour.md` (8) to the standard HTML comment format with full Page/Action/Focus/Alt metadata.

## Implementation Details

- 21 files changed, +201/-9 lines
- All placeholders mapped to actual web console routes verified against the route definitions in `client-apps/web/console/src/routes/index.ts`
- Runner section placeholders comply with the IP preservation guidelines — no internal architecture details, kind names, or port numbers in any description
- Placeholders placed adjacent to the paragraph that references the visual, following the convention of "before or after the referencing content"

## Benefits

- **Screenshot collection is now systematic** — a future contributor can grep for `SCREENSHOT:` and process each placeholder into an actual image
- **Alt text is pre-written** — accessibility compliance is built in from the start
- **Web console routes are documented** — each placeholder's `Page` field serves as a cross-reference between docs content and actual console pages
- **No user-visible changes** — HTML comments are invisible in rendered Markdown; the documentation reads identically

## Impact

This completes the final task of Week 1 (Documentation Structure and Scaffolding) of the Planton docs overhaul project. The documentation site now has:

- 53 pages across 8 sections
- 89 screenshot placeholders ready for image collection
- Zero broken links
- Zero blockquote-style placeholders (all standardized)
- Consistent formatting across all pages

Week 2 deep content work can now begin with the full structural foundation in place.

## Related Work

- Phase 1-9 of the docs overhaul (sessions 1-11)
- `coding-guidelines/docs-format-conventions.md` — Screenshot placeholder convention definition
- `coding-guidelines/runner-ip-preservation.md` — Runner content restrictions

---

**Status**: Live
**Timeline**: Single session

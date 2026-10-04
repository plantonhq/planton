# Docs Mobile UX and Page Actions Overhaul

**Date**: February 14, 2026
**Type**: Enhancement
**Components**: Documentation, UI Components, Layout Components, Responsive Design

## Summary

Fixed mobile horizontal scroll, repositioned the copy/page-actions button next to the page title, replaced the Snackbar copy-success notification with inline green checkmark feedback, and added an "Open Raw" option to the page actions dropdown. These changes improve the mobile documentation reading experience and bring the page actions UX in line with world-class docs sites.

## Problem Statement / Motivation

The docs site had several mobile UX issues carried over from Session 17's partially-complete responsive work:

### Pain Points

- **Horizontal scroll on real mobile devices**: Despite `overflow-x: clip` on `<html>` and `overflow-x-hidden` on the content area, users could still scroll horizontally on phones. The root cause was the footer's terms links row overflowing on screens <= ~378px wide.
- **Copy button occupied its own line on mobile**: The MUI `IconButton` (34px touch target) sat in the metadata row below the title, wasting vertical space on small screens.
- **Snackbar feedback was disconnected**: After copying, a heavy MUI Snackbar popped up at the bottom of the page — far from where the user's eye was focused.
- **"Open Raw" was buried**: To view the raw markdown source in a new tab, users had to first open the "View as Markdown" dialog, then click a small icon in the dialog header.

## Solution / What's New

### 1. Mobile Horizontal Scroll Fix

Identified and fixed the root cause: footer terms links in an unwrapping `flex-row` with `gap-5` (20px) and `whitespace-nowrap` on each link. On a 375px phone, the row needed ~338px but only had ~335px available.

**Fix**: Added `flex-wrap` and `gap-x-5 gap-y-1` to the terms row. Hidden dot separators below `sm` breakpoint (dots wrapping to the start of a new line looked awkward). Added `overflow-x: clip` to `body` alongside `html` for defense-in-depth.

### 2. Copy Button in Title Row

Moved `DocsPageActions` from the metadata row (date/author) into the title row. The title and copy action now share a flex container with the title taking `flex-1` and the action right-aligned. On mobile, the copy icon is a subtle 14px icon; on desktop, the full "Copy page" button with chevron.

### 3. Inline Copy Feedback

Replaced the MUI `CopySnackbar` component with inline icon swapping. When content is copied, the `ContentCopy` icon becomes a green `Check` icon for 2 seconds, then reverts. This pattern matches `HeadingWithAnchor` and `CodeBlock` in the codebase.

The `MarkdownViewDialog` also received its own local copy feedback — clicking the copy icon in the dialog header shows the green checkmark in-place rather than leaking feedback to the hidden parent button.

### 4. "Open Raw" Dropdown Option

Added a third option to the page actions dropdown: "Open Raw" with an `OpenInNew` icon. Opens `${path}.md` in a new browser tab, same as GitHub's "Raw" button. The existing `handleOpenSourceContent` function was wired through as `onOpenRaw`.

## Implementation Details

**Files changed**: 8 modified, 1 deleted, 1 changelog created

| File | Change |
|------|--------|
| `src/components/layout/footer.tsx` | flex-wrap on terms row, hide dots on mobile |
| `src/app/globals.css` | `overflow-x: clip` on body alongside html |
| `src/lib/MDXRenderer.tsx` | Move DocsPageActions to title row, simplify metadata condition |
| `src/app/(root)/docs/components/DocsPageActions/CopyButton.tsx` | `copied` prop, green Check icon, 14px mobile icon |
| `src/app/(root)/docs/components/DocsPageActions/ActionsMenu.tsx` | `onOpenRaw` prop, "Open Raw" menu item |
| `src/app/(root)/docs/components/DocsPageActions/index.tsx` | useEffect auto-reset, remove Snackbar, wire onOpenRaw |
| `src/app/(root)/docs/components/DocsPageActions/MarkdownViewDialog.tsx` | Local copied state with inline checkmark feedback |
| `src/app/(root)/docs/components/DocsPageActions/CopySnackbar.tsx` | **Deleted** — no longer needed |

## Benefits

- **No horizontal scroll** on any mobile device, including iPhone SE (320px)
- **Cleaner mobile layout** — copy icon sits next to the title instead of occupying its own line
- **Immediate feedback** — green checkmark appears right where the user clicked, not at the bottom of the page
- **Direct raw access** — "Open Raw" is one click from the dropdown, no dialog required
- **Less MUI overhead** — removed Snackbar + Alert components, replaced mobile IconButton with plain HTML button
- **Defense-in-depth** — body overflow protection prevents future regressions from any element in the layout

## Impact

- All 50 documentation pages benefit from the mobile fixes
- Footer fix applies site-wide (all pages, not just docs)
- Page actions dropdown now has three options: Copy as Markdown, View as Markdown, Open Raw
- Deleted CopySnackbar.tsx reduces component count

## Related Work

- Session 17 (2026-02-14): Initial mobile responsiveness work — layout, typography, double hamburger fix
- Session 16 (2026-02-13): Final review and terminology sweep — the last content-focused session before mobile UX work
- `_changelog/2026-02/2026-02-14-050229-docs-mobile-responsiveness.md`: Covers Session 17's partial mobile work

---

**Status**: Live
**Timeline**: 1 session (~2 hours)

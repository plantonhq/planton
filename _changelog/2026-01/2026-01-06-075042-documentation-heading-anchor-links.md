# Documentation Heading Anchor Links

**Date**: January 6, 2026
**Type**: Feature
**Components**: Documentation, UI Components, MDXRenderer

## Summary

Added clickable anchor links to all documentation headings (h1-h6). Hovering over a heading reveals a link icon; clicking copies the section URL to clipboard and shows a green checkmark confirmation. Also added documentation guidelines rule and excluded `_rules` folder from sidebar indexing.

## Problem Statement

Users couldn't easily share links to specific documentation sections. There was no visual affordance for copying section URLs, which is a standard feature in modern documentation sites like GitHub, MDN, and Notion.

### Pain Points

- No way to link directly to a specific heading
- Users had to manually construct URLs with anchor fragments
- No visual feedback when copying links

## Solution

Created a reusable `HeadingWithAnchor` component that wraps all markdown headings with anchor link functionality.

### Behavior

1. **Hover**: Link icon appears to the left of the heading
2. **Click**: 
   - Updates URL hash (`#section-name`)
   - Smooth scrolls to section (with header offset)
   - Copies full URL to clipboard
3. **Feedback**: Green checkmark replaces link icon for 2 seconds

## Implementation Details

### New Component

**File**: `src/components/docs/HeadingWithAnchor.tsx`

```tsx
// Key features:
- LinkIcon and CheckIcon SVG components
- useState for copied state with auto-reset timer
- scroll-mt-24 for fixed header offset (96px)
- Group hover pattern for icon visibility
```

### MDXRenderer Integration

**File**: `src/lib/MDXRenderer.tsx`

Replaced inline heading components with `HeadingWithAnchor`:

```tsx
h2: ({ children }) => (
  <HeadingWithAnchor
    id={generateHeadingId(children)}
    level={2}
    className="text-2xl font-bold text-white mt-6 mb-3"
  >
    {children}
  </HeadingWithAnchor>
),
```

### Documentation Guidelines

**File**: `public/docs/_rules/general-planton-ai-docs-guidelines.mdc`

Created always-apply rule with:
- Truth over speculation principle
- Heading anchor documentation
- Content structure guidelines
- Platform-specific instruction ordering
- Quick checklist for documentation quality

### Sidebar Exclusion

**File**: `src/app/(root)/docs/utils/fileSystem.ts`

Added filter to exclude underscore-prefixed directories:

```typescript
if (item.startsWith('.') || item.startsWith('_')) {
  continue;
}
```

## Benefits

### For Users
- Share exact documentation sections with teammates
- Visual feedback confirms successful copy
- Smooth scroll experience with proper header offset
- Standard UX pattern familiar from other doc sites

### For Documentation Authors
- Auto-generated anchors from heading text
- No manual anchor ID management
- Guidelines ensure consistent documentation quality

## Files Changed

| File | Change |
|------|--------|
| `src/components/docs/HeadingWithAnchor.tsx` | Created - anchor link component |
| `src/components/docs/index.ts` | Created - component exports |
| `src/lib/MDXRenderer.tsx` | Modified - use HeadingWithAnchor for h1-h6 |
| `src/app/(root)/docs/utils/fileSystem.ts` | Modified - exclude `_` prefixed dirs |
| `public/docs/_rules/general-planton-ai-docs-guidelines.mdc` | Created - doc guidelines |

## Related Work

- CLI Installation Documentation (`_changelog/2026-01/2026-01-06-065300-cli-installation-documentation.md`)
- Documentation system uses same MDXRenderer across docs, blog, and tutorials

---

**Status**: ✅ Live
**Timeline**: ~45 minutes implementation


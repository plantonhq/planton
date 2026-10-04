# Header Redesign, Inter Font Migration, and Dropdown Menu Polish

**Date**: March 27, 2026
**Type**: Design
**Components**: Navigation, Design System, Documentation, UI Components

## Summary

Replaced the "Join Beta" modal with GitHub-style "Sign in" / "Sign up" direct links, migrated the site-wide font from Work Sans to Inter, made the Discord icon monochrome, and polished all three dropdown menus for consistent visual density. Also tuned Inter's rendering with proper letter-spacing, disabled MUI's uppercase button convention, and added tight tracking to the docs content wrapper.

## Problem Statement / Motivation

The header carried several design artifacts from an earlier era of the site that no longer matched the monochrome, professional identity established in the March 2026 rebrand:

### Pain Points

- **"Join Beta" modal** added friction to signup — users had to click a button, read a dialog, then click again to reach the console. GitHub, Vercel, and cursor.com all use direct auth links.
- **Purple Discord icon** (`#6665D2`) broke the monochrome theme established across the site.
- **Work Sans** is a wider, humanist sans-serif that reads as "consumer product" rather than "engineering tool." Every peer platform (Vercel, Linear, Cursor, Tailscale) uses a tight geometric sans-serif.
- **Dropdown menus had a data-modeling bug**: Solutions, Resources, and Explore items put their display text in `subLabel` (gray `#666`) instead of `label` (white), making them appear washed out.
- **Resources dropdown** was bare (6 text links in a narrow column) compared to the rich Product dropdown.
- **MUI buttons rendered in UPPERCASE** (`text-transform: uppercase`), which in Inter looked aggressively heavy.

## Solution / What's New

### Header Auth Buttons

Removed the `JoinBetaBtn` component (which triggered a `BetaDialog` modal) and replaced it with two direct links:
- **Sign in** → `planton.ai/login` (ghost button style)
- **Sign up** → `planton.ai/signup` (filled white button)

Both desktop and mobile drawer updated. The `JoinBetaBtn` was removed from the barrel export in `src/components/landing-page/index.ts` but left in the legacy v1 components file to avoid breaking unused-but-compiled v1 code.

### Discord Monochrome

Changed `public/images/discord.svg` fill from `#6665D2` to `currentColor`. The icon now inherits the parent button's text color (`#999999` via `text-text-secondary`), fitting the monochrome theme without any component code changes.

### Inter Font Migration

Swapped Work Sans for Inter across four configuration layers:
- `src/app/layout.tsx` — `next/font/google` loader
- `tailwind.config.ts` — font family key
- `src/theme/theme.ts` — MUI theme typography
- `src/app/globals.css` — body font-family fallback

Updated `font-work-sans` → `font-inter` Tailwind class in 5 page/component files (DocsLayout, TutorialsPageClient, changelog pages, blog page).

### Inter Size Compensation

Inter has a ~7% taller x-height than Work Sans. At identical font-size declarations, everything appeared visually larger. Compensated with:
- **MUI theme**: Added `MuiButton.textTransform: 'none'` (normal-case buttons), changed `letterSpacing` from `"0 !important"` to `"-0.011em"` (Inter's recommended tight tracking)
- **Header nav**: Reduced all nav titles from `text-base` (16px) to `text-sm` (14px) in ProductMenu.tsx and header.tsx
- **Header auth buttons**: Added `!text-sm` overrides for consistent 14px across Discord, Sign in, and Sign up

### Dropdown Menu Polish

- **Fixed label/subLabel data bug**: Moved display text from `subLabel` to `label` in `menuExplorer`, `menuByUseCases`, `menuBySize`, and `menuByRole` — items now render in white instead of gray.
- **Enriched Resources dropdown**: Added MUI icons (MenuBook, School, Article, NewReleases, Explore, PlayCircle) and brief descriptions to all 6 items. Widened dropdown from `min-w-[180px]` to `w-[300px]`.
- **Improved Runner icon**: Replaced `PlayCircleOutline` (media play button) with `SyncAlt` (bidirectional arrows representing orchestration).
- **Product dropdown descriptions**: Changed from `font-medium` to `font-normal` for clearer title/description hierarchy.

### Docs Tracking

Added `tracking-tight` to the docs content wrapper in `MDXRenderer.tsx`. The MUI theme's letter-spacing only affects MUI Typography components, not the plain HTML tags used by `react-markdown`. This makes docs text feel as crisp as the rest of the site without reducing the 16px body text size (which is the industry standard for documentation).

## Implementation Details

### Files Changed (14 files, net -6 lines)

| File | Change |
|------|--------|
| `src/components/layout/header/header.tsx` | Auth buttons, menu data arrays, icon imports, dropdown width |
| `src/components/layout/header/ProductMenu.tsx` | Nav title size, description weight |
| `src/theme/theme.ts` | MuiButton textTransform, letterSpacing |
| `src/app/layout.tsx` | Work_Sans → Inter |
| `tailwind.config.ts` | font-work-sans → font-inter |
| `src/app/globals.css` | font-family fallback |
| `public/images/discord.svg` | fill → currentColor |
| `src/components/landing-page/index.ts` | Removed JoinBetaBtn export |
| `src/lib/MDXRenderer.tsx` | tracking-tight on content wrapper |
| 5 page files | font-work-sans → font-inter class |

### Design Decisions

- **Body text stays at 16px**: Inter at 16px is the universal standard for documentation (Stripe, Vercel, Tailwind, MDN). The "looks bigger" perception comes from Inter's taller x-height vs Work Sans, not from an actual sizing problem. Tighter letter-spacing addresses the perceived looseness.
- **Header-specific size overrides, not global Btn changes**: The `Btn` component uses `md:text-base` globally. Changing it would cascade to CTAs, pricing, and feature pages. Header buttons get `!text-sm` overrides instead.
- **Solutions items don't get icons**: They're navigational categories under clear section headers. Icons would add visual noise without information. GitHub and Vercel both use text-only category menus.

## Benefits

- **Reduced signup friction**: One click to console instead of two (button → modal → link)
- **Consistent monochrome identity**: Discord icon, button casing, and color palette all align
- **Compact, modern typography**: Inter at proper sizes matches cursor.com/Vercel caliber
- **Dropdown menus with proper contrast**: Solutions and Resources items visible in white instead of faded gray
- **Resources dropdown at parity with Product**: Icons + descriptions + proper width

## Impact

- Every page on the site is affected by the font change (Inter replaces Work Sans globally)
- The header is visible on all 122 static pages — auth button and nav changes are site-wide
- Documentation pages benefit from tighter tracking
- No breaking changes to URLs, auth flows, or external integrations
- All auth links still point to `planton.ai`

## Related Work

- Follows the monochrome rebrand (`2026-03-25-144804-black-and-white-theme-redesign`)
- Builds on the soften-white pass (`2026-03-25-151117-soften-white-and-remove-compliance-badges`)
- Continues the cursor.com-inspired aesthetic direction from the rebrand series

---

**Status**: ✅ Live (pending deploy)
**Timeline**: Single session

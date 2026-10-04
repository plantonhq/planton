# Soften White Tones & Remove Premature Compliance Badges

**Date**: March 25, 2026
**Type**: Design
**Components**: Design System, Landing Page, Navigation, Pricing Page, Agents Page

## Summary

Dialed down the white color intensity across the entire site from pure `#ffffff` to a softer `#ededed` to match cursor.com's subtler aesthetic. Also removed the premature SOC 2 and GDPR compliance badges from the SecurityTrustBar below the hero.

## Problem Statement / Motivation

After the black-and-white theme redesign (changelog: `2026-03-25-144804`), a side-by-side comparison of planton.ai and cursor.com revealed that Planton's white was noticeably "shinier" — pure `#ffffff` text on a `#0a0a0a` background creates harsh contrast that feels less refined than cursor.com's softer off-white approach.

Separately, the SecurityTrustBar was displaying SOC 2 Type I and GDPR Compliant badges with "Q1 2026" timelines — compliance certifications that are not yet achieved.

### Pain Points

- Pure `#ffffff` text on dark backgrounds is visually harsh compared to industry peers
- The brightness mismatch was noticeable in direct comparison with cursor.com
- SOC 2 and GDPR badges implied compliance milestones that haven't been reached
- Showing "Q1 2026" timelines on compliance badges was premature

## Solution / What's New

### White Color Override

Overrode Tailwind's built-in `white` color from `#ffffff` to `#ededed` (93% brightness). This single config change cascades through every `text-white`, `border-white`, and `bg-white/opacity` usage site-wide — no need to touch individual section files.

CTA buttons (Start Free Trial, Join Beta, Get Started, Contact Sales, Most Popular banner) were pinned to `bg-[#fff]` to keep their backgrounds pure white, since these are action elements that benefit from maximum contrast.

### Compliance Badge Cleanup

Removed the two unavailable compliance badges from the `securityBadges` array in SecurityTrustBar:
- ~~SOC 2 Type I (Q1 2026)~~
- ~~GDPR Compliant (Q1 2026)~~

Kept the two available badges: Zero-Trust Architecture and Open Source Audit.

## Implementation Details

### Tailwind Config (`tailwind.config.ts`)

Added `white: '#ededed'` to `theme.extend.colors`, which overrides Tailwind's default `white: '#fff'`. Also updated the `primary` and `secondary` color scales to use `#ededed` at steps 50 and 100 (previously `#ffffff`).

### Design Tokens (`shared.tsx`)

Updated `textPrimary` from `#ffffff` to `#ededed`. Changed PrimaryButton from `bg-white` to `bg-[#fff]` to keep the CTA button pure white.

### MUI Theme (`src/theme/colors.ts`)

Updated primary scale steps 50 and 100 from `#ffffff` to `#ededed`, which feeds into MUI's `palette.primary.main`.

### Button Background Fixes (10 instances across 8 files)

Every solid `bg-white` on a CTA button or badge was changed to `bg-[#fff]` to bypass the Tailwind override:
- `header.tsx` — Join Beta button
- `agents/hero-section.tsx` — hero CTA + hover state
- `agents/cta-section.tsx` — CTA button
- `pricing-calculator/calculation.tsx` — Get Started + Contact Sales buttons
- `pricing/plans.tsx` — Most Popular banner + plan CTA button
- `pricing/ready-to-try.tsx` — hover state
- `features/service-hub/page.tsx` — CTA button
- `v1/components.tsx` — Beta badge

### SecurityTrustBar

Removed two entries from the `securityBadges` array. No structural changes to the component.

## Benefits

- **Subtler aesthetic**: `#ededed` on `#0a0a0a` matches the refinement of cursor.com's dark theme
- **Site-wide cascade**: Single Tailwind override affects all `text-white`/`border-white` usages without touching 55+ files
- **Honest messaging**: Only displays compliance certifications that are actually achieved
- **CTA contrast preserved**: Buttons remain pure white for maximum click-through visibility

## Impact

- **12 files** modified
- **Zero build errors** — clean `next build` exit code 0
- All pages, pricing, agents, features, and navigation affected by the softer white
- SecurityTrustBar now shows 2 badges instead of 4
- Tour, Demo, Invest, Meets, ACME sections unaffected (out of scope)

## Related Work

- Follows the monochrome theme redesign (changelog: `2026-03-25-144804`)
- Part of the ongoing cursor.com-inspired visual refinement

---

**Status**: Live
**Timeline**: Single session

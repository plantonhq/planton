# Unified Domain UX Polish

**Date**: March 29, 2026
**Type**: Enhancement
**Components**: Navigation, Landing Page, UI Components, Layout

## Summary

Five UX fixes addressing leftovers from the unified domain migration (`planton.ai` serving both the marketing site and console app). Added a query-parameter escape hatch so logged-in users can view the landing page without incognito, replaced the deprecated plantoncloud mobile logo with the P icon, fixed the mobile hamburger menu not closing on navigation, and converted all CTA signup links from absolute URLs with `target="_blank"` to relative same-tab navigation.

## Problem Statement / Motivation

After consolidating `planton.ai` and `console.planton.ai` onto a single domain, several two-domain-era artifacts remained in the marketing site:

### Pain Points

- Logged-in users visiting `planton.ai/` were always redirected to `/dashboard` with no way to view the landing page without opening an incognito tab
- Clicking the header P logo while logged in triggered the same dashboard redirect
- The mobile hamburger drawer displayed a deprecated "plantoncloud" logo instead of the current P icon
- The mobile hamburger drawer stayed open after selecting a menu item, requiring a manual close
- CTA buttons ("Start Free Trial") used absolute `https://planton.ai/signup` URLs with `target="_blank"`, opening signup in a new tab -- a leftover from when signup lived on a separate domain

## Solution / What's New

### 1. Query Parameter Redirect Bypass

The inline `<head>` redirect script now checks `!window.location.search` before redirecting. Any query string (e.g., `/?preview`) suppresses the redirect and shows the landing page.

### 2. Session-Aware Logo Link

The `HeaderLogo` component now uses the `useLoggedIn` hook to link to `/?preview` when logged in and `/` when logged out. This applies to both desktop and mobile headers.

### 3. Mobile P Logo

Replaced the deprecated `<img src="/_site/images/header-logo-mobile.svg" />` with the `HeaderLogo` component, giving mobile the same P icon as desktop. Deleted the deprecated SVG file from the repo.

### 4. Mobile Drawer Auto-Close

Added an `onClick` handler on the drawer's content `Stack` that closes the drawer when any link (`<a>` element) is clicked. Accordion expand/collapse still works because the handler checks `(e.target as HTMLElement).closest('a')` -- only link clicks trigger the close.

### 5. Relative CTA Links

Converted all `https://planton.ai/signup` references to `/signup` and removed `target="_blank"`. Signup now navigates in the same tab as expected on a unified domain.

## Implementation Details

**Files modified:**

| File | Change |
|------|--------|
| `src/app/layout.tsx` | Added `&&!window.location.search` to redirect script |
| `src/components/layout/header/header.tsx` | `HeaderLogo` uses `useLoggedIn` for conditional href; mobile drawer uses `HeaderLogo` instead of `<img>`; added `onClick` handler for auto-close on link navigation |
| `src/components/landing-page/v3-2026-01-02-1000/HeroSection.tsx` | `href="/signup"` without `target="_blank"` |
| `src/components/landing-page/v3-2026-01-02-1000/ROICalculator.tsx` | `href="/signup"` without `target="_blank"` |
| `src/components/common/typography.tsx` | `GetStartedBtn` uses `href="/signup"`, removed `target="_self"` |
| `src/components/layout/footer.tsx` | Footer signup URL changed to `/signup` |

**Files deleted:**

| File | Reason |
|------|--------|
| `public/_site/images/header-logo-mobile.svg` | Deprecated plantoncloud logo, zero remaining references |

## Benefits

- Logged-in team members and developers can view the marketing site by visiting `planton.ai/?preview`
- Mobile experience matches desktop with consistent P logo branding
- Mobile navigation works as expected -- select a page, drawer closes
- CTA buttons navigate in-tab, consistent with single-domain architecture
- Removed deprecated asset that was no longer referenced

## Impact

- **Logged-in users**: Can now access the landing page via `/?preview` or by clicking the P logo
- **Mobile users**: Improved navigation flow and consistent branding
- **All visitors**: CTA buttons no longer open unnecessary new tabs
- **Codebase**: Removed deprecated plantoncloud SVG asset

## Related Work

- Unified domain migration: `planton/_projects/.completed/20260314.01.unified-domain-migration/`
- Session redirect and nav toggle (initial implementation): `_changelog/2026-03/2026-03-28-164803-session-redirect-and-nav-toggle.md`
- Context document updated: `planton/_team/agents/_knowledge-base/context/product/how-are-both-planton-website-and-console-served-on-one-domain.md`

---

**Status**: Production Ready (pending deploy to GitHub Pages)
**Timeline**: 30 minutes

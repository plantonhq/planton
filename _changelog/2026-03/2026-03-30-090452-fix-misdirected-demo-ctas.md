# Fix Misdirected Demo CTAs

**Date**: March 30, 2026
**Type**: Bug Fix
**Components**: Navigation, Tour/Walkthrough, Solutions Pages

## Summary

Fixed three misdirected "Book Demo" CTAs across the planton.ai site: the footer's "Book a Demo" link pointed to `/demo` (interactive tour) instead of `/book-demo` (lead capture form), the TourPage's scheduling buttons opened a placeholder Calendly URL, and two CTA buttons on the developers solutions page were completely inert with no navigation wired.

## Problem Statement / Motivation

After the `/book-demo` page and site-wide link migration (T02/T03) were completed, a user reported being taken to the interactive product tour (`/demo`) after clicking a "Book Demo" button. Investigation revealed three CTAs that were missed or pre-dated the migration.

### Pain Points

- Footer "Book a Demo" link visible on every page sent users to the wrong destination (`/demo` interactive tour instead of `/book-demo` lead form)
- TourPage "Book Demo" and "Schedule demo" buttons opened `https://calendly.com/your-demo-link` -- a placeholder URL that was never replaced
- Developers solutions hero had two dead buttons ("Start Building Today" and "Request a Demo") with no `href` or `onClick`

## Solution / What's New

### Footer Link Fix

Changed the footer "Book a Demo" URL from `/demo` to `/book-demo` in the `groups` data array.

### TourPage Navigation Fix

Replaced the `handleScheduleDemo` handler that opened a Calendly placeholder with `router.push('/book-demo')`. Added `useRouter` from `next/navigation`. This handler is used by two buttons:
- "Book Demo" on the tour closing screen (after tour completion)
- "Schedule demo" on the mid-tour prompt

Both now navigate to `/book-demo` in the same tab, which is appropriate since the handler already resets all tour state.

### Dead Buttons Fix

Wired the two inert MUI `Button` components in `BuildFasterDeploySmarter` using the same `LinkComponent={Link}` + `href` pattern established by the sibling `ReadyToElevate` component in the same directory:
- "Start Building Today" now navigates to `/signup`
- "Request a Demo" now navigates to `/book-demo`

## Implementation Details

| File | Change |
|------|--------|
| `src/components/layout/footer.tsx` | `url: '/demo'` → `url: '/book-demo'` |
| `src/components/tour/TourPage.tsx` | Added `useRouter` import, replaced `window.open('https://calendly.com/your-demo-link', '_blank')` with `router.push('/book-demo')` |
| `src/components/solutions/by-role/developers/build-faster-deploy-smarter.tsx` | Added `Link` to MUI imports, added `LinkComponent={Link}` and `href` props to both buttons |

## Benefits

- Every "Book a Demo" CTA across the site now correctly routes to the `/book-demo` lead capture page
- The tour's scheduling flow uses the same book-demo experience as the rest of the site (consistent UX)
- Two previously dead buttons are now functional, reducing user friction on the developers solutions page

## Impact

- **Footer fix**: Affects every page on the site -- this was the most visible bug
- **Tour fix**: Affects users who complete or pause the interactive product tour
- **Dead buttons fix**: Affects the `/solutions/by-role/developers` page

## Related Work

- T02/T03: `/book-demo` page build and site-wide link migration (completed 2026-03-29)
- Changelog: `2026-03-29-181206-book-demo-page-and-link-migration.md`

---

**Status**: ✅ Live (pending commit and deploy)

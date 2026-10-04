# Disable Auto-Redirect for Local Development

**Date**: March 29, 2026
**Type**: Fix
**Components**: Navigation, Layout

## Summary

Added a hostname guard to the session-aware redirect script so it only fires on `planton.ai`, never on localhost or other dev hostnames. Also removed the `/?preview` workaround from the header logo — the logo now always links to `/`.

## Problem Statement / Motivation

After the unified domain migration (2026-03-28), an inline `<head>` script was added to redirect logged-in users from `/` to `/dashboard`. This works correctly on production where the Origin Rule routes `/dashboard` to the K8s console app. However, when a developer runs both the console app and the marketing site on localhost, the `planton_logged_in` cookie (set by the console's NextAuth handler) is shared across ports on the same hostname. The redirect fires on the marketing site's root and sends the developer to `/dashboard`, which doesn't exist on the marketing dev server.

### Pain Points

- Marketing site homepage unreachable during local development when also running the console app
- The `/?preview` query parameter on the logo was a workaround, not a proper fix
- The workaround leaked into production UX (logged-in users clicking the logo landed on `/?preview` instead of `/`)

## Solution / What's New

### Hostname Guard on Redirect Script

Prepended `window.location.hostname==="planton.ai"` as the first condition in the inline redirect script. The JavaScript engine short-circuits the `&&` chain on any non-production hostname, so the redirect never fires on localhost, staging, or other dev environments.

This mirrors the Cloudflare Origin Rule hostname scoping fix from 2026-03-28, where `http.host eq "planton.ai"` was added to prevent the rule from matching subdomains. Same pattern, same reason — scope to the exact production hostname.

### Logo Always Links to `/`

Removed the conditional `href={loggedIn ? '/?preview' : '/'}` from `HeaderLogo`. The logo now always links to `/`. On production, logged-in users clicking the logo are redirected to `/dashboard` by the script (standard behavior matching GitHub, Linear, Vercel). On localhost, no redirect fires — devs see the marketing homepage as expected.

The `/?preview` bypass is preserved for explicit use: anyone who manually navigates to `/?preview` on production will see the marketing homepage because the `!window.location.search` condition in the redirect script returns false when a query string is present.

## Implementation Details

### Files Modified

**`src/app/layout.tsx`** — Added `window.location.hostname==="planton.ai"` to the redirect script condition chain.

**`src/components/layout/header/header.tsx`** — Simplified `HeaderLogo` to always link to `/`. Removed the `useLoggedIn()` call from the component (the hook is still used by `DesktopAuthButtons` and `MobileAuthButtons` for the Dashboard vs Sign in/Sign up toggle).

## Benefits

- Local development of the marketing site works without interference from the session redirect
- Logo behavior is standard across platforms (always goes to `/`)
- No workaround query parameters leaking into production URLs
- The `/?preview` escape hatch still works for intentional use on production

## Impact

- **Developers**: Can run the marketing site and console app locally without redirect conflicts
- **Production users**: Logo click now goes to `/` (redirect to `/dashboard` if logged in) instead of `/?preview`
- **Existing behavior**: `useLoggedIn()` hook, cookie lifecycle, and auth button toggle unchanged

## Related Work

- Session redirect cookie: `_changelog/2026-03/2026-03-28-164803-session-redirect-and-nav-toggle.md`
- Origin Rule hostname scoping: `planton/_changelog/2026-03/2026-03-28-162434-origin-rule-hostname-scoping-fix.md`
- Unified domain migration: `planton/_projects/.completed/20260314.01.unified-domain-migration/`

---

**Status**: Live (pending deploy)

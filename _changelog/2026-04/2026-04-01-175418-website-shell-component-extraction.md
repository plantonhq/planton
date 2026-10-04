# Website Shell Component Extraction — MUI-Only Header, Footer, and Layout

**Date**: April 1, 2026
**Type**: Feature
**Components**: Navigation, Layout, Build System, Design System

## Summary

Extracted the planton.ai website shell (Header, Footer, Layout) from Tailwind-based local components into MUI-only components inside the `@plantonhq/website-shell` npm package. The website now consumes its own shell from the workspace package, establishing the foundation for the console app to render public-facing pages with the marketing website's visual identity.

## Problem Statement / Motivation

The planton.ai console app needs to render public-facing pages (Deployment Store, future forum, platform catalog) with the marketing website's visual identity for unauthenticated users. The website's header, footer, and layout were built with a mix of Tailwind CSS classes and MUI components — unusable by the console app which has no Tailwind.

### Pain Points

- Console app cannot import website shell components (Tailwind dependency)
- No npm package exists for the website shell
- Header/Footer components had circular import dependencies with marketing content components
- Footer links had stale URLs diverging from header
- Social media links were all placeholders pointing to `/`
- Two breakpoint systems (Tailwind `md:` = 768px vs MUI default `md` = 900px) would cause layout shifts

## Solution / What's New

### Shell Components (MUI-Only)

17 new files in `packages/website-shell/src/` covering the complete website chrome:

- **WebsiteShell** — Full layout wrapper: ThemeProvider + Header + content area + Footer
- **WebsiteHeader** — Fixed header with responsive desktop mega-menus and mobile hamburger drawer
- **WebsiteFooter** — Full footer with 5-column link groups, terms, and copyright
- **WebsiteLogo** — Inline SVG with next/link
- **MegaMenu/MegaMenuItem** — Dropdown navigation with sections
- **MenuAccordion** — Mobile drawer accordion sections
- **AuthButtons** — Cookie-based Sign in/Sign up vs Dashboard
- **DiscordButton** — Inline SVG icon (no svgr dependency)
- **ShellButton** — Base button primitives

### Architecture Decisions

1. **Tailwind-aligned breakpoints** added to `websiteTheme`: `{ xs: 0, sm: 640, md: 768, lg: 1024, xl: 1280 }` — ensures mobile/desktop switch at 768px matches original Tailwind behavior
2. **`'use client'` banner** in tsup config — Next.js App Router requires the directive at the top of the built bundle; tsup strips per-file directives during bundling
3. **Inline styles for critical colors** — nested ThemeProviders cause CSS specificity collisions; auth buttons and Discord button use `style` prop for colors that must override any theme
4. **Localhost returns `false` for auth** — `useLoggedIn` hook skips cookie check on localhost since the website dev server has no console backend
5. **Footer social links removed** — all four were placeholders with `/_site/` image paths that would not resolve in the console
6. **Footer link bugs fixed** — Agent Fleet and CLI now use header's canonical `/features/` URLs

### Build Pipeline

- `yarn build` in root now runs `yarn build:packages` first (builds website-shell before Next.js build)
- ESLint config updated to ignore `packages/*/dist/**` generated output
- GitHub Actions publish workflow created for GitHub Packages

## Implementation Details

### Files Modified

- `eslint.config.mjs` — ignore `packages/*/dist/**`
- `package.json` — added `build:packages` script, updated `build` to run it first
- `packages/website-shell/package.json` — added `next`, `@mui/icons-material` peer deps
- `packages/website-shell/tsup.config.ts` — added `'use client'` banner
- `packages/website-shell/src/theme/websiteTheme.ts` — Tailwind-aligned breakpoints
- `packages/website-shell/src/index.ts` — exports all shell components
- `src/app/(root)/layout.tsx` — switched to `WebsiteShell` from package

### Files Created

- 17 component/hook/data files in `packages/website-shell/src/`
- `.github/workflows/publish-website-shell.yml` — publish workflow

## Benefits

- **Console can import `WebsiteShell`** — single `<WebsiteShell>{children}</WebsiteShell>` API
- **Zero Tailwind dependency** — pure MUI `sx` props and inline styles
- **Navigation data centralized** — single source of truth in `data/navigation.ts`
- **Clean build pipeline** — `make clean-build` works from scratch
- **Visual parity** — header matches production at pixel level (logo size, font, colors, breakpoints)

## Impact

- Website now renders shell from the npm package (126 pages, zero errors)
- No visual regression on production pages (verified side-by-side)
- Foundation ready for Phase 2: console app dual-shell integration

## Related Work

- **T01**: Website shell package scaffold (workspace setup, MUI theme, build infrastructure)
- **Project**: `20260401.01.website-shell-public-pages` — Phase 1 complete
- **ADR**: Unified Domain via Cloudflare Origin Rule — the architectural context for why the shell needs to be shared

---

**Status**: ✅ Live (in workspace, not yet published to GitHub Packages)
**Timeline**: T03 completed in single session

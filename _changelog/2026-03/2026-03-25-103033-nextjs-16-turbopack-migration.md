# Next.js 16.2 + Turbopack Migration

**Date**: March 25, 2026
**Type**: Enhancement
**Components**: Build System, UI Components, Documentation

## Summary

Migrated the planton.ai site from Next.js 15.1.0 (webpack) to Next.js 16.2.1 (Turbopack). The build is now fully clean — zero ESLint errors, zero ESLint warnings, zero Turbopack warnings — with 122 static pages generated in under 5 seconds.

## Problem Statement / Motivation

The planton monorepo had already migrated to Next.js 16.2 but was forced to stay on webpack due to a Turbopack `transpilePackages` incompatibility (#85315). The planton.ai site has no workspace dependencies, so Turbopack could be adopted without that constraint. The goals were:

- Align Next.js versions across all Planton repositories
- Adopt Turbopack as the default bundler for faster dev and build cycles
- Eliminate all build warnings for a clean foundation

### Pain Points

- Next.js 15.1.0 was two major versions behind, accumulating drift from the monorepo
- Webpack bundler with `config.cache = false` workaround was slow and noisy
- `next lint` was deprecated in Next.js 16, breaking the existing lint pipeline
- ESLint config used the `FlatCompat` bridge which broke with `eslint-config-next@16`

## Solution / What's New

### Dependency Upgrades

| Package | Before | After |
|---------|--------|-------|
| `next` | 15.1.0 | 16.2.1 |
| `eslint-config-next` | 15.1.0 | 16.2.1 |
| `@next/third-parties` | `latest` | `^16.2.1` |

Removed `@eslint/eslintrc` (no longer needed after ESLint config rewrite).

### Turbopack Bundler Switch

Replaced the webpack callback in `next.config.ts` with Turbopack-native configuration:

```ts
// Before: webpack callback with SVG loader + cache hack
webpack(config) {
  config.module.rules.push({ test: /\.svg$/, use: ["@svgr/webpack"] });
  config.cache = false;
  return config;
},

// After: Turbopack rules block
turbopack: {
  rules: {
    '*.svg': { loaders: ['@svgr/webpack'], as: '*.js' },
  },
},
```

Also removed the `eslint: { ignoreDuringBuilds: true }` block — this config key was removed in Next.js 16.

### ESLint Config Rewrite

`eslint-config-next@16.2.1` ships native ESLint 9 flat config arrays. The old `FlatCompat` bridge from `@eslint/eslintrc` caused a `TypeError: Converting circular structure to JSON` because the React plugin creates circular references that the bridge's JSON validator couldn't handle.

**Fix**: Import the configs directly — no bridge needed:

```js
import coreWebVitals from "eslint-config-next/core-web-vitals";
import typescript from "eslint-config-next/typescript";
```

### React 19 Hooks Compliance

`eslint-config-next@16` introduces two new React hooks rules: `react-hooks/set-state-in-effect` and `react-hooks/refs`. These flagged 13 pre-existing code patterns across 10 kinds. Each was fixed with the appropriate React 19 pattern:

| Pattern | Fix Applied | Files |
|---------|-------------|-------|
| Static browser value in effect | `useSyncExternalStore` | SearchTrigger |
| Derived state in effect | `useMemo` | TableOfContents, TutorialsPageClient |
| Derived config from props | `useMemo` with override layers | InfraJobLogger |
| URL hash sync on mount | `requestAnimationFrame` callback | InvestorDeckV2, MeetsDeck, useExplainerState |
| State reset on prop change | `requestAnimationFrame` callback | DocsSidebar, MarkdownViewDialog, TourPage |
| DOM position via setState | Direct ref-based DOM manipulation | Tooltip |
| Ref read during render | Simplified conditional | Header (mobile) |
| Filter reset | Handler-level reset | TutorialsPageClient |

### Turbopack NFT Tracing Fixes

Two Turbopack build warnings were eliminated:

1. **Broad file pattern** (13,966 files matched): `fileSystem.ts` does recursive directory traversal for the docs system. Added `/* turbopackIgnore: true */` comments to the dynamic `path.join` calls inside `buildStructure`, `getMarkdownContent`, and `resolveDocFilePath` so Turbopack skips file pattern analysis on those paths.

2. **`next.config.ts` in NFT list**: The docs module's `DOCS_DIRECTORY` was imported from `@/lib/constants` — Turbopack couldn't statically determine the scope through the cross-module import. Fixed by inlining the constant in `fileSystem.ts` with the `turbopackIgnore` comment, and adding `outputFileTracingExcludes` to exclude `next.config.ts` from route-level traces.

## Implementation Details

**Files changed**: 15 modified + 1 new (`.nvmrc`)

Key changes by file:
- **`next.config.ts`**: Turbopack rules, remove webpack callback, remove eslint block, add `outputFileTracingExcludes`
- **`eslint.config.mjs`**: Native flat config imports, remove `FlatCompat` bridge
- **`package.json`**: Bump next/eslint-config-next, pin `@next/third-parties`, remove `@eslint/eslintrc`, change lint script to `eslint .`
- **`fileSystem.ts`**: Inline `DOCS_DIRECTORY` constant, add `turbopackIgnore` comments
- **10 kind files**: React 19 hooks compliance fixes
- **`.nvmrc`**: Pin Node 24.14.0

## Benefits

- **Faster builds**: Turbopack compiles in ~5s (was ~7s+ with webpack)
- **Clean build output**: Zero warnings of any kind — lint, TypeScript, Turbopack
- **Modern React patterns**: All hooks usage now follows React 19 best practices
- **Version alignment**: Next.js 16.2.1 across planton.ai and the monorepo
- **Node version pinned**: `.nvmrc` ensures consistent environments

## Impact

- All 122 static pages build and generate correctly
- Pagefind search indexing continues to work (57 pages indexed)
- No user-facing behavior changes — this is purely a build toolchain upgrade
- The `constants.tsx` module still exports `DOCS_DIRECTORY` for other consumers (blog, changelog, tutorials)

## Related Work

- Planton monorepo Next.js 16 migration: `_projects/.completed/20260313.01.nextjs-build-optimization`
- The monorepo was forced to stay on webpack (`--webpack` flag) due to Turbopack issue #85315 with `transpilePackages`. The planton.ai site has no workspace dependencies, so Turbopack works cleanly here.

---

**Status**: Live
**Timeline**: Single session

# Infra Catalog Feature Page and Platform Stats Centralization

**Date**: April 3, 2026
**Type**: Feature
**Components**: Features Pages, Navigation, Design System, Landing Page, Solutions Pages

## Summary

Added a dedicated Infra Catalog feature page at `/features/infra-catalog` with hero, capabilities, and CTA sections following the established feature page pattern. Centralized all hardcoded platform statistics (deployment module count, cloud provider count) into a single source-of-truth constants file, fixing inconsistent values scattered across 16 files. Updated the website-shell navigation to include Infra Catalog between ServiceHub and Runner in both the Product mega-menu and footer.

## Problem Statement / Motivation

Infra Catalog — the browsable catalog of deployment modules and Infra Charts — had no presence on the marketing website. Users who discovered the catalog through the console had no feature page explaining what it is, why it exists, or how it fits into the Planton product suite.

### Pain Points

- No marketing/feature page for Infra Catalog — zero SEO surface for a flagship public feature
- Platform statistics were hardcoded as string literals across 16 files with three conflicting values: `360+` (11 files), `350+` (1 file), and `500+` (1 file) for the same deployment module count
- Cloud provider count was hardcoded as `17` in two files — actual count is `10+`
- Updating a count required finding and editing every occurrence manually
- Infra Catalog was missing from the header Product mega-menu and footer navigation

## Solution / What's New

### Infra Catalog Feature Page

Created `/features/infra-catalog` following the established 3-section pattern:

- **Hero**: Badge, headline ("Find it. Deploy it. Done."), pain/solution copy, "Explore Catalog" primary CTA linking to `/infra-catalog`, terminal animation showing the catalog-to-deploy flow, module type and provider chips
- **Capabilities**: MetricsStrip (350+ modules, 10+ providers, 50+ charts, <5 min deploy), Two Module Types showcase (Catalog Kinds vs Infra Charts), BentoGrid (filter by provider, presets, open source, no signup required), YAML deep dive with terminal, deploy flow steps
- **CTA**: RelatedModules (infra-hub, runner, open-source), "Start exploring" card with Explore Catalog and Read the Docs buttons

### Platform Stats Constants

Created `src/data/platform-stats.ts`:

```typescript
export const PLATFORM_STATS = {
  DEPLOYMENT_MODULE_COUNT: '350+',
  CLOUD_PROVIDER_COUNT: '10+',
  INFRA_CHART_COUNT: '50+',
} as const;
```

### Navigation Update

Updated `@planton/website-shell` navigation data:
- Added Infra Catalog to `menuProduct` between ServiceHub and Runner with sublabel "Browse and deploy infrastructure modules"
- Added Infra Catalog to footer Product column
- Bumped package version to `0.1.2`

## Implementation Details

### Files Created

| File | Purpose |
|------|---------|
| `src/data/platform-stats.ts` | Single source of truth for platform-wide statistics |
| `src/app/(root)/features/infra-catalog/page.tsx` | Route page with metadata |
| `src/components/product/infra-catalog/index.tsx` | Barrel export |
| `src/components/product/infra-catalog/hero.tsx` | Hero section with terminal animation |
| `src/components/product/infra-catalog/capabilities.tsx` | MetricsStrip, module types, BentoGrid, YAML deep dive, deploy flow |
| `src/components/product/infra-catalog/cta.tsx` | RelatedModules and CTA card |

### Files Modified (Stats Centralization — 16 files)

All updated from hardcoded `360+` / `500+` / `17` to `PLATFORM_STATS.*` constants:

- 4 landing page components (InfraHub, HowItWorks, ComparisonTable, OpenSourceFoundation)
- 5 product/feature components (infra-hub hero + capabilities, open-source capabilities, overview modules-grid, shared RelatedModules)
- 3 solutions pages (startups, growing-teams, engineering-leaders)
- 1 book-demo page
- 1 demo component (OpenMcfIntro)
- 1 agents page (technology-section)
- 1 route metadata (infra-hub page.tsx)

### Files Modified (Navigation)

- `packages/website-shell/src/data/navigation.ts` — menuProduct and footerGroups
- `packages/website-shell/package.json` — version 0.1.1 → 0.1.2
- `src/app/(root)/features/layout.tsx` — features subnav paths

## Benefits

- **Single source of truth**: Next stat update is one line in one file instead of hunting through 16 files
- **Correct numbers**: All pages now show `350+` modules and `10+` providers (was `360+` and `17`)
- **SEO surface**: Infra Catalog has a dedicated, indexable feature page
- **Navigation presence**: Infra Catalog is discoverable from the header and footer on every page
- **Conversion funnel**: "Explore Catalog" CTA on the feature page drives visitors to the live catalog at `/infra-catalog`

## Impact

- New page at `/features/infra-catalog` (static export count: 135 → 136)
- Infra Catalog appears in Product mega-menu and footer on all pages (after website-shell publish)
- Features subnav now shows: Product | InfraHub | ServiceHub | **Infra Catalog** | Runner | Security | Agent Fleet | CLI | Open Source
- All marketing pages reflect corrected platform statistics

## Related Work

- **Console route migration**: `/platform/infra-catalog` → `/infra-catalog` (planton monorepo, same session)
- **Website-shell-public-pages project**: This feature page is the marketing complement to the public Infra Catalog delivered in that project
- **Previous**: Feature pages structural overhaul (2026-03-25) established the Hero + Capabilities + CTA pattern

---

**Status**: Live (pending deployment)
**Timeline**: Completed in single session

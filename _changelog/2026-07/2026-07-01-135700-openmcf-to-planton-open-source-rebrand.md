# OpenMCF → Planton open source Rebrand (Site-Wide)

**Date**: July 1, 2026
**Type**: Refactor / Content
**Components**: Navigation, Demo, Landing, Product, Solutions, Docs, Tutorials, Legal, Invest/Meets

## Summary

Retired the "OpenMCF" brand across the entire planton.ai website. The open-source
foundation is now consistently called **Planton open source** (bare "Planton"
continues to mean the commercial Platform). External identifiers moved to the
unified brand: repo `github.com/plantonhq/planton`, domain `planton.dev`, CLI
`planton` (installed via `brew install plantonhq/tap/planton`), and the Infra
Charts link now points at the merged `charts/` folder in the `planton` repo.

## What Changed

### Brand + identifiers
- Prose "OpenMCF" / "Open Multi-Cloud Framework" → "Planton open source" (lowercase
  "open source" — a descriptor, not a sub-brand).
- GitHub links `plantonhq/openmcf` → `plantonhq/planton`; `openmcf.org` → `planton.dev`.
- CLI examples `openmcf …` → `planton …`; Homebrew tap → `plantonhq/tap/planton`.
- Footer "Infra Charts" link → `github.com/plantonhq/planton/tree/main/charts`.

### apiVersion (two-domain split preserved)
- Open-source cloud-provider kinds `<provider>.openmcf.org/v1` → `<provider>.planton.dev/v1`.
- Platform hub kinds (`infra-hub`, `service-hub`, `connect`) stay on `*.planton.ai/v1`
  — those are the Platform's own kinds and were intentionally left untouched.

### Live navigation + chrome
- `packages/website-shell/src/data/navigation.ts` — the live nav/footer data
  (Open Source subLabel, footer link label, Infra Charts URL).

### Demo journey
- Renamed the 4 `OpenMcf*` demo components to `Planton*`; updated `journeys.ts`
  (journey id/screens/description) and `DemoPage.tsx` (imports, `DemoScreen` union
  `planton-*`, flows, switch); updated demo JSON apiVersions, forms, hooks, and
  concept docs.

### Feature / landing / product / solutions
- Open Source feature page (hero/capabilities/cta) and the live v3 landing,
  product, and solutions pages rebranded.

### Docs, tutorials, legal, sitemap
- Renamed the docs page `infrastructure/openmcf.md` → `infrastructure/open-source.md`
  (slug `/docs/infrastructure/open-source`, matching `/features/open-source`).
- Added a client-side compatibility redirect at `/docs/infrastructure/openmcf` so
  the old URL does not 404 (static-export-safe, uses the existing `RedirectPage`).
- Updated infra + ci-cd docs cross-links, 5 tutorials, `content/legal/terms.md`
  (open-source license references), and `public/sitemap.xml`.

### Investor + meeting decks
- `src/components/invest/**` and `src/components/meets/**` rebranded (live routes).

## Dead-Code Removal (cleaner codebase)

Confirmed-unused duplicates were deleted after verifying zero importers:

- `src/components/layout/MainLayout.tsx` and the legacy `header.tsx` / `footer.tsx`
  (superseded by the `@planton/website-shell` chrome). The one still-used export,
  `HeaderLogo`, was relocated to `src/components/layout/header/HeaderLogo.tsx`.
- The dead `src/components/landing-page/v2-2025-12-31-0900/` folder (20 files;
  the live homepage renders v3).

This makes the "zero OpenMCF" result genuine rather than requiring dead-code
exclusions, and removes divergent copies of the nav/landing that could mislead
future contributors.

## Verification

- `yarn lint` — clean.
- `yarn build` (static export) — succeeded; TypeScript passed; 144 pages generated.
- Source audit: `openmcf` remains only in preserved dated history (`_changelog/`,
  `content/copywriting/_stage-area/`, `workspace/transcripts/`).
- Built-HTML audit: no `openmcf` in any exported page except the intentional
  redirect stub; `/docs/infrastructure/openmcf` redirects to `.../open-source`.

---

**Status**: ✅ Complete (build green; not yet deployed)

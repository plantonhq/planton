# Homepage Hero Unification and Streamlining: From Two Competing Heroes to One Focused Flow

**Date**: March 26, 2026
**Type**: Refactoring | Design
**Components**: Landing Page, Features Pages, UI Components, Navigation, Design System

## Summary

Unified the homepage and product overview heroes so the homepage carries the strongest positioning message ("One platform. Infra and apps. Any cloud.") paired with a multi-scenario animated terminal that demonstrates platform breadth — not just one module. Streamlined the homepage from 18 sections to 10 by merging overlapping content and removing sections that duplicate dedicated product pages. Transformed the `/features` product overview from a second landing page into a navigation hub. Added cross-module navigation to all 7 product pages.

## Problem Statement / Motivation

The website had two competing hero messages on two different pages:

- **Homepage** (`/`): "Deploy Infrastructure in Minutes, Not Weeks" — an animated terminal cycling through 4 variations of the same InfraHub `chart install` command. This undersold Planton as just a faster Terraform, missing ServiceHub, Agent Fleet, Security, and the "runs in your cloud" differentiator.
- **Product Overview** (`/features`): "One platform. Infra and apps. Any cloud." — correct messaging but visually dead: badge + headline + paragraph + two buttons. No animation, no proof, no visual substance.

Additionally, the homepage had 18 sections with significant overlap: two security sections (SecurityTrustBar and SecurityCompliance), three social proof sections (SocialProofBar, WallOfLove, CustomerStories), and homepage teaser sections for InfraHub, ServiceHub, and AgentFleet that duplicated their dedicated product pages.

### Pain Points

- Homepage hero positioned Planton as only an infrastructure deployment tool, missing 6 of 7 product modules
- Terminal animation showed 4 variants of the same operation (InfraHub chart install) — visually impressive but narrow
- `/features` page duplicated the homepage hero with a second pitch, CTAs, and "100 automation minutes free" footnote
- 18 homepage sections created scroll fatigue and diluted conversion focus
- SecurityTrustBar + SecurityCompliance covered the same topic with different depth
- WallOfLove + CustomerStories both showed testimonials/case studies as separate sections
- InfraHub, ServiceHub, and AgentFleet homepage sections were watered-down versions of their dedicated product pages
- Product pages had no cross-navigation — visitors reaching the bottom of one module had no path to adjacent modules
- Homepage sections had no scroll animations despite Framer Motion being available

## Solution / What's New

### Unified Homepage Hero

The homepage hero now uses the features page's messaging with the homepage's visual richness:

- **Headline**: "One platform. Infra and apps. Any cloud."
- **Subtitle**: "Infrastructure deployment, application CI/CD, and AI operations — all in your cloud. Open source foundation. Zero vendor lock-in."
- **Multi-scenario terminal** cycling through 3 scenarios spanning platform breadth:
  - **Infra**: `planton chart install aws-ecs --name api --env dev --values values.yaml`
  - **Services**: `git push origin main` triggering a full CI/CD pipeline
  - **CLI**: `planton apply -f service.yaml` for manifest-driven deployment
- **Clickable scenario tabs** in the terminal header (Infra / Services / CLI) for direct navigation between demos
- **Trust bar** (Minutes Not Weeks / Your Cloud / Open Source / Enterprise Security) integrated below the provider logos
- Provider logos and CTAs preserved

### Homepage Streamlined to 10 Sections

| # | Section | Source |
|---|---------|--------|
| 1 | HeroSection (rewritten) | New |
| 2 | SocialProofBar | Kept as-is |
| 3 | HowItWorks | Kept as-is |
| 4 | ComparisonTable | Kept as-is |
| 5 | SocialProof | Merged from WallOfLove + CustomerStories |
| 6 | Security | Merged from SecurityTrustBar + SecurityCompliance |
| 7 | ROICalculator | Kept as-is |
| 8 | PricingSimplified | Kept as-is |
| 9 | OpenSourceFoundation | Kept as-is |
| 10 | FinalCTA | Kept as-is |

**Removed**: ProblemSolution, InfraHub, ServiceHub, AgentFleet, BuiltByDevOps, OpenStandards (8 sections cut).

**Merged**: SocialProof combines the 2 customer story cards with 3 testimonial cards in one section. Security combines the thin trust strip pills with the 3 security model cards and IAM JSON example.

### Product Overview Transformed

`/features` is now a navigation hub, not a second landing page:

- **Compact page header** replaces the full hero — "Everything you need to deploy and operate infrastructure" with a one-line subtitle, no CTAs
- **Trust bar removed** from features page (now in homepage hero)
- **Journey FlowSteps** added between module grid and architecture diagram: Connect Cloud → Deploy Infra → Ship Code → Secure → Automate

### Cross-Module Navigation

New `RelatedModules` component added to all 7 product page CTAs:

| Page | Related Modules |
|------|----------------|
| InfraHub | Runner, Security, Open Source |
| ServiceHub | InfraHub, CLI, Runner |
| Runner | InfraHub, Security, CLI |
| Security | Runner, Open Source, InfraHub |
| Agent Fleet | ServiceHub, CLI, Security |
| CLI | InfraHub, ServiceHub, Open Source |
| Open Source | InfraHub, CLI, Runner |

### Scroll Animations

All homepage sections (except Hero and SocialProofBar) wrapped in `ScrollReveal` for fade-up entrance animations on scroll, matching the animation language already used on product pages.

## Implementation Details

### New Components

| Component | Path | Purpose |
|-----------|------|---------|
| `SocialProof` | `src/components/landing-page/v3-2026-01-02-1000/SocialProof.tsx` | Merged customer stories + testimonials |
| `Security` | `src/components/landing-page/v3-2026-01-02-1000/Security.tsx` | Merged security trust strip + compliance |
| `ProductJourney` | `src/components/product/overview/journey.tsx` | FlowSteps journey visualization |
| `RelatedModules` | `src/components/product/shared/RelatedModules.tsx` | Cross-module navigation strip |

### Key Design Decisions

- **Terminal tabs labeled "Infra / Services / CLI"** instead of "InfraHub / ServiceHub / CLI" — shorter labels that communicate function, not product module names, to first-time visitors who don't know the internal naming
- **Homepage doesn't include a product module grid** — that's the product overview's job. The homepage sells outcomes and proof; `/features` catalogs modules
- **Legacy component exports preserved** in the barrel file — old components (WallOfLove, CustomerStories, SecurityTrustBar, etc.) are still exported for the investor deck and other consumers that import them directly

## Benefits

- **Unified first impression**: Every visitor sees the full platform positioning immediately, not just an infrastructure speed pitch
- **Platform breadth demonstrated**: The terminal now shows infra, services, and CLI — three distinct product capabilities in one interactive element
- **Reduced scroll fatigue**: 10 focused sections vs. 18 with overlap. Every section earns its place with a distinct conversion purpose
- **Navigation web**: Visitors on any product page can discover adjacent modules through the Related Modules strip
- **Clean product hierarchy**: Homepage sells outcomes → `/features` catalogs modules → individual pages go deep
- **Consistent animation language**: Homepage now uses the same ScrollReveal animations as product pages

## Impact

### Pages Affected

- **1 homepage** rewritten (18 sections → 10)
- **1 product overview** transformed (full hero → compact header + journey)
- **7 product page CTAs** updated with RelatedModules
- **4 new components** created
- **Build passes** cleanly with zero errors

### Files Changed

~17 files across:
- `src/components/landing-page/v3-2026-01-02-1000/` — HeroSection rewrite, new SocialProof + Security, barrel update
- `src/app/(root)/page.tsx` — homepage section composition
- `src/app/(root)/features/page.tsx` — product overview composition
- `src/components/product/overview/` — compact hero, new journey, barrel update
- `src/components/product/shared/` — new RelatedModules, barrel update
- `src/components/product/*/cta.tsx` — 7 CTA files with cross-links

## Related Work

- Website copywriting overhaul (March 25, 2026) — updated content to reflect current product
- Feature pages structural overhaul (March 25, 2026) — added AnimatedTerminals, CodeTabs, BentoGrids, FlowSteps to product pages
- Website accuracy & quality overhaul (March 26, 2026) — fixed factual errors, removed placeholders, standardized CTAs

---

**Status**: ✅ Live
**Timeline**: Single session

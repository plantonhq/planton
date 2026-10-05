# Meets Deck Creation: ClearRoute

**Date**: August 12, 2026
**Type**: Feature
**Components**: Meets Micro-App, ClearRoute Presentation

## Summary

Created a 9-slide meeting deck at `/meets/clear-route` for ClearRoute, a global
platform engineering consultancy (London, founded 2022, SOC 2, clients including
Commonwealth Bank, Collinson, Howden and the UK National Energy System Operator).

The narrative positions Planton as an **AI-native self-service DevOps platform
that ClearRoute deploys for their own clients**, rather than a tool ClearRoute
buys for itself. It hands off to a live Infra Hub demo on the desktop app.

Also closed a search-indexing exposure across the whole micro-app — see below.

## Guest Information

- **Slug**: `clear-route`
- **Name**: ClearRoute
- **Type**: Company (platform engineering consultancy / prospective partner)
- **Meeting Date**: `2026-08-12-1400`
- **Location**: Virtual

## Slides Created

| # | Slide Name | Purpose |
|---|------------|---------|
| S01 | Cover | Title, positioning line, meeting details |
| S02 | We Know ClearRoute | QCE framework, Route to Live, client roster, India growth |
| S03 | The Bottleneck | Their own homepage quote plus three friction cards |
| S04 | What Is Planton | Infra Hub and Service Hub, plus the anti-abstraction stance |
| S05 | Why This Fits You | Five consultancy-specific business arguments |
| S06 | Runs In Their Kubernetes | One-command Helm install, secret-free connections, preview honesty |
| S07 | AI-Native | Determinism argument, 245 MCP tools, AI teammates |
| S08 | Demo Handoff | Frames the desktop app, scopes the demo to Infra Hub |
| S09 | Next Steps | Design-partner ask and Q&A |

Slides S03 and S07 are designated live cut points; dropping either leaves the
narrative intact for a 7-slide fast path.

## Framework Change: decks are no longer indexable

`public/robots.txt` allows all crawlers and the meets layout declared no robots
directive, so every deck — including the investor deck at `/meets/nirav`, which
carries fundraising terms — was fully crawlable and indexable.

Added `robots: { index: false, follow: false, nocache: true }` to
`src/app/(micro-apps)/meets/layout.tsx`. Placing it in the layout means all
existing and future decks inherit it. Verified in the build output: both
`out/meets/clear-route.html` and `out/meets/nirav.html` now emit
`<meta name="robots" content="noindex, nofollow, nocache"/>`.

Deliberately did **not** add `Disallow: /meets/` to `robots.txt`. A Disallow
blocks crawling, which prevents crawlers from ever reading the noindex
directive, leaving links that were shared by URL eligible for indexing.

## Content Accuracy Decisions

- **"600+ catalog kinds across 17 clouds"** rather than an exact count. Three
  internal sources disagree because they count different things: the API
  versioning ADR (2026-08-06) says 613 kinds, the monetization plan says 617,
  and a direct count of `kind_meta` annotations gives 676.
- **17 Infra Charts**, not the 49 claimed in the public `planton` README. The
  chart catalog was deliberately clean-slated and rebuilt as a curated set.
- **No "$150K DevOps hire" or "$50K saved" claims** — both are banned as
  unvalidated projections. Used the sanctioned framing instead: run production
  infrastructure without senior DevOps expertise or a larger ops org.
- **Self-hosted stated as active preview.** Overclaiming maturity to a
  consultancy that would deploy it at a bank was judged the single largest
  relationship risk, so the ask became a design-partner invitation.
- **Demo scoped to Infra Hub.** Service Hub does not run on the desktop
  instance today; S08 states this so the gap is never discovered live.
- **No pricing.** Packaging is mid-rework and every figure is an unapproved
  anchor.

## Files Created

| File | Purpose |
|------|---------|
| `guests/clear-route/config.ts` | Slide array, presenter notes, guest config |
| `guests/clear-route/slides/S01Cover.tsx` | Cover |
| `guests/clear-route/slides/S02WeKnowClearRoute.tsx` | Guest research |
| `guests/clear-route/slides/S03BottleneckIsRouteToLive.tsx` | Problem |
| `guests/clear-route/slides/S04WhatIsPlanton.tsx` | Product |
| `guests/clear-route/slides/S05WhyFitsClearRoute.tsx` | Business fit |
| `guests/clear-route/slides/S06RunsInClientKubernetes.tsx` | Deployment |
| `guests/clear-route/slides/S07AINativeDeterministic.tsx` | AI differentiator |
| `guests/clear-route/slides/S08DemoHandoff.tsx` | Demo transition |
| `guests/clear-route/slides/S09NextSteps.tsx` | Ask and Q&A |

## Files Modified

| File | Change |
|------|--------|
| `src/components/meets/guests/index.ts` | Registry entry `clear-route/2026-08-12-1400` |
| `src/app/(micro-apps)/meets/layout.tsx` | Robots noindex directive |

## Known Wart (not addressed)

`src/app/(micro-apps)/meets/page.tsx` still hardcodes `redirect('/meets/sep')`,
so the bare `/meets` index points at a January meeting. The correct fix is to
redirect to the most recent meeting by date; repointing it at ClearRoute would
only relocate the staleness.

---

**Status**: Live
**URL**: https://planton.ai/meets/clear-route

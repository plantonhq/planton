# Rebrand "Planton" to "Planton" + Google OAuth Verification Fix

**Date**: March 25, 2026
**Type**: Enhancement
**Components**: Landing Page, Solutions Pages, Features Pages, Demo, CLI, Hackathon, Agents, Pricing, Legal Pages, SEO/Performance

## Summary

Site-wide rebrand from "Planton" to "Planton" across ~60 rendered files, plus root metadata and hero section updates to fix two Google OAuth consent screen verification failures. The legal entity name "Planton Cloud, Inc." / "Planton Cloud Inc." is preserved in the footer copyright, privacy policy, and terms of service.

## Problem Statement / Motivation

Two issues blocking Google OAuth consent screen approval:

1. **"Your homepage does not explain the purpose of your app"** — the root `<title>` and `<meta description>` were both just "Planton" with no description of what the platform does.
2. **"The app name 'Planton' configured for your OAuth consent screen does not match the app name on your homepage"** — "Planton" didn't appear in the homepage H1 or meta title in a way Google's verification bot could match.

Additionally, the product branding had shifted from "Planton" to "Planton" but the website still used the old name in ~100+ places across solutions, features, demo, CLI, hackathon, agents, pricing, and other pages.

### Pain Points

- Google OAuth consent screen verification was rejected
- Inconsistent branding across the site ("Planton" in some places, "Planton" in others)
- Meta title and description provided zero information about the platform

## Solution / What's New

### Root Metadata (Google OAuth Fix #1)

Updated `src/app/layout.tsx` metadata from the uninformative:

```typescript
title: 'Planton',
description: 'Planton',
```

To a descriptive:

```typescript
title: 'Planton — Deploy Production Infrastructure in Minutes, Not Weeks',
description: 'Planton is a DevOps automation platform...',
```

### Hero Section (Google OAuth Fix #2)

Changed the homepage H1 from "What if DevOps Didn't Block Your Developers?" (no mention of Planton) to "Planton / Deploy Infrastructure in Minutes, Not Weeks." with the gradient styling on "Planton". Subheadline updated to explicitly describe the platform capabilities.

### Site-Wide Rebrand

Replaced "Planton" → "Planton" across all rendered pages:

- Solutions pages (by-role, by-size, by-use-case, all)
- Features pages (all features, service-hub, auditable-intelligence, iac-workflows)
- Landing page components (hero, v1 legacy components)
- Demo pages and concepts
- CLI components
- Hackathon pages (MobileVibe 2025)
- Agents page and technology section
- Pricing page
- Kubernetes dashboard components
- Invest/investor pages and slides
- Legal micro-app layouts
- Tour layout
- Blog post (getting-started-with-planton.md)
- Privacy/terms meta descriptions

### Preserved (Legal Entity Name)

- Footer copyright: "Planton Cloud Inc."
- Privacy policy body: "Planton Cloud, Inc."
- Terms of service body: "Planton Cloud, Inc."
- Historical changelogs and internal copywriting staging

## Benefits

- Unblocks Google OAuth consent screen verification
- Consistent "Planton" branding across the entire site
- Homepage `<title>` and `<meta description>` now provide meaningful SEO value
- "Planton" appears as the first word in the H1 and page title, making the app name match unambiguous for Google

## Impact

- **Google OAuth**: Both verification issues should be resolved — homepage explains purpose and app name matches
- **SEO**: Descriptive title and meta description replace the bare "Planton" string
- **Branding**: ~100+ occurrences updated across ~60 files
- **Build**: Verified — `yarn build` passes, correct `<title>` and `<meta description>` confirmed in generated HTML

---

**Status**: Live
**Timeline**: Single session

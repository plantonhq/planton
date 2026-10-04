# Investor Pitch Deck - Y Combinator Style Seed Stage Deck

**Date**: December 17, 2025
**Type**: Feature
**Components**: Interactive Demo, UI Components, Navigation, Micro-Apps

## Summary

Created a complete Y Combinator-style investor pitch deck available at `planton.ai/invest`. The deck features 10 interactive slides with keyboard navigation, smooth animations, and a toggleable SAFE investment explainer. Also restructured the app folder organization from `(tour-demo)` to `(micro-apps)` for clarity.

## Problem Statement / Motivation

Planton needed a professional, shareable investor pitch deck for seed-stage fundraising meetings. The requirements were specific:

### Pain Points

- No existing investor-facing presentation on the website
- Need for a concise, YC-style deck that tells a compelling story
- Complex investment terms (SAFE notes) need simple explanations for investors unfamiliar with startup financing
- The deck needed to be web-native, not a static PDF, allowing for easy updates and sharing

## Solution / What's New

Built a complete 10-slide interactive pitch deck using the same component architecture as the existing demo at `planton.ai/demo`.

### Slides Created

| # | Slide | Purpose |
|---|-------|---------|
| 1 | **Cover** | Planton branding, "DevOps in a Box", 3+ years building, $500K+ self-funded |
| 2 | **Problem** | Cloud infrastructure pain points: weeks to deploy, $150K+ DevOps cost, multi-cloud chaos |
| 3 | **Solution** | InfraHub + ServiceHub with "replaces X" messaging (Terraform Enterprise, GitHub Actions, etc.) |
| 4 | **Product** | Three-step workflow: Connect → Configure → Deploy |
| 5 | **Traction** | Key metrics: 3+ years, $500K+ invested, $800 MRR, 5 team members |
| 6 | **Market** | $150B+ cloud spend, $20B+ DevOps tools market, "Why Now" timing |
| 7 | **Roadmap** | Four phases: Now (Live), Q1 2026 (Observability), 2026 (AI Agent Fleet), Long Term (Cloud Provider) |
| 8 | **Team** | Core team with roles and "Why Us" strengths |
| 9 | **Ask** | $500K SAFE with use of funds breakdown and interactive SAFE explainer |
| 10 | **Close** | Call to action with contact info and links |

### Interactive SAFE Explainer

The Ask slide includes a "How does a SAFE work?" toggle button that reveals:

- **For Planton**: Receive funds now, no valuation negotiation, focus on building
- **For Investor**: Invest early at discount, converts at Series A, protected by cap
- **Math Example**: Shows how $500K converts to ~7% equity at $7M cap, potential 14x return

### Key Messaging

**Solution Slide Positioning:**
- InfraHub → "Replaces Terraform Enterprise / Pulumi Cloud"
- ServiceHub → "Replaces GitHub Actions / Jenkins / GitLab Pipelines"
- Tagline: "Save $ and eliminate integration chaos"

**Cloud Provider Vision:**
- Clear aspiration: "Join AWS, GCP, Azure, DigitalOcean, Civo"
- Not confusing "software cloud provider" language

**Roadmap with Observability:**
- Added Q1 2026 Observability milestone based on customer demand
- Design complete, development underway

## Implementation Details

### File Structure

```
src/app/(micro-apps)/invest/
├── invest.css          # Custom slide styling
├── layout.tsx          # Metadata and Inter font
└── page.tsx            # Entry point

src/components/invest/
├── InvestorDeckPage.tsx    # Main orchestrator with navigation
└── slides/
    ├── SlideAsk.tsx        # Funding ask with SAFE explainer
    ├── SlideClose.tsx      # Contact and links
    ├── SlideCover.tsx      # Opening slide
    ├── SlideMarket.tsx     # Market opportunity
    ├── SlideProblem.tsx    # Pain points
    ├── SlideProduct.tsx    # How it works
    ├── SlideRoadmap.tsx    # Four-phase roadmap
    ├── SlideSolution.tsx   # InfraHub + ServiceHub
    ├── SlideTeam.tsx       # Team members
    └── SlideTraction.tsx   # Metrics and milestones
```

### Navigation Features

- **Keyboard**: Arrow keys (←/→), Home key returns to cover
- **Click**: Next/Previous buttons
- **Animations**: Framer Motion for smooth slide transitions
- **State**: React useState for current slide index

### Folder Restructuring

Renamed route group for clarity:

```
src/app/(tour-demo)/  →  src/app/(micro-apps)/
├── demo/     → planton.ai/demo
├── tour/     → planton.ai/tour
└── invest/   → planton.ai/invest
```

Also renamed component folder:
```
src/components/investor-deck/  →  src/components/invest/
```

### Technologies Used

- **React/Next.js**: App router with route groups
- **Framer Motion**: Slide animations and transitions
- **Tailwind CSS**: Styling with custom gradients
- **Lucide React**: Icons throughout slides
- **TypeScript**: Full type safety

## Benefits

### For Fundraising
- Professional, shareable deck at a memorable URL (`planton.ai/invest`)
- Web-native means always up-to-date, no stale PDFs
- Interactive SAFE explainer educates investors unfamiliar with startup terms

### For Storytelling
- YC-style concise format: few words per slide, easy to consume
- Clear flow: Problem → Solution → Proof → Ask
- Consistent visual design with the rest of planton.ai

### For Maintenance
- Component-based: each slide is independent
- Easy to update content without touching structure
- Same patterns as existing demo page

## Impact

### User Experience
- Investors get a professional, interactive presentation
- One URL to share: `planton.ai/invest`
- Works on any device (responsive design)

### Code Organization
- `(micro-apps)` route group clearly identifies standalone mini-applications
- Consistent naming: `invest` everywhere (URL, component folder, CSS)
- Shared layout provides consistent branding across tour/demo/invest

### Content Updates Made

| Item | Change |
|------|--------|
| Date | December 2025 (not 2024) |
| Email | swarup@planton.ai |
| Website | planton.ai |
| Open Source | openmcf.org |
| Team | Fixed spelling: Suresh Attaluri |
| Roadmap | Added Observability (Q1 2026) |
| Cloud Provider | "Join AWS, GCP, Azure, DigitalOcean, Civo" |

## Related Work

- **Demo Page** (`planton.ai/demo`): Used as architectural reference for navigation patterns and slide structure
- **Tour Page** (`planton.ai/tour`): Shares the same route group and minimal header layout

## Files Changed

### New Files (12)
- `src/app/(micro-apps)/invest/page.tsx`
- `src/app/(micro-apps)/invest/layout.tsx`
- `src/app/(micro-apps)/invest/invest.css`
- `src/components/invest/InvestorDeckPage.tsx`
- `src/components/invest/slides/SlideAsk.tsx`
- `src/components/invest/slides/SlideClose.tsx`
- `src/components/invest/slides/SlideCover.tsx`
- `src/components/invest/slides/SlideMarket.tsx`
- `src/components/invest/slides/SlideProblem.tsx`
- `src/components/invest/slides/SlideProduct.tsx`
- `src/components/invest/slides/SlideRoadmap.tsx`
- `src/components/invest/slides/SlideSolution.tsx`
- `src/components/invest/slides/SlideTeam.tsx`
- `src/components/invest/slides/SlideTraction.tsx`

### Renamed (Route Group)
- `src/app/(tour-demo)/` → `src/app/(micro-apps)/`

---

**Status**: ✅ Live
**URL**: https://planton.ai/invest
**Timeline**: ~3 hours of conversation and iteration

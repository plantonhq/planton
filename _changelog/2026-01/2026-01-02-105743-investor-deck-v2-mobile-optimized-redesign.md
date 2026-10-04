# Investor Deck v2: Mobile-Optimized YC-Style Redesign

**Date**: January 2, 2026
**Type**: Feature
**Components**: Investor Deck, UI Components, Design System, Navigation

## Summary

Redesigned the investor pitch deck at `/invest` with a mobile-first, YC-style approach featuring 13 slides (up from 10), shareable URLs via hash navigation, versioned routing (`/invest`, `/invest/v1`, `/invest/v2`), and updated content reflecting the December 2025 landing page positioning and customer traction. Each slide is optimized to fit within a single mobile viewport without scrolling.

## Problem Statement / Motivation

The initial investor deck (v1) received positive feedback but had a key usability issue: slides didn't fit within a single mobile screen, requiring scrolling and reducing the impact of the presentation. Additionally, the content needed updating to reflect the improved copywriting, customer stories, and competitive positioning developed during the December 2025 landing page optimization work.

### Pain Points

- **Mobile overflow**: Slides required scrolling on mobile devices, breaking the presentation flow
- **Outdated content**: Metrics and messaging didn't reflect latest customer traction (450+ deployments, 100% retention)
- **Missing competitive context**: No comparison slide showing Planton vs Terraform/Pulumi/Vercel
- **No customer stories**: Real customer names and metrics (Jai.CX, TynyBay, iorta TechNext) weren't featured
- **No shareable slides**: Couldn't link directly to specific slides like "The Ask"
- **No version history**: No way to access the original deck after updates

## Solution / What's New

Created a complete v2 investor deck with:

1. **Mobile-first design**: Every slide uses `100dvh` and responsive typography to fit single viewport
2. **13 slides** (vs 10 in v1): Added Comparison, Customers, and "Why Invest Now" slides
3. **Hash navigation**: URLs like `/invest/#ask` for sharing specific slides
4. **Version routing**: `/invest` serves v2, `/invest/v1` preserved for history
5. **Updated metrics**: 450+ deployments, 100% retention, $450+ MRR, 5 paying customers
6. **Real customer stories**: Jai.CX (<1hr deploy), TynyBay (8+ projects), iorta (96% savings)
7. **Competitive positioning**: Comparison table vs Terraform, Vercel

### Slide Structure

| # | Slide ID | Title | Key Content |
|---|----------|-------|-------------|
| 1 | `cover` | Cover | Logo, tagline, key metrics (100% retention) |
| 2 | `problem` | Problem | 3 pain points (compact cards) |
| 3 | `solution` | Solution | InfraHub + ServiceHub (2-column) |
| 4 | `product` | How It Works | 3 steps visual |
| 5 | `comparison` | Why Planton | vs Terraform/Pulumi/Vercel (NEW) |
| 6 | `traction` | Traction | Updated metrics grid |
| 7 | `customers` | Customers | 3 customer stories (NEW) |
| 8 | `market` | Market | TAM/SAM numbers |
| 9 | `roadmap` | Roadmap | 4 phases horizontal timeline |
| 10 | `team` | Team | 5 members compact grid |
| 11 | `ask` | The Ask | $500K SAFE with collapsible explainer |
| 12 | `why-now` | Why Invest Now | 3 compelling reasons (NEW) |
| 13 | `close` | Let's Talk | Contact, links |

## Implementation Details

### File Structure

```
src/components/invest/
├── v1/                           # Existing deck (preserved)
│   ├── InvestorDeckPage.tsx
│   └── slides/...
├── v2/                           # New mobile-optimized deck
│   ├── InvestorDeckV2.tsx        # Main orchestrator
│   ├── shared.tsx                # Design system
│   └── slides/
│       ├── SlideCover.tsx
│       ├── SlideProblem.tsx
│       ├── SlideSolution.tsx
│       ├── SlideProduct.tsx
│       ├── SlideComparison.tsx   # NEW
│       ├── SlideTraction.tsx
│       ├── SlideCustomers.tsx    # NEW
│       ├── SlideMarket.tsx
│       ├── SlideRoadmap.tsx
│       ├── SlideTeam.tsx
│       ├── SlideAsk.tsx
│       ├── SlideWhy.tsx          # NEW
│       └── SlideClose.tsx

src/app/(micro-apps)/invest/
├── page.tsx                      # Serves v2 deck
├── v1/page.tsx                   # v1 route
└── v2/page.tsx                   # Explicit v2 route
```

### Hash Navigation

The orchestrator component handles URL hash for shareable slides:

```typescript
// Get initial slide from URL hash
function getInitialSlideIndex(): number {
  if (typeof window === 'undefined') return 0;
  const hash = window.location.hash.slice(1);
  if (hash) {
    const index = slideConfig.findIndex(s => s.id === hash);
    if (index !== -1) return index;
  }
  return 0;
}

// Update hash when navigating
const updateHash = useCallback((slideId: string) => {
  window.history.replaceState(null, '', `#${slideId}`);
}, []);
```

### Mobile-First CSS

Each slide uses viewport units and responsive typography:

```typescript
<div className="min-h-[100dvh] max-h-[100dvh] overflow-hidden">
  <motion.div className="w-full max-w-5xl mx-auto text-center">
    {children}
  </motion.div>
</div>

// Typography scaling
<h2 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-bold">
```

### Hybrid Design System

Combined landing-page-v2 colors with pitch deck drama:

```typescript
export const colors = {
  // Primary gradient (from landing-page-v2)
  gradientStart: '#7c3aed',  // Purple
  gradientEnd: '#0ea5e9',    // Blue
  accentGreen: '#10b981',
  
  // Pitch deck dark backgrounds
  bgPrimary: '#0a0a0f',
  bgSecondary: '#0f0f1a',
};
```

### Key Components Created

1. **`shared.tsx`**: Full design system with `Slide`, `SlideTitle`, `Card`, `Metric`, `Badge`, `Grid`, `Callout`, `TeamMember`, `CustomerCard`, `RoadmapItem`, `FundsItem`, `ComparisonRow` components

2. **`SlideComparison.tsx`**: Competitive comparison table with checkmarks, warnings, and X icons for feature comparison

3. **`SlideCustomers.tsx`**: Customer stories with metrics (Jai.CX, TynyBay, iorta TechNext) and testimonial quote

4. **`SlideAsk.tsx`**: $500K SAFE ask with collapsible SAFE explainer showing investment flow

5. **`SlideWhy.tsx`**: "Why Invest Now" with three compelling reasons (Proven Product, Defensible Moat, Massive Market)

## Benefits

### User Experience
- **Single-screen slides**: Each slide fits mobile viewport without scrolling
- **Touch navigation**: Swipe left/right on mobile devices
- **Keyboard navigation**: Arrow keys, Space, Home, End keys
- **Progress indicator**: Visual progress bar at top + slide counter

### Shareability
- **Deep links**: Share `/invest/#ask` to jump directly to The Ask slide
- **Browser history**: Back/forward navigation works with slides

### Content Quality
- **Updated metrics**: Reflects latest customer traction
- **Real customer stories**: Named companies with specific outcomes
- **Competitive positioning**: Clear differentiation from alternatives

### Maintainability
- **Version history**: v1 preserved at `/invest/v1`
- **Component reuse**: Shared design system across slides
- **Type safety**: Full TypeScript throughout

## Impact

### Investors
- More compelling presentation with real customer proof points
- Easy to share specific slides (e.g., `/invest/#ask` for investment terms)
- Mobile-friendly for reviewing on the go

### Team
- Can reference specific slides in conversations
- Version history allows comparing deck evolution
- Reusable design system for future slides

### Marketing
- Consistent visual language with landing page
- Updated messaging reflects "DevOps in a Box" positioning
- Customer stories build credibility

## Visual Design

### Cover Slide
- Planton branding with gradient
- "DevOps in a Box" tagline
- Key metrics: 3+ Years, $500K+ Self-Funded, 100% Retention
- "Seed Stage • Raising $500K" badge

### Problem Slide
- Red-tinted background for urgency
- 3 compact problem cards with icons
- "80% of cloud services rebuilt from scratch" callout

### Comparison Slide
- 4-column table (Feature, Planton, Terraform, Vercel)
- Color-coded status (green checks, amber warnings, red X)
- Key differentiator callout

### Ask Slide
- $500K in prominent gradient box
- Collapsible SAFE explainer with investment flow
- 3-column use of funds (60% Engineering, 25% AI R&D, 15% GTM)
- 18-month milestones

## Related Work

- **Landing Page Copywriting Revamp** (`2025-12-19-104351`): Source of "DevOps in a Box" positioning
- **Landing Page Marketing Optimization** (`2025-12-31-101158`): Customer stories and metrics
- **Original Investor Deck** (`2025-12-17-112501`): v1 implementation now at `/invest/v1`

---

**Status**: ✅ Live
**Timeline**: ~3 hours implementation
**Routes**: `/invest`, `/invest/v1`, `/invest/v2`


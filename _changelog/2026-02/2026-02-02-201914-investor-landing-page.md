# Investor Landing Page

**Date**: February 2, 2026
**Type**: Feature
**Components**: Investor Pages, Navigation, Layout, UI Components

## Summary

Created a dedicated landing page at `/invest/` that serves as the entry point for all investor resources. Moved the pitch deck to `/invest/deck`, added navigation cards for quick access to key resources, and implemented a Home button in the header for easy navigation back to the investor hub from any investor-related page.

## Problem Statement / Motivation

The investor experience at Planton.ai had grown significantly over time, expanding from a single pitch deck to include:
- Pitch deck (13 slides)
- Opportunity page (market context, alternatives)
- Process page (how to invest via Carta)
- Carta walkthrough (screenshots)
- Explainer pages (What You Get, What We Look For)
- Investor updates timeline

With so many resources, investors arriving at `/invest/` needed a central hub to navigate to the right content based on their intent, rather than landing directly on the deck.

### Pain Points

- No single entry point for the investor experience
- Returning investors had no easy way to find updates
- New investors couldn't easily explore all available resources
- No quick way to navigate back to the investor hub from sub-pages

## Solution / What's New

### Landing Page at `/invest/`

Created a professional landing page with:

1. **Hero Section**: Bold "Invest in Planton" headline with gradient text, key terms badge ($500K SAFE at $7M Cap), and primary CTA to view the deck

2. **Primary Navigation Cards (2x2 grid)**:
   - The Pitch Deck → `/invest/deck`
   - The Opportunity → `/invest/opportunity`
   - How to Invest → `/invest/process`
   - Investor Updates → `/legal/investor-updates`

3. **Secondary Cards (3-column row)**:
   - What You Get → `/invest/and-you-get`
   - What We Look For → `/invest/if-you-are`
   - Carta Walkthrough → `/invest/steps/carta`

4. **Credibility Strip**: 4 metrics (3+ Years Building, $500K+ Self-Funded, 100% Customer Retention, DE C-Corp)

5. **Ready to Talk CTA**: Direct email contact for interested investors

### Deck Moved to `/invest/deck`

The pitch deck that was previously at `/invest` is now accessible at `/invest/deck`, keeping the landing page as the primary entry point.

### Home Button in Header

Added a Home button (using Lucide's Home icon) to the top-right corner of all investor pages that navigates back to `/invest`. The button:
- Appears on all `/invest/*` sub-pages (but not on the landing page itself)
- Appears on `/legal/investor-updates` pages
- Provides consistent navigation across the investor experience

## Implementation Details

### New Files

| File | Purpose |
|------|---------|
| `src/app/(micro-apps)/invest/deck/page.tsx` | Route for pitch deck at new URL |
| `src/components/invest/landing/InvestLandingPage.tsx` | Main landing page component |
| `src/components/invest/InvestHeader.tsx` | Home button component |

### Modified Files

| File | Change |
|------|--------|
| `src/app/(micro-apps)/invest/page.tsx` | Now renders landing page instead of deck |
| `src/app/(micro-apps)/invest/layout.tsx` | Updated metadata, added InvestHeader |
| `src/app/(micro-apps)/legal/layout.tsx` | Added InvestHeader with alwaysShow |
| `src/components/invest/explainer/layout/Footer.tsx` | Added "Home" link, updated "Deck" to `/invest/deck` |

### Component Architecture

```
src/components/invest/
├── InvestHeader.tsx           # Home button for header
└── landing/
    └── InvestLandingPage.tsx  # Landing page with all sections
```

The landing page is built with these internal sections:
- `HeroSection` - Gradient headline, badge, CTA
- `NavigationCard` - Reusable card component for primary resources
- `NavigationCardsSection` - 2x2 grid of primary cards
- `SecondaryCardItem` - Smaller card for secondary resources
- `SecondaryCardsSection` - 3-column row of secondary cards
- `CredibilityStripSection` - 4 metrics in a strip
- `ReadyToTalkSection` - Email CTA

### URL Structure

```
/invest/              → Landing page (NEW)
/invest/deck          → Pitch deck (MOVED from /invest)
/invest/opportunity   → Market opportunity (unchanged)
/invest/process       → Investment process (unchanged)
/invest/steps/carta   → Carta walkthrough (unchanged)
/invest/and-you-get   → What you get (unchanged)
/invest/if-you-are    → What we look for (unchanged)
/legal/investor-updates → Updates timeline (unchanged)
```

## Benefits

### For Investors
- **Clear navigation**: Find the right resource based on intent
- **Quick access**: Primary cards surface most important content
- **Easy return**: Home button always visible for navigation
- **Professional impression**: Polished landing page builds confidence

### For the Company
- **Single shareable URL**: Send `/invest` to anyone interested
- **Analytics opportunity**: Track which paths investors take
- **Scalable structure**: Easy to add new resources to the hub

### Technical Benefits
- **Consistent patterns**: Reuses existing design system components
- **Clean architecture**: Dedicated landing component, separate header component
- **Maintainable**: Footer links auto-updated with new structure

## Impact

### Pages Added
- `/invest` - New landing page
- `/invest/deck` - Deck at new URL

### User Experience
- Investors have a clear entry point to explore resources
- Returning investors can quickly find updates
- All investor pages have consistent Home navigation

### Navigation
- Footer updated with "Home" link pointing to `/invest`
- Footer "Deck" link now points to `/invest/deck`
- Header Home button on all sub-pages

## Related Work

- **Previous**: [`2026-02-02-194524-investor-opportunity-page-enhancement.md`](./2026-02-02-194524-investor-opportunity-page-enhancement.md) - Enhanced opportunity page
- **Previous**: [`2026-02-02-113733-investor-experience-pages-and-header-standardization.md`](./2026-02-02-113733-investor-experience-pages-and-header-standardization.md) - Header standardization
- **Previous**: [`2026-01-30-150743-investor-explainer-pages.md`](../2026-01/2026-01-30-150743-investor-explainer-pages.md) - Explainer pages

---

**Status**: ✅ Live
**Timeline**: ~1 hour implementation

# Investor Experience Pages and Header Standardization

**Date**: February 2, 2026
**Type**: Feature
**Components**: Investor Pages, Investor Updates, Navigation, Layout, UI Components

## Summary

Created three new investor-focused pages (`/invest/opportunity`, `/invest/process`, `/legal/investor-updates`) that continue the philosophy of professional, shareable investor communication. Also standardized headers across all micro-app pages to show only the "P" logo in the top-left corner, removing inconsistent navigation elements.

## Problem Statement / Motivation

Following interest from potential angel investors (like Jeff's $10K interest), there was a need to provide more comprehensive investor communication:

1. **Market Context Gap**: Investors wanted to understand where Planton sits in the competitive landscape
2. **Process Uncertainty**: No clear documentation of how to actually invest (SAFE via Carta)
3. **Transparency Expectations**: Investors wanted to know what happens after they invest (updates, visibility)
4. **Inconsistent Headers**: Different micro-app pages had different header implementations (some with links, progress indicators, or missing logos)

### Pain Points

- No place for investors to understand the market opportunity and competition
- Investment process via Carta wasn't documented publicly
- No mechanism for ongoing investor updates/transparency
- `/invest` deck had no logo, `/legal/investor-updates` had extra navigation links
- Header implementations were fragmented across micro-app pages

## Solution / What's New

### Three New Investor Pages

#### 1. `/invest/opportunity` - Market Opportunity Snapshot

A page for comparing Planton against competitors (Porter.run, Flightcontrol.dev, Qovery.com) with:
- Market context and platform engineering trends
- Competitor cards with funding, positioning, and differentiation (placeholder structure)
- "Why Planton" thesis and technical moat
- Current momentum section (what we did last month, what's next)

#### 2. `/invest/process` - Investment Process via Carta

Step-by-step guide for investors:
- 6-step process from expressing interest to Carta portal access
- Visual indicators for who takes each action (You, Swarup, Carta)
- Infrastructure overview (Mercury, Carta, Stripe, Stripe Atlas)
- Terms summary and process FAQ
- Prepared for Carta screenshots (to be added later)

#### 3. `/legal/investor-updates` - Timeline Updates (Fastlane-Style)

Modeled after donepudi.me/fastlane:
- Accordion expand/collapse timeline
- Date prefix **kept** in URLs (e.g., `/legal/investor-updates/2026-02-01-first-update`)
- Copy link and open in new tab actions
- Individual update pages with full content
- Markdown content stored in `public/investor-updates/`

### Header Standardization

Unified all micro-app pages to use a consistent header:
- Logo on top-left corner only (from parent layout)
- No navigation links in header
- No progress indicators in header (except deck progress bar)
- Cross-references moved to page content/footer

## Implementation Details

### New Files Created

**Opportunity Page**:
- `src/app/(micro-apps)/invest/opportunity/page.tsx` - Route with metadata
- `src/components/invest/opportunity/OpportunityPage.tsx` - Full page component

**Process Page**:
- `src/app/(micro-apps)/invest/process/page.tsx` - Route with metadata
- `src/app/(micro-apps)/invest/steps/page.tsx` - Alias route
- `src/components/invest/process/ProcessPage.tsx` - Full page component

**Investor Updates**:
- `src/app/(micro-apps)/legal/layout.tsx` - Layout for legal pages
- `src/app/(micro-apps)/legal/investor-updates/page.tsx` - Timeline listing
- `src/app/(micro-apps)/legal/investor-updates/[slug]/page.tsx` - Individual update
- `src/lib/investor-updates.ts` - Content utilities (adapted from fastlane.ts)
- `src/components/investor-updates/InvestorUpdatesTimeline.tsx` - Timeline component
- `public/investor-updates/README.md` - Content documentation
- `public/investor-updates/2026-02-01-placeholder-first-update.md` - Sample update

### Files Modified

**Header Standardization**:
- `src/app/(micro-apps)/invest/layout.tsx` - Removed CSS hack hiding parent logo
- `src/app/(micro-apps)/legal/layout.tsx` - Removed CSS hack hiding parent logo
- `src/components/invest/explainer/layout/Footer.tsx` - Added cross-navigation links
- `src/components/invest/explainer/pages/AndYouGetPage.tsx` - Removed Header component
- `src/components/invest/explainer/pages/IfYouArePage.tsx` - Removed Header component
- `src/components/invest/v2/InvestorDeckV2.tsx` - Removed Home button, kept progress dots below header

### Key Technical Decisions

1. **Date Prefix in URLs**: Unlike Fastlane (which strips dates), investor updates keep the date prefix for chronological clarity in shared links

2. **Parent Layout for Headers**: Instead of each page having its own header, all micro-apps inherit the logo from `(micro-apps)/layout.tsx`

3. **Progress Dots Position**: For the investor deck, progress dots moved from header area to `top-16` (below the logo)

4. **Placeholder Structure**: Competitor data and momentum items use HTML comment placeholders for easy filling later

## Benefits

### For Investors
- Clear market context and competitive positioning
- Transparent, documented investment process
- Ongoing visibility via investor updates
- Professional, consistent experience across all pages

### For Planton
- Answer common investor questions before meetings
- Reduce back-and-forth on process questions
- Build confidence through transparency
- Single source of truth for investor communication

### Technical Benefits
- Consistent header behavior across all micro-apps
- Reusable investor-updates library (similar to fastlane)
- Clean separation of content (markdown) from presentation
- Placeholder structure enables iterative content improvement

## Impact

### Pages Added
- `/invest/opportunity` - Market opportunity and competitor comparison
- `/invest/process` - Investment process documentation
- `/invest/steps` - Alias for `/invest/process`
- `/legal/investor-updates` - Timeline listing
- `/legal/investor-updates/[slug]` - Individual updates

### UX Improvements
- All micro-app pages now have consistent logo-only header
- Navigation moved to page content/footer where appropriate
- Investor deck progress dots visible but unobtrusive

## Related Work

- **Previous**: [`2026-01-30-150743-investor-explainer-pages.md`](../2026-01/2026-01-30-150743-investor-explainer-pages.md) - Original explainer pages (`/and-you-get`, `/if-you-are`)
- **Inspiration**: donepudi.me Fastlane feature for timeline-style updates
- **Business Context**: Jeff's investment interest prompted need for clearer process documentation

## Future Enhancements

- [ ] Fill in competitor data placeholders (funding, positioning, differentiation)
- [ ] Add actual momentum items (last month accomplishments, next month goals)
- [ ] Create first real investor update with actual metrics
- [ ] Add Carta screenshots when ready
- [ ] Consider adding RSS feed for investor updates

---

**Status**: ✅ Live
**Routes**:
- https://planton.ai/invest/opportunity
- https://planton.ai/invest/process
- https://planton.ai/invest/steps
- https://planton.ai/legal/investor-updates

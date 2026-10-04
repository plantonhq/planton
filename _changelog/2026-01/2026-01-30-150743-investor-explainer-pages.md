# Interactive Investor Explainer Pages

**Date**: January 30, 2026
**Type**: Feature
**Components**: Investor Pages, Interactive Demo, UI Components, Design System

## Summary

Implemented two interactive, mobile-first investor explainer pages (`/invest/and-you-get` and `/invest/if-you-are`) with path-based progressive disclosure, URL-synced state, live investment calculator, and comprehensive content tailored to different audience backgrounds. These pages provide shareable links that answer "What do I get if I invest?" and "What are you looking for in an investor?" with content adapting based on the visitor's familiarity with startup investing.

**Update (v2)**: Enhanced with UX polish including simplified header (logo-only navigation), hash-based URLs for cleaner sharing, ARR/MRR context alongside valuations, compact landing screen layout, and animated path selection with confirmation states.

## Problem Statement / Motivation

Investors come from diverse backgrounds with varying levels of familiarity with startup investing concepts like SAFEs, valuation caps, and dilution. A one-size-fits-all explanation either overwhelms beginners with jargon or bores experienced investors with basics.

### Pain Points

- **Knowledge gap**: Friends/family investors need fundamentally different explanations than VCs
- **Static content**: PDF pitch decks can't adapt to the reader's background
- **Shareability**: Need direct links that answer specific investor questions
- **Currency context**: Indian investors need INR formatting and context
- **Complex calculations**: SAFE terms and ownership math are confusing without interactive examples

## Solution / What's New

Built a modular, component-based explainer system with:

### Page 1: `/invest/and-you-get` (What You Get When You Invest)

Four knowledge paths with tailored content:
- **Beginner**: Plain English explanations, everyday analogies, no jargon
- **Intermediate**: Standard startup terms with clear definitions
- **Advanced**: Technical details, scenario analysis, edge cases
- **Friend**: Relationship-focused messaging for personal connections

### Page 2: `/invest/if-you-are` (What We're Looking For)

Five investor backgrounds with relevant content:
- **VC / Angel**: Due diligence focus, governance expectations
- **Technical**: Product roadmap, architecture decisions
- **Friend / Family**: Relationship dynamics, realistic expectations
- **Customer**: Strategic partnership value, integration opportunities
- **General**: Universal expectations and deal mechanics

### Shared Features

- **Live Investment Calculator**: Real-time ownership %, return value, and multiple calculations
- **Currency Toggle**: USD/INR with localized formatting (K/M vs L/Cr)
- **URL State Sync**: `?path=beginner&currency=INR` for shareable links
- **Progressive Disclosure**: Content reveals based on selected path
- **Smooth Animations**: Framer Motion for page entry, section reveals, staggered content

### Route Aliases

For easier sharing:
- `/invest/why` → same content as `/invest/and-you-get`
- `/invest/if` → same content as `/invest/if-you-are`

## Implementation Details

### Architecture

```
src/components/invest/explainer/
├── shared.tsx              # Design tokens, primitives, utility functions
├── hooks/
│   ├── useExplainerState.ts    # URL-synced state management
│   └── useCalculator.ts        # Investment calculator logic
├── layout/
│   ├── Header.tsx          # Logo + currency toggle / back link
│   ├── Hero.tsx            # Page title and subtitle
│   ├── Footer.tsx          # Last updated timestamp
│   └── CTASection.tsx      # Call-to-action section
├── controls/
│   └── PathSelector.tsx    # Knowledge level / background selector
├── calculator/
│   └── Calculator.tsx      # Investment calculator + scenario table
├── content/
│   └── FAQ.tsx             # Accordion FAQ component
└── pages/
    ├── AndYouGetPage.tsx   # Full page 1 with all sections
    ├── IfYouArePage.tsx    # Full page 2 with all sections
    └── index.ts            # Barrel export
```

### Key Technical Decisions

1. **URL State Sync**: Used `useSearchParams` from Next.js App Router with custom `useExplainerState` hook to persist selections in URL for shareability

2. **Static Export Compatible Aliases**: Since `planton.ai` uses `output: 'export'`, aliases are implemented as separate page files rather than Next.js rewrites

3. **Calculator Logic**: Ported investment formulas from vanilla JS to TypeScript with proper typing:
   - SAFE ownership calculation with valuation cap
   - Post-money dilution modeling
   - Exit scenario projections

4. **Component Modularity**: Section components are self-contained with consistent props interface, allowing easy reordering and conditional rendering based on path

### Files Created

**Routes (4 files)**:
- `src/app/(micro-apps)/invest/and-you-get/page.tsx`
- `src/app/(micro-apps)/invest/if-you-are/page.tsx`
- `src/app/(micro-apps)/invest/why/page.tsx` (alias)
- `src/app/(micro-apps)/invest/if/page.tsx` (alias)

**Components (12 files)**:
- `src/components/invest/explainer/shared.tsx`
- `src/components/invest/explainer/hooks/useExplainerState.ts`
- `src/components/invest/explainer/hooks/useCalculator.ts`
- `src/components/invest/explainer/layout/Header.tsx`
- `src/components/invest/explainer/layout/Hero.tsx`
- `src/components/invest/explainer/layout/Footer.tsx`
- `src/components/invest/explainer/layout/CTASection.tsx`
- `src/components/invest/explainer/controls/PathSelector.tsx`
- `src/components/invest/explainer/calculator/Calculator.tsx`
- `src/components/invest/explainer/content/FAQ.tsx`
- `src/components/invest/explainer/pages/AndYouGetPage.tsx`
- `src/components/invest/explainer/pages/IfYouArePage.tsx`

**Barrel Exports (4 files)**:
- `src/components/invest/explainer/hooks/index.ts`
- `src/components/invest/explainer/layout/index.ts`
- `src/components/invest/explainer/controls/index.ts`
- `src/components/invest/explainer/pages/index.ts`

## Benefits

### For Investors
- Get exactly the information relevant to their background
- Interactive calculator to understand potential outcomes
- Shareable links to specific content sections
- Currency-appropriate formatting (USD for US, INR for India)

### For Planton
- Answer common investor questions before meetings
- Filter investors by sending appropriate links
- Track engagement via URL parameters
- Reduce time spent on basic explanations

### Technical Benefits
- Modular architecture enables easy content updates
- URL state enables analytics tracking of path preferences
- TypeScript ensures type safety across calculations
- Framer Motion provides polished, professional feel

## Impact

### User Experience
- Investors can self-serve answers to common questions
- Content adapts to knowledge level, reducing cognitive load
- Interactive calculator makes abstract terms concrete

### Shareability
- Direct links: `planton.ai/invest/why?path=friend&currency=INR`
- Each path selection creates unique shareable URL
- Social previews with appropriate metadata

### Conversion
- Pre-qualified investors arrive better informed
- Reduced back-and-forth on basic questions
- Professional presentation builds credibility

## Related Work

- **Previous**: `2025-12-17-112501-investor-pitch-deck.md` - Original investor deck
- **Previous**: `2026-01-02-105743-investor-deck-v2-mobile-optimized-redesign.md` - Mobile-optimized deck
- **Source**: HTML/CSS/JS prototypes in `planton/_projects/20260130.02.investor-explainer-pages/`

## v2 Enhancements (Same Day)

Following user feedback on the initial implementation, significant UX improvements were made:

### 1. Simplified Header Navigation

**Before**: Header included "Investor Deck" text + inter-page navigation links ("What We Look For →")
**After**: Header shows only the "P" logo (linking to home) + progress indicator in center

This change focuses attention on the current content rather than fragmenting user attention across pages.

### 2. Hash-Based URLs (Cleaner Sharing)

**Before**: `/invest/and-you-get?path=beginner&currency=INR`
**After**: `/invest/and-you-get#beginner`

- Path selection now uses URL hash (`#intermediate`) for cleaner URLs
- Currency preference stored in `localStorage` instead of URL (user preference, not content state)
- Simpler, more shareable links

### 3. ARR/MRR Context for Valuations

Every valuation figure now includes corresponding ARR and MRR to help investors contextualize:

**Before**: `$7M valuation cap`
**After**: `$7M valuation cap (~$583K ARR / ~$49K MRR)`

This makes it immediately clear what revenue level justifies each valuation:
- $7M → ~$49K/month revenue
- $20M → ~$139K/month revenue
- $100M → ~$694K/month revenue

Uses industry-standard 12x revenue multiple for B2B DevTools.

### 4. Currency Toggle Relocation

**Before**: Currency toggle (USD/INR) in global header
**After**: Currency toggle inline with "Investment Calculator" title

Contextually placed where currency actually matters—when doing calculations.

### 5. Path Selection UX Overhaul

**Before**: Clicking a path just updated state; no visual confirmation
**After**: Rich interaction with:
- Smooth collapse animation of unselected options
- Compact confirmation card showing selected path
- "Change" button to reset selection
- Checkmark (✓) indicator for selected state
- Staggered entry animations

### 6. Compact Landing Screen

Reduced padding throughout to fit hero + path selector in single viewport:

| Element | Before | After |
|---------|--------|-------|
| Hero padding | `py-12/16/20` | `py-6/8/10` |
| PathSelector padding | `py-8/12` | `py-4/6` |
| Button padding | `p-4/5` | `p-3/4` |
| Button spacing | `space-y-3` | `space-y-2` |

Users no longer need to scroll to see all path options on first load.

### 7. Parent Layout Conflict Resolution

Fixed overlapping "P" logo issue caused by nested layout inheritance:
- `(micro-apps)/layout.tsx` renders a `HeaderLogo`
- `invest/layout.tsx` now injects CSS to hide the parent logo
- Allows invest pages to have their own minimal header design

### Files Modified

```
src/app/(micro-apps)/invest/layout.tsx                    # CSS override for parent logo
src/components/invest/explainer/calculator/Calculator.tsx # +formatARR, currency toggle
src/components/invest/explainer/controls/PathSelector.tsx # Collapse animation, confirmation
src/components/invest/explainer/hooks/useExplainerState.ts # Hash-based routing
src/components/invest/explainer/layout/Header.tsx         # Simplified to logo-only
src/components/invest/explainer/layout/Hero.tsx           # Reduced padding
src/components/invest/explainer/pages/AndYouGetPage.tsx   # ARR/MRR in content
src/components/invest/explainer/pages/IfYouArePage.tsx    # ARR/MRR in content
src/components/invest/explainer/shared.tsx                # +formatARR, valuationToMRR
```

---

**Status**: ✅ Live
**Routes**: 
- https://planton.ai/invest/and-you-get
- https://planton.ai/invest/if-you-are
- https://planton.ai/invest/why (alias)
- https://planton.ai/invest/if (alias)

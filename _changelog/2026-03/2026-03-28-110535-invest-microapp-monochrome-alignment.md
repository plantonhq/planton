# Invest Microapp Monochrome Design System Alignment

**Date**: March 28, 2026
**Type**: Design
**Components**: Invest Microapp, Design System, UI Components

## Summary

Brought the entire invest microapp — pitch deck (14 slides), landing page, and all explainer pages (Opportunity, Process, Carta Walkthrough, What You Get, What We Look For) — into full alignment with the monochrome design system established across the rest of planton.ai. This was a multi-phase effort spanning two separate component systems (`v2/shared.tsx` for the deck and `explainer/shared.tsx` for long-form pages) plus dozens of inline style overrides across 25+ files.

## Problem Statement / Motivation

After the monochrome redesign wave (11 changelogs from March 25-27, 2026), the main website — homepage, feature pages, docs, pricing — looked consistent and professional. But the invest microapp still felt like it belonged to a different product: decorative gradients, alpha-based colors, oversized typography, inconsistent card styles, and visual noise that clashed with the clean, brightness-based hierarchy used everywhere else.

### Pain Points

- Two separate component foundations (`v2/shared.tsx` and `explainer/shared.tsx`) had never been updated for the monochrome system
- Decorative gradient text, gradient backgrounds, gradient CTAs, and gradient orbs violated the "brightness, not hue" principle
- Alpha-based styling (`bg-white/5`, `border-white/10`, `text-white/60`) produced inconsistent rendering depending on backdrop
- Typography used `font-bold` and `font-extrabold` while the site standardized on `font-semibold`
- Cards, badges, and callouts used different patterns from the main site
- Slide content crowded against the dot navigation tracker
- Team photos loaded from external Cloudflare CDN causing slow loads
- "Planton" branding persisted on the cover slide despite the rebrand to "Planton"

## Solution / What's New

### Phase 1: Deck Foundation + Slides 1-7

Updated `src/components/invest/v2/shared.tsx` — the foundation for all 14 deck slides:
- Design tokens aligned to solid hex values matching the main site
- `SlideTitle`: `font-bold` to `font-semibold`, scale from `lg:text-5xl` to `lg:text-4xl`, added `tracking-tight`
- Card backgrounds from `bg-white/5 border-white/10` to `bg-[#151515] border-[#2a2a2a]` with hover states
- Badge from alpha-based to solid `bg-[#2a2a2a] text-[#a0a0a0] border-[#3a3a3a]`
- Check/X/Warning icons from `w-4 h-4` to `w-3.5 h-3.5` with `/70` opacity
- Slide backgrounds flattened from gradients to `bg-[#0a0a0a]`

Updated `InvestorDeckV2.tsx` deck shell:
- Footer from gradient overlay to solid `bg-[#0a0a0a]/90 backdrop-blur-sm border-t border-[#2a2a2a]`
- Dot navigation moved from `top-16` to `top-4`
- Swipe hint: emoji replaced with ChevronRight icon, pulse replaced with fade-in
- Back/Next buttons aligned with main site patterns

Slide-specific fixes for Cover through Customers:
- "Planton" renamed to "Planton"
- All gradient text replaced with solid `text-white`
- Green gradient callouts replaced with monochrome + semantic accent borders
- Negative margin hacks removed, proper top padding applied
- All `text-white/XX` converted to hex equivalents

### Phase 2: Slides 8-14

Applied the same treatment to Wall of Love, Market, Roadmap, Team, The Ask, Why Invest Now, and Close slides:
- Top padding standardized to `!pt-24 sm:!pt-28 md:!pt-32`
- All inline card backgrounds from alpha to solid hex
- Avatar fallback gradients replaced with solid `bg-[#2a2a2a]`
- SAFE modal: gradient background flattened, all internal boxes to solid hex
- CollegeBadge aligned to Badge component pattern
- Gradient text on "This Hard" replaced with solid white

### Phase 3: Explainer System + All Long-Form Pages

Updated `src/components/invest/explainer/shared.tsx` — the foundation for all explainer pages:
- Design tokens: `textSecondary` from `rgba(255,255,255,0.6)` to `#a0a0a0`, borders from rgba to hex
- Same Card/Badge/Callout/Metric alignment as deck shared.tsx
- Table borders, Step circles, List text all to hex
- Icons aligned to `w-3.5 h-3.5` with `strokeWidth={2.5}`

Updated layout components:
- `Hero.tsx`: gradient background and orb decoration removed, flat `bg-[#0a0a0a]`
- `Footer.tsx`: border and text colors to hex

Page-specific fixes across all explainer pages:
- `OpportunityPage.tsx`: 15+ inline fixes — decorative card borders removed, CTA gradient to solid white, comparison boxes to hex
- `ProcessPage.tsx`: gradient page bg, step number gradients, CTA gradient all fixed
- `CartaWalkthroughPage.tsx`: same pattern — gradient bg, step circle, CTA fixed
- `IfYouArePage.tsx`: gradient bg and 5 gradient SectionTitle overrides removed
- `AndYouGetPage.tsx`: gradient bg and 4 gradient SectionTitle overrides removed

### Phase 4: Team Images + Dead CSS

- Copied 5 optimized JPG team photos from planton monorepo to `public/images/team/`
- Updated all avatar URLs from `https://assets.planton.ai/team/*.png` to `/images/team/*.jpg`
- Team slide spacing tightened to prevent bottom cutoff
- Dead CSS in `invest.css` reduced to only `kbd` styling

## Implementation Details

### Files Modified

**Deck system (v2):**
- `src/components/invest/v2/shared.tsx`
- `src/components/invest/v2/InvestorDeckV2.tsx`
- `src/components/invest/v2/slides/SlideCover.tsx`
- `src/components/invest/v2/slides/SlideProblem.tsx`
- `src/components/invest/v2/slides/SlideSolution.tsx`
- `src/components/invest/v2/slides/SlideProduct.tsx`
- `src/components/invest/v2/slides/SlideComparison.tsx`
- `src/components/invest/v2/slides/SlideTraction.tsx`
- `src/components/invest/v2/slides/SlideCustomers.tsx`
- `src/components/invest/v2/slides/SlideWallOfLove.tsx`
- `src/components/invest/v2/slides/SlideMarket.tsx`
- `src/components/invest/v2/slides/SlideRoadmap.tsx`
- `src/components/invest/v2/slides/SlideTeam.tsx`
- `src/components/invest/v2/slides/SlideAsk.tsx`
- `src/components/invest/v2/slides/SlideWhy.tsx`
- `src/components/invest/v2/slides/SlideClose.tsx`

**Explainer system:**
- `src/components/invest/explainer/shared.tsx`
- `src/components/invest/explainer/layout/Hero.tsx`
- `src/components/invest/explainer/layout/Footer.tsx`

**Explainer pages:**
- `src/components/invest/opportunity/OpportunityPage.tsx`
- `src/components/invest/process/ProcessPage.tsx`
- `src/components/invest/carta-walkthrough/CartaWalkthroughPage.tsx`
- `src/components/invest/explainer/pages/IfYouArePage.tsx`
- `src/components/invest/explainer/pages/AndYouGetPage.tsx`

**Landing + CSS:**
- `src/components/invest/landing/InvestLandingPage.tsx`
- `src/app/(micro-apps)/invest/invest.css`

**Assets:**
- `public/images/team/swarup-donepudi.jpg` (new)
- `public/images/team/suresh-attaluri.jpg` (new)
- `public/images/team/irshad-ahmed.jpg` (new)
- `public/images/team/avinash-sana.jpg` (new)
- `public/images/team/satish-lakhani.jpg` (new)

## Benefits

- **Visual consistency**: The invest microapp now looks like it belongs to the same platform as the main website
- **Faster page loads**: Team photos served locally (~170-287KB JPGs) instead of external Cloudflare PNG round-trips
- **Maintainability**: Both component foundations now follow the same design token patterns as the main site
- **Reduced technical debt**: Dead CSS removed, vestigial variant names cleaned up, negative margin hacks eliminated
- **Correct branding**: "Planton" updated to "Planton" on the cover slide

## Impact

Every investor-facing page and the entire 14-slide pitch deck now render with the same monochrome design language as the main website. The invest microapp no longer feels like a separate product — it feels like a cohesive part of planton.ai.

## Related Work

- `2026-03-25-144804-black-and-white-theme-redesign.md` — established the monochrome design system
- `2026-03-25-170502-micro-apps-monochrome-theme.md` — initial micro-apps alignment (this work completes it)
- `2026-03-25-095619-rebrand-planton-cloud-to-planton-and-google-oauth-fix.md` — rebrand that this work enforces on the deck
- `2026-03-27-102536-header-redesign-inter-font-and-dropdown-polish.md` — header/typography standards this work follows

---

**Status**: Live
**Timeline**: Single session, multi-phase implementation

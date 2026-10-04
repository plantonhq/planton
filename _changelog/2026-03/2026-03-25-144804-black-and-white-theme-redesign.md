# Black & White Theme Redesign

**Date**: March 25, 2026
**Type**: Design
**Components**: Landing Page, Design System, Navigation, Pricing Page, Features Pages, Solutions Pages, Agents Page, UI Components

## Summary

Complete visual overhaul of planton.ai from a purple/blue gradient-heavy aesthetic to a clean monochrome black-and-white theme inspired by cursor.com. Stripped all decorative gradients, colored orbs, and rainbow accents from 55 files across the entire marketing site. Compacted typography scale to match cursor.com's tighter, more professional sizing. The result is a site that matches the Planton console's existing black-and-white identity.

## Problem Statement / Motivation

The planton.ai website had accumulated heavy purple (#7c3aed), sky blue (#0ea5e9), cyan, amber, and gold gradients across every section -- hero orbs, gradient text, colored badges, tinted cards, gradient buttons, and decorative glows. The visual identity felt "built by AI" and clashed with the Planton console app, which already used a clean black-and-white theme.

### Pain Points

- Purple-to-sky gradient text on nearly every section headline
- Decorative radial gradient orbs in the hero and CTA sections
- Rainbow badge variants (purple, green, amber) used decoratively rather than semantically
- PrimaryButton used a purple-to-sky gradient instead of a clean solid fill
- Navigation dropdown was light-themed (white background) creating a jarring contrast with the dark site
- Typography was oversized (hero text up to 96px, section titles up to 48px) compared to modern B2B SaaS sites
- Inconsistency between the marketing site's colorful aesthetic and the console's monochrome UI

## Solution / What's New

### Design Tokens (`shared.tsx`)

Rewrote the v3 design token system -- the single source of truth for all 18 homepage sections:

- **Colors**: Removed `gradientStart`, `gradientEnd`, `gradientAccent`, `accentPurple`, `accentBlue` tokens entirely. Kept only background layers, text scale, semantic status colors (green/red/amber for functional indicators), and border grays
- **PrimaryButton**: `bg-gradient-to-r from-[#7c3aed] to-[#0ea5e9]` replaced with `bg-white text-black`
- **SecondaryButton**: Purple hover replaced with `hover:border-white hover:bg-white/5`
- **Badge**: Purple variant now maps to the same neutral gray as default
- **Section gradient variant**: Previously added a blue-tinted `via-[#0f0f1a]` midpoint; now identical to default `#0a0a0a`
- **Quote**: Purple left border (`#7c3aed`) replaced with `border-white/30`
- **Metric/MetricCard**: Gradient text replaced with solid white

### Typography Compaction

Tightened the entire type scale to match cursor.com's compact aesthetic:

| Element | Before | After |
|---|---|---|
| Section padding | `py-16 md:py-24` | `py-12 md:py-16` |
| SectionTitle | 24px-48px | 20px-30px |
| Hero "Planton" | 48px-96px | 36px-60px |
| Hero tagline | 30px-60px | 24px-36px |
| CTA button | `text-lg px-10 py-5` | `text-sm px-8 py-3` |
| Card padding | `p-6 md:p-8` | `p-5 md:p-6` |
| TypoH2 (pricing) | 32px/86px | 24px/40px |

Font weights shifted from `font-bold` to `font-semibold` with `tracking-tight` letter-spacing throughout.

### Navigation Dropdown

Converted from light (white background) to dark theme:

- Paper background from white to `#111` with `border: 1px solid #2a2a2a`
- Menu items hover: from invisible white-on-white text to `hover:bg-white/10` background highlight
- Section dividers from `border-gray-200` to `border-[#2a2a2a]`
- Footer menu from `bg-[#e3e3e3]` to `bg-white/5`
- Section titles styled as `text-[#666] text-xs uppercase tracking-wider`

### Subpage Updates

- **Pricing**: Purple card gradients (`#8A3391`), gold "Most Popular" banner, blue CTA buttons all replaced with monochrome equivalents
- **Features/Solutions**: Blue text gradients, `StyledAiBtn` blue border rings, blurred orbs all neutralized
- **Agents**: All 7 section files updated -- blue/cyan/purple Tailwind classes replaced with white/neutral equivalents
- **v1 Components**: `Pill` gradient borders, `SectionTitle` gradient text, `JoinBetaBtn` dialog all neutralized

## Implementation Details

### Phase 1: Foundation (5 files)

- `src/components/landing-page/v3-2026-01-02-1000/shared.tsx` -- Complete design token rewrite
- `tailwind.config.ts` -- Primary/secondary color scales replaced with neutral grayscale
- `src/theme/colors.ts` + `src/theme/theme.ts` -- MUI `primary.main` from `#0095FF` to `#ffffff`
- `src/app/globals.css` -- Demo button gradient and prose link purples neutralized

### Phase 2: Homepage (18 files)

All v3 landing sections rewritten: HeroSection, SecurityTrustBar, SocialProofBar, HowItWorks, ComparisonTable, ProblemSolution, InfraHub, ServiceHub, WallOfLove, CustomerStories, ROICalculator, PricingSimplified, OpenStandards, OpenSourceFoundation, AgentFleet, BuiltByDevOps, SecurityCompliance, FinalCTA.

### Phase 3: Shell (1 file)

`ProductMenu.tsx` -- Dropdown converted to dark theme with proper hover states.

### Phase 4: Subpages (35 files)

Batch color replacement across pricing, features, solutions, agents, CLI, IaC workflows, Kubernetes dashboard, self-service DevOps, and v1 shared components.

### What Was Preserved

- Terminal syntax colors (green for success, amber for timing) -- semantically meaningful
- Comparison table status icons (green check, red X, amber warning) -- functional indicators
- hljs code highlighting -- industry standard
- Provider/brand logos -- original colors with opacity treatment
- Tour, Demo, Invest, Meets, ACME sections -- out of scope (separate products/experiences)
- Mermaid diagram theme colors -- data visualization

## Benefits

- **Brand consistency**: Marketing site now matches the console's monochrome identity
- **Professional appearance**: Clean, modern aesthetic comparable to cursor.com
- **Reduced visual noise**: Removing decorative gradients lets content breathe
- **Compact typography**: Tighter type scale feels more confident and less "AI-generated"
- **Dark dropdown**: Navigation menus are now usable and consistent with the site theme

## Impact

- **55 files** modified across the entire marketing site
- **Zero build errors** -- clean `next build` exit code 0
- **All pages** compile and render correctly
- Homepage, pricing, features, solutions, agents, and docs all affected
- No changes to Tour, Demo, Invest, Meets, or ACME (out of scope)

## Related Work

- Follows the Planton console's existing black-and-white theme
- Complements the "Planton" to "Planton" rebrand (changelog: `2026-03-25-095619`)

---

**Status**: Live
**Timeline**: Single session

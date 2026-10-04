# Monochrome Polish & Contrast Fixes

**Date**: March 25, 2026
**Type**: Design
**Components**: Landing Page, Invest Decks, Demo, Changelog, Docs, Legal, Common Components

## Summary

Follow-up pass after the micro-apps monochrome theme rollout to fix contrast issues, invisible buttons, and missed colored components. Resolved white-on-white artifacts from gradient neutralization, made comparison table icons subtler, neutralized the hero terminal green, and extended the monochrome treatment to the changelog, docs, and investor updates timeline — components that were not part of the original micro-apps plan.

## Problem Statement / Motivation

After the bulk monochrome conversion (changelog: `2026-03-25-170502`), visual QA revealed several issues:

### Pain Points

- **Invest deck "Next" button** was invisible — `bg-gradient-to-r from-white to-[#666]` background with `text-white` meant white text on a white background
- **Invest deck "Let's Talk" CTA card** had the same gradient-on-white invisibility — the email button was `bg-white text-white`
- **Demo welcome card** had an invisible "Start Personalized Demo" button — `btn-gradient` CSS class now produced `#ededed` bg (from Phase 1 CSS fix) on a `bg-white` card (also `#ededed` via Tailwind override)
- **Demo company selection cards** were light-themed (`bg-white`, `text-gray-900`) sitting on a dark `#0a0a0a` background — washed out and jarring
- **Comparison table icons** (check/X/warning) were 20px at full color intensity — too loud for the monochrome aesthetic
- **Hero terminal** still had green `$` prompt and success message text
- **Changelog page** had blue category badges, blue link accents, and blue hover states
- **Docs** had blue sidebar active states, search modal accents, MDX link colors, and colored status badges
- **Investor Updates Timeline** had pink date badges, violet tag pills, pink title hovers, and cyan markdown links

## Solution / What's New

### White-on-White Gradient Fixes

The `from-white to-[#666]` gradient pattern was a common artifact from the bulk neutralization — it replaced `from-pink-500 to-violet-500` but created invisible text when used as backgrounds with `text-white`. Fixed across:

- **SlideClose CTA card**: light gradient → `bg-[#1a1a1a] border border-[#2a2a2a]` dark card with `bg-[#fff] text-black` CTA button
- **InvestorDeckV2 "Next" button**: gradient → solid `bg-[#fff] text-black`
- **InvestorDeckV2 progress bar**: gradient → solid `bg-white`
- **InfraVersionHistory version circles**: gradient → `bg-[#1a1a1a] border border-[#2a2a2a]`
- **GitHubConnection connect button**: gradient → `bg-[#fff] text-black`
- **CompanyTypeSelection card colors**: `from-white to-[#666]` → `from-white/10 to-white/5`

### Demo Cards Dark Theme Flip

Two demo screens had light-themed cards (`bg-white`, `text-gray-900`) on dark backgrounds:

- **DemoPage welcome card**: flipped to `bg-[#1a1a1a]` dark card with white heading, `text-[#a0a0a0]` description, `bg-[#fff] text-black` CTA button, `bg-white/10` play icon circle
- **CompanyTypeSelection cards**: flipped to `bg-[#1a1a1a] border-[#2a2a2a]` with `text-white` titles, `text-[#a0a0a0]` descriptions, `bg-white/5 text-[#666]` pain point pills, `bg-white/10` icon circles, `ring-2 ring-white/50` selected state

### Subtler Comparison Table Icons

Shrunk from `w-5 h-5` (20px) to `w-3.5 h-3.5` (14px) at 70% opacity with slightly thicker strokes (`strokeWidth` 2 → 2.5) for readability at the smaller size. Applied to CheckIcon, XIcon, and WarningIcon in the shared design system.

### Hero Terminal Neutralization

Green `$` prompt (`text-green-400`) and success messages → `text-[#a0a0a0]` (secondary gray). The checkmark characters `✓` already convey success without needing green. macOS traffic light dots preserved (universally recognized UI pattern).

### Changelog Monochrome

- **Category badges** (Feature/Improvement/Fix/Breaking): green/blue/amber/red → brightness-based monochrome using `white/` opacity variants
- **Timeline**: blue link hovers, tag pills, search focus border, copy/open icon hovers → white/gray
- **Markdown body**: blue links and blockquote borders → white

### Docs Monochrome

- **Sidebar**: blue active states → white/gray; yellow "Experimental" badge → neutral; colored status badges preserved as semantic (green/red)
- **Search modal**: blue accents, spinner, highlights → white/gray
- **MDX renderer/components**: blue links → white; semantic red/green preserved
- **Not-found page**: blue CTA button → white

### Investor Updates Timeline

Pink date badges, violet tag pills, pink title hover, cyan markdown links, expanded content left border → all neutral monochrome.

## Implementation Details

Root cause of the invisible buttons: the Tailwind `white` override to `#ededed` (from the earlier softening pass) meant `bg-white` produces `#ededed`. When the bulk gradient neutralization converted `from-pink-500 to-violet-500` to `from-white to-[#666]`, any element using this as a background with `text-white` text became white-on-white. The fix was case-specific: buttons got `bg-[#fff] text-black`, cards got flipped to dark theme, progress bars got solid `bg-white`.

## Benefits

- **No more invisible buttons** — every CTA is now visible and clickable
- **Consistent dark theme** — demo cards match the rest of the site instead of jarring light panels on dark backgrounds
- **Subtler visual indicators** — comparison icons don't scream; they inform
- **Complete monochrome coverage** — changelog, docs, and investor updates now match the design system
- **Clean terminal aesthetic** — hero terminal uses neutral gray for all output

## Impact

- **~20 files** modified across 5 follow-up commits
- **Zero build errors** throughout
- Changelog, docs, and investor updates pages now fully monochrome
- All demo screens consistent with dark theme

## Related Work

- Follows the micro-apps monochrome theme (changelog: `2026-03-25-170502`)
- Follows the white softening (changelog: `2026-03-25-151117`)
- Follows the original black-and-white redesign (changelog: `2026-03-25-144804`)

---

**Status**: Live
**Timeline**: Same session as micro-apps rollout

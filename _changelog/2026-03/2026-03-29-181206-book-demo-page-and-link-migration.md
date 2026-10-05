# Book Demo Page and Site-Wide Link Migration

**Date**: March 29, 2026
**Type**: Feature
**Components**: Book Demo Page, Lead Capture Form, Cal.com Integration, Focus Layout, CTA Links

## Summary

Built the `/book-demo` page — a two-phase lead capture and scheduling experience — and replaced every external form URL (Google Form + Typeform) across the entire planton.ai site with internal `/book-demo` navigation. The page uses a focus-mode layout (no header/footer, X button, Escape to close) inspired by the console app's creation wizard.

## Problem Statement / Motivation

"Book Demo" CTAs across planton.ai redirected users to external services (Google Form in ~44 files, Typeform in 6 files). This broke the user's flow, provided no real-time team notification, and presented an unprofessional experience for a platform positioned as world-class infrastructure automation.

### Pain Points

- External forms break immersion — users leave the site
- No structured lead data capture or immediate team notification
- No scheduling integration — users must wait for follow-up
- Google Form and Typeform URLs scattered across 50+ UI components with no centralization

## Solution / What's New

### /book-demo Page (Two-Phase UX)

**Phase 1 — Lead Capture:**
- Two-column layout: value propositions (left) + form card (right)
- Form fields: Full Name, Work Email, Company, Job Title (dropdown), Company Size (dropdown), Message (optional)
- Client-side validation on blur + submit
- Submits to the `demo-request-receiver` Cloudflare Worker (`POST /submissions`)
- Team gets Discord notification immediately

**Phase 2 — Cal.com Scheduling (Optional):**
- On successful submission: confetti burst (white/gray + success green), smooth crossfade transition
- Left column: confirmation ("Thanks, {firstName}!"), promise ("We'll reach out via email within 1-2 business days"), soft bridge ("or"), invitation to self-schedule
- Right column: Cal.com inline embed (`swarup-donepudi/60min`, month view, dark theme, prefilled name + email)
- "What to expect" prep tips alongside the embed

### Focus-Mode Layout

New `(focus)` route group with minimal shell:
- No site header or footer — zero distraction
- Fixed X button (top-right) navigates back
- Escape key also navigates back
- Content vertically centered in viewport
- Subtle grid background pattern (matching HeroSection)

### Site-Wide Link Migration

Replaced ALL external form URLs across the entire codebase:
- **Google Form**: 44+ files, zero remaining
- **Typeform**: 6 files, zero remaining
- Removed `target="_blank"` / `target="_self"` attributes (now internal navigation)

## Implementation Details

### New Files

| File | Purpose |
|------|---------|
| `src/app/(focus)/layout.tsx` | Focus-mode layout: X button, Escape key, no chrome |
| `src/app/(focus)/book-demo/page.tsx` | Route entry with SEO metadata |
| `src/components/book-demo/BookDemoPage.tsx` | Main component: phase state, two-column layout, confetti, Phase 1/2 left columns |
| `src/components/book-demo/BookDemoForm.tsx` | Lead capture form: Tailwind-styled inputs, validation, submission, error states |
| `src/components/book-demo/BookDemoScheduler.tsx` | Cal.com embed with `getCalApi` initialization, dynamic import |
| `src/components/book-demo/types.ts` | Shared types, constants, dropdown options |

### Modified Files

- `package.json` + `yarn.lock` — added `@calcom/embed-react`
- 50+ component files — Google Form and Typeform URLs replaced with `/book-demo`

### Key Design Decisions

**Tailwind-styled native inputs over MUI TextField**: The site had zero existing form components. Native `<input>` and `<select>` elements styled with Tailwind give pixel-perfect control matching the monochrome design system (`bg-[#1a1a1a]`, `border-[#2a2a2a]`, `text-[#ededed]`).

**Focus-mode layout**: Inspired by the console app's creation wizard. The `/book-demo` page is a conversion-critical flow — removing site chrome eliminates distraction and creates focus.

**Two-phase UX with lead-first**: The team gets notified via Discord immediately after Phase 1, even if the user doesn't book. Cal.com scheduling is offered as a convenience ("Prefer to skip the wait?"), not demanded.

**Confetti**: Brief burst of white/gray + success green particles using existing `canvas-confetti` dependency. `disableForReducedMotion: true` for accessibility.

**Error states**: Elegant card-style banner (not alarming red box). Human-friendly copy ("We hit a snag"), no blame on the user, fallback email (`hello@planton.ai`) with click-to-copy and green tick feedback.

**Job Title as dropdown**: Common roles (CTO, VP Engineering, Platform Engineer, etc.) + "Other" option. Reduces friction vs free text.

### Form → Worker Flow

```
Browser (planton.ai/book-demo)
  → POST https://demo-request-receiver.planton.ai/submissions
  → Worker validates, posts Discord embed
  → 200 → confetti + Phase 2 transition
  → Error → elegant error banner with email fallback
```

### Cal.com Configuration

- `calLink`: `swarup-donepudi/60min`
- `namespace`: `60min`
- `hideEventTypeDetails`: `false`
- `layout`: `month_view`
- `useSlotsViewOnSmallScreen`: `true`
- Prefill: name + email from Phase 1
- Theme/brand colors configured at Cal.com dashboard level

## Benefits

- **Immediate team notification**: Demo requests appear as structured Discord embeds the moment the form is submitted
- **Lead capture before scheduling**: Team gets notified even if the user doesn't complete Cal.com booking
- **Professional, on-brand experience**: Dark monochrome design, focus layout, elegant error handling
- **Zero external dependencies in the user flow**: No more Google Form or Typeform redirects
- **Conversion-optimized**: Value props, social proof, and metrics visible alongside the form

## Impact

- **50+ files updated** across the entire planton.ai codebase
- **Every "Book Demo" CTA** now routes to `/book-demo`
- **Zero remaining external form URLs** in the source code
- **New (focus) route group** available for future distraction-free experiences

## Related Work

- **T01** (completed, previous session): `demo-request-receiver` Cloudflare Worker + origin rule update
- **Architecture plan**: `_projects/20260329.01.book-demo/plans/book-demo-architecture.md`

---

**Status**: ✅ Live (pending worker deployment with DISCORD_WEBHOOK_URL secret)

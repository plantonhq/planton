# Book Demo Form Polish and Cal.com Scroll Fix

**Date**: March 29, 2026
**Type**: Enhancement
**Components**: Book Demo Page, Lead Capture Form, Cal.com Integration, Design System

## Summary

Restructured the `/book-demo` lead capture form (split full name into first/last, removed message field, one-field-per-row layout), fixed the Cal.com embed scrolling on booking confirmation, hardened error handling to never leak raw server messages, and eliminated submission flicker with a minimum loading duration pattern.

## Problem Statement / Motivation

The initial `/book-demo` implementation shipped with several UX rough edges that needed polishing before production deployment.

### Pain Points

- Single "Full Name" field provided unstructured data — poor for CRM, Cal.com prefill relied on fragile `split(' ')[0]` hack
- Optional "How can we help?" textarea added form height for marginal conversion value
- Cal.com embed scrolled internally on booking confirmation due to `overflow: scroll` fighting the iframe's auto-resize
- Error banner displayed raw server error text ("Validation failed") instead of empathetic copy when errors fell outside the `network`/`unknown` map
- Form submission to the non-deployed worker endpoint caused a visible flicker from the near-instant `idle → submitting → error` state transition

## Solution / What's New

### Form Restructure

New layout replaces the previous side-by-side groupings with a cleaner one-field-per-row design:

```
[First Name     ] [Last Name      ]
[Work Email                        ]
[Company                           ]
[Job Title                       ▼ ]
[Company Size                    ▼ ]
[          Book Your Demo          ]
```

- First Name + Last Name paired on the same row (natural grouping, industry standard for B2B forms)
- Every other field on its own row — gives each field breathing room in the narrow right-column form card (~420px content width)
- Message textarea removed entirely — fewer fields, higher conversion

### Cal.com Scroll Fix

Removed `overflow: scroll`, `height: 100%`, and `overflow-hidden` from the Cal.com embed container. The inline embed now auto-resizes via Cal.com's built-in postMessage height management. No more inner scrollbar on booking confirmation.

### Error Handling Hardening

- `submitForm` catch block now maps all non-network errors to `'unknown'` — raw server messages like "Validation failed" never reach the UI
- `ErrorBanner` fallback changed from `{ heading: 'Submission failed', body: rawError }` to `ERROR_COPY['unknown']` — defense in depth
- Every error path now shows the empathetic "Something went wrong" copy with the `hello@planton.ai` click-to-copy fallback

### Submission Flicker Fix

- Minimum loading duration (800ms) enforced via parallel sleep — the spinner always shows long enough to register visually before transitioning to success or error
- Error banner entrance animated with a `fadeIn` CSS keyframe (300ms ease-out with subtle 4px slide-up)
- Form submission decoupled from the form event: button uses `type="button"` with `onClick` for direct invocation, form `onSubmit` retained for Enter-key accessibility

## Implementation Details

### New Files

None — all changes are modifications to existing components.

### Modified Files

| File | Changes |
|------|---------|
| `src/components/book-demo/types.ts` | `fullName` → `firstName` + `lastName`, removed `message` from `DemoFormData` and `FieldErrors` |
| `src/components/book-demo/BookDemoForm.tsx` | New form layout, split validation, removed message textarea, hardened error handling, minimum loading duration, `type="button"` submission, fade-in error banner |
| `src/components/book-demo/BookDemoPage.tsx` | `formData.firstName` replaces `fullName.split(' ')[0]` hack |
| `src/components/book-demo/BookDemoScheduler.tsx` | Removed `overflow: scroll/hidden` and `height: 100%`, Cal.com name prefill uses `${firstName} ${lastName}` |
| `src/app/globals.css` | Added `@keyframes fadeIn` for error banner animation |

### Key Design Decisions

**One-field-per-row for the narrow form card**: Job Title has options like "Director of Engineering" (23 chars) that would truncate in a half-width dropdown at ~195px. Full-width dropdowns give every field breathing room and create clean vertical rhythm.

**Minimum loading duration over instant error**: When a fetch fails in <100ms, the rapid `idle → submitting → error` transition creates a jarring flicker. Holding the spinner for 800ms gives the user time to register "something is happening" before the outcome appears. The fetch still runs immediately — the delay is purely visual.

**Native `<form>` over MUI `<Box component="form">`**: MUI's polymorphic `component` prop can silently fail to attach event handlers depending on the version. A native `<form>` element eliminates the abstraction layer entirely.

## Benefits

- **Structured lead data**: First + Last name fields produce CRM-quality data without string-splitting hacks
- **Higher conversion**: Removing the optional message field reduces form height and friction
- **No scrolling on booking confirmation**: Cal.com embed auto-resizes naturally
- **Graceful error states**: Every error path shows empathetic copy with an actionable fallback
- **Smooth submission UX**: No flicker, proper loading state, animated error entrance

## Impact

- 5 files modified in planton.ai
- Backend contract updated in parallel (planton monorepo: `validation.ts`, `client.ts`, `submission.ts`, `submit.json`)
- Zero remaining references to `fullName` or `message` in the book-demo component tree

## Related Work

- **T01** (completed): `demo-request-receiver` Cloudflare Worker
- **T02/T03** (completed): `/book-demo` page and site-wide link migration
- **Backend contract**: `firstName`/`lastName` split and `message` removal applied to the worker in the planton monorepo

---

**Status**: ✅ Live (pending worker deployment and PR merge)

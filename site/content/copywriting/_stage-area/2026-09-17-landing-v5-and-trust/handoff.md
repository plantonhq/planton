# Handoff: Landing v5 and the Trust section

**Routes**: `/` (landing v5), `/trust`, `/trust/verified-before-deploy`, `/trust/rules-and-approvals`, `/trust/the-record`, `/trust/security-posture`, `/trust/your-cloud-your-keys`.

## What changed about the workflow

The copy for these pages is not in a draft file and not in components. It is data: `src/data/story.ts` (the thirteen chapters, each with claim, proof, and never-say), `src/data/trust.ts` (each Trust page's lede, proof points, illustrated record, honesty statements, doors), `src/data/personas.ts`, `src/data/testimonials.ts`, `src/data/platform-stats.ts`, and `src/data/pricing.ts`. The canonical narrative these mirror is `company/marketing/positioning/the-planton-story.md` in the company repository. A copy change is a data edit; the sections re-render.

The same session wrote the data and the sections, so the rendered page was the preview. No `draft-N.md` or `preview-N.html` was produced.

## The argument (read before touching a word)

The page is the story in order. The headline is chapter 1's first sentence (the visitor's coding agent already creates infrastructure; Planton makes it verifiable, recorded, reusable), never the category label; the category is the eyebrow. Every chapter section puts one claim beside one illustrated record of what the product stamps, captioned as an illustration. The Trust section is the buyer's half of the same story: one proof per page, then the platform's own honesty grammar said out loud ("What We Will Not Say"), then two sibling pages and a demo door.

## Component mapping (paths only)

- `src/components/landing-page/v5-2026-09-17-2100/` — one file per chapter; `index.ts` records what changed from v4 and why; `landing-page/index.ts` points here.
- `src/components/trust/TrustPage.tsx`, `TrustIndex.tsx` — the template and the index.
- `src/components/marketing/record-window.tsx` — the illustrated record primitive both use.

## Must not claim

Any dollar-savings figure; "compliant" of any component; retention as "forever" or "immutable" without the record page backing it; an analogy for the whole product; a customer quote without written approval on record (four exist); real product captures where none exist (every record window says it is an illustration).

## Verification

`make build` green with every guard; six rounds with the independent reviewer recorded in the session changelog; captures of the final round under `content/design-review/captures/2026-09-17/`.

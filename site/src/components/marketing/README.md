# `components/marketing`

The planton.app landing page, composed by `app/page.tsx` in **three movements**,
each with its own visual language:

1. **Skimmer top** — `Hero` (desktop-first, with a rendered product preview),
   `CatalogStats` (breadth numbers), `TrustStrip` (the fast "what's the catch").
2. **The conversation** — `Interlude` (prologue) → `Conversation` → `Interlude`
   (bridge) → `Origin`. The trade-off dialogue is **agree-first**: Planton
   affirms each tool, then names the one honest catch, and never says "we" inside
   the dialogue (the praise is of the tool, so the focus stays on the reader).
3. **Feature showcase** — `Features` (alternating `FeatureRow`s in the "your win"
   voice; the desktop app leads, the CLI is its companion), `Payoff` (trust +
   deploys), `WhereItFits` (positioning), `Horizon`, `Catch`, `FinalCta` (the
   "make the easy thing the right thing" close).

The marketing surface does NOT teach CLI install — the desktop launch experience
owns CLI setup. The CLI is acknowledged only as a companion in the `Features` row.

## Primitives
- `StoryStation` (`StationRail` + `StationBody`) — **numbered stations belong to
  the conversation ONLY**. Do not number the other movements.
- `Interlude` — centered narration between movements (no speaker, no number).
- `Line` — one speaker-labelled / toned paragraph (the atom of the dialogue).
- `FeatureRow` — an alternating copy/media feature row (movement 3).
- `Conversation` — the trade-off ladder as data (one `TRADE_OFFS` array, one
  renderer) so the rungs can never style-drift apart.
- Shared: `Section` (spacing), `Speaker` (eyebrow), `InlineCode` (command chip).

Product visuals come from `components/showcase`. Copy is lifted from the approved
messaging spec — change wording deliberately.

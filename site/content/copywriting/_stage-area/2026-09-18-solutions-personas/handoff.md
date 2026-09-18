# Handoff: the Solutions section as five persona pages and five persona decks

**Routes**: `/solutions` and `/solutions/{platform-engineer,engineering-leader,it-consultancy,startup-founder,security-and-governance-leader}`; `/decks/<same five>` (unindexed). Every `/solutions/by-*` path is retired and forwards to the page written for the person who would have read it.

## Where the copy is

`draft-1.md` in this folder is the human-readable record; `src/data/personas.ts` is the source the pages and decks render from, and the two say the same thing. Every sentence in a record is a persona's framing of a chapter of `src/data/story.ts` (the canonical narrative is `company/marketing/positioning/the-planton-story.md`, v1.2, in the company repository); proof sentences are the chapter's own, chosen by index. Doors come from `src/data/doors.ts` by key; quotes from `src/data/testimonials.ts` by the person's name; counts from `src/data/platform-stats.ts`; seat numbers from `src/data/pricing.ts`. Titles and descriptions are in `src/data/site-pages.ts`.

There is no `preview.html`. The built page under the capture harness is the preview, and the independent reviewer grades pixels, not a mock-up.

## The argument

A visitor is a person, not a company size or a use case, so the three old axes become five people. Each page tells the story in that person's order at that person's depth: the wall they hit, the two or three proof points that decide it for them, the chapters that matter most to them as full sections with the proving page's record beside each, the rest as doors, the objections they will raise answered where they arise, then the people and the numbers, then the doors. The deck is the same record with every chapter as a slide, the objections as one slide, the roadmap with its disclosure, and the close.

## Component mapping

- `src/components/marketing/` gained `ChapterSection` (from the landing) and `PlatformCounts` (from the landing's proof section) so a persona page composes the same chapter frame and counts strip the home page does.
- `src/components/solutions/PersonaPage.tsx`, `SolutionsIndex.tsx`; `src/components/decks/PersonaDeck.tsx` and its slides on `src/components/deck/` (the engine gained `bindSlide` and a palette-native slide frame).
- `packages/website-shell/src/data/navigation.ts`, `DesktopNav.tsx`, `MobileNav.tsx`: the Solutions menu is one list of five.

## Revisions made on the cold read

Before any code, each record was read as the visitor it names and as a copywriter for developer-tools sales pages. Four changes: a "two in the morning" phrase appeared in a deciding proof and again in a beat on the platform engineer's page (the beat now says "the deploy nobody watched"); "complement, not replacement" appeared three times on the security leader's page (a deciding proof became the record, and an objection became the estate-coverage question, so the point is made once, in the compare beat); the engineering leader's savings-figure answer ends on the reader rather than the page; the record claims name their chapter beside each sentence.

## Revisions forced by the independent review (three rounds)

`src/data/personas.ts` is the record; `draft-1.md` is the draft it started from. Where the two differ, the review is why: the story's chapter 11 overclaimed ("rules written once over a fixed list of controls" is roadmap), so the story went to v1.3 and the platform engineer's abstraction-layer answer says what ships (typed schemas over open-source modules, controls reported against one fixed list). Chapter 1 is no longer a section on any page: the persona's wall is chapter 1 told for that person, and the hero and the deck cover carry it. The platform engineer's doors are the site's user pair (Start Free first). Every hero says what Planton is (the umbrella sentence) under its doors, and the index addresses the developer with a coding agent open on the first screen. The engineering leader "will not live in a console" and the console objection answers with the record and a sign-in. The consultancy's third question is cost, not repeatability (the deciding proof already covers it). The platform engineer's quotes are Sai Saketh and Rakesh Kandhi, the two sides of the page's promise. Every objection card has a door.

## Must not claim

Any dollar-savings figure; "compliant" of any component or deployment; a framework verdict; anything from chapter 13 on a page (rules over spec content, refusal records, estimate-versus-actual, service health, image scanning, the account scan, the mobile companion); a paraphrased quote or a quote from a persona we have not sold to; a competitor's name; a price as a literal in prose; a per-minute runner rate.

## Verification

`make build` green with every guard; the landing and every untouched page byte-identical under the capture harness; the independent reviewer's sales-lens verdict on the index, two persona pages, and one deck, recorded in the company repository's session changelog.

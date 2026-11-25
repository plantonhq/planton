# Handoff: the hero, the first app and the close

**Route:** `/`. **Composition:** `src/components/landing-page/v7-2026-11-25/` (v6 retired). **Words:** `src/data/homepage.ts`: `HOMEPAGE.headline`, `intro` and `caption` (the hero), `firstApp` (the first section), `close`; `SETUP_PROMPT_EXAMPLE` (the server's words, `SetupAgentPrompt.render`, for the demo company Slotwren). The diagram's words stay in `src/data/homepage-experience.ts` (`HERO`).

**The argument:** Planton's purpose is to help a company run its cloud without a DevOps team (story v3.0, chapter 1). The hero says so beside the diagram of the team it serves. The first section proves the first beat with the exact words the visitor's coding agent will read, and says plainly what stays theirs to do. The close ends on what they keep: their app, their cloud, their bill.

**Components:** the hero reuses `HeroExperience` and `HeroScene` unchanged. New: `components/marketing/SetupPromptExample.tsx` (renders the example: the code as a marked placeholder, the ask-first sentence in full ink, no copy button). Doors stay `ExperienceLink` and `DemoLink`, attributed `hero` and `close`.

**Revisions forced by the reads:** listed in `draft-1.md`, "The cold reads".

**What the page must not claim:** a cloud other than Google Cloud beside "running app"; that Codex or Cursor are proven end to end; that it is hands-off; any savings figure; contempt for people who do DevOps work.

**Revisions forced by the design review** (three rounds: 15 findings, then 11, then 7): the prompt spans the section with hanging indents, paths and placeholders that never break, and lines capped at a readable length; the section header and the lists share the hero's two columns; the close sits at the section-heading scale; the price ("Free for up to 3 seats. Never asks for a card.") replaced the slogan under the hero's doors; the hero line was shortened; the first section's intro stopped repeating the hero; the close's line says what "your own bill" means; the example is labelled as a demo company's; list labels are Title Case headings.

**Standing findings, with owners:** the page proves the input, not the result (major, open: the founder decides it before the deploy; the result arrives with the next sections and the hero film); the published design law's weights and dark-only palette table, and the social card's missing name and label (the frame session, when the law moves out of `/branding`); two soft line endings (accepted).

**Verification:** `make build` (every guard, including the new page-weight budget); `node scripts/check-homepage.mjs` (80 checks); the design reviewer at 1280, 1680 and phone; the founder reviewed it on localhost. It waits on the branch `website-one-page` until he says deploy (founder, 2026-11-25).

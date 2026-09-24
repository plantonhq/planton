# Handoff: the Compare page

**Route**: `/compare` (group `compare`, indexed). Reached from the Product menu's Explore column ("How Planton Compares"), the landing's close (one door under chapter 11's three contrasts), and the security leader's chapter-11 beat. The footer entry rides the shell slice that rebuilds the footer.

## Where the copy is

`draft-1.md` in this folder is the human-readable record; `src/data/compare.ts` is the source the page renders from. Every sentence is chapter 11's, another chapter's proof sentence quoted by index, or a fact a page of this site already states; the chapter is named beside each. The proving pages' records are shown beside the words through `src/data/artifacts.ts`, so the Compare page draws no illustration of its own. There is no `preview.html`; the built page under the capture harness is the preview, and the independent reviewer grades pixels.

## The argument

The visitor already runs a posture tool, a Terraform pipeline, or a portal, and asks two things: what does Planton do that these do not, and do I throw anything away. The headline is the spine (proof at creation, not observation after); the lede answers both questions in two sentences and says "you keep them"; three routes under the doors take the reader to their tool. Each category section says what the tool does as a fact, what the reader keeps when they run both (the section's largest sentence after its title), what Planton does at the same moment as three labeled proof points, and shows the proving page's record beside it. The questions are the ones people ask in demos (question bank entries 1, 7, and the abstraction-layer objection), answered in the chapters' words with one concession said plainly (nothing self-heals).

## Component mapping

- `src/components/compare/ComparePage.tsx`: `PageHero`, a split "Where the Difference Is" with the stack-job record, one `CategorySection` per record (text column and `PageArtifact`, alternating sides, phone order title, tool, keep, record, proofs, door), `QuestionCard` grid, `PageCard` pair, `Doors`.
- `src/components/marketing/question-card.tsx`: lifted from the persona page's objection card so both pages render a question the same way; `PersonaPage` resolves the door and renders it.
- `src/data/page-shapes.ts`: `illustratedFooter`, the one footer vocabulary every illustrated record now uses (the review found four phrasings for one fact and an unexplained `est.`).
- `src/data/site-pages.ts`: the `compare` group and `PAGE_GROUP_HEADINGS`; `scripts/generate-llms.mjs` iterates the headings and refuses a group they omit, and prints the Compare record's own words.
- `packages/website-shell/src/data/navigation.ts`: "How Planton Compares" in the Explore column.

## Revisions made on the cold read

Before code: "the schema refuses" became "validates" (a refusal record is roadmap); "kept forever" became "kept and queryable"; "the writing is still yours" was praise and now reads as the cost it is; the portals description lost a sneer.

## Revisions forced by the independent review (three rounds: FAIL 13, FAIL 12, PASS 6 minor)

The headline was a label ("How Planton Compares") and became the spine; the lede described the page's method and now answers the visitor's question; the umbrella sentence left the hero (the kicker and the lede carry what Planton is). "Where the Difference Is" was one 49-word sentence with no proof and now shows the stack-job record beside a two-sentence claim. Every category section gained the proving page's record; the "keep it" sentence moved above the proof points and up a size. The story went to v1.4: chapter 11 said every component states its controls where chapter 3 says every covered component does, and the review saw both sentences on one page. The engines are said one way (the Terraform module, which OpenTofu or Terraform runs; and Pulumi), matching the Open Source record's row. Repeated sentences were removed (typed schema, adopted not rewritten, the exit path, "with the record attached"). Every illustrated record's footer converged on one vocabulary that explains `est.`. The abstraction-layer question's door moved from Import to Catalog; its garbled sentence was rewritten with three concrete nouns.

## Standing after PASS, with owners

The record window's height against a taller text column (the chapter frame's law; a landing review); the control profile's exposed row with no mark and its `verdict · none` sentinel (the Security Posture record's own; the Trust records' convergence); `est.` on the same line as "verified" resolved only by the footer (the shared record vocabulary); the header's "Sign up" beside the hero's "Start Free" (the shell slice).

## Must not claim

A vendor's name; "DevSecOps"; "FinOps"; a table with a competitor column; "policy as code"; rules over spec content, refusal records, drift detection, or anything else from chapter 13 as shipped; "compliant" of any component; a dollar-savings figure; "zero lock-in" as a badge.

## Verification

`make build` green with every guard (the llms generator now also refuses a page group missing from the index); capture compare against the branch tip's build: every scene byte-identical except the three landing scenes (the close's door and chapter 11's corrected proof sentence), the platform engineer's page (the corrected objection), the security leader's page (the beat's door now reads "How Planton Compares"), and every page whose record footer converged; the reviewer's PASS recorded in the company repository's session changelog.

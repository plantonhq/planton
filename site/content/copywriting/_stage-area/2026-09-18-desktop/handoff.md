# Handoff: the desktop landing and download pages rebuilt from the story

**Routes**: `/desktop` and `/desktop/download` (Planton Desktop's permanent address; the `/features/desktop` paths are retired routes).

## Where the copy is

`draft-1.md` in this folder is the human-readable record and the claim check; `src/data/desktop.ts` is the source both pages render from. Every sentence names its chapter or document in a comment. What was already data stays where it was: the desktop line and its seven points in `positioning.ts`; the platforms, installers, install steps, verify commands, and the ways line in `desktop-download.ts`; seats in `pricing.ts`; counts in `platform-stats.ts`; the screenshot slots in `desktop-screenshots.ts` (all `null` until real captures exist). There is no `preview.html`; the built page under the capture harness is the preview.

## The argument

Unchanged from 2026-09-11: the reader already has two ways to do this (a platform that hides the cloud; their agent and the cloud CLI) and Planton is the honest third, on their machine, in their account, with rules the agent obeys and a record it leaves. What changed is where the words live and what the first screen proves: the headline is the story's derived two-sentence line, the sourced platform counts sit under the doors, the seven "what it adds" points are their own beat, the exit path is stated where a person deciding to install asks it, and every number carries its provenance on the page.

## Component mapping

- `src/components/desktop/DesktopLanding.tsx`: `PageHero` (the eyebrow to Distributions, the derived headline, "Planton Desktop" as kicker, the lede, the ways line, the doors), `PlatformCounts`, then sections on `Section`, `Card`, `Grid`, `FeatureTitle`, `BodyText`, `Metric`, `CommandBlock`, `PageCard`, `Doors`. No motion.
- `src/components/desktop/DesktopDownload.tsx` (the one platform selection), `DownloadHero.tsx` (on `PageHero`; the platform tabs, the platform card with compact one-line commands, the unavailable card), `AfterInstall.tsx` (the three beats; the follow-ons on `CommandTabs`, hidden below `sm` like every command a phone cannot run), `Verify.tsx` (the verify card, Updates as a centered paragraph, the two footer lines), `Screenshot.tsx`, `use-detected-platform.ts`.
- `src/components/marketing/command-tabs.tsx` (new): several things a person might paste, of which they pick one. `command-block.tsx` gained a `compact` density for a one-liner inside a list of steps.
- Deleted: `components/product/desktop/` (14 files) and `components/product/shared/` (8 files); `components/product/` holds the Product template and index only.
- `scripts/capture-pages.mjs`: the poster CSS no longer hides a hero whose eyebrow is a link, and hides the strip under a hero; both desktop posters republished.

## Claim check

Every sentence was read against the story (chapters 1, 2, 8, 12) and the desktop record (the CI/CD-on-your-laptop guide, the coding-agents and connections guides, the local-runtime, artifact-distribution, desktop-identity, and credential-autodetect knowledge articles, the desktop wiki). All held, including the two suspected before the session: the free device lane the assistant runs on is Planton-billed, and bring-your-own-key is the privacy choice in the demo record's own words. The "Read Next" block's 2025 sentences were the one thing the page said that the site no longer says; it now reads the registry.

## Revisions forced by the independent review (two rounds: FAIL 16, PASS 8 minor)

The H1 was the 21-word desktop line and became the story's derived two sentences; the first screen had no proof and gained the sourced counts strip; the assistant card was 60% empty beside the seven-point list and both cards are now short and equal with the points as their own beat; the ways line said macOS twice; the metric labels were lowercase and the 800 MiB figure had no provenance sentence; the counts card had no source or door; "the free 3" read as broken; "One Craft" and "rails you can inspect" were metaphors; the exit path was unstated; Updates was a half-empty card; the Linux card's one-liners sat in full terminal frames; the footer line held two unrelated questions; the agent command rendered on a phone. Four minors closed after the round: the Homebrew claim ("installs from Homebrew too"), one noun for the count ("Component Kinds"), the Checksums link's visibility, the Updates paragraph's alignment.

## Standing after PASS, with owners

The shared hero's title measure narrower than its lede (the hero primitive's law); the first screen's missing artifact (real captures, the founder's standing ask); the header's "Sign up" beside a lede that says no account (the shell slice); the SmartScreen install step rendered as prose rather than a callout (the download data's step shape gaining a warning kind; a small later slice).

## Must not claim

A platform whose installer is off; a one-command install that is `pending`; a price as a literal; the monthly free AI allowance as live; a vendor's name; anything from chapter 13 as shipped; a screenshot that is not a real capture.

## Verification

`make build` green with every guard; capture compare against the branch tip: exactly the ten desktop and download scenes differ, every other scene identical; the reviewer's PASS recorded in the company repository's session changelog.

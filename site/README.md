# planton.ai — Website and docs

This is the source for the planton.ai website, living in the `site/` folder of the Planton open-source repository. It contains:

- All written content and visuals for the public website: Home, Security, Pricing, Planton Desktop, the docs, the changelog and the legal pages
- File system–based content for the docs and the changelog (Markdown/MDX) alongside the components, templates, and logic used to render them
- All the components, styling, and configurations used to run the site as an application

## Table of contents

- [What this is](#what-this-is)
- [Quick start](#quick-start)
- [How the site is built](#how-the-site-is-built)
- [The laws](#the-laws)
- [Where things live](#where-things-live)
- [Guards that run in the build](#guards-that-run-in-the-build)
- [Proving a change](#proving-a-change)
- [Rules for agents](#rules-for-agents)
- [Health ledger](#health-ledger)
- [Contributing](#contributing)
- [Media management](#media-management)
- [Home](#home)

## What this is

The public website at planton.ai: a Next.js static export, deployed on every push to `main` that touches `site/` to the Cloudflare Worker `planton-website` (`wrangler.jsonc`; static assets only, no code). The push is the deploy, so the team's work lands on `main` and is pushed only after the founder has approved the page's after (its before/after); a contributor's pull request is merged the same way.

**Every pull request that changes `site/` gets a preview** at `https://review-<pull request number>.planton.dev`, the exact build production would serve. Every push updates it, the pull request's one "Website preview" comment links it, and it closes when the pull request is merged or closed. `.github/workflows/preview.site.yaml` builds it with no key, for every pull request including forks; `.github/workflows/publish.site-preview.yaml`, which runs from `main` only, uploads it. A contributor's pull request from a fork gets its preview once a maintainer approves the `website-preview-fork` deployment in its run, once per push; until then its comment says the preview is waiting. Previews carry `noindex`, and a link into the console (`/login`, `/signup`) shows the site's 404 there: the console is tested on planton.ai.

The site shares its domain with the console. A router at the edge (planton-platform's `apex-router`) hands a declared list of top-level path segments to this site; every other path goes to the console, which reads it as an organization's name. A page can build and deploy and still be unreachable at planton.ai if its top-level path is not on that list. The build has a guard for it (below).

The jobs that deploy hold a Cloudflare key, so they never install the website's dependencies: `site/deploy` holds only the pinned wrangler they run, and each key reaches its one Worker alone.

## Quick start

```bash
nvm use            # Node from .nvmrc
corepack enable    # Yarn from package.json's packageManager field
yarn install
yarn dev           # builds the website-shell package, generates the desktop release pointer, starts Next
```

`make build` is what CI runs: install, lint, typecheck, then the full build with every guard, then the asset-prefix copy. Run it before every push.

## How the site is built

The site is data first. Every sentence that states what Planton is or does lives in `src/data/`, and pages render it:

| File | Holds | Read by |
|------|-------|---------|
| `src/data/story.ts` | The thirteen chapters of the Planton story: claim, proof, never-say, its spine, and the stand (`STAND`, Home's headline and the 404's sentence). Mirrors `company/marketing/positioning/the-planton-story.md` in the company repository. | `homepage.ts`, `not-found.ts`, the registry's chapters, `llms.txt` |
| `src/data/homepage.ts` | Home, one record per section in page order, each naming its chapter; the setup prompt example; each cloud's agent run; the booking page's words. | `src/components/landing-page/`, `src/components/book-demo/`, `llms.txt` |
| `src/data/positioning.ts` | What Planton is in one sentence (never an analogy), the first action under it, and each hub's one plain line. | `story.ts`, `desktop.ts` |
| `src/data/security.ts` | The Security page: each question a reviewer asks, its answer, the real artifact behind it (a permission list, a trust policy, a table), and how to check it; each sentence names its chapter and claims-ledger row. | `src/components/security/SecurityPage.tsx`, `llms.txt` |
| `src/data/desktop.ts` | Planton Desktop's landing and download pages: every section's words, the measured figures with their provenance, the first-launch beats and follow-ons; each sentence names its chapter or document. | `src/components/desktop/`, `llms.txt` |
| `src/data/desktop-download.ts` | The one source for what can be downloaded: platforms, installers, install steps, one-command installs, verify commands, the CDN pointers, the platform detector, and the ways line derived from them. | `src/components/desktop/`, `doors.ts` |
| `src/data/desktop-screenshots.ts` | The desktop pages' screenshot slots: a real capture of the app or `null`, never a placeholder. | `src/components/desktop/` |
| `src/data/doors.ts` | Every call to action once: label and destination by key; Start Free leads every pair. | the `Doors` primitive; a record names its pair |
| `src/data/not-found.ts` | The 404's words and doors; the stand read from `story.ts`. | `src/app/not-found.tsx` |
| `src/data/artifact.ts` | The shapes of the real things story pages show beside their words: a code artifact, a table, and the check a reader can run. | `components/marketing/Artifact.tsx`, `homepage.ts`, `security.ts`, `llms.txt` |
| `src/data/clouds.ts` | The clouds Home's picker offers, in order, and the default. | `CloudChoice`, `llms.txt` |
| `src/data/site-pages.ts` | The route registry: every non-content page with its title, description, group, chapters, and index flag; and the heading each group gets in `llms.txt` (`PAGE_GROUP_HEADINGS`), so a new group cannot be forgotten by the index. | `sitemap.ts`, `robots.ts`, `lib/page-metadata.ts`, `generate-llms.mjs`, `check-apex-routing.mjs` |
| `src/data/pricing.ts` | Every price, cap, and free tier. | the pricing page and any sentence that names a price |
| `src/data/platform-stats.ts` | The counts a page prints, with how each was counted. | `PlatformCounts`, `llms.txt` |

Data files import each other with `.ts` extensions so Node can run them directly (the generators do); `tsconfig.json` allows it.

Components come from one library, `src/components/marketing/`, on the palette's role classes (`bg-canvas`, `bg-card`, `text-fg-secondary`, `border-edge`, `text-ok`). The library holds the primitives (card, badge, buttons, metric, step) and the page sections Pricing and Desktop compose (`PageHero`, `PageCard`, `Doors`, `CommandBlock`, `CommandTabs`, `PlatformCounts`); a page template is a thin composition of those in its reader's order and states nothing. Home and Security are the story pages: each has its own CSS module, and Security's composes Home's (`components/marketing/homepage.module.css`) rather than copying it. The website's palette is `src/theme/homepage.ts`, published as CSS variables on every page; the shell's own dark appearance is `packages/website-shell/src/theme/tokens.ts`; a component never types a hex. The design law, which the design reviewer grades every page against, is `design-system.md` beside this README.

The landing page is versioned: `src/components/landing-page/v<N>-<date>/` holds only that version's section composition, and `src/components/landing-page/index.ts` points at the active one. Only the active version stays in the tree; earlier ones live in git history. Nothing outside a versioned folder imports from inside one.

Decks run on one engine, `src/components/deck/` (hash navigation, keyboard, touch, presenter notes). Meeting decks are hand-written slides on the engine's slide kit, filed under `src/components/meetings/`.

## The laws

1. **A claim lives in `src/data/`, never in a component.** If you find yourself typing a sentence about the product into a component, stop and put it where the other claims live, naming its chapter.
2. **Public sentences are verified.** No dollar-savings figure. No component called compliant (a component enforces controls). No analogy for the umbrella; each hub has exactly one. Testimonials verbatim and attributed to a person. Competitors described, never named. Nothing unshipped presented as live; illustrated records say they are illustrations and mark example figures `est.`.
3. **Numbers come from `platform-stats.ts` and prices from `pricing.ts`.** A literal in prose is a defect.
4. **Colors come from the palette.** Role classes only; no hex in a component; semantic hues only where they carry meaning.
5. **Every page is registered.** A new route goes into `site-pages.ts` or the build fails; a removed page leaves the registry with it, and the site's 404 answers its address.
6. **Every new top-level path is three declarations**: the registry, the router's website list (`site_roots` in planton-platform's `infrastructure/desktop/Infra.foundation.InfraStack.foundation-apex-router.yaml`, applied before the page merges), and the platform's reserved handles. The apex guard names what is missing.
7. **Nothing merges without the founder.** Build, lint, typecheck, the guards, the screenshot compare for a zero-visual-change commit, and a design review that reads the page as the visitor and as a copywriter.

## Where things live

```
src/app/(site)/          pages inside the website shell (header, footer, theme)
src/app/(standalone)/    surfaces without the shell: meeting decks, book-demo, the desktop handoff, investor pages
src/app/not-found.tsx    the 404 every address without a page answers, inside the shell
src/app/sitemap.ts       generated from the registry and the content folders
src/app/robots.ts        generated
src/components/marketing/   the primitive library
src/components/landing-page/ the versioned landing compositions
src/components/security/ the Security page and its CSS module
src/components/desktop/  the Planton Desktop landing and download templates and their sections
src/components/content/  the docs' author and table-of-contents pieces
src/components/deck/     the deck engine and the meeting decks' slide kit
src/components/meetings/ meeting decks and their registry
src/data/                every claim, price, count, route
src/lib/                 page-metadata, content-routes, assets, console-handoff
scripts/                 the guards and generators (below)
content/                 legal markdown, copywriting stage areas, design-review board and captures, the image inbox
public/docs, changelog   content pages, walked by the sitemap and llms generator
packages/website-shell/  the header, footer, navigation, palette, and theme the console also consumes
```

## Guards that run in the build

| Script | Proves |
|--------|--------|
| `scripts/check-displayed-vs-enforced.mjs` | every displayed plan limit and entitlement matches what the platform enforces |
| `scripts/check-apex-routing.mjs` | every top-level path is in the router's website list and reserved as a platform handle (reads the sibling `planton-platform` checkout; skips loudly without it) |
| `next build` with `tsc --noEmit` and `eslint --max-warnings 0` before it | types hold, and the accessibility, image, and no-`any` rules hold everywhere with zero warnings |
| `scripts/check-internal-links.mjs` | every internal link in `out/` resolves to a page or a static file, |
| `scripts/check-page-weight.mjs` | Home's HTML, scripts and stylesheets stay within their gzipped budget |
| `scripts/generate-llms.mjs` | `llms.txt`, `llms-full.txt`, and one Markdown per marketing page from the same data; fails when an exported route is unregistered |

## Proving a change

- `make build` green.
- For a change that must not alter pixels: capture a before set from a `main` build (`node scripts/capture-pages.mjs --export /path/to/main/out --out /tmp/before`), then `make capture OUT=/tmp/after COMPARE=/tmp/before`. The harness renders under a virtual clock so two builds of the same page produce byte-identical PNGs; a difference is a change.
- For a new or rebuilt page: captures at 1280 and 1680 go to the independent design reviewer with the page's first-glance question and the named visitor. PASS means nothing above minor. The reference board it grades against is `content/design-review/reference-board/`.

## Rules for agents

Rule files live beside what they govern:

- `_rules/docs/` — writing and formatting documentation pages
- `content/copywriting/_rules/` — changing what the site says: a draft with every sentence beside its source and a handoff, then the data edit, the build, the captures, and the review
- `content/assets/_rules/` — images through the R2 pipeline
- `public/docs/_rules/` — docs content rules
- `src/app/(standalone)/meets/_rules/` — creating and updating a meeting deck

## Health ledger

What a day-one architect would not have done, and which work retires it. The build carries no warnings: lint runs with `--max-warnings 0` and there is no legacy allowlist, so a rule holds in every folder or the build fails. An image the build cannot size (a document's own picture, a record's avatar) is a plain `<img>` with a one-line exception at the site saying why; everything else here is debt the ledger names.

| Legacy | Where | Retired by |
|--------|-------|-----------|
| Hex literals in components; MUI icons beside lucide; `'use client'` on components with no hooks | `pricing`, `docs`, `content` (the docs' author and contents pieces), `changelog`, `common`, `book-demo`, `legal`, `invest` | each of these page groups' own rebuild (pricing is data-driven and refreshed 2026-08-20; the content pages are the docs follow-up's; book-demo and legal are small standalone slices; invest is named below) |
| The investor deck and explainer on their own primitive set (including the deck's own button, which came from the retired tour's kit), and the one Tailwind color the pricing FAQ still reads (`text.secondary`) | `components/invest`, `tailwind.config.ts`, `components/pricing/faqs.tsx` | invest onto the deck engine, by the founder's decision a later slice; the color with pricing's rebuild |
| The meeting decks' slide kit types its own colors and gradients (`deck/primitives.tsx`, `meets/meets.css`) | `components/deck/primitives.tsx`, `app/(standalone)/meets/meets.css` | the meeting decks' convergence onto the marketing primitives |
| Images under `public/_site/` instead of the asset CDN; 87 literal `/_site/` paths, in the investor pages, the meeting decks, and the content pages | `public/_site/images`, `src/**` | invest and the meeting decks with their convergence slices; the content pages with the docs follow-up; the Open Graph posters and the logo stay under the prefix `lib/assets.ts` names once |
| The `/_site` asset prefix and the post-build copy of the bundle | `next.config.ts`, `Makefile` | the apex routing decision |
| MUI's styles ship in `@layer mui` and Tailwind v3's preflight ships unlayered, and an unlayered rule beats a layered one, so every border MUI declares (`sx` borders, `Divider`, outlined `Paper`) computes to zero on this site; the shell draws its edges with inline styles for that reason, and the footer's dividers and pricing's outlined panels are lost | `src/app/globals.css`, `src/providers/theme.tsx`, every MUI border on the site | Tailwind v4, whose border utilities carry their own style and whose layers order `theme, base, mui, components, utilities`; lowering the v3 preflight below MUI was tried and breaks every Tailwind border on a MUI button and every `all: unset` heading, so it is not the fix |
| The Windows install step that warns about SmartScreen is rendered as prose under the button it concerns; the download data's install-step shape has no warning kind | `src/data/desktop-download.ts`, `components/desktop/DownloadHero.tsx` | a `kind: 'warning'` on the step and a callout in the card, one small slice |

## Contributing

Contributions big and small are welcome. Follow the quick start to run the site locally; for a content edit, open the file on GitHub and use the edit icon; for bugs or content ideas, open an issue at `https://github.com/plantonhq/planton/issues/new`. Keep a change focused, say what it does in the PR description, and for docs and blog posts prefer copy-pastable examples with exact paths or commands. Commit and PR conventions follow the repository-wide rules under `_rules/git/` and `_rules/pull-requests/`.

## Media management

We store screenshots and image assets in Cloudflare R2, mirrored from the `content/` folder. This keeps the repo lean, provides CDN delivery, and ensures predictable URLs.

### Architecture

```
content/
├── assets/
│   ├── _inbox/           # Drop zone for raw images (gitignored)
│   ├── _rules/           # Cursor rules for asset management
│   └── images/           # Organized, processed images
└── ...

↓ syncs to ↓

assets.planton.ai/site/
├── assets/
│   └── images/           # Same structure as local
└── ...
```

**Mental model**: Everything in `content/assets/` is available at `https://assets.planton.ai/site/`.

### Adding images (recommended workflow)

1. **Drop images** in `content/assets/_inbox/`

2. **Invoke the cursor rule** with context:
   ```
   @process-planton-ai-images
   
   I've added screenshots of the deployment pipeline feature for Service Hub docs.
   ```

3. **The tool processes images**:
   - Compresses (jpegoptim for JPEG, pngquant for PNG)
   - Renames following convention: `YYYY-MM-DD-HHMMSS-context-description.ext`
   - Moves to appropriate folder
   - Uploads to R2

4. **Use the R2 URL** in the docs:
   ```md
   ![Pipeline overview](https://assets.planton.ai/site/images/service-hub/2025-12-31-103045-service-hub-pipeline-overview.png)
   ```

### One-time setup

```bash
# Install CLI tools
brew install jpegoptim pngquant awscli

# Create virtual environment and install Python dependencies
python3 -m venv .venv
source .venv/bin/activate
pip install -r tools/image_processor/requirements.txt

# Configure R2 access (get credentials from Cloudflare dashboard)
aws configure --profile r2
```

### Manual commands

```bash
# Activate virtual environment first
source .venv/bin/activate

# Check prerequisites
make check-images

# List inbox contents
make inbox

# Sync assets to R2
make sync-assets

# Process images with specific context
python -m tools.image_processor process -c "service-hub" -d "feature-name"
```

### Naming convention

```
YYYY-MM-DD-HHMMSS-context-description.ext
```

Examples:
- `2025-12-31-103045-service-hub-deployment-pipeline.png`
- `2025-12-31-143500-kubernetes-dashboard-pod-logs.png`

### Folder structure

| Local Path | R2 URL |
|------------|--------|
| `content/assets/images/service-hub/x.png` | `https://assets.planton.ai/site/images/service-hub/x.png` |

### Troubleshooting

- **Preflight check fails**: Run `make check-images` to see what's missing
- **R2 access denied**: Verify `aws configure list --profile r2` shows credentials
- **No images found**: Ensure images are in `content/assets/_inbox/` (not a subdirectory)

For detailed documentation, see `content/assets/_rules/README.md`.


## Home

Home is one page that tells one story, built one or two sections at a time; the active composition is `src/components/landing-page/v7-2026-11-25/`. Its words live in `src/data/homepage.ts`, one record per section in page order, each naming the story chapter it mirrors; Home's markdown and the llms.txt introduction render the same records. Each section's draft, cold reads and handoff live under `content/copywriting/_stage-area/` (the hero and the first app: `2026-11-25-hero/`; infrastructure: `2026-11-25-any-infrastructure/`).

- **The hero** is the stand and the team it serves. `HeroExperience` draws the team, Planton and your cloud with a pure `HeroScene`, labelled as an illustration; its words are `HERO` in `src/data/homepage-experience.ts`. Its clock is `workflows/useWorkflowPlayback.ts` over `workflows/timeline.ts`: hidden and offscreen time is excluded, reduced motion shows the finished state, and two GA events (`workflow_explainer_view`, `workflow_explainer_cycle_complete`) carry only the story id, version and motion mode. Its marks are embedded by `scripts/generate-workflow-icons.mjs` from the canonical SVGs, and the build checks them for drift.
- **The first app** shows the setup prompt as the server writes it (`SETUP_PROMPT_EXAMPLE`, rendered by `SetupPromptExample`): words to read, with the sign-in code as a marked placeholder and no copy button, because the real prompt carries a personal single-use code. When the server's wording changes, the example changes with it.
- **Infrastructure** shows what a real coding agent wrote when a person at a demo company asked it for a database, on the visitor's cloud (`AGENT_RUNS`, rendered by `InfrastructureSection`): the ask, the set's main file verbatim, the check it ran, and every kind in the set with its open-source module, deployed and torn down in a real account. Each cloud's record comes from a real run (the released CLI and skills, signed in to nothing); when a kind's fields or the CLI's output change, the run is repeated and the record replaced. The section carries the cloud picker (`CloudPicker`, `src/data/clouds.ts`): the visitor's pick is remembered in their browser and every section below reads it through `useCloud`. A cloud joins the picker when its run exists.
- **The close** repeats the doors.

`ExperienceLink` renders Start Free, the primary door in every pair, and `DemoLink` the demo door beside it; each records its placement (`hero`, `close`). All public website routes share the light appearance and one header, whose Start Free is attributed as `navigation`.

After `make build`, run `node scripts/check-homepage.mjs`: responsive captures, the hero's reduced-motion state and labels, the setup prompt, the infrastructure section on every cloud (the pick, its memory across a reload, a 320 px phone), the door attribution, the mobile drawer, and the mocked booking journey (CI runs it on pull requests). `node scripts/preview-homepage.mjs` previews the export at port 4177; `HOMEPAGE_CAPTURES` changes the capture directory. Generate the social card with `node scripts/capture-homepage-og.mjs` and commit the PNG: the hero, with the header's mark and the site's name set above it, since the card hides the header. It is also the default card for every page without its own (`SITE.defaultOgImage`).

`scripts/check-page-weight.mjs` runs inside the build: it sums Home's HTML and the scripts and stylesheets it references, gzipped, against a budget written in the script with its reason. A section that needs more raises the budget in the same change, with the reason beside the number.

Earlier landing versions are in git history, not in the tree. Other pages do not import version-specific components.

## Security

`/security` answers what a reviewer asks before anyone connects a cloud: what Planton can touch, that it stores no keys, where the work runs and what Planton keeps, what the GitHub App sees, who approves what, the record, and how to cut Planton off. Each answer sits beside the real thing behind it (a catalog permission list, the trust policy the setup writes, a table) and a way to check it. Its words live in `src/data/security.ts`, each naming its chapter and claims-ledger row; the page's Markdown twin renders the same records. Its draft and cold reads are in `content/copywriting/_stage-area/2026-11-25-frame/`. When a fact behind an answer changes (the GitHub App's permissions, a cloud's revoke step), the record changes with it, and the claims ledger is re-checked first.

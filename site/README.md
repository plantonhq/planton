# planton.ai — Website, docs, and blog

This is the source for the planton.ai website, living in the `site/` folder of the Planton open-source repository. It contains:

- All written content and visuals for the public website: product pages, docs, blog, case studies, tutorials, and marketing pages
- File system–based content for docs and blog (Markdown/MDX) alongside the components, templates, and logic used to render them
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

## What this is

The public website at planton.ai: a Next.js static export deployed to GitHub Pages on every push to `main` that touches `site/`. The push is the deploy, so all work happens on a branch and merges only when the founder approves.

The site shares its domain with the console. At the edge, a fixed list of path prefixes passes through to this static site; every other path goes to the console. A page can build and deploy and still be unreachable at planton.ai if its top-level path is not on that list. The build has a guard for it (below).

## Quick start

```bash
nvm use            # Node from .nvmrc
corepack enable    # Yarn from package.json's packageManager field
yarn install
yarn dev           # builds the website-shell package, generates the desktop release pointer, starts Next
```

`make build` is what CI runs: install, lint, typecheck, then the full build with every guard, then the asset-prefix copy. Run it before opening a PR.

## How the site is built

The site is data first. Every sentence that states what Planton is or does lives in `src/data/`, and pages render it:

| File | Holds | Read by |
|------|-------|---------|
| `src/data/story.ts` | The thirteen chapters of the Planton story: claim, proof, never-say. Mirrors `company/marketing/positioning/the-planton-story.md` in the company repository. | the landing page, the Trust pages, `llms.txt` |
| `src/data/positioning.ts` | The vocabulary law: the umbrella tagline (never an analogy) and each hub's one line and one analogy. | `story.ts`, the hero, the hub cards |
| `src/data/personas.ts` | The five people the story is told to, with their chapter order. | the persona router and cards; the persona pages and decks when they land |
| `src/data/trust.ts` | The five Trust pages: lede, proof points, the illustrated record, the honesty statements, the sibling pages. | `src/components/trust/TrustPage.tsx` |
| `src/data/product.ts` | The seven Product pages: lede, how it works, proof points, the record or the commands, the Trust pages that prove them, the siblings. | `src/components/product/ProductPage.tsx` |
| `src/data/distributions.ts` | The hosted and self-hosted pages: lede, proof points, the record or the install, what stays yours. | `src/components/distributions/DistributionPage.tsx` |
| `src/data/doors.ts` | Every call to action once: label and destination by key; the user's pair and the buyer's pair. | the `Doors` primitive; a record names its pair |
| `src/data/page-shapes.ts` | The shapes the page records share: a proof point, a record or commands artifact, a step. | `trust.ts`, `product.ts`, `distributions.ts` |
| `src/data/site-pages.ts` | The route registry: every non-content page with its title, description, group, chapters, and index flag. | `sitemap.ts`, `robots.ts`, `lib/page-metadata.ts`, `generate-llms.mjs`, `check-apex-routing.mjs` |
| `src/data/retired-routes.ts` | Every retired path and the page that answers for it. | `RetiredRoute`, the link gate, the edge redirect declared in the estate |
| `src/data/pricing.ts` | Every price, cap, and free tier. | the pricing page and any sentence that names a price |
| `src/data/platform-stats.ts` | The counts a page prints, with how each was counted. | the proof strip, `llms.txt` |
| `src/data/testimonials.ts` | Customer quotes, verbatim, attributed, with approval on record. | the proof section |

Data files import each other with `.ts` extensions so Node can run them directly (the generators do); `tsconfig.json` allows it.

Components come from one library, `src/components/marketing/`, on the palette's role classes (`bg-canvas`, `bg-card`, `text-fg-secondary`, `border-edge`, `text-ok`). The library holds the primitives (card, badge, buttons, record window) and the page sections every story page composes (`PageHero`, `ProofList`, `PageCard`, `PageArtifact`, `Doors`, `CommandBlock`, `ProviderStrip`); a page template is a thin composition of those in its reader's order and states nothing. The palette is defined once in `packages/website-shell/src/theme/tokens.ts` and projected into the MUI theme and Tailwind; a component never types a hex. The published law is `public/branding/design-system.md`.

The landing page is versioned: `src/components/landing-page/v<N>-<date>/` holds only that version's section composition; `src/components/landing-page/index.ts` points at the active one and is the rollback switch. Nothing outside a versioned folder imports from inside one.

Decks run on one engine, `src/components/deck/` (hash navigation, keyboard, touch, presenter notes, the slide kit). Meeting decks compose it from `src/components/meetings/`; persona decks will sit beside them.

## The laws

1. **A claim lives in `src/data/`, never in a component.** If you find yourself typing a sentence about the product into a component, stop and put it where the other claims live, naming its chapter.
2. **Public sentences are verified.** No dollar-savings figure. No component called compliant (a component enforces controls). No analogy for the umbrella; each hub has exactly one. Testimonials verbatim and attributed to a person. Competitors described, never named. Nothing unshipped presented as live; illustrated records say they are illustrations and mark example figures `est.`.
3. **Numbers come from `platform-stats.ts` and prices from `pricing.ts`.** A literal in prose is a defect.
4. **Colors come from the palette.** Role classes only; no hex in a component; semantic hues only where they carry meaning.
5. **Every page is registered.** A new route goes into `site-pages.ts` (or `retired-routes.ts`) or the build fails.
6. **Every new top-level path is three declarations**: the registry, the edge passthrough list, and the platform's reserved handles. The apex guard names what is missing.
7. **A retired route is whole.** Every retired path has an eight-line stub, forwards to a live registered page (never to another retired path), and nothing in the export links to it: the forward exists for the outside world, our own links point at the live page. The link gate enforces all three.
8. **Nothing merges without the founder.** Build, lint, typecheck, the guards, the screenshot compare for a zero-visual-change commit, and a design review that reads the page as the visitor and as a copywriter.

## Where things live

```
src/app/(site)/          pages inside the website shell (header, footer, theme)
src/app/(standalone)/    surfaces without the shell: decks, the demo, book-demo, investor pages
src/app/sitemap.ts       generated from the registry and the content folders
src/app/robots.ts        generated
src/components/marketing/   the primitive library
src/components/landing-page/ the versioned landing compositions
src/components/trust/    the Trust template and index
src/components/product/  the Product template and index (and, until the Solutions pages are rebuilt, the 2025 solutions and desktop kits)
src/components/distributions/ the Distributions template and index
src/components/deck/     the deck engine and slide kit
src/components/meetings/ meeting decks and their registry
src/components/site/     site mechanics (RetiredRoute)
src/data/                every claim, price, count, route
src/lib/                 page-metadata, content-routes, assets, console-handoff
scripts/                 the guards and generators (below)
content/                 legal markdown, copywriting stage areas, design-review board and captures, the image inbox
public/docs, blog, changelog, tutorials   content pages, walked by the sitemap and llms generator
packages/website-shell/  the header, footer, navigation, palette, and theme the console also consumes
```

## Guards that run in the build

| Script | Proves |
|--------|--------|
| `scripts/check-displayed-vs-enforced.mjs` | every displayed plan limit and entitlement matches what the platform enforces |
| `scripts/check-apex-routing.mjs` | every top-level path is passed through at the edge and reserved as a platform handle (reads the sibling `planton-platform` checkout; skips loudly without it) |
| `next build` with `tsc --noEmit` and eslint before it | types and the accessibility and image rules hold |
| `scripts/check-internal-links.mjs` | every internal link in `out/` resolves to a page or a static file, and every retired route is whole (stub present, target live, nothing links to it) |
| `scripts/generate-llms.mjs` | `llms.txt`, `llms-full.txt`, and one Markdown per marketing page from the same data; fails when an exported route is unregistered |

## Proving a change

- `make build` green.
- For a change that must not alter pixels: capture a before set from a `main` build (`node scripts/capture-pages.mjs --export /path/to/main/out --out /tmp/before`), then `make capture OUT=/tmp/after COMPARE=/tmp/before`. The harness renders under a virtual clock so two builds of the same page produce byte-identical PNGs; a difference is a change.
- For a new or rebuilt page: captures at 1280 and 1680 go to the independent design reviewer with the page's first-glance question and the named visitor. PASS means nothing above minor. The reference board it grades against is `content/design-review/reference-board/`.

## Rules for agents

Rule files live beside what they govern:

- `_rules/docs/` — writing and formatting documentation pages
- `content/copywriting/_rules/` — the two-step copywriting workflow (draft and handoff, then implementation)
- `content/assets/_rules/` — images through the R2 pipeline
- `public/docs/_rules/` — docs content rules
- `src/app/(standalone)/meets/_rules/` — creating and updating a meeting deck

## Health ledger

What a day-one architect would not have done, and which work retires it. Nothing is added to the legacy allowlist in `eslint.config.mjs`; entries leave it as their folder is rebuilt.

| Legacy | Where | Retired by |
|--------|-------|-----------|
| Hex literals in components; MUI icons beside lucide; `'use client'` on components with no hooks | `components/product/solutions`, `components/product/shared`, `components/product/desktop`, `pricing`, `enterprise`, `blog`, `docs`, `tutorials`, `changelog`, `common`, `book-demo`, `legal`, `branding`, `invest`, `demo`, `tour`, `hackathon` | each page group as it is rebuilt from the story |
| Two landing versions kept for rollback (`v3`, `v4`) and the 2025 primitives (`v1`) used by the hackathon pages | `components/landing-page/` | v3 and v4 after v5 has held for one release; v1 with the hackathon's retirement |
| The interactive demo and the tour, on a light palette | `components/demo`, `components/tour` | a product decision: one surface on the design system, or real product recordings |
| The investor deck and explainer on their own primitive set | `components/invest` | invest onto the deck engine |
| Images under `public/_site/` instead of the asset CDN; 136 literal `/_site/` paths | `public/_site/images`, `src/**` | images to R2 as pages are rebuilt; `lib/assets.ts` names the prefix once |
| The `/_site` asset prefix and the post-build copy of the bundle | `next.config.ts`, `Makefile` | the apex routing decision |
| Retired paths served by a client-side forward only | `RetiredRoute` | the edge redirect declared in the estate |
| The desktop landing and download pages at `/features/desktop` | `app/(site)/features/desktop` | the founder's choice of Desktop's permanent address, which the release pipeline prints into every installer |
| The shell's product menu holds its own sub-labels beside the registry's descriptions | `packages/website-shell/src/data/navigation.ts` | the navigation slice (the shell cannot import from `src/`) |

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

4. **Use the R2 URL** in docs/blog:
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

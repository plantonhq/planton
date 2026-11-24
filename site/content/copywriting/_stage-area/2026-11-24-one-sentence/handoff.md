# Handoff: the one sentence, and batch 1 (what comes off)

## Routes

- **Every page:** the header and footer (`packages/website-shell`).
- **Home (`/`):** the eyebrow and caption, the door order, the illustration's people, the agents section, the FAQ and the close.
- **`/product` and its seven pages, `/trust` and its five pages, `/distributions/*`, `/desktop`, `/pricing`:** false and retired sentences corrected.
- **The docs:** the overview, platform, security, Infra Hub, Service Hub and billing pages, plus the Assistant mentions in seven other pages.
- **New:** the site-wide 404 (`src/app/not-found.tsx`).
- **Removed, now answered by the 404:** `/solutions`, `/solutions/<persona>` (five), `/compare`, `/decks/<persona>` (five), `/pricing/enterprise`.
- **Outside the site:** the open-source `README.md` pitch, and the organization's GitHub profile.

## Where the copy is

- **The sentence:** `positioning.ts` holds `umbrella.sentence` and `umbrella.firstAction`. Chapter 2 of `story.ts` mirrors the story's version 2.0.
- **Each corrected sentence** sits in the record that owns it, beside its chapter comment.
- **The 404's words** are in `src/data/not-found.ts`, and its doors are in `doors.ts` by key.
- **The header and footer** are data in `navigation.ts`. The components render a list and state nothing.

## The argument

A stranger should meet one true story. Planton deploys every push to GitHub into their own cloud, and shows the cost and permissions of every infrastructure change before it runs. Everything that contradicted that is gone: a self-service-platform umbrella, AI that designs infrastructure, an in-product Assistant with credit packs, five personas, a comparison page, an enterprise rate card with no plan behind it, and stale counts. Home's full rebuild around the sentence is the next batch. This batch is subtraction, plus a 404 that gives a stranger a way on.

## Component mapping

- **The header** is `HeaderLinks` plus `AuthButtons`. It is the same on every page, and the homepage variant is gone.
- **The footer** is five groups plus the legal line.
- **The 404** is `PageHero` plus `Doors` inside `MarketingShell`.
- **Nothing new joined `components/marketing/`.** The persona card left it.

## Revisions after the draft

- **More copy the sweep missed, the same kind as the approved changes, all corrected:**
  - Every docs page quotes the Infra Hub and Service Hub introductions, and they still carried the retired hub analogies.
  - The billing doc carried an AI-credits section.
  - The Service Hub overview named Cloudflare Workers as a deploy target. The service-delivery code registers only Kubernetes, Cloud Run and ECS.
- **The llms generator** still read the removed desktop Assistant card. Fixed.
- **The pricing page's subtitle** read "enterprise identity". It now reads "more seats and team controls".

## What the pages must not claim

The retired phrases, in-product AI or an Assistant, AI credits, enterprise tiers, SOC 2 or another framework by name, ready-made charts, Cloudflare Workers as a delivery target, production approval without "Team plan", or any count other than `platform-stats.ts`'s.

## Left for later batches

- **The docs' Cloudflare Workers delivery sections**, all false today: `connections/container-registries.md`, `ci-cd/what-is-a-service.md`, `deployment-targets.md`, `pipelines.md`, `build-methods.md`, `ingress.md`, `getting-started.md`, `platform/platform-tour.md`. Also the Day 1–3 journey in `platform/index.md` and the runner-first framing in `security/index.md`. All of these go to batch 4.
- **The refund policy's and terms' AI-credit and Assistant clauses** go to batch 3, along with the refund page's description.
- **`docs/self-hosting/index.md`** names a bundled image an operator must mirror. That stays until the self-hosting docs are reviewed against the never-public list.
- **The homepage video's "Proven charts" chapter** is baked into the render.

## Verification

- **`make -C site build` is green:** both pricing guards, the apex guard, lint, typecheck, 153 pages exported, the link gate (5,860 links) and llms coverage.
- **`make capture` was compared against the main tip's captures.** Every scene differs, because the header is on every page, and each changed scene was read. The 404, Home's top and the footer match the approved after.
- **An independent design review** graded the captures (its record is below).
- **Grep gates on the rendered export** found zero hits outside history pages (blog, changelog, meetings, investor and legal records) for the retired phrases, "SOC 2", "700+", the credit packs and the removed routes.

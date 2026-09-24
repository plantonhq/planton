# Handoff: the Product and Distributions sections

**Routes**: `/product` and `/product/{infra-hub,service-hub,coding-agents,cli,catalog,import,open-source}`; `/distributions` and `/distributions/{hosted,self-hosted}`. Every `/features/*` path is retired and forwards to its live page.

## Where the copy is

Data, as for the landing and Trust: `src/data/product.ts` (each page's lede, how it works, proof points, the record or the commands, the Trust pages that prove it, the siblings, the doors), `src/data/distributions.ts` (each shape's lede, proof points, the record or the install, what stays yours), `src/data/doors.ts` (every call to action once), `src/data/site-pages.ts` (titles and descriptions). Proof points are chapter sentences from `src/data/story.ts` wherever a chapter states the fact, and documented shipped behavior (the docs under `public/docs/`, the repository README, the company wiki) where a page needs a fact the story only summarizes. Every command is quoted from the guide the caption links. The canonical narrative is `company/marketing/positioning/the-planton-story.md` (v1.2) in the company repository.

## The argument

Product is the user's half of the story in the order a platform engineer meets it: what this is and who it is for, how it works as a person experiences it, what you get (proof points beside the record the product shows or the commands a person types), where the same claims are proven for the person who signs, then the siblings and the way to start. Distributions is the decider's order: the shape in two sentences, what you get, what stays yours said plainly, the other shapes. The mobile companion has no page: it is not shipped.

## Component mapping

- `src/components/marketing/` gained the page sections every story page composes: `page-hero`, `proof-list`, `page-card`, `page-artifact`, `doors`, `command-block`, `provider-strip`. Trust and landing v5 were recomposed on them first with zero visual change.
- `src/components/product/ProductPage.tsx`, `ProductIndex.tsx`; `src/components/distributions/DistributionPage.tsx`, `DistributionsIndex.tsx`.
- `packages/website-shell/src/data/navigation.ts` and `components/header/menu-icons.tsx` (one icon map for both headers).

## Must not claim

Any dollar-savings figure; "compliant" of any component; anything not in the public docs or the story as shipped (the account-wide browse, the failed-deploy rescue, and a "drift report" were removed by review); a bill where the product shows a verified monthly cost; a mobile page or store link; a competitor's name.

## Verification

`make build` green with every guard, including the retired-route law added to the link gate this session. Four rounds with the independent reviewer (Product) and three (Distributions); the final round's only major on either group is the standing ask for real product captures. Captures of the final round under `content/design-review/captures/2026-09-18/`.

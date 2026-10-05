# Cost Transparency — a Standing Duty

Most people composing charts here are solo practitioners and small teams
paying the cloud bill themselves. Cost is on their mind even when it is not
in their words. **Surface the cost picture of every architecture you propose
— without being asked.** Nobody else combines "composes the infrastructure"
with "tells me what it will cost and how to pay less"; this duty is a large
part of why the composer is loved.

## Where the numbers come from: read, never recall

Every dollar figure you state comes from the catalog's verified cost data,
read at answer time — the same law the multi-cloud-catalog skill applies to
schema facts extends to money. Whatever you remember about cloud pricing is
stale by construction; the catalog's figures were read from the provider's
own price documents, carry the source URL and the date they were verified,
and are re-verified by machine.

- **Files first (the fast path).** Covered kinds carry fact-sheets in
  the catalog skill's reference pack: `<provider>/<kind>/cost.yaml` (billing
  model, always-on baseline charges, the spec fields that move the bill,
  exclusions) and `_pricing/estimates/<kinddir>.yaml` (per-preset monthly
  estimates — exact quantities, unit prices, line totals, each line citing
  its price source and verification date). The catalog skill's
  research-recipes reference has the exact moves; resolve the pack through
  its ladder.
- **On the platform-tools arm** (no filesystem pack), the same answers come
  from the platform tools: `get_catalog_kind_cost` for the cost anatomy and
  preset estimates, `get_catalog_kind_control_posture` and
  `get_catalog_kind_permissions` for the posture and permissions sides of the
  fact-sheet.
- **When neither instrument is reachable, say so.** "I can't reach the
  verified cost data from here" is a complete, honest answer. Never fill
  the gap with a remembered rate — a confident stale number costs the user
  real money and your answer its trust.

## When and how

- **At the explain-after (Phase 4a)** — or at the plan, when the user asked
  to review one first: alongside the built (or proposed) resource list, give
  the monthly picture — the total order of magnitude and the two or three
  resources that dominate it, from the kinds' estimate documents. (The
  catalog-availability disclosure, when there is one, rides this same block
  — `catalog-availability.md`.) One
  short block, not a rate card:

  > Rough monthly cost: ~$150 at list prices. The EKS control plane and the
  > NAT gateway dominate; the rest is nodes and storage. (Figures from the
  > catalog's verified estimates — I can show the per-line breakdown with
  > sources.)

- **At every costly choice**: when a decision moves the bill meaningfully
  (NAT per-AZ vs single, instance sizes, multi-AZ databases, provisioned vs
  on-demand), read both configurations' figures and state the delta as part
  of recommending. The kind's `cost.yaml` names exactly which spec
  fields move the bill — that list IS the decision map.
- **At the finish**: the chart summary repeats the cost picture and names
  the params the user can turn when they want it cheaper.

## Honesty rules

- Numbers are **estimates at list prices** — say "roughly" / "~", state the
  region assumption, and note the exclusions the estimate document itself
  declares (traffic, data transfer, usage variance). Never present an
  estimate as a quote or a bill prediction.
- **Quote money verbatim.** Estimate figures are exact decimal strings;
  round only when presenting a total, and keep the exact figures available
  ("$16.4250/mo on the estimate" may read as "~$16/mo" with the exact line
  on request). Each line's `price_source` + `retrieved_on` is the citation
  to offer when precision matters.
- **A kind without published cost data is "not yet priced", never $0
  and never a guess.** Coverage is kind by kind; some covered
  kinds deliberately ship no estimate because their rate lives on a
  referenced resource — their cost.yaml notes say where the honest estimate
  happens. Relay that honesty instead of papering over it.
- **Usage-based kinds with a $0 committed estimate are good news to
  state precisely**: "nothing charges while idle; the meters that bill are
  named in the estimate's exclusions."

## Saving recommendations

Always pair the number with the lever. The classics that fit chart params:

- Single NAT gateway for non-production (the fleet's charts expose this) —
  read the NAT kind's estimate to state what each extra gateway adds.
- Right-size nodes and start with fewer; scaling up later is a param change.
- Public EKS endpoint for dev (private needs a standing runner — itself a
  cost); flip to private for production.
- Spot/ARM (Graviton) instance types where the workload tolerates them —
  spot presets estimate $0.00 committed with the market-price caveat stated.
- Turn off what the motive does not need: multi-AZ databases, provisioned
  IOPS, per-AZ redundancy in a sandbox — each of these is a cost driver
  named in its kind's cost.yaml, so the delta is readable, not
  guessable.

Frame savings against the user's motive (see `craft.discovery.md`): production
resilience is worth paying for; a learning sandbox is not.

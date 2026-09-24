---
name: multi-cloud-catalog
description: Research Planton's cloud component catalog from the file-based reference pack shipped inside this skill -- generated per-component pages (spec fields, validation rules, outputs, wiring, a validated example), provider indexes, the foreign-key graph, the authored wisdom layer, and verified fact-sheets (cost estimates, security postures, runner permission manifests). Use when composing or reviewing an architecture and you need component facts (what exists for a provider, what a component requires or exports, what can reference what, what it costs), judgment between components, or a compatible alternative -- and for giving wisdom back through the contribution workflow. The discipline is absolute -- every schema claim, price, posture, and permission is read from a pack file at answer time, never recalled from memory; with no pack reachable, fall back to planton explain and say what that cannot see. Not for editing schemas or generated pages, and not for deploying -- this is the research layer.
---

# Multi-Cloud Catalog

Planton's deployable building blocks -- cloud components spanning AWS, GCP,
Azure, Kubernetes, and every other supported provider -- are documented as a
reference pack: plain files you read and search. One page per component,
regenerated from the schemas so facts cannot drift, plus indexes, a
catalog-wide wiring graph, an authored wisdom layer, and -- for covered
components -- verified fact-sheets: what it costs (with per-preset dollar
estimates citing the provider price document and verification date), which
security controls it enforces with evidence, and exactly what permissions
its deployment runner needs. This skill is how you research that pack:
which file answers which question, in what order to look, and the
discipline that keeps answers grounded.

## The one law: facts are read, never recalled

Every field name, type, enum value, default, validation rule, and output
path you state must come from a pack file you read while answering -- or,
when no pack is on disk, from `planton explain` or from a component page
fetched into your workspace (`references/pack-layout.md` names the fallback
ladder). Whatever you remember about a component's schema is stale by
construction: schemas change and the pack regenerates with them. A confident
answer from memory that names one wrong field costs the user a failed
deployment; a ten-second file read costs nothing.

## Locate the pack first

Probe for a `components/` directory beside this file first: when this skill
ships self-contained, the pack travels inside it, published and versioned
with the skill, so a hit is always exactly as fresh as these instructions.
On a miss, resolve the pack through the ladder in
`references/pack-layout.md` -- a Planton repo checkout's `catalog/` (it may
carry unreleased pages), the release artifact, and the honest per-component
fallbacks when no pack is reachable at all. Everything below writes
`<pack-root>` for the resolved root -- the directory whose `_docs/` holds
`reference-commons.md`.

## Which file answers which question

| Question | Where the answer lives |
|---|---|
| What components exist for this provider / this need? | The provider's `reference-index.md` table; full-text search for capability words when the name is unknown |
| What does component K require? | K's `reference.md`: read `## Example` first (a validated manifest), then the `## Spec Fields` table's Required column and `## Validation Rules` |
| What does one field mean and accept? | The field's own `### spec.<path>` block in `## Field Details` -- docs, rules, allowed enum values |
| Where does a secret go on K (a password, API key, token)? | The marks on K's `## Spec Fields` rows: `(sensitive)` holds secret material (on Planton only as a `$secret/...` reference); `(no secrets: use <field>)` is read by every viewer and refuses a secret reference -- the named sibling is where it goes, and its field detail says why. `_docs/reference-commons.md` ("Where a secret may go") explains the rule once; K's `GUIDE.md` weighs the secret-bearing fields against each other |
| What does K export after deployment? | K's `## Outputs` table |
| How do components wire together? | K's `## References` (outbound) and `## Referenced By` (inbound) tables; `_docs/reference-graph.yaml` for catalog-wide edges |
| What is the judgment call before choosing K? | `GUIDE.md` beside K's page -- the page head links it when one exists -- and `_patterns/` for multi-component recipes |
| What does K cost per month? | K's `cost.yaml` (billing model, always-on baseline charges, the spec fields that move the bill, exclusions) + the per-preset dollar estimates at `_pricing/estimates/<kinddir>.yaml`, where `<kinddir>` is the kind lowercased (e.g. `awsalb.yaml`) |
| Which security controls does K enforce? | K's `controls.yaml` -- every control's stance with evidence; control ids resolve to names and statements in `_compliance/controls-catalog.yaml`; framework views (HIPAA, SOC 2, FedRAMP, CIS) via `_compliance/frameworks/*.yaml` -- honor each crosswalk's `spec.providers` scope: a scoped framework (e.g. CIS AWS -> aws) never answers for another provider's component, while an empty scope means provider-neutral |
| What permissions does K's runner need? | K's `iac/permissions.yaml` -- the least-privilege provisioning manifest per provider, each entry marked `derived` or `proven` |
| The asked-for software has no component of its own | The catalog guide (`_docs/GUIDE.md`) owns this workflow: search the pack by the name the user said, propose the compatible alternative openly, generic mechanisms only as the last resort |
| Manifest grammar: envelope, value/valueFrom, fieldPath spelling | `_docs/reference-commons.md` -- read it once per session; it also documents the search grammar these pages share |
| This session learned judgment the pack does not teach | `references/contributing-wisdom.md` -- offer once, verify, draft, and deliver it as a pull request |

The concrete command moves for each row -- with the friction traps already
learned so you do not relearn them -- are in
`references/research-recipes.md`.

## The research workflow

1. **Commons once.** Read `_docs/reference-commons.md` at the start of
   catalog work: the manifest grammar every component shares and the fixed
   heading vocabulary that makes the whole pack greppable.
2. **Index before grep.** For "what exists" questions, open the provider's
   `reference-index.md` before reaching for full-text search -- the table
   carries every kind with a one-line purpose, and its Guide column tells
   you where judgment has been written.
3. **Page before graph.** For one component's wiring, its own References /
   Referenced By tables are complete; the graph file is for catalog-wide
   questions ("every field anywhere that can point at X").
4. **Facts, then wisdom.** Settle the schema facts first, then read the
   component's guide and any pattern that composes it -- wisdom changes
   which component you pick and what else the architecture needs, not what
   the fields are.
5. **Verify before you assert.** Before stating a field path, an enum
   value, or an output name in an answer or a manifest, confirm the exact
   spelling in the page you have open. If a guide and a generated page ever
   disagree on a fact, the generated page wins -- and the disagreement is
   worth contributing back (`references/contributing-wisdom.md`), as is any
   judgment this session had to work out that no guide teaches.

## Money, posture, and permissions: the honesty grammar

The fact-sheet layer carries claims with real-world consequences, so the
one law gets four extra clauses when you read it:

- **Money is echoed verbatim, never recomputed or rounded silently.** Every
  dollar figure in an estimate is an exact decimal string with the provider
  price document it was read from and the date it was verified -- quote it
  as written (round only when presenting, and say so). Estimates are list
  prices for a named preset with stated exclusions -- never a bill
  prediction, never a quote.
- **Absence is "not yet published", never $0 and never a guess.** Coverage
  is component by component: a kind without fact-sheets has no cost.yaml
  and no estimate document, and a covered kind whose rate lives on a
  referenced resource ships no estimate on purpose (its cost.yaml notes say
  where the honest estimate happens). Say the data is not published;
  never fill the gap from memory.
- **Posture language never becomes "compliant."** controls.yaml states
  technical control posture with evidence; the crosswalks map it onto
  framework requirements. No component is ever "HIPAA-compliant" or "SOC 2
  compliant" by itself -- say "enforces/exposes these controls, which map
  to these requirements." FedRAMP binds hardest: it is an authorization
  program, so the crosswalk is evidence an assessment consumes, never an
  authorization claim -- "FedRAMP authorized" must never be said of any
  component, chart, or deployment.
- **A framework's provider scope is part of its truth.** Each crosswalk's
  `spec.providers` names the providers the framework itself claims (empty =
  provider-neutral, applies everywhere). Never map a provider-scoped
  benchmark onto another provider's component -- if asked, say the
  framework names only its own provider in scope.
- **Provenance is part of the answer.** Permission entries are `derived`
  (read from the official module sources by static analysis) or `proven`
  (observed from live provisioning runs) -- state which, and that manifests
  describe the official modules as shipped.

## Speak the user's language

The catalog's internal vocabulary -- kind names, `valueFrom`, `fieldPath`,
provisioner annotations -- is your building material, never the user's
curriculum. Deliver answers in the terms the user asked in; surface platform
constructs only where the user must see them (a manifest they will read) or
when they ask to learn. When you substitute a compatible alternative for
software they named, say so explicitly and explain the compatibility --
never substitute silently.

## Boundaries

- **Facts are never edited on the page.** A wrong or thin fact on a
  generated page is fixed at its source (the proto comment or validation
  rule it derives from), never by editing the page. Wisdom -- judgment in
  guides and patterns -- IS contributable from here: follow
  `references/contributing-wisdom.md`. Schema and generator changes remain
  out of scope.
- **No pack, no guessing.** Exhaust the fallback ladder first -- the CLI's
  explain report, or component pages fetched into your workspace
  (`references/pack-layout.md`). Only when every rung is out of reach do you
  say plainly that schema facts are unavailable -- never answer from memory.
- **One pack version per answer.** Files within one skill mount, one
  checkout, or one release zip are mutually consistent; never mix pages
  from two pack versions in a single answer.

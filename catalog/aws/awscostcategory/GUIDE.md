# AwsCostCategory — Kind Guide

Authored operational judgment for the cost-category component: the
design decisions behind the spec's shape, and what to know before
operating categories in production.

## Design decisions

- **The name is an explicit spec field** (spaces legal — "Cost
  Center") and create-only: renaming replaces the category and resets
  its history.
- **`rule_version` is module-pinned** to `CostCategoryExpression.v1`
  — the only value the AWS API accepts, so a spec knob would be dead
  configuration (recorded as a manifest exclusion).
- **Rules are an ORDERED list, first match wins** — the modules render
  them exactly in spec order, and each rule is exactly one shape
  (REGULAR expression or INHERITED_VALUE dimension), spec-enforced.
- **The rule expression is LEVELED (root → node → leaf)** — exactly
  the two composition levels AWS accepts below the root, so deeper
  nesting is inexpressible and neither engine needs depth checks.
- **The cost-allocation-tag activation toggle is deliberately NOT
  here.** `aws_ce_cost_allocation_tag` is a per-tag-key ACCOUNT
  feature with no schema edge to any category — folding it in would
  make many category instances fight over one account object. It is
  planned as its own account-settings kind.

## Operating categories in production

- **Categorization is retroactive to the month**: AWS re-categorizes
  the current and previous month on every rule change;
  `effectiveStart` accepts month boundaries only, up to twelve months
  back.
- **A DELETED category keeps answering reads until month end** — AWS
  end-dates it (`effective_end`) rather than erasing it, so reports
  stay coherent. The E2E verifier treats an end-dated definition as
  absent for the same reason.
- **Order rules most-specific first** — a broad first rule swallows
  spend later rules were meant to catch.
- **Inherited TAG rules only see ACTIVATED cost-allocation tags**;
  un-activated keys inherit nothing.
- **Split-charge sources must not be targets of other split rules**,
  and FIXED splits need one percentage per target summing to 100
  (spec-enforced presence; AWS enforces the sum).
- **Categories can build on categories** — the expression's
  cost_category leaf matches another category's values; deletion of a
  referenced category breaks the referencing rule at AWS, not at
  plan.
- **The import ID is the category ARN.**

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# AwsBudget — Kind Guide

Authored operational judgment for the budget component: the design
decisions behind the spec's shape, and what to know before operating
budgets in production.

## Design decisions

- **The name is an explicit spec field.** Budget names legally carry
  spaces and mixed case ("Engineering Monthly Spend") that
  `metadata.name` cannot; changing the name replaces the budget.
- **Exactly one funding shape, spec-enforced.** AWS accepts one of a
  fixed limit, planned per-period limits, or auto-adjustment per
  budget — the CEL rule keeps the failure at manifest time instead of
  apply time.
- **Two filter generations, mutually exclusive.** The modern
  `metric` + `filterExpression` pair and the legacy
  `costFilters`/`costTypes` are both modeled (total accounting), but
  AWS accepts one generation per budget. The modern pair must be set
  together (the provider's own RequiredWith).
- **The filter tree is LEVELED (root → node → leaf)** — exactly the
  two composition levels AWS accepts below the root. The shape makes
  deeper nesting inexpressible, so neither engine needs a depth
  check. The `ABSENT` match option is deliberately not offered on
  budgets: the provider rejects it at plan time (its API contract is
  self-contradictory for budgets).
- **Actions fold as name-keyed satellites.** An action exists only on
  its budget; the entry's `name` is Planton-side identity (the
  for_each key and the `action_ids` output-map key) — AWS identifies
  actions by a generated ID.

## Operating budgets in production

- **Budget math is decimal-normalized** — "100" and "100.0" are the
  same limit; the modules pass amounts as strings verbatim.
- **RI/Savings Plans budget types take PERCENTAGE limits** (usually
  100) — a dollar limit on RI_UTILIZATION is rejected by AWS.
- **MANUAL approval stages an action without executing it** — the
  console shows it pending; AUTOMATIC executes on breach and REVERSES
  when spend drops back under (IAM/SCP arms detach; SSM stops do not
  restart instances).
- **The execution role must trust budgets.amazonaws.com** and carry
  the permissions its arm implies — AWS's managed
  `AWSBudgetsActionsWithAWSResourceControlAccess` policy is the
  canonical grant. Action creation retries briefly on IAM propagation.
- **Notifications need at least one subscriber** (spec-enforced) and
  SNS topics need a policy allowing budgets.amazonaws.com to publish
  — silent alert loss otherwise.
- **Tag-based filter expressions only match ACTIVATED cost-allocation
  tags** — activate keys in the Billing console (or the future
  cost-allocation-tags kind) or the filter matches nothing.
- **Import IDs**: the budget as `account_id:budget_name`, each action
  as `account_id:action_id:budget_name`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

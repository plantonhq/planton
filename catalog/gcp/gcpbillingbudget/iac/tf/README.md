# GcpBillingBudget — Terraform Implementation

This directory contains the Terraform implementation for provisioning a
Cloud Billing budget from the Planton spec: one `google_billing_budget` on
the billing account.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration); the principal needs
  `roles/billing.costsManager` on the billing account

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | The bare billing account ID, display name (spec or metadata.name fallback), the `projects/{number}` prefixing of the project filter |
| `main.tf` | `google_billing_budget` with its amount, filter, threshold rules, and notification rule |
| `outputs.tf` | `name`, `budget_id`, `billing_account` |

## Send Posture

- **`billing_account`** -- the spec accepts the ID or `billingAccounts/{id}`;
  the module strips the prefix, which the provider adds back.
- **`amount`** -- exactly one arm (spec CEL): `specified_amount` with
  `units` as the string the API wants, `currency_code` only when set
  (Optional+Computed: unset takes the account's currency), `nanos` only
  when non-zero; or `last_period_amount = true`.
- **`budget_filter.projects`** -- flattened `GcpProject` references arrive
  as project numbers; the module prefixes `projects/` unless the literal
  already carries it, so a reference and a literal render the same.
- **Defaulted enums** -- `credit_types_treatment` (INCLUDE_ALL_CREDITS),
  `spend_basis` (CURRENT_SPEND), and `schema_version` (1.0) carry proto
  defaults the manifest loader applies; the `coalesce` is the same rule for
  a hand-written tfvars file, and the value is always sent (PARITY with the
  Pulumi module).
- **`notifications`** -- the spec's name for the provider's
  `all_updates_rule`; its two booleans are always sent, its topic and
  channels only when set.
- **`ownership_scope`** -- sent only when set; unset lets Google apply its
  default.
- **`deletion_policy`** -- DELETE (default), PREVENT, or ABANDON.

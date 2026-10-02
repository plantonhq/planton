# GcpBillingBudget — Pulumi Implementation

This directory contains the Pulumi implementation for provisioning a Cloud
Billing budget from the Planton spec: one `gcp.billing.Budget` on the
billing account.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `budget` |
| `module/locals.go` | The bare billing account ID, display name (spec or metadata.name fallback) |
| `module/budget.go` | Maps spec to `gcp.billing.Budget`; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `budget_id`, `billing_account`) |

## Send Posture (parity with Terraform)

- **`BillingAccount`** -- the bare ID (the `billingAccounts/` prefix
  stripped when given).
- **`Amount`** -- exactly one arm: `SpecifiedAmount` with `Units` as a
  string, `CurrencyCode` only when set, `Nanos` only when non-zero; or
  `LastPeriodAmount = true`.
- **`BudgetFilter.Projects`** -- resolved `GcpProject` references (project
  numbers) prefixed with `projects/` unless the literal already carries it.
- **Defaulted enums** -- `CreditTypesTreatment`, `SpendBasis`, and
  `SchemaVersion` are sent explicitly with their proto defaults when the
  spec leaves them empty, exactly as the Terraform module's `coalesce`.
- **`AllUpdatesRule`** -- the spec's `notifications`; the two booleans
  always sent, the topic and channels only when set.
- **`OwnershipScope`**, **`DeletionPolicy`** -- sent only when set.
- **`budget_id`** -- derived from the resource name's last segment, the
  shape the Terraform module exports.

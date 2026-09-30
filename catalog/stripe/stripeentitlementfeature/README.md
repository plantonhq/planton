# StripeEntitlementFeature

Declares one capability a customer can be entitled to in [Stripe Entitlements](https://docs.stripe.com/billing/entitlements) -- "api-access", "sso" -- identified by a lookup key your application checks.

## When to Use

- **Gate features on what a plan grants**: your code asks "does this customer have `api-access`?" instead of listing the prices that include it.
- **Move a capability between plans without a deploy**: a StripeProduct grants features by reference, so changing a plan is a manifest change.
- **The same feature names everywhere**: every environment's account gets the same lookup keys.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeEntitlementFeature
metadata:
  name: api-access
  org: acme-corp
  env: production
spec:
  lookupKey: api-access
  name: API access
```

## Fields

| Field | Description |
|---|---|
| `lookupKey` | The stable name your application checks (required, up to 80 characters). Changing it **replaces** the feature |
| `name` | Your team's label in the Dashboard (required), never shown to customers. Changes in place |
| `metadata` | Key-value pairs stored on the feature |

## Key Behaviors

- **Destroy archives the feature.** Stripe keeps it, archived; it can no longer be attached to new products, and this kind cannot reactivate it.
- **Products grant it.** List the feature under a StripeProduct's `features`.
- **A feature that cannot be read fails the plan**: the provider does not treat a missing feature as gone. Remove it from state and apply again.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The feature's Stripe id (`feat_...`), what a product's `features` reference |
| `lookup_key` | The stable name, as your application checks it |
| `active` | `false` once archived |

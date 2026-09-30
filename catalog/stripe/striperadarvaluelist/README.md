# StripeRadarValueList

Declares a [Radar list](https://docs.stripe.com/radar/lists) -- blocked countries, trusted customers, known fraudulent emails -- and every item in it, for Radar rules to reference by alias.

## When to Use

- **Block and allow lists in files**: reviewed like code, the same in every account.
- **Values checked before Stripe sees them**: a country list takes two-letter codes, an email list takes addresses.
- **Rules that change less than their values**: the rule names `@alias`; the list holds what it matches.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeRadarValueList
metadata:
  name: blocked-countries
  org: acme-corp
  env: production
spec:
  alias: blocked_countries
  name: Blocked countries
  itemType: country
  items:
    - KP
    - IR
```

## Fields

| Field | Description |
|---|---|
| `alias` | The name rules use after `@` (required). Changes in place |
| `name` | The list's name in the Dashboard (required). Changes in place |
| `itemType` | What the list holds; unset, `string`. Changing it **replaces** the list and every item |
| `items` | The values, each once. Adding creates, removing deletes |
| `metadata` | Key-value pairs stored on the list |

## Key Behaviors

- **Items are validated by type**: country, email, IP address, card BIN, customer id and account id lists refuse malformed values before Stripe sees them.
- **Destroy deletes the list and its items.** Stripe refuses while a Radar rule still uses the list.
- **The list is free; using it may not be**: custom rules need Radar for Fraud Teams, priced per screened transaction.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The list's Stripe id (`rsl_...`) |
| `alias` | The name rules reference |
| `item_ids` | Each item's value mapped to its Stripe id (`rsli_...`) |

# StripeBillingMeter

Declares a usage meter: the event name your application sends to Stripe for each unit of use (an API request, a GB stored), how those events add up per customer, and alerts when a customer's usage crosses a threshold. A metered StripePrice names the meter to bill what it counts. See Stripe's [usage-based billing](https://docs.stripe.com/billing/subscriptions/usage-based).

## When to Use

- **Usage-based pricing in one chart**: the meter, the metered price that names it, and your application's deployment that reads its event name.
- **Usage alerts** that tell you when a customer reaches a threshold.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeBillingMeter
metadata:
  name: api-requests
  org: acme-corp
  env: production
spec:
  displayName: API requests
  eventName: api_requests
  defaultAggregation:
    formula: sum
  alerts:
    - title: 10k requests
      gte: 10000
```

## Fields

| Field | Description |
|---|---|
| `displayName` | The meter's name in the Dashboard and on invoices (required). The only field that changes in place |
| `eventName` | The event your application sends (required). **Replaces** |
| `defaultAggregation.formula` | `count`, `sum` or `last` (required). **Replaces** |
| `customerMapping.eventPayloadKey` | Where each event names the customer. Unset: `stripe_customer_id`. **Replaces** |
| `valueSettings.eventPayloadKey` | Where each event carries its amount. Unset: `value`. **Replaces** |
| `eventTimeWindow` | Pre-aggregate by `hour` or `day`. **Replaces** |
| `alerts` | Thresholds keyed by title: `gte`, an optional `customer`, `recurrence`. Can't be changed; see below |

## Key Behaviors

- **Only the display name changes in Stripe.** Changing anything else deactivates the meter and creates a new one, and every metered price on it is replaced with it.
- **Your application reads the event name by reference** (`status.outputs.event_name`), so it follows a replacement.
- **Alerts are forgotten, not deleted.** Changing an alert creates a new one; removing it, or destroying the meter, leaves it active in Stripe until you archive it.
- **Destroy deactivates the meter.** It stops accepting events; Stripe keeps it.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The meter's Stripe id (`mtr_...`); it changes when the meter is replaced |
| `event_name` | The event name your application sends |
| `status` | `active`, or `inactive` once deactivated |
| `alert_ids` | Each alert's title mapped to its Stripe id |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

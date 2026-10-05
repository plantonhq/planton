# Stripe Billing Meter

Declares a usage meter -- API requests summed per customer, the latest seat count -- that a metered price bills, with alerts when a customer's usage crosses a threshold. One Infra Component per meter.

## What Gets Created

When you deploy this Infra Component, the OpenTofu module creates, in the Stripe account your Stripe connection's key belongs to:

- **The meter** -- the event name your application sends, and how events add up per customer
- **Its alerts** -- one per declared threshold, each sending `billing.alert.triggered` when a customer reaches it

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Billing Meters: Write**, and write on alerts when alerts are declared (Stripe Dashboard: Developers, API keys, the key's permissions).
- **One owner**: declare a meter here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Billing Meter**, and click **Deploy**. Start from the **API Requests With an Alert** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeBillingMeter
metadata:
  name: api-requests
  org: acme-corp
  env: prod
spec:
  displayName: API requests
  eventName: api_requests
  defaultAggregation:
    formula: sum
```

```shell
planton apply -f stripe-billing-meter.yaml
```

An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Only the display name changes in place** -- any other change creates a new meter, and every metered price on it is replaced with it.

**Alerts are forgotten** -- a changed, removed or destroyed alert stays active in Stripe until you archive it.

**Destroy deactivates** -- the meter stops accepting events; Stripe keeps it.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The meter's Stripe id | A metered price's `recurring.meter` |
| `event_name` | The event your application sends | Your application's environment |
| `status` | `active` or `inactive` | Audits |
| `alert_ids` | Each alert's title mapped to its id | Audits, archiving a stale alert |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Requests with an alert** -- a summed meter that warns at a threshold. Start from the **API Requests With an Alert** preset.

**Seats** -- the latest seat count, for per-seat usage billing. Start from the **Active Seats** preset.

**Price and application in one chart** -- a metered Stripe Price names the meter, and your application's deployment reads its event name, both by reference. The meter deploys first, and the name your application sends and the name the meter counts can't drift:

```yaml
# StripePrice
spec:
  recurring:
    interval: month
    usageType: metered
    meter:
      valueFrom:
        kind: StripeBillingMeter
        name: api-requests
        fieldPath: status.outputs.id
---
# KubernetesDeployment
spec:
  container:
    app:
      env:
        variables:
          - name: STRIPE_METER_EVENT_NAME
            valueFrom:
              kind: StripeBillingMeter
              name: api-requests
              fieldPath: status.outputs.event_name
```

## Works With

- [**Stripe Price**](/infra-catalog/stripe-price) -- a metered price bills what the meter counts.
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- your application reads the event name by reference.

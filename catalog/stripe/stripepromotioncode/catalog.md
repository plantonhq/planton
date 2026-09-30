# Stripe Promotion Code

Declares a code customers type to redeem a coupon -- `LAUNCH25` for first-time customers, `WELCOME50` on orders over 50 dollars -- with its own limits. One Cloud Resource per code.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one promotion code on the coupon it names, in the Stripe account your Stripe connection's key belongs to:

- **The code** -- what customers type, or one Stripe generates
- **Its limits** -- first-time customers, a minimum order, one customer, a cap, an expiry

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Promotion Codes: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **The coupon** -- a Stripe Coupon Cloud Resource, or an existing coupon's id.
- **One owner**: declare a code here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Promotion Code**, and click **Deploy**. Start from the **Launch Code** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePromotionCode
metadata:
  name: launch25
  org: acme-corp
  env: prod
spec:
  coupon:
    valueFrom:
      name: launch-25
  code: LAUNCH25
  maxRedemptions: 500
```

```shell
planton apply -f stripe-promotion-code.yaml
```

A Stack Job tracks the change in real time.

### InfraChart

When the coupon and this code deploy together, name the coupon with ValueFromRef:

```yaml
spec:
  coupon:
    valueFrom:
      kind: StripeCoupon
      name: launch-25
      fieldPath: status.outputs.id
```

The InfraPipeline deploys the coupon first, then this code. When a change replaces the coupon, the code follows it on the same apply.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Letters, digits and dashes** -- a code with a space or another character is refused before anything runs.

**Limits are fixed** -- changing the coupon, expiry, cap, customer or restrictions creates a new code; the same code keeps working.

**Destroy deactivates** -- the code stops working and can be declared again later.

## Outputs and Dependencies

### What This Component Consumes

| Field | Kind | Output |
|-------|------|--------|
| `coupon` | Stripe Coupon | `status.outputs.id` |

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The promotion code's Stripe id | Checkout sessions that apply the code |
| `code` | What customers type | Marketing copy, your site |
| `active` | `false` once deactivated | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Launch code** -- first-time customers, capped and expiring. Start from the **Launch Code** preset.

**Minimum order** -- a welcome code on orders over a threshold. Start from the **First Order Minimum** preset.

## Works With

- [**Stripe Coupon**](/cloud-catalog/stripe-coupon) -- the discount the code redeems.
- [**Stripe Payment Link**](/cloud-catalog/stripe-payment-link) -- a hosted page where customers enter it.
- [**Stripe Billing Portal Configuration**](/cloud-catalog/stripe-billing-portal-configuration) -- lets customers apply codes to subscriptions.

# Stripe Coupon

Declares a discount -- 25% off for three months, 10 dollars off a first order, a free month forever -- that customers redeem with a promotion code or your application applies. One Cloud Resource per coupon.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one coupon in the Stripe account your Stripe connection's key belongs to:

- **The discount** -- a percentage, or an amount in one or more currencies
- **How long it lasts** -- the first invoice, some months, or every invoice
- **Its limits** -- how many redemptions, until when, and on which products

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Coupons: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **One owner**: declare a coupon here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Coupon**, and click **Deploy**. Start from the **25% Off for Three Months** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeCoupon
metadata:
  name: launch-25
  org: acme-corp
  env: prod
spec:
  name: Launch 25% off
  percentOff: 25
  duration: repeating
  durationInMonths: 3
```

```shell
planton apply -f stripe-coupon.yaml
```

A Stack Job tracks the change in real time.

### InfraChart

When the coupon and the products it discounts deploy together, name each product with ValueFromRef:

```yaml
spec:
  appliesToProducts:
    - valueFrom:
        kind: StripeProduct
        name: pro-plan
        fieldPath: status.outputs.id
```

The InfraPipeline deploys the product first, then this coupon, then any promotion code that names it.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**A discount change is a new coupon** -- Stripe never changes a coupon's discount, duration or limits. Planton deletes the old coupon and creates the new one; its promotion codes are replaced with it.

**Destroy deletes** -- customers who already applied the coupon keep their discount; nobody new can redeem it.

**Pick one discount** -- a percentage, or an amount with its currency.

## Outputs and Dependencies

### What This Component Consumes

| Field | Kind | Output |
|-------|------|--------|
| `appliesToProducts` | Stripe Product | `status.outputs.id` |

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The coupon's Stripe id | A promotion code's coupon, subscriptions, Checkout sessions |
| `valid` | Whether new customers can still redeem it | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Launch discount** -- a percentage off one plan for three months. Start from the **25% Off for Three Months** preset.

**First-order credit** -- a fixed amount off once, in several currencies. Start from the **Ten Off Once** preset.

## Works With

- [**Stripe Promotion Code**](/cloud-catalog/stripe-promotion-code) -- the code customers type to redeem the coupon.
- [**Stripe Product**](/cloud-catalog/stripe-product) -- the products the coupon discounts.
- [**Stripe Payment Link**](/cloud-catalog/stripe-payment-link) -- a hosted page that accepts promotion codes.

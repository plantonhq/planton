# Stripe Payment Method Configuration

Declares which payment methods your checkout offers -- cards, wallets, bank debits, buy now pay later, local methods -- method by method, as a configuration of your own that every checkout names. One Cloud Resource per configuration.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one payment-method configuration in the Stripe account your Stripe connection's key belongs to:

- **A preference per method** -- on, off, or left to Stripe, for any of 59 methods
- **What is really available** -- the methods Stripe reports on and with their capability active

The account's default configuration is never adopted or changed.

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Payment Method Configurations: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **Capabilities** for the methods you turn on (Stripe Dashboard: Settings, Payment methods).

## Deploy

### Console

Open the deployment store, find **Stripe Payment Method Configuration**, and click **Deploy**. Start from the **Cards and Wallets** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentMethodConfiguration
metadata:
  name: checkout-methods
  org: acme-corp
  env: prod
spec:
  name: Checkout methods
  card:
    preference: "on"
  applePay:
    preference: "on"
  googlePay:
    preference: "on"
```

```shell
planton apply -f stripe-payment-method-configuration.yaml
```

Name `status.outputs.id` as `payment_method_configuration` when your application creates a Checkout Session. A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Quote `"on"` and `"off"` for other tools** -- Planton reads them bare, but YAML 1.1 tools such as PyYAML read them as booleans.

**Off means off, left out means Stripe decides** -- a method you do not list keeps Stripe's default for the account; set `"off"` to hide it.

**Check what is available** -- `status.outputs.availablePaymentMethods` shows the methods Stripe will really offer; a method set on but missing there needs its capability turned on.

**Destroy deactivates** -- Stripe keeps the configuration, inactive, forever.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The configuration's Stripe id | Checkout Sessions, and a Stripe Billing Portal Configuration's `paymentMethodUpdate`, by reference |
| `isDefault` | Whether it is the account default | Audits |
| `active` | Whether payments may use it | Monitoring |
| `availablePaymentMethods` | The methods Stripe will offer | Checking capabilities |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cards and wallets** -- Cards plus Apple Pay, Google Pay and Link. Start from the **Cards and Wallets** preset.

**Cards only** -- Cards, with wallets and deferred payments off. Start from the **Cards Only** preset.

## Works With

- [**Stripe Billing Portal Configuration**](/cloud-catalog/stripe-billing-portal-configuration) -- reference this configuration so the portal offers the same methods.
- [**Stripe Webhook Endpoint**](/cloud-catalog/stripe-webhook-endpoint) -- where the payments these methods complete arrive as events.

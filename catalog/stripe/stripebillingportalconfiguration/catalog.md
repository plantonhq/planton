# Stripe Billing Portal Configuration

Declares what Stripe's customer portal lets your customers do -- cancel, switch plans, update payment methods and details, download invoices -- as a configuration of your own that every portal session names. One Cloud Resource per configuration.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one portal configuration in the Stripe account your Stripe connection's key belongs to:

- **The features** -- exactly what a customer may do, each off unless enabled
- **The policies** -- cancellation timing and reasons, which plans a customer may switch to, how changes are prorated
- **The branding** -- the headline and the privacy and terms links

The account's default configuration is never adopted or changed.

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Customer portal: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **Products and prices** to switch between, if you enable plan changes.

## Deploy

### Console

Open the deployment store, find **Stripe Billing Portal Configuration**, and click **Deploy**. Start from the **Self-Serve Portal, Cancel at Period End** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeBillingPortalConfiguration
metadata:
  name: customer-portal
  org: acme-corp
  env: prod
spec:
  name: Customer portal
  features:
    invoiceHistory:
      enabled: true
    paymentMethodUpdate:
      enabled: true
    subscriptionCancel:
      enabled: true
      mode: at_period_end
```

```shell
planton apply -f stripe-billing-portal-configuration.yaml
```

Name `status.outputs.id` as `configuration` when your application opens a portal session. A Stack Job tracks the change in real time.

### InfraChart

When the payment-method configuration and this portal deploy together, wire the portal to the configuration with ValueFromRef, so a customer adding a card in the portal is offered the same methods as checkout:

```yaml
spec:
  features:
    paymentMethodUpdate:
      enabled: true
      paymentMethodConfiguration:
        valueFrom:
          kind: StripePaymentMethodConfiguration
          name: checkout-methods
          fieldPath: status.outputs.id
```

The InfraPipeline deploys the payment-method configuration first, then this portal.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cancel when the paid period ends** -- `subscriptionCancel.mode: at_period_end` keeps what the customer paid for; `immediately` needs a `prorationBehavior` decision.

**Plans to switch between** -- `subscriptionUpdate.products` lists each product and the prices a customer may choose; allowing `price` changes needs at least one.

**The same methods as checkout** -- reference a Stripe Payment Method Configuration in `paymentMethodUpdate`.

**Destroy deactivates** -- Stripe keeps the configuration, inactive, forever; a session that names it afterwards is refused.

## Outputs and Dependencies

### What This Component Consumes

| Field | Kind | Output |
|-------|------|--------|
| `features.paymentMethodUpdate.paymentMethodConfiguration` | Stripe Payment Method Configuration | `status.outputs.id` |

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The configuration's Stripe id | Your application's portal sessions, by reference |
| `isDefault` | Whether it is the account default | Audits |
| `active` | Whether sessions may use it | Monitoring |
| `loginPageUrl` | The shareable sign-in URL | Emails and help pages |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Self-serve, cancel at period end** -- Details, payment methods, invoices and cancellation with a reason. Start from the **Self-Serve Portal, Cancel at Period End** preset.

**Plan switching** -- Upgrades at once, downgrades at period end. Start from the **Plan Switching Portal** preset.

## Works With

- [**Stripe Payment Method Configuration**](/cloud-catalog/stripe-payment-method-configuration) -- the methods a customer may add in the portal.
- [**Stripe Webhook Endpoint**](/cloud-catalog/stripe-webhook-endpoint) -- where the subscription changes customers make here arrive as events.

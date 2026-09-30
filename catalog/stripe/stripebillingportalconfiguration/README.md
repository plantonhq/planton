# StripeBillingPortalConfiguration

Declares what Stripe's [customer portal](https://docs.stripe.com/customer-management) lets a customer do: cancel (now or at period end, with a reason), switch plans and quantities, update payment methods and details, and download invoices.

## When to Use

- **The same portal in every environment**: what a customer may do is a file, so test and live never drift apart.
- **Your own configuration, not the account default**: your application names this configuration's id when it opens a portal session, and the account's default stays untouched.
- **A readable policy**: anyone can read in the manifest whether customers can cancel, and how.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeBillingPortalConfiguration
metadata:
  name: customer-portal
  org: acme-corp
  env: production
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

## Fields

| Field | Description |
|---|---|
| `features` | What a customer may do (required): `customerUpdate`, `invoiceHistory`, `paymentMethodUpdate`, `subscriptionCancel`, `subscriptionUpdate`. Every feature is off unless enabled |
| `features.subscriptionCancel` | `mode` (`at_period_end`, `immediately`), `prorationBehavior` (immediate cancellations only), `cancellationReason` with Stripe's fixed reasons |
| `features.subscriptionUpdate` | `defaultAllowedUpdates` (`price`, `quantity`, `promotion_code`), `products` (up to ten, each a StripeProduct reference with its StripePrice references and a quantity range), `prorationBehavior`, `billingCycleAnchor`, `trialUpdateBehavior`, `scheduleAtPeriodEnd` |
| `features.paymentMethodUpdate.paymentMethodConfiguration` | The payment-method configuration whose methods the portal offers; reference a StripePaymentMethodConfiguration |
| `businessProfile` | The headline and the privacy and terms links the portal shows |
| `defaultReturnUrl` | Where the back link leads when a session names none |
| `loginPage` | A shareable sign-in URL (`status.outputs.loginPageUrl`) |
| `name`, `metadata` | A team-facing label and key-value pairs |
| `active` | Defaults to `true`; `false` deactivates without destroying |

## Key Behaviors

- **Destroy deactivates**: Stripe never deletes a configuration, so destroy sets it inactive and Stripe keeps it forever.
- **A removed setting is not reset**: the provider sends only values that are set. Turn a feature off with `enabled: false` rather than deleting it from the manifest.
- **Never the account default**: this kind creates its own configuration; name its `id` in every portal session.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The configuration's Stripe id (`bpc_...`), passed as `configuration` when a portal session is created |
| `isDefault` | Whether it is the account's default |
| `active` | Whether portal sessions may use it |
| `loginPageUrl` | The shareable sign-in URL, when enabled |

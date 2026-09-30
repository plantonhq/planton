# StripePaymentMethodDomain

Registers a domain where your own checkout pages may show wallet buttons -- Apple Pay, Google Pay, Link, PayPal, Amazon Pay, Klarna -- through Stripe Elements or embedded Checkout. See Stripe's [domain registration](https://docs.stripe.com/payments/payment-methods/pmd-registration).

## When to Use

- **Wallet buttons on your domain**: a checkout page you serve yourself.
- **See why a button is missing**: each wallet's status and Stripe's reason are outputs.
- **The same domains in every account**: test and live register the same way.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentMethodDomain
metadata:
  name: checkout-domain
  org: acme-corp
  env: production
spec:
  domainName: pay.acme.com
```

## Fields

| Field | Description |
|---|---|
| `domainName` | The hostname (required), no scheme or path. Changing it **replaces** the registration |
| `enabled` | Whether wallets may appear (default `true`). Changes in place |

## Key Behaviors

- **Destroy only forgets.** The domain stays registered, and enabled, in Stripe. Apply `enabled: false` before deleting to turn wallets off.
- **Stripe validates per wallet.** Apple Pay needs Stripe's domain association file served from the domain.
- **A domain that cannot be read fails the plan**: remove it from state and apply again.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The registration's Stripe id (`pmd_...`) |
| `enabled` | Whether wallets may appear |
| `<wallet>_status`, `<wallet>_error_message` | For `apple_pay`, `google_pay`, `link`, `paypal`, `amazon_pay`, `klarna`: `active` or `inactive`, and Stripe's reason |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

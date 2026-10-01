# StripePaymentMethodConfiguration

Declares which [payment methods](https://docs.stripe.com/payments/payment-method-configurations) checkout offers, method by method: `"on"`, `"off"`, or `"none"` to leave it to Stripe.

## When to Use

- **The methods you meant, everywhere**: checkout, and a portal that references this configuration, offer exactly the methods declared.
- **Your own configuration, not the account default**: your application names this configuration's id when it creates a Checkout Session or PaymentIntent.
- **See what is really on**: `status.outputs.availablePaymentMethods` lists the methods Stripe reports available, so a method set on but missing its capability shows up.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentMethodConfiguration
metadata:
  name: checkout-methods
  org: acme-corp
  env: production
spec:
  name: Checkout methods
  card:
    preference: "on"
  applePay:
    preference: "on"
  googlePay:
    preference: "on"
```

## Fields

| Field | Description |
|---|---|
| `name` | A team-facing label |
| `active` | Defaults to `true`; `false` deactivates without destroying |
| `parent` | For a Connect platform, the parent configuration (`pmc_...`) a child inherits from. Changing it **replaces** the configuration |
| `<method>.preference` | One per payment method -- 59 in all: `card`, `applePay`, `googlePay`, `link`, `sepaDebit`, `usBankAccount`, `klarna`, `affirm`, `ideal`, `pix`, ... Values `on`, `off`, `none` |

## Key Behaviors

- **Quote `on` and `off` for other tools**: Planton reads them bare, but YAML 1.1 tools (PyYAML, older linters) read them as booleans.
- **Destroy deactivates**: Stripe never deletes a configuration, so destroy sets it inactive and Stripe keeps it forever.
- **A removed method keeps its last preference**: the provider sends only values that are set. Set it to `"none"` to hand it back to Stripe's default.
- **Capabilities still apply**: a method set on appears only where its capability is active on the account and the payment qualifies.
- **Naming a method the account is not offered**: creating and changing the configuration both work, and the method is simply not available. Such a configuration can't be adopted by import, though: Stripe returns no preference for that method, and it refuses the first update that names it ("not available to this account").
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The configuration's Stripe id (`pmc_...`), passed as `payment_method_configuration` |
| `isDefault` | Whether it is the account's default |
| `active` | Whether payments may use it |
| `availablePaymentMethods` | The methods Stripe reports available |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

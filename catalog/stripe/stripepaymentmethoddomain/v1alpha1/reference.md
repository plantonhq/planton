# StripePaymentMethodDomain

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

StripePaymentMethodDomainSpec registers a domain where the account's own checkout pages may
show wallet buttons -- Apple Pay, Google Pay, Link, PayPal, Amazon Pay, Klarna -- through Stripe
Elements or embedded Checkout. Stripe validates the domain for each wallet; status.outputs
reports each wallet's state with Stripe's own error message, so a wallet that stays inactive
says why (Apple Pay, for one, needs Stripe's association file served from the domain).

Stripe keeps a registered domain: destroy only forgets it. The domain stays registered, and
enabled, in Stripe after Planton deletes the resource. To turn the wallets off, set enabled
false and apply, then delete. A domain removed outside Planton makes the next plan fail (the
provider does not treat a missing domain as gone); remove it from state and apply again.

The key needs "Payment Method Domains" write (iac/permissions.yaml).

https://docs.stripe.com/payments/payment-methods/pmd-registration
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/payment_method_domain

## Example

```yaml
# The canonical example: the domain a checkout page is served from.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentMethodDomain
metadata:
  name: checkout-domain
  org: e2e-org
  env: testing
spec:
  domainName: pay.example.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.domainName` | `string` | yes |  |  |
| `spec.enabled` | `bool` |  | `true` |  |

## Field Details

### spec.domainName

`string` · required

domain_name is the hostname the wallet buttons appear on ("pay.example.com"), without a
scheme or path. Changing it REPLACES the registration: the new domain is registered and the old
one is forgotten, still registered in Stripe.

- rule: {"required":true,"string":{"hostname":true}}

### spec.enabled

`bool` · optional (explicit presence)

enabled is whether wallets may appear on the domain. It updates in place. Set it false and
apply before deleting the resource, because deleting leaves the domain enabled in Stripe.

- default: `true`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripePaymentMethodDomain, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the registration's Stripe id (pmd_...). |
| `status.outputs.enabled` | `bool` | enabled is whether wallets may appear on the domain. |
| `status.outputs.apple_pay_status` | `string` | apple_pay_status is "active" or "inactive". |
| `status.outputs.apple_pay_error_message` | `string` | apple_pay_error_message is Stripe's reason Apple Pay is inactive, when it is. |
| `status.outputs.google_pay_status` | `string` | google_pay_status is "active" or "inactive". |
| `status.outputs.google_pay_error_message` | `string` | google_pay_error_message is Stripe's reason Google Pay is inactive, when it is. |
| `status.outputs.link_status` | `string` | link_status is "active" or "inactive". |
| `status.outputs.link_error_message` | `string` | link_error_message is Stripe's reason Link is inactive, when it is. |
| `status.outputs.paypal_status` | `string` | paypal_status is "active" or "inactive". |
| `status.outputs.paypal_error_message` | `string` | paypal_error_message is Stripe's reason PayPal is inactive, when it is. |
| `status.outputs.amazon_pay_status` | `string` | amazon_pay_status is "active" or "inactive". |
| `status.outputs.amazon_pay_error_message` | `string` | amazon_pay_error_message is Stripe's reason Amazon Pay is inactive, when it is. |
| `status.outputs.klarna_status` | `string` | klarna_status is "active" or "inactive". |
| `status.outputs.klarna_error_message` | `string` | klarna_error_message is Stripe's reason Klarna is inactive, when it is. |

## See Also

- [Overview](../README.md)

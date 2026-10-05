# Stripe Payment Method Configuration Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout


## Payment Method Security Notes

### Methods Change Your Risk and Refund Story

Bank debits settle over days and can be disputed long after; buy-now-pay-later methods move repayment risk to the lender; wallets carry the device's authentication. Turn on a method when your fulfillment, refunds and reconciliation handle it, not because Stripe offers it.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, deactivate | Payment Method Configurations: Write | `POST` on `/v1/payment_method_configurations`; there is no delete |
| Read | Payment Method Configurations: Write | Write includes the read every refresh makes |

## Judgment
## Your Configuration, Never the Default

An account holds many payment-method configurations, and one is its default: the one Stripe uses when a checkout names none, and the one the Dashboard's payment-methods page edits. This kind never adopts or edits it. It creates a configuration of its own, and your application passes its `id` as `payment_method_configuration` every time it creates a Checkout Session or PaymentIntent. A Stripe Billing Portal Configuration can reference the same `id`, so checkout and the portal offer the same methods.

## Preference, Availability, and Capability

- **Preference** is what you declare: `"on"`, `"off"`, or `"none"` (Stripe's default, or a Connect parent's).
- **Capability** is whether the account may accept the method at all, turned on in the Dashboard.
- **Availability** is both together, and it is what `status.outputs.availablePaymentMethods` lists. Stripe reports no availability for cards.
- Even an available method appears only for payments that qualify: the currency, the customer's country, the amount.

## What Destroy and Removal Do

- **Destroy deactivates**: Stripe has no delete for a configuration. Destroy sets `active: false` and Stripe keeps it forever.
- **A removed method stays as it was**: the provider sends only values that are set. Set the method to `"none"` to hand it back to Stripe's default.

## Traps

- **A bare `on` means different things to different tools**: Planton reads `preference: on` as the preference, but YAML 1.1 tools (PyYAML, older linters) read it as `true`. Quote it (`"on"`) in files other tools also process.
- **Apple Pay Later is write-only**: Stripe never returns its setting, so it is sent on every apply and never shows drift.
- **French meal vouchers are missing**: Stripe's API accepts `fr_meal_voucher_conecs`, but the pinned provider does not expose it; it stays at Stripe's default.
- **`parent` replaces**: a Connect child configuration's parent cannot change in place.
- **A configuration that cannot be read fails the plan**: the provider has no handling for it. Remove it from state and apply again for a new one.

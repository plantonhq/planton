# Stripe Shipping Rate Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Shipping addresses**: collected and stored by Stripe on the Checkout session, never by this component

## Shipping Rate Security Notes

### Rates Are Public

A shipping rate's name and amount are shown on every page that offers it. Nothing about it is secret.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, deactivate | Shipping Rates: Write | `POST` on `/v1/shipping_rates`; there is no delete |
| Read | Shipping Rates: Write | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a shipping rate here only if nothing else creates or edits it. Rates your application computes per order belong to your application; a fixed menu of shipping options suits a chart you review like code.

## Changing a Shipping Rate

Stripe fixes a shipping rate's name, amount, currency, delivery window and tax code. Changing any of them replaces the rate: Planton deactivates the old one and creates the new one, with a new id. A StripePaymentLink that offers the rate by reference is replaced on the same apply, with a new address.

`taxBehavior`, `fixedAmount.currencyOptions`, `active` and `metadata` change in place. Once the tax behavior is `inclusive` or `exclusive`, Stripe refuses to change it.

## What Destroy Does

Destroy deactivates the rate; Stripe never deletes one. New purchases can't choose it, and orders already placed keep it.

## Importing a Rate Made in the Dashboard

Import the rate by its id (`shr_...`). Every field marked as replacing must match the Dashboard's rate exactly, or the first apply replaces it.

## Traps

- **Free is 0, not unset**: `amount: 0` is a real, free rate; the amount is always sent.
- **An open-ended window**: set only `minimum` ("at least 3 days") or only `maximum` ("up to a week").
- **A rate that cannot be read fails the plan**: the provider does not treat a missing rate as gone. Recover with `tofu state rm stripe_shipping_rate.this` and an apply, which creates a new one.

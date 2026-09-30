# Stripe Tax Rate Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Tax Rate Security Notes

### A Rate Is Not Tax Advice

Planton declares the rate you give it. Whether that rate is right for a customer, a product and a date is a question for your tax adviser, or for Stripe Tax, which calculates it.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, deactivate | Tax Rates: Write | `POST` on `/v1/tax_rates`; there is no delete |
| Read | Tax Rates: Write | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a tax rate here only if nothing else creates or edits it. When Stripe Tax calculates your tax, you need no manual rates at all; when your application creates rates per customer, they belong to your application.

## Changing a Tax Rate

Stripe fixes a rate's `percentage` and `inclusive`. Changing either replaces the rate: Planton deactivates the old one and creates the new one, with a new id. Subscriptions and invoices that use the old rate keep it until your code or the Dashboard moves them to the new one, because a rate is applied by id.

`displayName`, `description`, `country`, `state`, `jurisdiction`, `taxType`, `active` and `metadata` change in place.

## What Destroy Does

Destroy deactivates the rate; Stripe never deletes one. It can't be added to anything new, and it still applies to the subscriptions and invoices that already use it.

## Importing a Rate Made in the Dashboard

Import the rate by its id (`txr_...`). `percentage` and `inclusive` must match the Dashboard's rate exactly, or the first apply replaces it.

## Traps

- **A new rate doesn't move subscriptions**: after a rate change, existing subscriptions still carry the old, deactivated rate. Move them in your code or the Dashboard.
- **The provider accepts 14 tax types**: amusement, communications, GST, HST, IGST, JCT, lease, PST, QST, retail delivery fee, RST, sales, service and VAT. Stripe's reference lists more; the pinned provider refuses them.
- **A rate that cannot be read fails the plan**: the provider does not treat a missing rate as gone. Recover with `tofu state rm stripe_tax_rate.this` and an apply, which creates a new one.

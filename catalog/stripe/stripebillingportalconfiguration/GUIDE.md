# Stripe Billing Portal Configuration Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout


## Portal Security Notes

### A Session Is a Key to the Customer's Billing

A portal session URL lets whoever holds it act on that customer's subscriptions and payment methods until it expires. Open sessions only for the signed-in customer, from your server, and never log their URLs.

### The Login Page Is Public

`loginPage` creates a shareable URL where a customer signs in by email. Enable it only when your application cannot open sessions itself.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, deactivate | Customer portal: Write | `POST` on `/v1/billing_portal/configurations`; there is no delete |
| Read | Customer portal: Write | Write includes the read every refresh makes |

## Judgment
## Your Configuration, Never the Default

An account holds many portal configurations, and one is its default: the one Stripe uses when a session names none, and the one the Dashboard's portal settings page edits. This kind never adopts or edits it. It creates a configuration of its own, and your application passes its `id` as `configuration` every time it opens a session. So every customer sees what the manifest declares, whatever anyone does to the default.

## What Destroy and Removal Do

- **Destroy deactivates**: Stripe has no delete for a configuration. Destroy sets `active: false`, Stripe keeps the configuration forever, and a session that names it is refused.
- **A removed setting stays in Stripe**: the provider sends only values that are set, so deleting a feature from the manifest leaves it as it was. Set `enabled: false` instead. The same holds for every optional field.
- **Deactivated outside Planton**: the next apply reactivates it, since `active` defaults to `true`.

## Traps

- **Prorations need an immediate cancellation**: `prorationBehavior` `always_invoice` or `create_prorations` settles unused time, which only `mode: immediately` leaves. The spec refuses the combination.
- **Price changes need prices**: allowing `price` in `defaultAllowedUpdates` without listing `products` leaves nothing to switch to. The spec refuses it.
- **Product and price ids are the account's own**: a test-mode `price_...` does not exist in live mode. Reference the environment's own Stripe Product and Stripe Price resources rather than pasting ids, and each environment's portal names its own.
- **A replaced price leaves the portal pointing at the archived one until the portal applies again**: a price whose amount changed has a new id. A reference picks it up on the portal's next apply; a pasted id never does.
- **A configuration that cannot be read fails the plan**: the provider has no handling for it. Remove it from state and apply again for a new one.

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
- **The account's first configuration becomes its default, and the default can't be deactivated**: in an account (or sandbox) that has no portal configuration yet, Stripe makes the first one it is given the default, and destroy then fails with Stripe's "You cannot set `active: false` on your default PortalConfiguration". An account that already has a default (every account whose customers have opened a portal) never makes this kind's configuration the default. In a fresh account the first configuration stays the default: its destroy keeps failing, and the way out is `tofu state rm stripe_billing_portal_configuration.this`, which leaves it active in Stripe as the account's default.

## Traps

- **Prorations need an immediate cancellation**: `prorationBehavior` `always_invoice` or `create_prorations` settles unused time, which only `mode: immediately` leaves. The spec refuses the combination.
- **Subscription changes can't be declared on the pinned provider**: Stripe requires switchable products whenever `subscriptionUpdate` is on, even to change only quantities, and returns a configuration's products only when a read expands them. The provider never expands, so a create that names products succeeds in Stripe and then fails, every time. The spec refuses `subscriptionUpdate.products` and `subscriptionUpdate.enabled`. Cancellation, payment methods, customer details and invoices all work.
- **A configuration that cannot be read fails the plan**: the provider has no handling for it. Remove it from state and apply again for a new one.

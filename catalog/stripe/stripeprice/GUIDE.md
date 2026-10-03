# Stripe Price Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Price Security Notes

### Prices Are Public, Lookup Keys Are Not Secrets

A price's amount and currency are shown on every page that sells it. A lookup key is a stable name, not a credential: anyone with your secret key can list prices by it, and nobody without it can.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, archive | Prices: Write | `POST` on `/v1/prices`; there is no delete |
| Read | Prices: Write | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a price here only if nothing else creates or edits it. When your application's own code, a billing system, or the Dashboard already owns your price list, leave its products and prices there: two owners overwrite each other on every apply, and neither notices until a customer is charged the wrong amount. Declared products and prices suit a catalog you review like code; a catalog your application generates belongs to your application.

## Changing a Price's Amount

Stripe never changes a price's amount, currency, product, scheme, tiers or schedule. Changing any of them in the manifest replaces the price:

1. Planton creates the new price.
2. With `transferLookupKey: true`, the lookup key moves from the old price to the new one in that same call, so code that finds the price by key never sees a gap.
3. Planton archives the old price. Customers already subscribed keep paying the old amount; new checkouts get the new one.

Without `transferLookupKey`, creating the new price fails while the old one still holds the key, and nothing changes. Set it together with the amount change. The provider sends it on every apply while it is set; whether Stripe accepts it on an apply that changes nothing else is proven in the live lanes before any preset leaves it on permanently.

## What Destroy and Other Changes Do

- **Destroy archives**: subscriptions on the price keep billing, new purchases are refused, and Stripe keeps it.
- **In place**: `nickname`, `lookupKey`, `metadata`, `currencyOptions`, `taxBehavior` and `active` change without a new price.
- **Tax behavior is fixed once chosen**: after `inclusive` or `exclusive`, Stripe refuses to change it.
- **Currency options are not read back**: Stripe does not return them, so a change made in the Dashboard is not detected, and a currency removed from the manifest stays in Stripe.

## Importing a Price Made in the Dashboard

Import the price by its id (`price_...`). Every field that shapes the charge must match the Dashboard's price exactly, or the first apply replaces it.

## Traps

- **No hidden products**: the provider can create a product inside a price, but that product would be invisible to Planton and never archived. Reference a StripeProduct instead.
- **A portal lists prices by id**: after a replacement, a portal that pasted the old id still offers the archived price. Reference the StripePrice so the portal follows it on its next apply.
- **Metered prices need a meter**: set `recurring.usageType: metered` and name the StripeBillingMeter that counts usage by reference. A metered price without a meter is refused before anything runs, and a replaced meter replaces the price with it.
- **A price that cannot be read fails the plan**: the provider does not treat a missing price as gone. Recover with `tofu state rm stripe_price.this` and an apply, which creates a new one.

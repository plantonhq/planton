# Stripe Payment Link Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the payment page and the API are HTTPS-only
- **Card data**: collected and stored by Stripe on its hosted page, never by this component or your site
- **What the page collects** (addresses, names, phone numbers, tax ids, custom answers) is stored by Stripe on each Checkout session, never by this component

## Payment Link Security Notes

### The Page Is Public, and It Takes Payments

A payment link is a public page by design: anyone with its address can pay for the declared prices, from the moment it is applied. Applying it moves no money and creates no customer record; buyers do both. Control its reach with where you publish the address, `active`, and `restrictions.completedSessions` (deactivate after a number of payments). Destroy deactivates it.

### Connect Fields Decide Where Money Goes

`transferData.destination` sends each payment's funds to a connected account, `onBehalfOf` makes that account the settlement merchant (its name is shown to the buyer), and `applicationFeeAmount` or `applicationFeePercent` is what the platform keeps. Review these like a bank account change: they decide where every future buyer's money goes. Changing any of them creates a new link.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, deactivate | Payment Links: Write | `POST` on `/v1/payment_links`; there is no delete |
| Read | Payment Links: Write | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a link here only if nothing else creates or edits it. Links your application creates per customer or per order belong to your application, or to Checkout sessions; a few public pages your site links to suit a chart you review like code.

## Changing the Price a Link Sells

Stripe can't change which price a line item sells, and the pinned provider never stores a line item's quantity, so neither change could be sent to an existing link. Planton replaces the link instead:

1. Planton creates a new link with the new prices and quantities. It has a new address.
2. Your site, reading `status.outputs.url` by reference, picks up the new address on the same apply.
3. Planton deactivates the old link. Visitors to the old address see `inactiveMessage`, so write it for them: "This plan has moved. Visit example.com/pricing."

A referenced price that is itself replaced (an amount change) replaces the link too, because the link sells the declared price. Adjustable quantities, optional items and every other field change in place, except the ones the provider replaces: `currency`, `consentCollection`, `managedPayments`, `paymentIntentData.captureMethod` and `setupFutureUsage`, `shippingOptions`, `subscriptionData.description`, and the Connect fields.

## Rules Stripe Checks at Apply

Validation can't see the price behind a reference, so three of Stripe's rules are checked when the link is created:

- `applicationFeeAmount` needs a link with no recurring prices; `applicationFeePercent` needs one with a recurring price.
- `paymentMethodCollection` and `subscriptionData` need a subscription (a recurring price).
- `consentCollection.termsOfService: required` needs your terms of service URL in the Dashboard's public details.

## What Destroy Does

Destroy deactivates the link; Stripe never deletes one. Its address shows `inactiveMessage`, and payments already taken, with their customers and subscriptions, stay.

## Importing a Link Made in the Dashboard

Import the link by its id (`plink_...`). The first apply after the import creates the module's line-item tracker, which holds nothing in Stripe, and leaves the link untouched. Every field the provider replaces must match the Dashboard's link exactly, or that first apply replaces it.

## Traps

- **An old address stops working**: anywhere the address is pasted rather than read by reference (an email, a printed QR code) shows `inactiveMessage` after a price change.
- **Quantities are never read back**: the provider doesn't store them, so a quantity changed in the Dashboard is not detected.
- **A link that cannot be read fails the plan**: the provider does not treat a missing link as gone. Recover with `tofu state rm stripe_payment_link.this` and an apply, which creates a new one with a new address.

# Stripe Coupon Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Coupon Security Notes

### A Coupon Is Not a Secret

A coupon's id is not a credential: your application applies it with your secret key, and customers never see it. Customers redeem a discount through a promotion code, and a code's limits (first-time customers, one customer, a cap, an expiry) are what keep it from being shared too widely.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, delete | Coupons: Write | `POST` and `DELETE` on `/v1/coupons` |
| Read | Coupons: Write | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a coupon here only if nothing else creates or edits it. When your application's own code, a billing system, or the Dashboard already owns your discounts, leave them there: two owners overwrite each other on every apply. Declared coupons suit a campaign you review like code; coupons your application creates per customer belong to your application.

## Changing a Coupon

Stripe never changes a coupon's discount, currency, duration, months, redemption limit, redeem-by date or products. Changing any of them in the manifest replaces the coupon:

1. Planton deletes the old coupon. Customers and subscriptions that already applied it keep their discount.
2. Planton creates the new coupon, with a new id.
3. Every StripePromotionCode that references the coupon is replaced on the same apply, so its code keeps working on the new coupon.

A coupon's `name` and `metadata` change in place. Changing `currencyOptions` replaces it too: Stripe refuses a new amount for a currency the coupon already has.

Stripe also never hands a coupon's products or its amounts in other currencies back to the provider's read. The module keeps both in a small tracker beside the coupon, so a change to either still replaces it, and an imported coupon keeps its id: the first apply after an import creates the tracker and leaves the coupon untouched.

## What Destroy Does

Destroy deletes the coupon in Stripe. Stripe keeps the discount on every customer and subscription that already applied it; nobody new can redeem it (Stripe's API reference, "Delete a coupon"). Delete the promotion codes first, or together in one chart, so no code points at a deleted coupon.

## Importing a Coupon Made in the Dashboard

Import the coupon by its id. Coupons made in the Dashboard or the API may carry an id the account chose; a literal coupon id in a promotion code is accepted as it is. Every field marked as replacing must match the Dashboard's coupon exactly, or the first apply replaces it.

## Traps

- **Dates are Unix seconds**: `redeemBy` takes seconds since 1970 (1798761599 is 2026-12-31 23:59:59 UTC; on Linux, `date -u -d 2026-12-31T23:59:59Z +%s`).
- **A code can't outlast its coupon**: a promotion code's expiry and redemption cap can't exceed the coupon's `redeemBy` and `maxRedemptions`; Stripe refuses either at apply.
- **A coupon that cannot be read fails the plan**: the provider does not treat a missing coupon as gone. Recover with `tofu state rm stripe_coupon.this` and an apply, which creates a new one.

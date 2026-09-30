# Stripe Promotion Code Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Promotion Code Security Notes

### A Public Code Is Public

Anyone who learns an unrestricted code can redeem it, and codes spread. The code's limits are its access control: `restrictions.firstTimeTransaction`, `customer`, `maxRedemptions` and `expiresAt`. Give every public code at least a cap or an expiry.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, deactivate | Promotion Codes: Write | `POST` on `/v1/promotion_codes`; there is no delete |
| Read | Promotion Codes: Write | Write includes the read every refresh makes |

## Judgment
## One Owner per Object

Declare a code here only if nothing else creates or edits it. Codes your application mints per customer (a referral code, a support credit) belong to your application; codes a campaign publishes suit a chart you review like code.

## Changing a Code

Stripe fixes everything that decides who may redeem a code. Changing the coupon, the code, the customer, the expiry, the cap or the restrictions replaces the code:

1. Planton deactivates the old code. Discounts it already applied stay.
2. Planton creates the new code. The same text works, because a code only has to be unique among active codes.

`active` and `metadata` change in place. Setting `active: false` pauses a code; setting it back to `true` reactivates it, unless another active code has taken the same text meanwhile.

## What Destroy Does

Destroy deactivates the code; Stripe never deletes one. Nobody can redeem it, discounts already applied stay, and the same code can be declared again later.

## Limits the Coupon Sets

A code's `expiresAt` can't be later than its coupon's `redeemBy`, and its `maxRedemptions` can't exceed the coupon's (Stripe's API reference). Validation can't see the coupon behind a reference, so Stripe refuses a code that breaks either at apply.

## Importing a Code Made in the Dashboard

Import the code by its id (`promo_...`), not by the text customers type. Every field marked as replacing must match the Dashboard's code exactly, or the first apply replaces it.

## Traps

- **Codes are case-insensitive for customers**: `launch25` and `LAUNCH25` are the same code to Stripe's uniqueness rule.
- **Dates are Unix seconds**: `expiresAt` takes seconds since 1970 (1798761599 is 2026-12-31 23:59:59 UTC).
- **A code that cannot be read fails the plan**: the provider does not treat a missing code as gone. Recover with `tofu state rm stripe_promotion_code.this` and an apply, which creates a new one.

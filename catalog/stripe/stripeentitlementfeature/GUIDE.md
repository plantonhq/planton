# Stripe Entitlement Feature Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Entitlement Feature Security Notes

### Check Entitlements on Your Server

A customer's active entitlements come from Stripe's API with your secret key. Check them on your server, never trust a client that says what it is entitled to.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update, archive | Entitlements: Write | `POST` on `/v1/entitlements/features`; there is no delete |
| Read | Entitlements: Write | Write includes the read every refresh makes |

## Judgment
## Features, Products and Your Code

A feature is a name for a capability. A product grants it: every customer subscribed to a product that lists the feature holds it, and Stripe reports that customer's active entitlements to your application. So your code checks `api-access`, and which plans include the API is a manifest decision, not a code change.

## What Destroy and Changes Do

- **Destroy archives**: Stripe has no delete for a feature. Destroy sets it inactive; Stripe keeps it, it drops out of the features list, and it can no longer be attached to new products.
- **No way back**: the provider never sends `active`, so an archived feature -- by destroy or in the Dashboard -- stays archived. Declare a new one with a new lookup key.
- **A new lookup key replaces the feature**: the old one is archived, and every product that referenced this resource attaches the new one on its next apply.

## Traps

- **Reusing an archived lookup key**: whether Stripe lets a new feature take an archived feature's lookup key is not documented. Choose a new key rather than recreating an old one.
- **A feature that cannot be read fails the plan**: the provider does not treat a missing feature as gone. Recover with `tofu state rm stripe_entitlements_feature.this` and an apply, which creates a new one.

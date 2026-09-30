# Stripe Payment Method Domain Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Payment Method Domain Security Notes

### Register Only Domains You Serve

A registration lets wallet buttons appear on pages from that host with your account's payments behind them. Register only hosts you control, and remove a registration (enabled false, then delete) when a host changes hands.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Register, update | Payment Method Domains: Write | `POST` on `/v1/payment_method_domains`; destroy makes no call |
| Read | Payment Method Domains: Write | Write includes the read every refresh makes |

## Judgment
## What Deleting Does, and Does Not Do

Stripe keeps a registered domain. Destroying this resource only removes it from Planton's state: the provider makes no API call, and the domain stays registered, and enabled, in Stripe. Wallet buttons keep appearing on it.

To retire a domain:

1. Set `enabled: false` and apply. Wallets stop appearing.
2. Delete the resource.

To bring a forgotten domain back under Planton, import it by its id (`pmd_...`) -- Stripe still has it.

## Why a Wallet Stays Inactive

Stripe validates the domain for each wallet on its own. Each `<wallet>_status` output reads `active` or `inactive`, and `<wallet>_error_message` carries Stripe's reason. The most common one: Apple Pay needs Stripe's domain association file served from the domain before it goes active. The provider never asks Stripe to validate again; Stripe re-checks on its own schedule.

## Traps

- **Changing `domainName` leaves the old domain registered**: the new domain is registered and the old one only forgotten.
- **Registering a domain again**: whether Stripe accepts a second registration of a domain it already holds is not documented; import the existing one instead.
- **A domain that cannot be read fails the plan**: the provider does not treat a missing domain as gone. Recover with `tofu state rm stripe_payment_method_domain.this` and an apply, which creates a new one.

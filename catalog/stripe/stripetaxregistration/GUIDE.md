# Stripe Tax Registration Guide

## Security
## Platform Security Posture

The certifications below are Stripe's own published claims about its platform (verify current status at docs.stripe.com/security). They describe the vendor's service -- never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Stripe's published certifications and security standards:

- PCI DSS Level 1 Service Provider (audited annually by an independent Qualified Security Assessor)
- SOC 1 and SOC 2 Type II reports

## Data Protection

- **Encryption in transit**: the API and every Stripe-hosted page are HTTPS-only
- **Card data**: collected and stored by Stripe, never by this component or your application when you use Stripe-hosted checkout

## Tax Registration Security Notes

### A Registration Is Not Tax Advice

Planton declares the registration you give it. Whether and when you must register in a place is a question for your tax adviser. Declaring a registration tells Stripe you hold it; it does not register you with the tax authority.

## Permissions
## Restricted-Key Permissions

| Operation | Permission | Description |
|-----------|------------|-------------|
| Create, update | Tax Registrations: Write | `POST` on `/v1/tax/registrations`; there is no delete |
| Read | Tax Registrations: Write | Write includes the read every refresh makes |

## Judgment
## What Applying Does

From `activeFrom`, Stripe Tax calculates and collects tax in the registered place on every payment that uses automatic tax. Stripe Tax itself (turning it on, your origin address, your default tax code) is your account's setting in the Dashboard; this kind only declares where you are registered.

## Choosing the Type

The country decides which types exist:

- **An EU member state**: `standard` (registered in that country), `oss_union` or `oss_non_union` (the One-Stop Shop, declared in the country you registered it in), or `ioss` (the Import One-Stop Shop). A `standard` registration may set `placeOfSupplyScheme`: `standard`, `small_seller` or `inbound_goods`.
- **Canada**: `standard` (GST/HST), `province_standard` (a province's own tax, with `province`) or `simplified`.
- **The United States**: `state_sales_tax`, `state_communications_tax` or `state_retail_delivery_fee`, each with `state`; or `local_amusement_tax` or `local_lease_tax`, with `state` and the local jurisdiction's FIPS code in `jurisdiction`. A `state_sales_tax` registration may carry `stateSalesTaxElections`. Each state offers its own elections, and Stripe refuses one the state doesn't take, naming those it does: Texas takes only `single_local_use_tax`.
- **Standard only**: AE, AL, AO, AU, AW, BA, BB, BD, BF, BH, BS, CD, CH, ET, GB, GN, IS, JP, ME, MK, MR, NO, NZ, OM, RS, SG, SR, UY, ZA, ZW. A registration may set `placeOfSupplyScheme` to `standard` or `inbound_goods`.
- **Simplified only**: AM, AZ, BJ, BY, CL, CM, CO, CR, CV, EC, EG, GE, ID, IN, KE, KG, KH, KR, KZ, LA, LK, MA, MD, MX, MY, NG, NP, PE, PH, RU, SA, SN, TH, TJ, TR, TW, TZ, UA, UG, UZ, VN, ZM.

The EU member states are AT, BE, BG, CY, CZ, DE, DK, EE, ES, FI, FR, GR, HR, HU, IE, IT, LT, LU, LV, MT, NL, PL, PT, RO, SE, SI and SK. These 101 countries are the ones the pinned provider accepts. Validation refuses any other country, and any type or field a country does not take.

## Dates

Both dates are Unix seconds (`date -u -d 2026-01-01T00:00:00Z +%s` on Linux gives `1767225600`).

- **`activeFrom` must be now or later, and at most five years ahead, when the registration is created.** Stripe refuses a start in the past, and one further out than five years. Recreating a registration in a new account from an old file means moving its start date forward first. An imported registration keeps Stripe's own start date.
- **`expiresAt` can't be more than five years ahead either.**
- **`expiresAt` is the only way to stop collecting.** Set it and apply; the registration stops at that time.
- **Removing `expiresAt` does not clear it.** The stored date stays. To change an expiry, set a new date.

## Changing a Registration

Only `activeFrom` and `expiresAt` update in place. Changing the country, type, place-of-supply scheme, province, state, jurisdiction or elections REPLACES the registration: Planton creates a new one and forgets the old one, which stays active in Stripe. So a change is two steps:

1. Set `expiresAt` on the old registration and apply.
2. Declare the new registration as its own resource.

## What Destroy Does

Stripe never deletes a registration. Destroy only removes it from Planton's state: the provider makes no call, and Stripe keeps collecting. Set `expiresAt` and apply before destroying if collection should stop.

## Importing a Registration Made in the Dashboard

Import the registration by its id (`taxreg_...`, listed by `GET /v1/tax/registrations`). The manifest must carry the registration's own `activeFrom`, country and options: a different start date is sent on the first apply, and a different type or option replaces the registration.

## One Owner per Object

Declare a registration here only if nothing else creates or expires it. When your team manages registrations in the Dashboard, leave them there. Manual tax rates (Stripe Tax Rate) are the alternative for accounts that do not use Stripe Tax.

## Traps

- **A changed type leaves two registrations collecting**: the old one is forgotten, not expired. Expire it first.
- **A start date in the past is refused at creation**: move it to now or later.
- **A registration that cannot be read fails the plan**: the provider does not treat a missing registration as gone. Recover with `tofu state rm stripe_tax_registration.this` and an apply, which creates a new one.

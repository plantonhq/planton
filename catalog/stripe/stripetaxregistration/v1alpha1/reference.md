# StripeTaxRegistration

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeTaxRegistrationSpec declares one place the account is registered to collect tax with
Stripe Tax -- VAT in Germany under the EU's One-Stop Shop, sales tax in Texas -- so the account's
registrations live in files, the same in every environment.

Applying it tells Stripe the account is registered there: from active_from, Stripe Tax
calculates and collects tax in that place on every payment that uses automatic tax. Stripe Tax
itself (turning it on, the origin address, the default tax code) is the account's setting in the
Dashboard, outside this kind.

Stripe never deletes a registration. Destroy only removes it from Planton's state: the provider
makes no call, and Stripe keeps collecting. To stop collecting, set expires_at and apply, then
destroy.

Only active_from and expires_at update in place. Changing the country, type, place-of-supply
scheme, province, state, jurisdiction or elections REPLACES the registration: Planton creates a
new one and forgets the old one, which stays active in Stripe. To change a registration, set
expires_at on the old one and declare the new one as its own resource.

A registration that cannot be read makes the plan fail (the provider does not treat a missing
registration as gone); remove it from state and apply again.

The key needs "Tax Registrations" write (iac/permissions.yaml).

https://docs.stripe.com/tax/registering
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/tax_registration

## Example

```yaml
# The canonical example: the EU One-Stop Shop, declared in Germany, starting
# 1 January 2028 and expiring a day later. Stripe accepts a start from now to
# five years ahead, and nothing deletes a registration; the live scenarios
# take their dates from the run's clock.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeTaxRegistration
metadata:
  name: de-oss
  org: e2e-org
  env: testing
spec:
  country: DE
  type: oss_union
  activeFrom: 1830297600
  expiresAt: 1830384000
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.country` | `string` | yes |  |  |
| `spec.type` | `enum` |  |  |  |
| `spec.activeFrom` | `int64` |  |  |  |
| `spec.expiresAt` | `int64` |  |  |  |
| `spec.placeOfSupplyScheme` | `enum` |  |  |  |
| `spec.province` | `string` |  |  |  |
| `spec.state` | `string` |  |  |  |
| `spec.jurisdiction` | `string` |  |  |  |
| `spec.stateSalesTaxElections` | `[]StripeTaxRegistrationStateSalesTaxElection` |  |  |  |
| `spec.stateSalesTaxElections[].type` | `enum` |  |  |  |
| `spec.stateSalesTaxElections[].jurisdiction` | `string` |  |  |  |

## Field Details

### spec.country

`string` · required

country is the two-letter ISO country code of the registration, in capitals ("DE", "US").
Changing it REPLACES the registration.

- rule: country is a two-letter ISO country code in capitals, such as DE
- rule: {"required":true}

### spec.type

`enum`

type is the kind of registration, and the country decides which types exist:
- an EU member state: standard (registered in that country), oss_union or oss_non_union (the
  One-Stop Shop, declared in the country the account registered it in), or ioss (the Import
  One-Stop Shop);
- Canada: standard (GST/HST), province_standard (a province's own tax, with province) or
  simplified;
- the United States: state_sales_tax, state_communications_tax or state_retail_delivery_fee
  (with state), or local_amusement_tax or local_lease_tax (with state and jurisdiction);
- every other supported country takes exactly one: standard or simplified, as Stripe lists it.
Changing it REPLACES the registration.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `type_unspecified`
- `standard`
- `simplified`
- `ioss`
- `oss_union`
- `oss_non_union`
- `province_standard`
- `state_sales_tax`
- `local_amusement_tax`
- `local_lease_tax`
- `state_communications_tax`
- `state_retail_delivery_fee`

### spec.activeFrom

`int64`

active_from is when the registration starts, in Unix seconds (1767225600 is
2026-01-01T00:00:00Z; on Linux `date -u -d 2026-01-01T00:00:00Z +%s` computes one). Stripe
accepts only now or a future time no more than five years ahead when the registration is
created, so recreating a registration in a new account means moving a start date that has
passed forward; an imported registration keeps Stripe's own value. It updates in place.

- rule: {"int64":{"gt":"0"}}

### spec.expiresAt

`int64` · optional (explicit presence)

expires_at is when the registration stops, in Unix seconds, and the only way to stop Stripe
collecting in that place. Stripe refuses one more than five years ahead. Unset, the
registration never expires. It updates in place, but
removing it from the manifest does not clear it in Stripe: the stored date stays, and a new
date is the only change.

- rule: {"int64":{"gt":"0"}}

### spec.placeOfSupplyScheme

`enum`

place_of_supply_scheme refines a standard registration in a country that has one: standard,
inbound_goods (goods imported into the country), or, in the EU only, small_seller. Unset,
Stripe applies its default. Changing it REPLACES the registration.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `stripe_tax_registration_place_of_supply_scheme_unspecified`
- `standard`
- `small_seller`
- `inbound_goods`

### spec.province

`string`

province is the two-letter Canadian province code of a province_standard registration ("QC").
Changing it REPLACES the registration.

- rule: province is a two-letter Canadian province code in capitals, such as QC

### spec.state

`string`

state is the two-letter US state code of a United States registration ("TX"), required there.
Changing it REPLACES the registration.

- rule: state is a two-letter US state code in capitals, such as TX

### spec.jurisdiction

`string`

jurisdiction is the FIPS code of the local jurisdiction of a local_amusement_tax or
local_lease_tax registration, required for those two types. Changing it REPLACES the
registration.

### spec.stateSalesTaxElections

`[]StripeTaxRegistrationStateSalesTaxElection`

state_sales_tax_elections are the local-tax elections of a state_sales_tax registration, for
states that offer them. Changing them REPLACES the registration.

### spec.stateSalesTaxElections[].type

`enum`

type is the election: local_use_tax, simplified_sellers_use_tax or single_local_use_tax.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `election_type_unspecified`
- `local_use_tax`
- `simplified_sellers_use_tax`
- `single_local_use_tax`

### spec.stateSalesTaxElections[].jurisdiction

`string`

jurisdiction is the FIPS code of the local jurisdiction the election applies to, where the
election is local.

## Validation Rules

- `spec.country.supported`: country is not one Stripe Tax registers in: the pinned provider accepts registrations in 101 countries, listed in this kind's GUIDE
- `spec.type.eu`: an EU registration's type is standard, oss_union, oss_non_union or ioss
- `spec.type.standard_only`: a registration in this country is standard: Stripe offers no other type there
- `spec.type.simplified_only`: a registration in this country is simplified: Stripe offers no other type there
- `spec.type.ca`: a Canadian registration's type is standard, province_standard or simplified
- `spec.type.us`: a US registration's type is state_sales_tax, local_amusement_tax, local_lease_tax, state_communications_tax or state_retail_delivery_fee
- `spec.place_of_supply_scheme.where`: place_of_supply_scheme refines a standard registration in a country that has one: the EU, AE, AL, AO, AU, AW, BA, BB, BD, BF, BH, BS, CD, CH, ET, GB, GN, IS, JP, ME, MK, MR, NO, NZ, OM, RS, SG, SR, UY, ZA or ZW
- `spec.place_of_supply_scheme.small_seller_eu`: place_of_supply_scheme small_seller exists only for EU registrations
- `spec.province.province_standard`: a province_standard registration names its province, and only it takes one
- `spec.state.us`: a US registration names its state, and only a US registration takes one
- `spec.jurisdiction.local_tax`: a local_amusement_tax or local_lease_tax registration names its jurisdiction's FIPS code, and only those two take one
- `spec.state_sales_tax_elections.state_sales_tax`: state_sales_tax_elections belong to a state_sales_tax registration only
- `spec.expires_at.after_active_from`: expires_at is later than active_from

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeTaxRegistration, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the registration's Stripe id (taxreg_...). It changes when the registration is replaced. |
| `status.outputs.status` | `string` | status is Stripe's reading of the dates: scheduled (active_from is still ahead), active, or expired (expires_at has passed). |

## See Also

- [Overview](../README.md)

# StripeTaxRate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeTaxRateSpec declares a manual tax rate -- VAT at 19% in Germany, a state sales tax -- that
invoices, subscriptions, Checkout sessions and payment links apply when the account calculates
tax itself rather than with Stripe Tax.

A tax rate's percentage and whether it is inclusive can never change in Stripe. Changing either
REPLACES the rate: Planton deactivates the old one and creates the new one. Its name,
description, country, state, jurisdiction, tax type, active and metadata update in place.

One owner per object: declare a tax rate here only if nothing else creates or edits it.

Stripe never deletes a tax rate. Destroy deactivates it (active = false): it can't be added to
anything new, and it still applies to the subscriptions and invoices that already use it. A rate
that cannot be read makes the plan fail (the provider does not treat a missing rate as gone);
remove it from state and apply again.

The key needs "Tax Rates" write (iac/permissions.yaml).

https://docs.stripe.com/billing/taxes/tax-rates
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/tax_rate

## Example

```yaml
# The canonical example: German VAT at 19%, added on top.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeTaxRate
metadata:
  name: de-vat
  org: e2e-org
  env: testing
spec:
  displayName: VAT
  percentage: 19
  inclusive: false
  country: DE
  jurisdiction: DE
  description: German VAT, standard rate
  taxType: vat
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.displayName` | `string` | yes |  |  |
| `spec.percentage` | `double` |  |  |  |
| `spec.inclusive` | `bool` |  |  |  |
| `spec.country` | `string` |  |  |  |
| `spec.state` | `string` |  |  |  |
| `spec.jurisdiction` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.taxType` | `enum` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.displayName

`string` · required

display_name is the tax's short name as customers see it on invoices and receipts ("VAT",
"Sales tax"). It updates in place.

- rule: {"required":true}

### spec.percentage

`double`

percentage is the rate, 0 to 100 ("19" is 19%). Changing it REPLACES the rate.

- rule: {"double":{"lte":100,"gte":0}}

### spec.inclusive

`bool`

inclusive is whether the tax is already included in the amount it applies to (true), or
added on top of it (false). Changing it REPLACES the rate.

### spec.country

`string`

country is the two-letter ISO country code the tax applies in ("DE"). It updates in place.

- rule: country is a two-letter ISO country code in capitals, such as DE

### spec.state

`string`

state is the ISO 3166-2 subdivision code the tax applies in, without the country prefix
("NY" for New York). It updates in place.

### spec.jurisdiction

`string`

jurisdiction is the tax authority's name, shown on invoices for the account's records. It
updates in place.

### spec.description

`string`

description is a note about the rate for the account's team, hidden from customers. It
updates in place.

### spec.taxType

`enum`

tax_type is the kind of tax. Unset, Stripe records none. It updates in place.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `tax_type_unspecified`
- `amusement_tax`
- `communications_tax`
- `gst`
- `hst`
- `igst`
- `jct`
- `lease_tax`
- `pst`
- `qst`
- `retail_delivery_fee`
- `rst`
- `sales_tax`
- `service_tax`
- `vat`

### spec.active

`bool` · optional (explicit presence)

active is whether the rate can be added to new invoices, subscriptions and sessions. Setting
it false deactivates the rate without destroying the resource; it still applies wherever it
is already used.

- default: `true`

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the rate. It updates in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeTaxRate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the tax rate's Stripe id (txr_...), the value an invoice, a subscription or a Checkout session's tax rates reference. It changes when the rate is replaced. |
| `status.outputs.active` | `bool` | active is false once the rate is deactivated (by destroy, by spec.active, or in the Dashboard). |

## See Also

- [Overview](../README.md)

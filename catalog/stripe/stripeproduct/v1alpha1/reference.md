# StripeProduct

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeProductSpec declares one thing an account sells -- a plan, a seat, a physical good -- as
Checkout, Payment Links, subscriptions and pricing tables show it, and the entitlement features
buying it grants. Its prices are separate resources: each StripePrice names this product.

One owner per object: declare a product here only if nothing else creates or edits it. When the
application's own code or the Dashboard owns the price list, leave its products and prices
there -- two owners overwrite each other on every apply.

Stripe never deletes a product this kind created. Destroy archives it (active = false): it can
no longer be bought, and existing subscriptions keep billing. A setting removed from the
manifest is not reset in Stripe -- the provider sends only values that are set -- so change a
setting to a new value rather than deleting it. A product that cannot be read makes the plan
fail (the provider does not treat a missing product as gone); remove it from state and apply
again.

The provider's default_price_data (a price created inside the product) is not offered: that
price would be invisible to Planton, never archived on destroy, and its drift never detected.
Declare a StripePrice that names this product instead.

The key needs "Products" write, and "Entitlements" write when features are granted
(iac/permissions.yaml).

https://docs.stripe.com/products-prices/overview
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/product

## Example

```yaml
# The canonical example: a SaaS plan as customers see it. Its prices are
# StripePrice resources that name it.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeProduct
metadata:
  name: pro-plan
  org: e2e-org
  env: testing
spec:
  name: Pro
  description: For growing teams
  unitLabel: seat
  marketingFeatures:
    - name: Unlimited projects
    - name: Priority support
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.type` | `enum` |  |  |  |
| `spec.images` | `[]string` |  |  |  |
| `spec.marketingFeatures` | `[]StripeProductMarketingFeature` |  |  |  |
| `spec.marketingFeatures[].name` | `string` | yes |  |  |
| `spec.packageDimensions` | `StripeProductPackageDimensions` |  |  |  |
| `spec.packageDimensions.height` | `double` |  |  |  |
| `spec.packageDimensions.length` | `double` |  |  |  |
| `spec.packageDimensions.weight` | `double` |  |  |  |
| `spec.packageDimensions.width` | `double` |  |  |  |
| `spec.shippable` | `bool` |  |  |  |
| `spec.statementDescriptor` | `string` |  |  |  |
| `spec.taxCode` | `string` |  |  |  |
| `spec.unitLabel` | `string` |  |  |  |
| `spec.url` | `string` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |
| `spec.features` | `[]string \| valueFrom` |  |  | StripeEntitlementFeature (`status.outputs.id`) |

## Field Details

### spec.name

`string` · required

name is the product's name as customers see it in Checkout, invoices, receipts and the
customer portal.

- rule: {"required":true}

### spec.description

`string`

description is the product's long-form explanation, shown to customers.

### spec.active

`bool` · optional (explicit presence)

active is whether the product can be bought. Setting it false archives the product without
destroying the resource: existing subscriptions keep billing, new purchases are refused.

- default: `true`

### spec.type

`enum`

type is what the product is. Unset, Stripe creates a service. Changing it REPLACES the
product: the old one is archived, a new one is created, and every price must move to the new
one (a StripePrice that references this resource is replaced with it).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `type_unspecified`
- `good` -- good is a physical or digital good, for Orders.
- `service` -- service is sold through subscriptions, Checkout and Payment Links (Stripe's default).

### spec.images

`[]string`

images are up to 8 public image URLs shown to customers.

- rule: {"repeated":{"maxItems":"8","unique":true,"items":{"string":{"uri":true}}}}

### spec.marketingFeatures

`[]StripeProductMarketingFeature`

marketing_features are up to 15 lines shown under the product in Stripe pricing tables.
They are display text only; the capabilities a purchase grants are features.

- rule: {"repeated":{"maxItems":"15"}}

### spec.marketingFeatures[].name

`string` · required

name is the line's text, up to 80 characters.

- rule: {"required":true,"string":{"maxLen":"80"}}

### spec.packageDimensions

`StripeProductPackageDimensions`

package_dimensions are the product's shipping dimensions, for physical goods.

### spec.packageDimensions.height

`double`

height in inches.

- rule: {"double":{"gt":0}}

### spec.packageDimensions.length

`double`

length in inches.

- rule: {"double":{"gt":0}}

### spec.packageDimensions.weight

`double`

weight in ounces.

- rule: {"double":{"gt":0}}

### spec.packageDimensions.width

`double`

width in inches.

- rule: {"double":{"gt":0}}

### spec.shippable

`bool` · optional (explicit presence)

shippable is whether the product is shipped (a physical good).

### spec.statementDescriptor

`string`

statement_descriptor is the text on the customer's card statement for subscription payments
of this product: up to 22 characters, at least one letter, none of < > \ " ', shown in
capitals with non-ASCII characters stripped. Only a service may carry one.

- rule: statement_descriptor is up to 22 characters, contains at least one letter, and has none of < > \ " '

### spec.taxCode

`string`

tax_code is the Stripe Tax category the product is taxed as (txcd_..., for example
txcd_10103001 for software as a service). Unset, the account's preset tax code applies.

- rule: tax_code is a Stripe tax code id such as txcd_10103001

### spec.unitLabel

`string`

unit_label names what a quantity counts ("seat", "GB"), shown on receipts, invoices,
Checkout and the customer portal. Only a service may carry one.

### spec.url

`string`

url is a public web page for the product.

- rule: url is an https:// or http:// address

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the product. A key removed here is removed in
Stripe on the next apply; removing the whole map leaves Stripe's keys in place.

### spec.features

`[]string | valueFrom`

features are the entitlement features buying this product grants (feat_...). Reference a
StripeEntitlementFeature. Each one is attached as its own link in Stripe: adding a feature
attaches it, removing one detaches it (the link is deleted), and a subscribed customer's
active entitlements follow. An archived feature cannot be attached.

- references: StripeEntitlementFeature (`status.outputs.id`)
- rule: {"repeated":{"items":{"cel":[{"id":"spec.features.format","message":"a feature is an entitlement feature id (feat_...), or a reference to a StripeEntitlementFeature","expression":"!has(this.value) || this.value.startsWith('feat_')"}]}}}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripeEntitlementFeature, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

## Validation Rules

- `spec.features.unique`: each feature is granted once: remove the repeated feature id
- `spec.service_only_fields`: statement_descriptor and unit_label apply only to a service: remove them, or leave type unset (a service)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeProduct, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the product's Stripe id (prod_...), the value a StripePrice's product references. |
| `status.outputs.active` | `bool` | active is false once the product is archived (by destroy, by spec.active, or in the Dashboard). |
| `status.outputs.default_price` | `string` | default_price is the price Stripe treats as the product's default (price_...), when one is set. This kind never sets it; it is observed so tools that read it see what Stripe reports. |
| `status.outputs.product_feature_ids` | `map<string, string>` | product_feature_ids maps each granted entitlement feature's id (feat_...) to the id of the link that attaches it to this product -- the id Stripe needs to import or detach that link. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.features` | StripeEntitlementFeature | `status.outputs.id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| StripeBillingPortalConfiguration | `spec.features.subscriptionUpdate.products[].product` | `status.outputs.id` |
| StripePrice | `spec.product` | `status.outputs.id` |

## See Also

- [Overview](../README.md)

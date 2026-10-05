# StripeEntitlementFeature

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

StripeEntitlementFeatureSpec declares one capability a customer can be entitled to in Stripe
Entitlements -- "api-access", "sso", "priority-support". A StripeProduct grants it by listing the
feature under its features; Stripe then reports which features each subscribed customer holds
(their active entitlements), so the application checks a feature's lookup key instead of
hard-coding product or price ids.

Stripe never deletes a feature. Destroy archives it (active = false) and Stripe keeps it,
archived, forever: an archived feature cannot be attached to new products and drops out of the
features list. This kind has no way to reactivate an archived feature -- the provider never
sends active -- so a feature archived in the Dashboard stays archived after the next apply. A
feature that cannot be read makes the plan fail (the provider does not treat a missing feature
as gone); remove it from state and apply again.

The key needs "Entitlements" write (iac/permissions.yaml).

https://docs.stripe.com/billing/entitlements
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/entitlements_feature

## Example

```yaml
# The canonical example: one capability a plan grants, checked by the
# application through its lookup key.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeEntitlementFeature
metadata:
  name: api-access
  org: e2e-org
  env: testing
spec:
  lookupKey: api-access
  name: API access
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.lookupKey` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.lookupKey

`string` · required

lookup_key is the feature's stable name, the one the application checks a customer's active
entitlements for ("api-access"). Up to 80 characters. Changing it REPLACES the feature: the old
one is archived, a new one is created, and every product that granted the old feature must
grant the new one (a product that references this resource does so on its next apply).

- rule: {"required":true,"string":{"maxLen":"80"}}

### spec.name

`string` · required

name is the feature's label for the account's team, shown in the Dashboard. It is not meant
for customers. Changing it updates the feature in place.

- rule: {"required":true}

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the feature. A key removed here is removed in
Stripe on the next apply; removing the whole map leaves Stripe's keys in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeEntitlementFeature, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the feature's Stripe id (feat_...), the value a StripeProduct's features reference. |
| `status.outputs.lookup_key` | `string` | lookup_key is the feature's stable name, as the application checks it. |
| `status.outputs.active` | `bool` | active is false once the feature is archived (by destroy, or in the Dashboard). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| StripeProduct | `spec.features` | `status.outputs.id` |

## See Also

- [Overview](../README.md)

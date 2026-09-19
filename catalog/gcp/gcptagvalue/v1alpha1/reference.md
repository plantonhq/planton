# GcpTagValue

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpTagValueSpec creates one Resource Manager tag value: the VALUE half
of a tag -- `prod` under the `environment` key, `pci` under
`data-classification`. A value belongs to exactly one tag key
(GcpTagKey) and is what gets bound to resources (GcpTagBinding) and what
organization policies, IAM conditions, and firewall rules test for.

Declare one GcpTagValue per allowed value of a key: `environment/prod`,
`environment/staging`, `environment/dev`. A key's vocabulary is the set
of values declared under it (unless the key has an allowed_values_regex,
in which case bindings may also carry undeclared matching values).

A value's key and short name are IMMUTABLE; only the description changes
in place. A value cannot be deleted while any binding uses it -- a chart
that declares the bindings by reference destroys them first.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagValue
metadata:
  name: prod
spec:
  # The key this value belongs to: a GcpTagKey reference (its name output,
  # tagKeys/{id}) or that literal. Immutable.
  tagKey:
    value: tagKeys/281475647562788

  # The value as written in conditions; defaults to metadata.name. 1-256
  # characters, none of / \ ' ". Immutable; reserved 30 days after deletion.
  shortName: prod

  description: Customer-facing workloads

  # DELETE (default; refused while bindings exist), PREVENT, ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.tagKey` | `string \| valueFrom` | yes |  | GcpTagKey (`status.outputs.name`) |
| `spec.shortName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.tagKey

`string | valueFrom` · required

The key this value belongs to: a reference to a GcpTagKey resource
(its `name` output, `tagKeys/{id}`) or that name as a literal.
Immutable.

- references: GcpTagKey (`status.outputs.name`)
- rule: tag_key must be the key's resource name in the form tagKeys/{numeric_id} (a GcpTagKey's name output) or a reference to a GcpTagKey
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpTagKey, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.shortName

`string`

The value as written in tag conditions and shown in the console:
`prod`, `pci`. Defaults to metadata.name when empty. Unique among the
key's values; 1-256 characters of any UTF-8 except the four Google
forbids -- `/`, `\`, `'`, `"`. Must match the key's
allowed_values_regex when the key has one. Immutable: a rename is a
delete and a create.

A deleted value's short name stays reserved for the 30-day soft-delete
window; a fresh value cannot reuse it under the same key until then.

- rule: short_name must be 1-256 characters and must not contain a slash, backslash, or quote -- e.g. prod or pci

### spec.description

`string`

What the value means, shown in the console. At most 256 characters.
Mutable.

- rule: {"string":{"maxLen":"256"}}

### spec.deletionPolicy

`string`

What destroying this resource does to the value in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the value is deleted; fails while any binding still
               uses it (destroy the bindings first -- a chart's
               dependency order does this when they reference the value)
  "PREVENT" -- destroy FAILS; the guard for a value every policy tests
  "ABANDON" -- the value is removed from management but stays live in
               GCP with its bindings

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpTagValue, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The value's resource name, `tagValues/{numeric_id}` -- what a GcpTagBinding's tag_value references, what a GcpFolder's or GcpProject's create-time `tags` map uses as its value, and what `resource.matchTagId(..., 'tagValues/...')` tests. |
| `status.outputs.namespaced_name` | `string` | The value's namespaced name, `{parent}/{key_short_name}/{short_name}` -- e.g. `123456789012/environment/prod` -- the human-readable handle `gcloud` and the console show. |
| `status.outputs.tag_value_id` | `string` | The numeric ID alone (the part after `tagValues/`). |
| `status.outputs.create_time` | `string` | When the value was created (RFC 3339 UTC). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.tagKey` | GcpTagKey | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpTagBinding | `spec.tagValue` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

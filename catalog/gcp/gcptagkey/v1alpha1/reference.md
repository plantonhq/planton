# GcpTagKey

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpTagKeySpec creates one Resource Manager tag key: the NAME half of a
tag such as `environment`, `cost-center`, or `data-classification`. A tag
key owns a set of tag values (GcpTagValue: `prod`, `staging`), and a tag
value is bound to a resource (GcpTagBinding) so that organization
policies (`resource.matchTag`), IAM conditions, and firewall policies can
key on it. Tags differ from labels in exactly this: they are governed
(created centrally, permissioned, immutable in name) and the platform's
policy engines evaluate them; labels are free-form metadata nobody
enforces anything on.

A tag key belongs to the organization (visible everywhere beneath it) or
to one project (visible to that project's resources only). A folder is
NOT a valid owner -- Google's rule. Keys owned by the organization are
the normal landing-zone choice; a project-scoped key is for tags one
application manages for itself.

Almost everything about a key is IMMUTABLE once created: its owner, its
short name, and its purpose. Only the description and the allowed-values
regex change in place. Rename by creating a new key. A key cannot be
deleted while any of its values exist, and a value cannot be deleted
while any binding uses it -- a chart that declares all three by reference
destroys them in the right order.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagKey
metadata:
  name: environment
spec:
  # Who owns the key: exactly one of organizationId (values bindable
  # estate-wide) or projectId (a literal or GcpProject reference; values
  # bindable only inside that project). A folder cannot own a key.
  # Immutable.
  parent:
    organizationId: "123456789012"

  # The name written in tag conditions; defaults to metadata.name. 1-256
  # characters, none of / \ ' ". Immutable; reserved 30 days after deletion.
  shortName: environment

  description: The environment a resource belongs to

  # Optional system purpose: GCE_FIREWALL (secure tags for firewall policy
  # rules; needs purposeData.network) or DATA_GOVERNANCE. Immutable.
  # purpose: GCE_FIREWALL
  # purposeData:
  #   network: my-gcp-project-123/shared-vpc

  # RE2 every value must match; also makes the key dynamic.
  # allowedValuesRegex: ^(prod|staging|dev)$

  # DELETE (default; refused while values exist), PREVENT, ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.parent` | `GcpTagKeyParent` | yes |  |  |
| `spec.parent.organizationId` | `string` |  |  |  |
| `spec.parent.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.shortName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.purpose` | `string` |  |  |  |
| `spec.purposeData` | `map<string, string>` |  |  |  |
| `spec.allowedValuesRegex` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.parent

`GcpTagKeyParent` · required

Who owns the key: the organization or one project. Exactly one arm.
Immutable.

- rule: {"required":true}
- rule: set exactly one of organization_id or project_id -- a tag key is owned by the organization or by one project (never a folder)

### spec.parent.organizationId

`string`

An organization-owned key: the numeric organization ID, without the
`organizations/` prefix. Its values can be bound to any resource in
the organization -- the landing-zone default.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.parent.projectId

`string | valueFrom`

A project-owned key: a literal project ID (Google also accepts the
number; the provider treats the two as equal) or a reference to a
GcpProject resource. Its values can be bound only to resources in that
project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.shortName

`string`

The key's name as written in tag conditions and shown in the console:
`environment`, `cost-center`. Defaults to metadata.name when empty.
Unique among the owner's keys; 1-256 characters of any UTF-8 except
the four Google forbids -- `/`, `\`, `'`, `"` -- because
`{parent}/{key}/{value}` and `resource.matchTag('{org}/{key}',
'{value}')` address tags by short name. Immutable: a rename is a
delete and a create.

A deleted key's short name stays reserved for the 30-day soft-delete
window; a fresh key cannot reuse it under the same owner until then.

- rule: short_name must be 1-256 characters and must not contain a slash, backslash, or quote -- e.g. environment or cost-center

### spec.description

`string`

What the key is for, shown in the console. At most 256 characters.
Mutable.

- rule: {"string":{"maxLen":"256"}}

### spec.purpose

`string`

Reserve the key for one of Google's system uses, which changes how the
key behaves:
  ""                -- an ordinary tag key (the usual case)
  "GCE_FIREWALL"    -- the key's values may be used as targets and
                       sources in network firewall policy rules
                       (secure tags); requires purpose_data.network
  "DATA_GOVERNANCE" -- the key's values classify data for Sensitive
                       Data Protection and BigQuery policy tags
Immutable: a purpose cannot be added, changed, or removed once the key
exists.

- rule: purpose must be empty, GCE_FIREWALL, or DATA_GOVERNANCE

### spec.purposeData

`map<string, string>`

Data the purpose needs, as Google defines it per purpose. For
GCE_FIREWALL exactly one entry, `network`, naming the VPC the secure
tags are scoped to as `{project_id}/{network_name}` (or the network's
full resource name). Only meaningful with a purpose; immutable with it.

### spec.allowedValuesRegex

`string`

An RE2 regular expression every value's short name must match. When
set, the key also becomes a DYNAMIC key: bindings may carry values
that were never declared, as long as they match the regex (the
pattern for high-cardinality tags such as a ticket number or a team
code). Leave empty for a fixed vocabulary of declared GcpTagValues.
Mutable. Google compiles it on apply; an invalid expression is
rejected there.

### spec.deletionPolicy

`string`

What destroying this resource does to the key in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the key is deleted; fails while any of its values still
               exist (destroy the values first -- a chart's dependency
               order does this when the values reference the key)
  "PREVENT" -- destroy FAILS; the guard for a key every policy keys on
  "ABANDON" -- the key is removed from management but stays live in
               GCP with its values and bindings

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `purpose_data_needs_purpose`: purpose_data is only meaningful with a purpose (GCE_FIREWALL needs purpose_data.network) -- set purpose or drop purpose_data

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpTagKey, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The key's resource name, `tagKeys/{numeric_id}` -- what a GcpTagValue's tag_key references, what a GcpFolder's or GcpProject's create-time `tags` map uses as its key, and what `resource.matchTagId('tagKeys/...', ...)` tests. |
| `status.outputs.namespaced_name` | `string` | The key's namespaced name, `{org_id}/{short_name}` or `{project_id}/{short_name}` -- the form `resource.matchTag('{org_id}/{short_name}', ...)` tests, and the form `gcloud` shows. |
| `status.outputs.tag_key_id` | `string` | The numeric ID alone (the part after `tagKeys/`). |
| `status.outputs.create_time` | `string` | When the key was created (RFC 3339 UTC). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.parent.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpTagValue | `spec.tagKey` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

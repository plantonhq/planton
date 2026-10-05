# CloudflareList

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareListSpec provisions an account-scoped Cloudflare List: a reusable,
named collection referenced from rule expressions (e.g. a WAF/custom rule's
`ip.src in $my_list`, or a Bulk Redirect ruleset's `from_list`). The list is
the container; its entries are managed as CloudflareListItem resources, so a
single list can hold anything from a handful of curated IPs to a large,
independently-managed set. (The list's `kind` and `name` are fixed at
creation; changing either replaces the list.)

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareList
metadata:
  name: test-list
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  kind: ip
  name: office_allowlist
  description: Corporate office egress IPs allowed through WAF custom rules
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.kind` | `enum` |  |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this list.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.kind

`enum`

The list type, fixing which kind of item the list accepts. Immutable —
changing it replaces the list.

- rule: kind must be one of ip, redirect, hostname, asn
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `list_kind_unspecified`
- `ip`
- `redirect`
- `hostname`
- `asn`

### spec.name

`string` · required

An informative name for the list, unique within the account. This is the
identifier used in filter and rule expressions (e.g. `$name`), so prefer a
short, lowercase, expression-friendly name. Immutable — changing it
replaces the list.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"50","pattern":"^[a-zA-Z][a-zA-Z0-9_]*$"}}

### spec.description

`string`

An optional human-readable summary of the list's purpose.

- rule: {"string":{"maxLen":"500"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareList, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.list_id` | `string` | The Cloudflare-assigned identifier of the list. A CloudflareListItem references this value to add entries to the list. |
| `status.outputs.name` | `string` | The list name (echoed; this is the identifier used in rule expressions). |
| `status.outputs.kind` | `string` | The list kind (ip, redirect, hostname, or asn). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareListItem | `spec.listId` | `status.outputs.list_id` |
| CloudflareRuleset | `spec.rules[].actionParameters.fromList.name` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

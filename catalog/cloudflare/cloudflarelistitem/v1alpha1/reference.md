# CloudflareListItem

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareListItemSpec declares a single entry inside a Cloudflare List. The
entry's shape must match the parent list's kind (an ip/CIDR, an ASN, a
hostname, or a redirect). Items have independent lifecycles, so a list can be
grown or trimmed one entry at a time without rewriting the whole set. (Item
values are immutable in the provider: changing an entry replaces it.)

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareListItem
metadata:
  name: test-list-item
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  listId:
    value: "2c0fc9fa937b11eaa1b71c4d701ab86e"
  ip: "203.0.113.0/24"
  comment: Datacenter egress range
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.listId` | `string \| valueFrom` | yes |  | CloudflareList (`status.outputs.list_id`) |
| `spec.ip` | `string` | yes |  |  |
| `spec.asn` | `int64` |  |  |  |
| `spec.hostname` | `CloudflareListItemHostname` |  |  |  |
| `spec.hostname.urlHostname` | `string` | yes |  |  |
| `spec.hostname.excludeExactHostname` | `bool` |  |  |  |
| `spec.redirect` | `CloudflareListItemRedirect` |  |  |  |
| `spec.redirect.sourceUrl` | `string` | yes |  |  |
| `spec.redirect.targetUrl` | `string` | yes |  |  |
| `spec.redirect.statusCode` | `int64` |  |  |  |
| `spec.redirect.includeSubdomains` | `bool` |  |  |  |
| `spec.redirect.preservePathSuffix` | `bool` |  |  |  |
| `spec.redirect.preserveQueryString` | `bool` |  |  |  |
| `spec.redirect.subpathMatching` | `bool` |  |  |  |
| `spec.comment` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns the parent list.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.listId

`string | valueFrom` · required

The list this entry is written to. A literal list ID or a reference to a
CloudflareList resource (defaulting to that list's status.outputs.list_id).

- references: CloudflareList (`status.outputs.list_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareList, name: <that resource's name>, fieldPath: status.outputs.list_id}} -- a bare string does not parse

### spec.ip

`string` · required

An IPv4 address, IPv4 CIDR, IPv6 address, or IPv6 CIDR (ip-kind lists).

- rule: {"string":{"minLen":"1"}}

### spec.asn

`int64`

A non-negative 32-bit Autonomous System Number (asn-kind lists).

- rule: asn must be a non-negative 32-bit integer

### spec.hostname

`CloudflareListItemHostname`

A hostname (hostname-kind lists).

- rule: exclude_exact_hostname is required for wildcard hostnames (those starting with '*')
- rule: exclude_exact_hostname is only allowed for wildcard hostnames (those starting with '*')

### spec.hostname.urlHostname

`string` · required

The hostname. Valid characters are a-z, 0-9, the hyphen, and a leading
wildcard (e.g. "*.example.com").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.hostname.excludeExactHostname

`bool` · optional (explicit presence)

Only meaningful for wildcard hostnames (e.g. "*.example.com"): when true
(the wildcard default), only subdomains match; when false, the apex domain
matches too. Required for wildcard hostnames and must be omitted for
non-wildcard hostnames.

### spec.redirect

`CloudflareListItemRedirect`

A redirect rule (redirect-kind lists).

### spec.redirect.sourceUrl

`string` · required

The source URL to match (e.g. "example.com/old"). A trailing path is
required; append "/" if redirecting the apex.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.redirect.targetUrl

`string` · required

The destination URL to redirect to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.redirect.statusCode

`int64`

HTTP redirect status code. Leave 0 for the provider default (301); otherwise
one of 301, 302, 307, 308.

- rule: status_code must be 0 (default 301) or one of 301, 302, 307, 308

### spec.redirect.includeSubdomains

`bool`

Also redirect subdomains of source_url.

### spec.redirect.preservePathSuffix

`bool`

Append the matched path suffix from source_url onto target_url.

### spec.redirect.preserveQueryString

`bool`

Preserve the original query string on the redirect.

### spec.redirect.subpathMatching

`bool`

Match source_url as a path prefix (subpath matching).

### spec.comment

`string`

An optional informative summary of this entry.

- rule: {"string":{"maxLen":"500"}}

## Validation Rules

- `item.exactly_one_value`: exactly one of ip, asn, hostname, or redirect must be set (matching the list's kind)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareListItem, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.item_id` | `string` | The Cloudflare-assigned identifier of the list item. |
| `status.outputs.list_id` | `string` | The list ID the entry was written to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.listId` | CloudflareList | `status.outputs.list_id` |

## See Also

- [Overview](../README.md)

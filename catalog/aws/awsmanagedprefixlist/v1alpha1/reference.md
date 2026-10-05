# AwsManagedPrefixList

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsManagedPrefixListSpec defines one customer-managed prefix list: a
named, versioned set of CIDR blocks that security-group rules, NACL
rules, and route tables reference as a single object (the pl-...
id). Update the list once and every referencing rule follows.

The list's name in AWS is metadata.name. Entries are managed
in-line as the single declarative owner - the standalone
aws_ec2_managed_prefix_list_entry resource carries the identical
payload and fights the in-line form, so this kind never uses it.

## Example

```yaml
# Canonical AwsManagedPrefixList example (hack/dev manifest and refgen
# Example source): an IPv4 list of office CIDRs with headroom for two
# more entries.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsManagedPrefixList
metadata:
  name: office-cidrs
  id: office-cidrs
  org: test-org
  env: dev
spec:
  region: us-west-2
  addressFamily: IPv4
  maxEntries: 5
  entries:
    - cidr: 203.0.113.0/24
      description: hq office
    - cidr: 198.51.100.0/24
      description: branch office
    - cidr: 192.0.2.0/24
      description: vpn egress
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.addressFamily` | `string` |  |  |  |
| `spec.maxEntries` | `int64` |  |  |  |
| `spec.entries` | `[]AwsManagedPrefixListEntry` |  |  |  |
| `spec.entries[].cidr` | `string` | yes |  |  |
| `spec.entries[].description` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the prefix list lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.addressFamily

`string`

The address family of every CIDR in this list. AWS fixes it for
life - changing it replaces the list (and every reference to the
old pl- id breaks with it).

- rule: {"string":{"in":["IPv4","IPv6"]}}

### spec.maxEntries

`int64`

The list's capacity - how many entries it may EVER hold without a
resize. Size it deliberately: a security-group rule referencing
this list consumes max_entries rule-quota slots regardless of how
many entries actually exist. Grows and shrinks in place (AWS
applies capacity increases before entry changes and decreases
after, so a resize never transiently strands entries).

- rule: {"int64":{"gte":"1"}}

### spec.entries

`[]AwsManagedPrefixListEntry`

The CIDR entries, each optionally described. Managed as the
complete set - an entry removed here is removed at AWS. AWS
versions the list on every entry change (the version
output).

- rule: entries must have unique cidr values

### spec.entries[].cidr

`string` · required

The CIDR block. Example: "10.20.0.0/16" (IPv4) or "2001:db8::/32"
(IPv6). Must match the list's address_family.

- rule: {"string":{"minLen":"1","pattern":"^.+/[0-9]+$"}}

### spec.entries[].description

`string`

What this CIDR is (an office, a partner VPC, a scanner fleet...).
Shown wherever the list is inspected; keep it current.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

## Validation Rules

- `spec.entries_within_max`: entries cannot exceed max_entries - raise max_entries or drop entries
- `spec.entries_match_address_family`: every entry's cidr must match address_family - IPv6 entries contain ':', IPv4 entries do not

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsManagedPrefixList, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.prefix_list_id` | `string` | The prefix list's id (pl-...) - what security-group rules, NACL rules, and route tables reference, and the provider's import ID. |
| `status.outputs.prefix_list_arn` | `string` | The prefix list's ARN. |
| `status.outputs.owner_id` | `string` | The AWS account that owns the list. |
| `status.outputs.version` | `string` | The list's current version - AWS increments it on every entry change (its optimistic-concurrency token). |

## See Also

- [Overview](../README.md)

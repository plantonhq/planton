# CloudflareZeroTrustList

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustListSpec defines a reusable Zero Trust list: a named set of
values (domains, IPs, URLs, emails, serial numbers, and kin) that Gateway
policies and device-posture rules reference by ID instead of repeating the
values inline. Centralizing the values in a list lets many policies share one
source of truth that evolves in one place.

Lists are account-scoped. The list TYPE is immutable at Cloudflare (changing it
replaces the list, which breaks any policy referencing the old list ID) -- pick
the type deliberately.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustList
metadata:
  name: test-zt-list
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: blocked-domains
  type: DOMAIN
  description: Domains Gateway policies block for all users
  items:
    - value: gambling.example.com
      description: policy-blocked domain
    - value: casino.example.net
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.items` | `[]CloudflareZeroTrustListItem` |  |  |  |
| `spec.items[].value` | `string` | yes |  |  |
| `spec.items[].description` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this list.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The display name of the list.

- rule: {"string":{"minLen":"1"}}

### spec.type

`string` · required

The list type -- what kind of values the items carry. IMMUTABLE at
Cloudflare: changing the type replaces the list (new list ID), breaking any
Gateway policy or posture rule that references the old ID. Use the
canonical UPPERCASE form (the API stores it uppercase; lowercase input
would round-trip as permanent drift).

- rule: {"required":true,"string":{"in":["SERIAL","URL","DOMAIN","EMAIL","IP","CATEGORY","LOCATION","DEVICE","AAGUID"]}}

### spec.description

`string`

Optional description of the list's purpose.

### spec.items

`[]CloudflareZeroTrustListItem`

The list's entries. Cloudflare treats items as a SET: order is not
significant and is not preserved -- two lists with the same values in a
different order are the same list. URL-type values are normalized by the
API (a known upstream drift source); prefer already-normalized URLs.

### spec.items[].value

`string` · required

The item value, interpreted per the list's type (a domain for DOMAIN lists,
an IP for IP lists, and so on). Required here even though the API tolerates
value-less items -- an entry without a value matches nothing and can only
be a mistake.

- rule: {"required":true}

### spec.items[].description

`string`

Optional description of this entry.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustList, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.list_id` | `string` | The UUID of the created list -- what Gateway policies and device-posture rules reference. |

## See Also

- [Overview](../README.md)

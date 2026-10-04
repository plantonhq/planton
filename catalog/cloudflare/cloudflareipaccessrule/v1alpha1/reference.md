# CloudflareIpAccessRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareIpAccessRuleSpec defines one IP Access rule: an allow/block/challenge
decision applied to traffic matching an IP address, IP range, ASN, or country
before it reaches the zone's other security products. Rules live either on the
whole account (every zone) or on a single zone -- exactly one scope must be set.

Cloudflare's API applies account rules across all zones and lets a zone rule
override an account rule for that zone. The provider accepts both scopes at
once and silently prefers the account; this spec requires exactly one so the
manifest states its intent unambiguously.

Only `mode` and `notes` can change in place. Cloudflare's API does not honor
editing the matched target/value of an existing rule even though it accepts the
request shape -- to change WHAT a rule matches, delete and recreate it (change
the configuration in a manifest and the module plans an in-place update that
will not stick; plan a replace instead by changing the resource identity).

## Example

```yaml
# A complete, protovalidate-valid CloudflareIpAccessRule example: an
# account-wide managed challenge on a suspicious network range. Exactly one
# scope (account_id or zone_id) is set -- the spec enforces it.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareIpAccessRule
metadata:
  name: challenge-suspicious-range
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  mode: managed_challenge
  configuration:
    target: ip_range
    value: "203.0.113.0/24"
  notes: "Repeated credential-stuffing attempts from this range"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` |  |  |  |
| `spec.zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.mode` | `string` | yes |  |  |
| `spec.configuration` | `CloudflareIpAccessRuleConfiguration` | yes |  |  |
| `spec.configuration.target` | `string` | yes |  |  |
| `spec.configuration.value` | `string` | yes |  |  |
| `spec.notes` | `string` |  |  |  |

## Field Details

### spec.accountId

`string`

The Cloudflare account to apply the rule account-wide (all zones). Set this
OR zone_id, never both. 32-character hex account ID.

- rule: account_id must be a 32-character hex string

### spec.zoneId

`string | valueFrom`

The zone to apply the rule to (single-zone rule). Set this OR account_id,
never both.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.mode

`string` · required

The action taken on matching traffic.
  - block: refuse the request outright.
  - challenge: interactive challenge page.
  - whitelist: allow, bypassing other security features (Cloudflare's legacy
    name for the allow action).
  - js_challenge: non-interactive JavaScript challenge.
  - managed_challenge: Cloudflare picks the least intrusive challenge that
    confirms the visitor is human (the recommended challenge mode).

- rule: {"required":true,"string":{"in":["block","challenge","whitelist","js_challenge","managed_challenge"]}}

### spec.configuration

`CloudflareIpAccessRuleConfiguration` · required

What the rule matches. Changing target/value on an existing rule does not
stick at the API -- see the message comment above.

- rule: {"required":true}
- rule: for target ip, value must be a plain IPv4 address (use target ip6 for IPv6, ip_range for CIDR blocks)
- rule: for target ip6, value must be an IPv6 address in fully-expanded long form (eight colon-separated groups, e.g. 2001:0db8:0000:0000:0000:0000:0000:0001) -- Cloudflare rejects compressed :: notation here
- rule: for target ip_range, Cloudflare only accepts IPv4 ranges of /16 or /24, and IPv6 ranges of /32, /48, or /64
- rule: for target country, value must be a two-character country code (ISO 3166-1 alpha-2, e.g. US -- Cloudflare also accepts T1 for Tor)
- rule: for target asn, value must be an AS number in the form AS13335

### spec.configuration.target

`string` · required

The kind of selector:
  - ip: a single IPv4 address.
  - ip6: a single IPv6 address (fully-expanded long form).
  - ip_range: a CIDR block (IPv4 /16 or /24; IPv6 /32, /48, or /64 -- the
    only prefix lengths Cloudflare accepts).
  - asn: an autonomous system, e.g. AS13335.
  - country: a two-letter country code, e.g. US (T1 matches Tor exit nodes).

- rule: {"required":true,"string":{"in":["ip","ip6","ip_range","asn","country"]}}

### spec.configuration.value

`string` · required

The value matched, interpreted per target (an IP, a CIDR, an AS number, or a
country code).

- rule: {"required":true}

### spec.notes

`string`

Free-form note about why the rule exists (shown in the dashboard's rule list).

## Validation Rules

- `spec.scope_exactly_one`: set exactly one of account_id (account-wide rule) or zone_id (single-zone rule)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareIpAccessRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.rule_id` | `string` | The ID of the created rule. |
| `status.outputs.zone_id` | `string` | The zone the rule was created on (empty for account-wide rules). Published so zone-scoped consumers and the import recipe can derive scope without re-reading the manifest. |
| `status.outputs.account_id` | `string` | The account the rule was created on (empty for zone-scoped rules). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)

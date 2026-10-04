# CloudflareAuthenticatedOriginPulls

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareAuthenticatedOriginPullsSpec manages a zone's Authenticated
Origin Pulls enablement: the zone-wide toggle that makes Cloudflare present
a client certificate when pulling from your origin, and the per-hostname
associations that pin specific hostnames to specific uploaded certificates.
The origin server completes the story by requiring and validating that
client certificate -- Cloudflare's side alone protects nothing.

Destroy semantics differ per surface, and neither is a delete: the
zone-wide toggle has NO delete at Cloudflare (destroying this resource
abandons whatever value is live), and a hostname association is removed by
a revert write (enabled: null) that voids it. Plan teardown accordingly --
disable explicitly before destroying when the zone outlives this resource.

## Example

```yaml
# A complete, protovalidate-valid CloudflareAuthenticatedOriginPulls
# example: the zone-wide toggle plus one hostname pinned to an uploaded
# client certificate and one hostname kept present but inactive. Every
# association row carries its certificate id -- Cloudflare rejects an
# association write without one (400 code 1404), even when zone-level
# certificate material exists. Destroying this resource abandons the zone
# toggle (no delete at Cloudflare) and reverts the associations.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAuthenticatedOriginPulls
metadata:
  name: www-origin-pulls
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  zone_enabled: true
  hostname_associations:
    - hostname: app.example.com
      certificate_id:
        value: "2458ce5a0c354c7f82c78e9487d3ff60"
    - hostname: legacy.example.com
      certificate_id:
        value: "7f31b9c4d2a54e6f91c30ab8d5e2c477"
      enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.zoneEnabled` | `bool` |  |  |  |
| `spec.hostnameAssociations` | `[]CloudflareAuthenticatedOriginPullsHostnameAssociation` |  |  |  |
| `spec.hostnameAssociations[].hostname` | `string` | yes |  |  |
| `spec.hostnameAssociations[].certificateId` | `string \| valueFrom` | yes |  | CloudflareAuthenticatedOriginPullsCertificate (`status.outputs.certificate_id`) |
| `spec.hostnameAssociations[].enabled` | `bool` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone whose Authenticated Origin Pulls surface is managed.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.zoneEnabled

`bool` · optional (explicit presence)

The zone-wide Authenticated Origin Pulls toggle. Leave unset to not
manage the toggle at all (associations can be managed independently);
set true/false to assert it. The toggle uses the zone's Universal SSL
client certificate unless hostname associations pin uploaded ones.

### spec.hostnameAssociations

`[]CloudflareAuthenticatedOriginPullsHostnameAssociation`

Per-hostname associations pinning hostnames to uploaded client
certificates (see CloudflareAuthenticatedOriginPullsCertificate). Each
row is its own association at Cloudflare; hostnames must be unique.

### spec.hostnameAssociations[].hostname

`string` · required

The hostname the association covers. Must belong to the zone.

- rule: {"string":{"minLen":"1"}}

### spec.hostnameAssociations[].certificateId

`string | valueFrom` · required

The uploaded client certificate the hostname uses (a hostname-scoped
upload from CloudflareAuthenticatedOriginPullsCertificate). REQUIRED at
the API: Cloudflare rejects an association write without a certificate
id (400 code 1404 "Certificate ID required in the request", measured
2026-08-28 -- unconditionally, even when zone-level certificate
material exists), so a row without one could never deploy.
When using value_from, defaults to CloudflareAuthenticatedOriginPullsCertificate
kind and status.outputs.certificate_id field path.

- references: CloudflareAuthenticatedOriginPullsCertificate (`status.outputs.certificate_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAuthenticatedOriginPullsCertificate, name: <that resource's name>, fieldPath: status.outputs.certificate_id}} -- a bare string does not parse

### spec.hostnameAssociations[].enabled

`bool` · optional (explicit presence)

Whether the association is active. Unset means active (true) -- both
engines send true when omitted, because Cloudflare treats a null here as
"void the association" and a declared row is meant to exist. Set false
to keep the row present but inactive.

## Validation Rules

- `spec.at_least_one_surface`: set zone_enabled and/or at least one hostname association -- an empty spec manages nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareAuthenticatedOriginPulls, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The zone whose Authenticated Origin Pulls surface is managed. The surface is zone-singleton shaped -- the zone id IS its identity. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.hostnameAssociations[].certificateId` | CloudflareAuthenticatedOriginPullsCertificate | `status.outputs.certificate_id` |

## See Also

- [Overview](../README.md)

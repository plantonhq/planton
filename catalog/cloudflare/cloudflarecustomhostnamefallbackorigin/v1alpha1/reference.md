# CloudflareCustomHostnameFallbackOrigin

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareCustomHostnameFallbackOriginSpec sets the default origin for a
Cloudflare for SaaS zone: the backend that all of the zone's custom hostnames
route to unless a hostname overrides it. It is a zone-level singleton (one per
zone) and a prerequisite for serving traffic to any custom hostname in the zone.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomHostnameFallbackOrigin
metadata:
  name: test-fallback-origin
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  origin:
    value: origin.example.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.origin` | `string \| valueFrom` | yes |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The Cloudflare Zone ID of the SaaS zone. A literal value or a reference to a
CloudflareDnsZone resource (its status.outputs.zone_id).

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.origin

`string | valueFrom` · required

The fallback origin hostname — a record within the SaaS zone that points at the
backend (e.g. "origin.helpdesk.io"). A literal hostname or a reference to
another resource's output (a backend endpoint).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareCustomHostnameFallbackOrigin, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.created_at` | `string` | RFC3339 timestamp of when the fallback origin was created. |
| `status.outputs.updated_at` | `string` | RFC3339 timestamp of when the fallback origin was last updated. |
| `status.outputs.zone_id` | `string` | The Cloudflare Zone ID this singleton belongs to. The fallback origin has no resource id of its own -- its API identity IS the zone (GET zones/{zone_id}/custom_hostnames/fallback_origin) -- so this is the handle verification, import, and chart blocks consume. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)

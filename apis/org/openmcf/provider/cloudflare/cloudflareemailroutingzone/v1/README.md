# CloudflareEmailRoutingZone

Enable Cloudflare Email Routing on a zone — the anchor of the Email Routing
family. Enabling provisions the zone's required MX/SPF/DKIM DNS records
automatically. The single per-zone catch-all rule is folded in; individual
routing rules and destination addresses are separate kinds.

## When to use

- Turning on Email Routing for a domain so it can forward mail to verified
  destinations or hand it to an Email Worker.

Pair with `CloudflareEmailRoutingRule` (per-address rules) and
`CloudflareEmailRoutingAddress` (verified destinations).

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareEmailRoutingZone
metadata:
  name: example-com-email
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: example-com
      fieldPath: status.outputs.zone_id
  catchAll:
    enabled: true
    type: forward
    forwardTo:
      - value: ops@example.com
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `zoneId` | yes | Zone ID, or a reference to a `CloudflareDnsZone` |
| `catchAll` | no | Folded catch-all rule (see below). Omit to leave Cloudflare's default |
| `lockDnsRecords` | no | Lock the Email Routing DNS records (default false) |

**catchAll**: `{ enabled, type (drop|forward|worker), forwardTo[] (→ CloudflareEmailRoutingAddress), worker (→ CloudflareWorker) }`. `forward` requires `forwardTo`; `worker` requires `worker`.

## Outputs

| Output | Description |
|---|---|
| `zone_id` | The zone Email Routing was enabled on |
| `enabled` | Whether Email Routing is enabled |
| `status` | Configuration status (ready, unconfigured, ...) |
| `name` | The zone's domain name |

> Enabling Email Routing rewrites the zone's MX/SPF/DKIM records. Only enable it
> on a zone whose mail you intend Cloudflare to handle.

## Related components

- `CloudflareDnsZone` — the zone this enables Email Routing on.
- `CloudflareEmailRoutingRule` — per-address routing rules.
- `CloudflareEmailRoutingAddress` — verified destination addresses.

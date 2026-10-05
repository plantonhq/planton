# Cloudflare Custom Hostname Fallback Origin

Sets the default origin for a Cloudflare-for-SaaS zone: the backend that all of the zone's custom hostnames route to unless an individual hostname overrides it. It is a zone-level singleton (one per zone) and a prerequisite for serving traffic to any custom hostname in the zone. The fallback origin has no resource ID of its own -- its API identity is the zone -- so two manifests targeting the same zone are the same object, and the second apply overwrites the first.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Fallback Origin** -- the zone's default backend for custom-hostname traffic. The write is asynchronous: status moves through `pending_deployment` before `active`, so a just-applied origin is not yet serving.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has SSL and Certificates edit access. Note that some tokens can read this singleton but are forbidden to write it -- a 403 on apply after a successful read is a credential scope problem, not a spec problem.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A SaaS zone** -- the fallback origin is set on an existing zone (`zoneId`) configured for Cloudflare for SaaS.
- **A resolvable origin hostname** -- Cloudflare expects the `origin` hostname to already exist as a DNS record on this zone (typically a proxied A/CNAME to the real backend). This kind does not create that record; without it the origin sits pending or errors.

## Deploy

### Console

Open the deployment store, find **Cloudflare Custom Hostname Fallback Origin**, and click **Deploy**. The creation wizard captures the zone and the default backend origin. Start from the **SaaS Fallback Origin** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomHostnameFallbackOrigin
metadata:
  name: saas-fallback
  org: acme-corp
  env: prod
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: saas-zone
      fieldPath: status.outputs.zone_id
  origin:
    value: origin.helpdesk.io
```

```shell
planton apply -f cloudflare-custom-hostname-fallback-origin.yaml
```

This points every custom hostname in the SaaS zone at `origin.helpdesk.io` by default. An Infra Job tracks the provisioning in real time.

### InfraChart

In a SaaS zone chart, wire the singleton to the zone with ValueFromRef -- and include exactly one of these per zone:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: saas-zone
      fieldPath: status.outputs.zone_id
  origin:
    value: origin.helpdesk.io
```

The InfraPipeline resolves the dependency graph, creates the zone (and the DNS record backing the origin hostname) first, then sets the fallback origin.

## Key Configuration

These are the most important decisions when configuring a fallback origin. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One per zone -- the identity IS the zone** -- there is no separate resource ID. A second manifest against the same zone silently overwrites the first. Put exactly one fallback origin in a SaaS zone's chart, and treat `zoneId` as immutable: changing it replaces the resource.

**Origin (`origin`)** -- the default backend hostname, editable in place. It must be a record within the SaaS zone that points at the real backend; this component will not create that record for you. A literal hostname or a reference to a backend endpoint output.

**Create equals update, and it is asynchronous** -- the write path is a PUT, so re-applying an identical spec is idempotent. But status passes through `pending_deployment` before `active` -- read the deployment status (and errors) from the Cloudflare API or dashboard before assuming custom-hostname traffic flows.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

`origin` is also a value-or-reference field and may point at another resource's output that resolves to a backend hostname.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `created_at` | RFC3339 creation timestamp | Auditing |
| `updated_at` | RFC3339 last-updated timestamp | Auditing |
| `zone_id` | The zone this singleton belongs to | The only API handle the fallback origin has -- verification and import tooling consume it |

Deployment status and the deployment-errors list are deliberately not outputs — both transition asynchronously server-side; read them from the Cloudflare API or dashboard.

## Common Patterns

**Single shared backend** -- route all customer hostnames in the SaaS zone to one origin, and use per-hostname `customOriginServer` overrides only for the customers that genuinely need a dedicated backend. This keeps the zone's routing story in one place. Start from the **SaaS Fallback Origin** preset.

**Zone-chart singleton** -- deploy the fallback origin in the same InfraChart as the SaaS zone and the DNS record backing the origin hostname, so the prerequisite chain (zone, record, fallback origin, then custom hostnames) resolves in one pipeline.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the SaaS zone this origin serves; `zoneId` references its output
- [**Cloudflare DNS Record**](/infra-catalog/cloudflare-dns-record) -- creates the in-zone record the origin hostname must resolve to
- [**Cloudflare Custom Hostname**](/infra-catalog/cloudflare-custom-hostname) -- the per-customer hostnames that route to this fallback origin

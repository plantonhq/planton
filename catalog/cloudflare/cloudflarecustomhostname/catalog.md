# Cloudflare Custom Hostname

Attaches a customer's own domain to a Cloudflare-for-SaaS zone. It extends Cloudflare's edge -- TLS termination, caching, WAF -- onto a hostname your customer owns (e.g. `support.acme.com`), with a per-customer certificate that Cloudflare provisions and auto-renews. Onboarding alone does not make the hostname live: the customer must CNAME their hostname to the SaaS zone and prove control using the ownership-verification records exported in the outputs, and the hostname sits in `pending` or `pending_validation` until they do.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Custom Hostname** -- a customer hostname onboarded to the SaaS zone
- **Per-hostname Certificate** -- a managed DV certificate (by default), auto-renewed by Cloudflare; Cloudflare derives the ownership-verification TXT and HTTP records the customer needs, which the module exports as outputs

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has SSL and Certificates edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A SaaS zone with a fallback origin** -- the zone needs a `CloudflareCustomHostnameFallbackOrigin` so traffic has a default backend (unless this hostname sets its own `customOriginServer` override).
- **Enterprise plan (only for BYO certificates)** -- uploading your own certificate via `ssl.customCertificate` or `ssl.customCertBundle` is Enterprise-gated; the API returns 403 on other plans even though the manifest validates.

## Deploy

### Console

Open the deployment store, find **Cloudflare Custom Hostname**, and click **Deploy**. The creation wizard captures the SaaS zone and hostname, an optional per-hostname origin override (with a live connection diagram), and an optional TLS/SSL block. Start from the **SaaS Vanity Domain (recommended)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomHostname
metadata:
  name: acme-support-hostname
  org: acme-corp
  env: prod
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: saas-zone
      fieldPath: status.outputs.zone_id
  hostname: support.acme.com
```

```shell
planton apply -f cloudflare-custom-hostname.yaml
```

This onboards `support.acme.com` to the SaaS zone with a managed DV certificate. The customer then CNAMEs their hostname to the zone and adds the ownership-verification record from the outputs. An Infra Job tracks the provisioning in real time.

### InfraChart

When the SaaS zone is deployed in the same InfraPipeline, wire the hostname to it with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: saas-zone
      fieldPath: status.outputs.zone_id
  hostname: support.acme.com
  customMetadata:
    tenant: acme
```

The InfraPipeline resolves the dependency graph, creates the zone (and its fallback origin) first, then onboards the hostname with the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring a custom hostname. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Zone and hostname (`zoneId`, `hostname`) are immutable** -- changing either replaces the resource, which restarts the customer-facing verification cycle. Everything under `ssl` and the origin override are editable in place.

**Ownership proof is the customer's job** -- the outputs carry the TXT record (and HTTP alternative) the customer must publish on *their* DNS before the hostname activates. Build your onboarding hand-off around those outputs; nothing on your side can complete this step for them.

**Custom Origin Server (`customOriginServer`)** -- leave it empty and traffic routes to the zone's fallback origin, which is a per-zone singleton (`CloudflareCustomHostnameFallbackOrigin`) -- set it once for the zone, not on each hostname. Set the override only when a specific customer needs a dedicated backend. Note `customOriginSni` is not configurable when the hostname uses the fallback origin.

**SSL (`ssl`)** -- unset means a Cloudflare-managed DV certificate with sensible defaults, which is the right choice for almost every hostname. The SSL block is a nested lifecycle: changing validation method, bundle method, or TLS settings edits in place rather than replacing the hostname. Bring-your-own certificates (`customCertificate`, `customCertBundle`) are Enterprise-gated -- keep them off the default path. The private key fields are sensitive and handled as managed secrets.

**Custom Metadata (`customMetadata`)** -- arbitrary key/values (e.g. a tenant id) readable by Workers and rules at request time. This is the hook for per-tenant routing logic at the edge.

**Deletion is soft** -- a hostname in `pending_deletion` or `deleted` still answers API reads. Automation that treats "the id exists" as "the hostname is live" will lie after a destroy; key on the hostname's activation status read from the Cloudflare API instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

`customOriginServer` is also a value-or-reference field and may point at any resource output that resolves to a backend hostname (e.g. a load balancer). When it is unset, the hostname relies on the zone's **CloudflareCustomHostnameFallbackOrigin**.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `custom_hostname_id` | The Cloudflare-assigned identifier | Verification tooling and imports -- paired with `zone_id`, since the hostname's API identity is the (zone, hostname) tuple |
| `ownership_verification_name`, `ownership_verification_type`, `ownership_verification_value` | The DNS record the customer adds to prove control | Customer onboarding hand-off (tickets, emails, portals) |
| `ownership_verification_http_url`, `ownership_verification_http_body` | The HTTP alternative for ownership verification | Customer onboarding hand-off when the customer cannot edit DNS |
| `created_at` | RFC3339 creation timestamp | Auditing |
| `zone_id` | The SaaS zone the hostname was onboarded onto | Completes the hostname's API identity for tooling that composes on it |

Activation status and verification errors are deliberately not outputs — both transition asynchronously server-side (status walks `pending` → `pending_validation` → `active`; the verification-errors list fills and empties as verification progresses); read them from the Cloudflare API or dashboard.

## Common Patterns

**Managed DV onboarding** -- onboard the customer hostname with an auto-provisioned certificate, hand them the ownership-verification record from the outputs, and let the zone's fallback origin serve the traffic. This is the shape for nearly all SaaS vanity domains. Start from the **SaaS Vanity Domain (recommended)** preset.

**Dedicated backend per customer** -- set `customOriginServer` on the hostnames of customers who need isolation (their own cluster, their own region) while everyone else rides the fallback origin.

**Bring your own certificate** -- Enterprise accounts that must serve a customer-supplied certificate upload it in the `ssl` block. Start from the **Bring Your Own Certificate (Enterprise)** preset -- and verify the account is actually Enterprise first, because the API rejects the upload with 403 on other plans.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the SaaS zone the hostname is added to; `zoneId` references its output
- [**Cloudflare Custom Hostname Fallback Origin**](/infra-catalog/cloudflare-custom-hostname-fallback-origin) -- the zone-level default backend this hostname routes to when no override is set
- [**Cloudflare Load Balancer**](/infra-catalog/cloudflare-load-balancer) -- a common `customOriginServer` target for customers with dedicated backends

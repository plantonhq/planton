# GCP Backend Service

Deploys a Compute Engine backend service — the hub of GCP's load-balancing family, global (the default) or regional when `region` is set. A backend service owns HOW traffic reaches a set of backends: which instance groups or network endpoint groups receive requests, how they are health-checked, how sessions stick, whether responses cache at Google's edge (Cloud CDN), whether Identity-Aware Proxy gates access, and how requests are logged. URL maps route host/path patterns to backend services; target proxies and forwarding rules sit in front of the URL map.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Compute Engine API enablement** (`compute.googleapis.com`) on the target project (never disabled on destroy)
- **Backend Service** -- global: for the classic external ALB (EXTERNAL), the envoy-based external ALB (EXTERNAL_MANAGED), the cross-region internal ALB (INTERNAL_MANAGED), or Traffic Director (INTERNAL_SELF_MANAGED); regional (when `region` is set): for the regional external ALB (EXTERNAL_MANAGED), the regional internal ALB (INTERNAL_MANAGED), or the internal and external passthrough Network Load Balancers (INTERNAL, EXTERNAL)
- **Backend attachments** -- each configured group with its balancing mode and capacity dials
- **Attached policies** -- Cloud CDN caching, Cloud Armor references, IAP, logging, and the scheme's traffic policies

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the service will be created. Provide the project ID directly or reference a GcpProject Cloud Resource via ValueFromRef.
- **A health check** (GcpHealthCheck) for instance-group backends — serverless and internet NEG backends manage their own health.
- **The backends** (GcpRegionNetworkEndpointGroup or instance groups) — or create the service health-check-only first and attach backends as they come online.

## Deploy

### Console

Open the deployment store, find **GCP Backend Service**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **External Web Backend** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBackendService
metadata:
  name: web-backend-service
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  healthCheck:
    value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/global/healthChecks/web-hc"
  backends:
    - group:
        value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/regions/us-central1/networkEndpointGroups/orders-neg"
      balancingMode: RATE
      maxRate: 100
```

```shell
planton apply -f backend-service.yaml
```

This creates the classic external web backend: default scheme and protocol, one NEG backend with a rate contract, and a health check. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the whole backend tier:

```yaml
spec:
  healthCheck:
    valueFrom:
      kind: GcpHealthCheck
      name: web-hc
      fieldPath: status.outputs.self_link
  backends:
    - group:
        valueFrom:
          kind: GcpRegionNetworkEndpointGroup
          name: orders-neg
          fieldPath: status.outputs.self_link
      balancingMode: RATE
      maxRate: 100
```

The InfraPipeline resolves the dependency graph — health check and NEG first, then the service — and a downstream GcpUrlMap references this service's `self_link`.

## Key Configuration

These are the most important decisions when configuring a backend service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- `region` empty builds the global backend service; a region name builds the regional one, which adds the passthrough Network Load Balancer levers (network, backend failover, connection tracking, HA policy, zonal affinity) and lacks the global edge features. A regional ALB needs a regional health check; a passthrough NLB's forwarding rule names the service directly. Immutable.

**Load balancer family** -- `loadBalancingScheme` (default EXTERNAL) decides which capabilities exist: Cloud CDN is external-only; circuit breakers, consistent hash, and stream limits are Traffic Director-only; outlier detection needs INTERNAL_SELF_MANAGED or EXTERNAL_MANAGED; backend `preference` needs a non-EXTERNAL scheme. A service cannot change families — the EXTERNAL → EXTERNAL_MANAGED canary (`externalManagedMigrationState`) is the only in-place transition. Immutable in practice.

**Backends** -- Each row names a group (a NEG by reference, or an instance-group self-link) with a balancing mode and its capacity targets: RATE requires a rate dial, CONNECTION a connection dial, CUSTOM_METRICS at least one ORCA metric. One service never mixes instance groups with NEGs. The `capacityScaler` is the drain/blue-green lever (0 drains without removing). Mutable — adding, removing, and re-weighing backends is the normal day-2 life.

**Session affinity** -- Default NONE. `strongSessionAffinityCookie` customizes STRONG_COOKIE_AFFINITY only (omit it for GCP's default cookie); `affinityCookieTtlSec` only with GENERATED_COOKIE; no affinity applies with the UDP protocol. Best-effort, never a guarantee.

**Cloud CDN** -- `enableCdn` plus `cdnPolicy`: the cache mode, TTLs, negative caching, request coalescing, the rich cache-key policy (host/protocol/query toggles, whitelist/blacklist, headers, named cookies), and up to 3 signed-URL keys (each value handled as a secret) for serving private content from the edge.

**Security stack** -- `securityPolicy` (Cloud Armor WAF, after the cache) and `edgeSecurityPolicy` (CLOUD_ARMOR_EDGE, before it); `iap` with the paired OAuth client (the secret never stored in plaintext); Traffic Director mTLS via `securitySettings`; AWS SigV4 signing for internet-NEG origins; protocol-gated backend TLS pinning via `tlsSettings`.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpHealthCheck** | `healthCheck` | `status.outputs.self_link` |
| **GcpRegionNetworkEndpointGroup** | `backends[].group` | `status.outputs.self_link` |
| **GcpCloudArmorPolicy** | `securityPolicy` | `status.outputs.policy_self_link` |
| **GcpCloudArmorPolicy** | `edgeSecurityPolicy` | `status.outputs.policy_self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the backend service | GcpUrlMap default service and route rules via ValueFromRef |
| `backend_service_name` | Name as it exists in GCP | Audit, fleet inventory |
| `generated_id` | GCP's numeric identifier | API-level integrations |
| `fingerprint` | The optimistic-locking token | Concurrent-update tooling |
| `region` | Region of a regional backend service; empty for global | Scope checks on downstream blocks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**External web backend** -- The classic shape: default scheme, a health check, NEG backends. Start from the **External Web Backend** preset.

**CDN-cached API** -- Edge caching with a tuned cache key for read-heavy APIs. Start from the **CDN-Cached API** preset.

**IAP-protected internal tool** -- Zero-trust access with Google identities in front of an internal app. Start from the **IAP-Protected Internal Tool** preset.

**Internal passthrough NLB backend** -- A regional INTERNAL backend service with a failover pool, connection tracking, and zonal affinity, named directly by a regional forwarding rule. Start from the **Internal Passthrough NLB Backend** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the service is created
- [**GCP Health Check**](/cloud-catalog/gcp-health-check) -- the probe referenced by `healthCheck`
- [**GCP Region Network Endpoint Group**](/cloud-catalog/gcp-region-network-endpoint-group) -- the serverless/PSC/internet backends referenced by `backends[].group`
- [**GCP URL Map**](/cloud-catalog/gcp-url-map) -- consumes this service's `self_link` in its routes
- [**GCP Backend Bucket**](/cloud-catalog/gcp-backend-bucket) -- the static-content sibling on the same URL map

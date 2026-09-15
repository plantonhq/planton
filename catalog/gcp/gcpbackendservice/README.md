# GCP Backend Service

Deploys a global Compute Engine backend service (`google_compute_backend_service`) — the hub of GCP's L7 load balancing family. A backend service owns HOW traffic reaches a set of backends: which instance groups or network endpoint groups receive requests, how they are health-checked, how sessions stick, whether responses are cached by Cloud CDN, whether Identity-Aware Proxy gates access, and how requests are logged. URL maps route host/path patterns to backend services; target proxies and forwarding rules sit in front of the URL map.

## What Gets Created

When you deploy a GcpBackendService resource, Planton provisions:

- **Backend Service** — a global `google_compute_backend_service` with its backends, health check, session affinity, CDN policy, IAP, Cloud Armor attachments, logging, and traffic policy
- **Signed-URL Keys** (optional) — one `google_compute_backend_service_signed_url_key` per entry in `signedUrlKeys`, for serving private CDN content with expiring, tamper-proof links

The health check, Cloud Armor policies, and backend groups are deliberately NOT created here — they are their own composable nodes, attached by reference. One health check is commonly shared by many backend services; one Cloud Armor policy protects many backends.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A health check** — referenced via `healthCheck` (a GcpHealthCheck resource or a literal self-link); required unless every backend is an internet or serverless NEG
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `backend-service.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBackendService
metadata:
  name: web-backend
spec:
  projectId:
    value: my-gcp-project-123
  healthCheck:
    valueFrom:
      kind: GcpHealthCheck
      name: web-hc
      fieldPath: status.outputs.self_link
  backends:
    - group:
        value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/zones/us-central1-a/instanceGroups/web-ig
      balancingMode: UTILIZATION
      maxUtilization: 0.8
```

Deploy:

```shell
planton apply -f backend-service.yaml
```

## Configuration Reference

### Optional Fields (all fields are optional — a health-checked service with no backends is a valid creation order)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | Project that owns the backend service. Immutable. |
| `backendServiceName` | `string` | `metadata.name` | Cloud-side name (RFC1035). Immutable. |
| `description` | `string` | `""` | What this service fronts and which URL maps route to it. |
| `protocol` | `string` | `HTTP` | LB→backend protocol: `HTTP`, `HTTPS`, `HTTP2`, `H2C`, `TCP`, `SSL`, `UDP`, `GRPC`. |
| `loadBalancingScheme` | `string` | `EXTERNAL` | `EXTERNAL`, `EXTERNAL_MANAGED`, `INTERNAL_MANAGED`, or `INTERNAL_SELF_MANAGED` (Traffic Director). Immutable except the EXTERNAL→EXTERNAL_MANAGED canary. |
| `portName` | `string` | none | Named port on the instance groups (EXTERNAL + instance-group backends). |
| `timeoutSec` | `int` | `30` | Backend response timeout; raise well above the longest stream for WebSockets/gRPC. |
| `connectionDrainingTimeoutSec` | `int` | `300` | Seconds draining instances keep existing connections. |
| `healthCheck` | `StringValueOrRef` | none | The ONE health check (GCP caps at one). Can reference a GcpHealthCheck. Required unless all backends are internet/serverless NEGs. |
| `backends` | `list(object)` | `[]` | The instance groups / NEGs serving traffic — see below. |
| `sessionAffinity` | `string` | `NONE` | `CLIENT_IP*`, `GENERATED_COOKIE`, `HEADER_FIELD`, `HTTP_COOKIE`, `STRONG_COOKIE_AFFINITY`. |
| `affinityCookieTtlSec` | `int` | session cookie | Generated-cookie lifetime (GENERATED_COOKIE only, ≤ 86400). |
| `strongSessionAffinityCookie` | object | none | Customizes the cookie of STRONG_COOKIE_AFFINITY (omit for GCP's default cookie; not valid with other modes). |
| `localityLbPolicy` | `string` | `ROUND_ROBIN` | Within-group algorithm incl. `LEAST_REQUEST`, `RING_HASH`, `MAGLEV`, `WEIGHTED_ROUND_ROBIN`. |
| `localityLbPolicies` | `list(object)` | `[]` | Ordered built-in/custom xDS policy list (Traffic Director). |
| `consistentHash` | object | none | Hash key parameters (INTERNAL_SELF_MANAGED + MAGLEV/RING_HASH). |
| `enableCdn` | `bool` | `false` | Cache responses at Google's edge (external schemes only). |
| `cdnPolicy` | object | GCP defaults | How responses are cached — see below. |
| `securityPolicy` | `StringValueOrRef` | none | Cloud Armor backend policy (type `CLOUD_ARMOR`), evaluated after the CDN cache. |
| `edgeSecurityPolicy` | `StringValueOrRef` | none | Cloud Armor EDGE policy (type `CLOUD_ARMOR_EDGE`), evaluated before the CDN cache. |
| `iap` | object | off | Identity-Aware Proxy; `oauth2ClientSecret` is secret material. |
| `logConfig` | object | off | Request logging with sampling and optional-field control. |
| `customRequestHeaders` / `customResponseHeaders` | `list(string)` | `[]` | Headers the LB adds, `"Header-Name: value"` form; values may use variables like `{client_ip}` / `{cdn_cache_status}`. |
| `compressionMode` | `string` | disabled | `AUTOMATIC` (gzip/brotli) or `DISABLED`. |
| `circuitBreakers` | object | GCP defaults | Connection-volume limits (INTERNAL_SELF_MANAGED only). |
| `outlierDetection` | object | GCP defaults | Passive health ejection (INTERNAL_SELF_MANAGED / EXTERNAL_MANAGED). |
| `maxStreamDuration` | object | no limit | Default stream timeout (INTERNAL_SELF_MANAGED only). |
| `securitySettings` | object | none | Traffic Director mTLS (client TLS policy, SANs) + AWS SigV4 origin signing; `awsV4Authentication.accessKey` is secret material. |
| `tlsSettings` | object | none | Backend TLS: authentication config, SNI, acceptable SANs (protocol SSL/HTTPS/HTTP2 only). |
| `ipAddressSelectionPolicy` | `string` | IPv4 | `IPV4_ONLY`, `PREFER_IPV6`, `IPV6_ONLY` toward dual-stack backends. |
| `externalManagedMigrationState` / `...TestingPercentage` | `string` / `double` | none | EXTERNAL→EXTERNAL_MANAGED canary controls. |
| `customMetrics` | `list(object)` | `[]` | Service-level ORCA metrics for WEIGHTED_ROUND_ROBIN. |
| `serviceLbPolicy` | `string` | none | Self-link of a networkservices ServiceLbPolicy. |
| `signedUrlKeys` | `list(object)` | `[]` | Up to 3 named signing keys for Cloud CDN signed URLs/cookies. `keyValue` is secret material. |
| `resourceManagerTags` | `map<string,string>` | `{}` | Create-time Resource Manager tags (`tagKeys/{id}` → `tagValues/{id}`) for org policy and IAM conditions. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does to the backend service AND its signed-URL keys: `DELETE`, `PREVENT` (refuse), or `ABANDON` (keep serving, drop from management). |

### Backends

| Field | Default | Description |
|-------|---------|-------------|
| `group` | required | Instance-group or NEG self-link (`StringValueOrRef`). All backends of one service must be the same family. |
| `balancingMode` | `UTILIZATION` | `UTILIZATION` (instance CPU), `RATE` (req/s), `CONNECTION` (TCP/SSL), `CUSTOM_METRICS` (ORCA). NEGs must use RATE or CUSTOM_METRICS. |
| `capacityScaler` | `1.0` | Fraction of capacity accepted; `0` drains the backend without removing it. |
| `maxRate` / `maxRatePerInstance` / `maxRatePerEndpoint` | — | RATE-mode targets (one required in RATE mode). |
| `maxConnections` / `...PerInstance` / `...PerEndpoint` | — | CONNECTION-mode targets (one required in CONNECTION mode). |
| `maxUtilization` | `0.8` | UTILIZATION-mode CPU target (instance groups only — GCP strips it from NEG backends). |
| `preference` | `DEFAULT` | `PREFERRED` backends fill before `DEFAULT` ones (not valid on EXTERNAL scheme). |
| `customMetrics` | `[]` | Per-backend ORCA metrics for CUSTOM_METRICS mode. |

### CDN Policy

Identical semantics to the backend bucket's CDN policy (cache modes, TTLs, negative caching, serve-while-stale, coalescing, signed-URL cache age, bypass headers), with a richer cache key: `includeHost`, `includeProtocol`, `includeQueryString`, `queryStringWhitelist`/`queryStringBlacklist` (mutually exclusive), `includeHttpHeaders`, and `includeNamedCookies`.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value URL maps reference as a default service or path-rule target |
| `backend_service_name` | `string` | Name of the backend service in GCP |
| `generated_id` | `string` | Server-assigned numeric ID |
| `fingerprint` | `string` | Optimistic-concurrency fingerprint for out-of-band updates |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Immutability**: `backendServiceName` and `projectId` are ForceNew; `loadBalancingScheme` cannot change in place except via the EXTERNAL→EXTERNAL_MANAGED canary migration. Everything else — backends, CDN policy, affinity, IAP — updates in place.
- **One health check**: GCP allows at most one health check per backend service, so the spec models it singular. A service with no health check is only valid when every backend is an internet or serverless NEG.
- **Instance groups and NEGs don't mix**: all backends of one service must be the same family; GCP rejects mixed backend lists.
- **Secrets**: `iap.oauth2ClientSecret`, `securitySettings.awsV4Authentication.accessKey`, and each `signedUrlKeys[].keyValue` are secret material — reference-only in the control plane, marked secret in Pulumi state, never in stack outputs.
- **Scheme applicability is validated pre-deploy**: CDN only on external schemes; circuit breakers and max-stream-duration only on INTERNAL_SELF_MANAGED; outlier detection on INTERNAL_SELF_MANAGED/EXTERNAL_MANAGED; consistent hash only with MAGLEV/RING_HASH. The spec rejects incoherent combinations before GCP would.
- **Cloud Armor attach semantics**: the provider applies `securityPolicy`/`edgeSecurityPolicy` via dedicated set-policy API calls — changing only a policy reference is a targeted update, not a full resource patch.

## Related Components

- [GcpHealthCheck](/docs/catalog/gcp/gcphealthcheck) — the probe deciding which backends receive traffic
- [GcpBackendBucket](/docs/catalog/gcp/gcpbackendbucket) — the static-content counterpart for GCS origins
- [GcpCloudArmorPolicy](/docs/catalog/gcp/gcpcloudarmorpolicy) — WAF/rate-limiting policies attached by reference
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the backend service

## Additional Resources

- [Backend services overview](https://cloud.google.com/load-balancing/docs/backend-service)
- [Balancing modes](https://cloud.google.com/load-balancing/docs/backend-service#balancing-mode)
- [Identity-Aware Proxy for load balancers](https://cloud.google.com/iap/docs/enabling-compute-howto)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

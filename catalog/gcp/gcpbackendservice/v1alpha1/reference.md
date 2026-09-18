# GcpBackendService

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBackendServiceSpec defines a global Compute Engine backend service — the
hub of GCP's L7 load balancing family. A backend service owns HOW traffic
reaches a set of backends: which instance groups or network endpoint groups
receive requests, how they are health-checked, how sessions stick, whether
responses are cached by Cloud CDN, whether Identity-Aware Proxy gates
access, and how requests are logged. URL maps route host/path patterns to
backend services; target proxies and forwarding rules sit in front of the
URL map. Each piece is its own resource, referenced by self-link.

This kind models the GLOBAL backend service — the backend of the global
external Application Load Balancer, Traffic Director / service-mesh
(INTERNAL_SELF_MANAGED), and the cross-region internal ALB
(INTERNAL_MANAGED). The regional backend service is a different GCP
resource with different capabilities (failover policy, connection
tracking, network scoping) and is deliberately not folded in here.

Cloud CDN is a policy ON this resource, not a separate GCP object:
enable_cdn turns edge caching on and cdn_policy tunes how responses are
cached and keyed. Cloud Armor attaches by reference through
security_policy / edge_security_policy.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBackendService
metadata:
  name: my-sample-backend-service
spec:
  # GCP project that owns the backend service.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  backendServiceName: web-backend

  description: Serves the web application through the global external ALB

  # LB→backend protocol; the frontend protocol is the target proxy's concern.
  protocol: HTTP

  # The classic global external Application Load Balancer.
  loadBalancingScheme: EXTERNAL

  # Named port defined on the instance groups below.
  portName: http

  timeoutSec: 30

  # The health check deciding which backends receive traffic
  # (or reference a GcpHealthCheck).
  healthCheck:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/healthChecks/web-hc

  # One instance-group backend balanced on CPU utilization.
  backends:
    - group:
        value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/zones/us-central1-a/instanceGroups/web-ig
      balancingMode: UTILIZATION
      maxUtilization: 0.8
      description: primary web pool

  # Stick clients to a backend via a generated cookie for one hour.
  sessionAffinity: GENERATED_COOKIE
  affinityCookieTtlSec: 3600

  # Cache static responses at Google's edge.
  enableCdn: true
  cdnPolicy:
    cacheMode: CACHE_ALL_STATIC
    defaultTtl: 3600
    clientTtl: 1800
    # Cache 404s briefly so missing-asset storms don't hammer the backends.
    negativeCaching: true
    negativeCachingPolicy:
      - code: 404
        ttl: 60
    requestCoalescing: true
    cacheKeyPolicy:
      includeHost: true
      includeProtocol: true
      includeQueryString: true
      # Only these parameters change the response; everything else shares a
      # cache entry.
      queryStringWhitelist:
        - v

  # Compress compressible content types for clients that ask.
  compressionMode: AUTOMATIC

  # Log a 10% sample of requests.
  logConfig:
    enable: true
    sampleRate: 0.1

  # Surface the CDN verdict for debugging.
  customResponseHeaders:
    - "X-Cache-Status: {cdn_cache_status}"

  # Ephemeral test fixture: delete the backend service (and any signed-URL
  # keys) on destroy.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.backendServiceName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.protocol` | `string` |  | `HTTP` |  |
| `spec.loadBalancingScheme` | `string` |  | `EXTERNAL` |  |
| `spec.portName` | `string` |  |  |  |
| `spec.timeoutSec` | `int32` |  | `30` |  |
| `spec.connectionDrainingTimeoutSec` | `int32` |  | `300` |  |
| `spec.healthCheck` | `string \| valueFrom` |  |  | GcpHealthCheck (`status.outputs.self_link`) |
| `spec.backends` | `[]GcpBackendServiceBackend` |  |  |  |
| `spec.backends[].group` | `string \| valueFrom` | yes |  | GcpRegionNetworkEndpointGroup (`status.outputs.self_link`) |
| `spec.backends[].balancingMode` | `string` |  | `UTILIZATION` |  |
| `spec.backends[].capacityScaler` | `double` |  | `1.0` |  |
| `spec.backends[].description` | `string` |  |  |  |
| `spec.backends[].maxConnections` | `int32` |  |  |  |
| `spec.backends[].maxConnectionsPerInstance` | `int32` |  |  |  |
| `spec.backends[].maxConnectionsPerEndpoint` | `int32` |  |  |  |
| `spec.backends[].maxRate` | `int32` |  |  |  |
| `spec.backends[].maxRatePerInstance` | `double` |  |  |  |
| `spec.backends[].maxRatePerEndpoint` | `double` |  |  |  |
| `spec.backends[].maxUtilization` | `double` |  |  |  |
| `spec.backends[].preference` | `string` |  |  |  |
| `spec.backends[].customMetrics` | `[]GcpBackendServiceBackendCustomMetric` |  |  |  |
| `spec.backends[].customMetrics[].name` | `string` | yes |  |  |
| `spec.backends[].customMetrics[].dryRun` | `bool` |  |  |  |
| `spec.backends[].customMetrics[].maxUtilization` | `double` |  | `0.8` |  |
| `spec.sessionAffinity` | `string` |  | `NONE` |  |
| `spec.affinityCookieTtlSec` | `int32` |  |  |  |
| `spec.strongSessionAffinityCookie` | `GcpBackendServiceStrongSessionAffinityCookie` |  |  |  |
| `spec.strongSessionAffinityCookie.name` | `string` |  |  |  |
| `spec.strongSessionAffinityCookie.path` | `string` |  |  |  |
| `spec.strongSessionAffinityCookie.ttl` | `GcpBackendServiceDuration` |  |  |  |
| `spec.strongSessionAffinityCookie.ttl.seconds` | `int64` |  |  |  |
| `spec.strongSessionAffinityCookie.ttl.nanos` | `int32` |  |  |  |
| `spec.localityLbPolicy` | `string` |  |  |  |
| `spec.localityLbPolicies` | `[]GcpBackendServiceLocalityLbPolicyConfig` |  |  |  |
| `spec.localityLbPolicies[].policy` | `GcpBackendServiceLocalityLbPolicy` |  |  |  |
| `spec.localityLbPolicies[].policy.name` | `string` | yes |  |  |
| `spec.localityLbPolicies[].customPolicy` | `GcpBackendServiceLocalityLbCustomPolicy` |  |  |  |
| `spec.localityLbPolicies[].customPolicy.name` | `string` | yes |  |  |
| `spec.localityLbPolicies[].customPolicy.data` | `string` |  |  |  |
| `spec.consistentHash` | `GcpBackendServiceConsistentHash` |  |  |  |
| `spec.consistentHash.httpCookie` | `GcpBackendServiceConsistentHashHttpCookie` |  |  |  |
| `spec.consistentHash.httpCookie.name` | `string` |  |  |  |
| `spec.consistentHash.httpCookie.path` | `string` |  |  |  |
| `spec.consistentHash.httpCookie.ttl` | `GcpBackendServiceDuration` |  |  |  |
| `spec.consistentHash.httpCookie.ttl.seconds` | `int64` |  |  |  |
| `spec.consistentHash.httpCookie.ttl.nanos` | `int32` |  |  |  |
| `spec.consistentHash.httpHeaderName` | `string` |  |  |  |
| `spec.consistentHash.minimumRingSize` | `int64` |  | `1024` |  |
| `spec.enableCdn` | `bool` |  |  |  |
| `spec.cdnPolicy` | `GcpBackendServiceCdnPolicy` |  |  |  |
| `spec.cdnPolicy.cacheMode` | `string` |  |  |  |
| `spec.cdnPolicy.clientTtl` | `int32` |  |  |  |
| `spec.cdnPolicy.defaultTtl` | `int32` |  |  |  |
| `spec.cdnPolicy.maxTtl` | `int32` |  |  |  |
| `spec.cdnPolicy.negativeCaching` | `bool` |  |  |  |
| `spec.cdnPolicy.negativeCachingPolicy` | `[]GcpBackendServiceNegativeCachingPolicy` |  |  |  |
| `spec.cdnPolicy.negativeCachingPolicy[].code` | `int32` | yes |  |  |
| `spec.cdnPolicy.negativeCachingPolicy[].ttl` | `int32` |  |  |  |
| `spec.cdnPolicy.serveWhileStale` | `int32` |  |  |  |
| `spec.cdnPolicy.requestCoalescing` | `bool` |  |  |  |
| `spec.cdnPolicy.signedUrlCacheMaxAgeSec` | `int32` |  | `3600` |  |
| `spec.cdnPolicy.cacheKeyPolicy` | `GcpBackendServiceCdnCacheKeyPolicy` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.includeHost` | `bool` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.includeProtocol` | `bool` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.includeQueryString` | `bool` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.queryStringWhitelist` | `[]string` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.queryStringBlacklist` | `[]string` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.includeHttpHeaders` | `[]string` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.includeNamedCookies` | `[]string` |  |  |  |
| `spec.cdnPolicy.bypassCacheOnRequestHeaders` | `[]GcpBackendServiceBypassCacheOnRequestHeader` |  |  |  |
| `spec.cdnPolicy.bypassCacheOnRequestHeaders[].headerName` | `string` | yes |  |  |
| `spec.securityPolicy` | `string \| valueFrom` |  |  | GcpCloudArmorPolicy (`status.outputs.policy_self_link`) |
| `spec.edgeSecurityPolicy` | `string \| valueFrom` |  |  | GcpCloudArmorPolicy (`status.outputs.policy_self_link`) |
| `spec.iap` | `GcpBackendServiceIap` |  |  |  |
| `spec.iap.enabled` | `bool` |  |  |  |
| `spec.iap.oauth2ClientId` | `string` |  |  |  |
| `spec.iap.oauth2ClientSecret` | `string` (sensitive) |  |  |  |
| `spec.logConfig` | `GcpBackendServiceLogConfig` |  |  |  |
| `spec.logConfig.enable` | `bool` |  |  |  |
| `spec.logConfig.sampleRate` | `double` |  | `1.0` |  |
| `spec.logConfig.optionalMode` | `string` |  |  |  |
| `spec.logConfig.optionalFields` | `[]string` |  |  |  |
| `spec.logConfig.requestHeaders` | `[]string` |  |  |  |
| `spec.logConfig.responseHeaders` | `[]string` |  |  |  |
| `spec.customRequestHeaders` | `[]string` |  |  |  |
| `spec.customResponseHeaders` | `[]string` |  |  |  |
| `spec.compressionMode` | `string` |  |  |  |
| `spec.circuitBreakers` | `GcpBackendServiceCircuitBreakers` |  |  |  |
| `spec.circuitBreakers.maxConnections` | `int32` |  | `1024` |  |
| `spec.circuitBreakers.maxPendingRequests` | `int32` |  | `1024` |  |
| `spec.circuitBreakers.maxRequests` | `int32` |  | `1024` |  |
| `spec.circuitBreakers.maxRequestsPerConnection` | `int32` |  |  |  |
| `spec.circuitBreakers.maxRetries` | `int32` |  | `3` |  |
| `spec.outlierDetection` | `GcpBackendServiceOutlierDetection` |  |  |  |
| `spec.outlierDetection.baseEjectionTime` | `GcpBackendServiceDuration` |  |  |  |
| `spec.outlierDetection.baseEjectionTime.seconds` | `int64` |  |  |  |
| `spec.outlierDetection.baseEjectionTime.nanos` | `int32` |  |  |  |
| `spec.outlierDetection.consecutiveErrors` | `int32` |  |  |  |
| `spec.outlierDetection.consecutiveGatewayFailure` | `int32` |  |  |  |
| `spec.outlierDetection.enforcingConsecutiveErrors` | `int32` |  |  |  |
| `spec.outlierDetection.enforcingConsecutiveGatewayFailure` | `int32` |  |  |  |
| `spec.outlierDetection.enforcingSuccessRate` | `int32` |  |  |  |
| `spec.outlierDetection.interval` | `GcpBackendServiceDuration` |  |  |  |
| `spec.outlierDetection.interval.seconds` | `int64` |  |  |  |
| `spec.outlierDetection.interval.nanos` | `int32` |  |  |  |
| `spec.outlierDetection.maxEjectionPercent` | `int32` |  |  |  |
| `spec.outlierDetection.successRateMinimumHosts` | `int32` |  |  |  |
| `spec.outlierDetection.successRateRequestVolume` | `int32` |  |  |  |
| `spec.outlierDetection.successRateStdevFactor` | `int32` |  |  |  |
| `spec.maxStreamDuration` | `GcpBackendServiceDuration` |  |  |  |
| `spec.maxStreamDuration.seconds` | `int64` |  |  |  |
| `spec.maxStreamDuration.nanos` | `int32` |  |  |  |
| `spec.securitySettings` | `GcpBackendServiceSecuritySettings` |  |  |  |
| `spec.securitySettings.clientTlsPolicy` | `string` |  |  |  |
| `spec.securitySettings.subjectAltNames` | `[]string` |  |  |  |
| `spec.securitySettings.awsV4Authentication` | `GcpBackendServiceAwsV4Authentication` |  |  |  |
| `spec.securitySettings.awsV4Authentication.accessKeyId` | `string` |  |  |  |
| `spec.securitySettings.awsV4Authentication.accessKey` | `string` (sensitive) |  |  |  |
| `spec.securitySettings.awsV4Authentication.accessKeyVersion` | `string` |  |  |  |
| `spec.securitySettings.awsV4Authentication.originRegion` | `string` |  |  |  |
| `spec.tlsSettings` | `GcpBackendServiceTlsSettings` |  |  |  |
| `spec.tlsSettings.authenticationConfig` | `string` |  |  |  |
| `spec.tlsSettings.sni` | `string` |  |  |  |
| `spec.tlsSettings.subjectAltNames` | `[]GcpBackendServiceTlsSubjectAltName` |  |  |  |
| `spec.tlsSettings.subjectAltNames[].dnsName` | `string` |  |  |  |
| `spec.tlsSettings.subjectAltNames[].uniformResourceIdentifier` | `string` |  |  |  |
| `spec.ipAddressSelectionPolicy` | `string` |  |  |  |
| `spec.externalManagedMigrationState` | `string` |  |  |  |
| `spec.externalManagedMigrationTestingPercentage` | `double` |  |  |  |
| `spec.customMetrics` | `[]GcpBackendServiceCustomMetric` |  |  |  |
| `spec.customMetrics[].name` | `string` | yes |  |  |
| `spec.customMetrics[].dryRun` | `bool` |  |  |  |
| `spec.serviceLbPolicy` | `string` |  |  |  |
| `spec.signedUrlKeys` | `[]GcpBackendServiceSignedUrlKey` |  |  |  |
| `spec.signedUrlKeys[].name` | `string` | yes |  |  |
| `spec.signedUrlKeys[].keyValue` | `string` (sensitive) | yes |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the backend service.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the backend service.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.backendServiceName

`string`

Name of the backend service in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the backend service,
briefly breaking every URL map that references the old self_link.

- rule: backend_service_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.description

`string`

What this backend service fronts and which URL maps route to it — write
it for the operator tracing a request path later. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.protocol

`string` · optional (explicit presence)

The protocol the load balancer uses to talk to the backends (default
HTTP). This is the LB→backend leg, independent of what clients speak to
the load balancer: an HTTPS frontend commonly forwards to HTTP backends.
H2C is HTTP/2 over cleartext. Must be GRPC when the backend service is
referenced by a URL map bound to a target gRPC proxy. Mutable, but
switching protocol families usually also means changing the health
check and backend ports.

- default: `HTTP`
- rule: protocol must be one of HTTP, HTTPS, HTTP2, H2C, TCP, SSL, UDP, GRPC, or UNSPECIFIED

### spec.loadBalancingScheme

`string` · optional (explicit presence)

Which load balancer family this backend service serves (default
EXTERNAL, the classic global external Application LB). EXTERNAL_MANAGED
is the newer envoy-based global external ALB; INTERNAL_MANAGED is the
cross-region internal ALB; INTERNAL_SELF_MANAGED is Traffic Director /
service mesh. A backend service created for one family cannot serve
another — the only in-place transition GCP supports is the canary
migration EXTERNAL → EXTERNAL_MANAGED driven by
external_managed_migration_state.

- default: `EXTERNAL`
- rule: load_balancing_scheme must be one of EXTERNAL, EXTERNAL_MANAGED, INTERNAL_MANAGED, or INTERNAL_SELF_MANAGED

### spec.portName

`string`

Name of the backend port to use for instance-group backends. The same
named port must be defined on every instance group this service
references — each group maps the logical name to its own port number.
Required by GCP when the scheme is EXTERNAL and the backends are
instance groups; ignored for NEG backends (endpoints carry their own
ports). Mutable.

- rule: port_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.timeoutSec

`int32` · optional (explicit presence)

Seconds the load balancer waits for a backend to fully respond before
giving up on the request (default 30). For streaming workloads
(WebSockets, gRPC streams, long polling) raise this well above the
longest expected stream duration. Not used by serverless NEG backends —
Cloud Run/Functions manage their own request timeouts. Mutable.

- default: `30`
- rule: {"int32":{"gt":0}}

### spec.connectionDrainingTimeoutSec

`int32` · optional (explicit presence)

Seconds an instance being removed or unhealthy keeps its existing
connections open to finish in-flight requests (default 300). Lower it
for fast-draining stateless services; raise it for long-lived
connections. Mutable.

- default: `300`
- rule: {"int32":{"gte":0}}

### spec.healthCheck

`string | valueFrom`

The health check that decides which backends receive traffic. GCP
allows at most ONE health check per backend service, so this is a
single reference, not a list. Reference a GcpHealthCheck resource or
provide a health check self-link directly. Required by GCP unless every
backend is an internet or serverless NEG — serverless platforms manage
their own health. Mutable.

- references: GcpHealthCheck (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpHealthCheck, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.backends

`[]GcpBackendServiceBackend`

The backends that actually serve traffic — instance groups or network
endpoint groups, each with its own balancing mode and capacity dials.
A backend service may mix backends of the same family but cannot mix
instance groups with NEGs. May be empty: a backend service with only a
health check is valid and is the natural creation order before instance
groups or NEGs exist. Mutable — adding and removing backends is the
normal scaling/blue-green operation.

- rule: with balancing_mode RATE, set one of max_rate, max_rate_per_instance, or max_rate_per_endpoint
- rule: with balancing_mode CONNECTION, set one of max_connections, max_connections_per_instance, or max_connections_per_endpoint
- rule: with balancing_mode CUSTOM_METRICS, define at least one entry in the backend's custom_metrics

### spec.backends[].group

`string | valueFrom` · required

Fully-qualified URL of the instance group or network endpoint group
serving this backend. Accepts an instance group (zonal or regional) or
a NEG self-link; all backends of one service must be the same family —
GCP rejects mixing instance groups with NEGs. Provide the URL directly
or reference the resource that owns it: the default reference kind is a
GcpRegionNetworkEndpointGroup (the serverless/PSC/internet backend
bridge), but any group producer can be referenced explicitly by kind.
For NEG backends GCP ignores utilization-based settings.

- references: GcpRegionNetworkEndpointGroup (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpRegionNetworkEndpointGroup, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.backends[].balancingMode

`string` · optional (explicit presence)

How this backend's capacity is measured: UTILIZATION (instance CPU,
the default — instance groups only), RATE (HTTP requests per second),
CONNECTION (open connections, for TCP/SSL), CUSTOM_METRICS
(backend-reported ORCA metrics), or IN_FLIGHT (concurrent in-flight
requests). IN_FLIGHT's target dial is not surfaced by the pinned
provider, so IN_FLIGHT backends ride the API's default target.
NEG backends must use RATE (or CUSTOM_METRICS); serverless NEGs
ignore balancing entirely. Mutable.

- default: `UTILIZATION`
- rule: balancing_mode must be one of UTILIZATION, RATE, CONNECTION, CUSTOM_METRICS, or IN_FLIGHT

### spec.backends[].capacityScaler

`double` · optional (explicit presence)

Fraction of the configured capacity this backend actually accepts
(default 1.0 = 100%). 0 drains the backend without removing it — the
standard lever for maintenance and gradual rollouts. Mutable.

- default: `1.0`
- rule: {"double":{"lte":1,"gte":0}}

### spec.backends[].description

`string`

What this backend is (e.g. "blue pool, us-central1") — write it for
the operator reading a capacity page later. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.backends[].maxConnections

`int32`

Max simultaneous open connections for the whole backend (CONNECTION
mode target; optional ceiling in UTILIZATION mode). Mutable.

- rule: {"int32":{"gte":0}}

### spec.backends[].maxConnectionsPerInstance

`int32`

Max simultaneous open connections per instance-group instance. Mutable.

- rule: {"int32":{"gte":0}}

### spec.backends[].maxConnectionsPerEndpoint

`int32`

Max simultaneous open connections per NEG endpoint. Mutable.

- rule: {"int32":{"gte":0}}

### spec.backends[].maxRate

`int32`

Max HTTP requests per second for the whole backend (RATE mode target;
optional ceiling in UTILIZATION mode). Mutable.

- rule: {"int32":{"gte":0}}

### spec.backends[].maxRatePerInstance

`double`

Max HTTP requests per second per instance-group instance. Fractional
rates let small instances take partial shares. Mutable.

- rule: {"double":{"gte":0}}

### spec.backends[].maxRatePerEndpoint

`double`

Max HTTP requests per second per NEG endpoint. Mutable.

- rule: {"double":{"gte":0}}

### spec.backends[].maxUtilization

`double`

Target CPU utilization (0.0-1.0) for UTILIZATION mode — the balancer
shifts new requests away as instances approach it. GCP's default is
0.8. Ignored (and stripped by GCP) for NEG backends. Mutable.

- rule: {"double":{"lte":1,"gte":0}}

### spec.backends[].preference

`string`

Whether this backend is PREFERRED (filled to capacity before DEFAULT
backends receive traffic) — the primary/spillover pattern. Cannot be
set when the service's load_balancing_scheme is EXTERNAL. Mutable.

- rule: preference must be PREFERRED or DEFAULT

### spec.backends[].customMetrics

`[]GcpBackendServiceBackendCustomMetric`

Per-backend custom metrics for CUSTOM_METRICS balancing mode, reported
by this backend via ORCA. Each can run dry (reported but not acted on)
while being validated.

### spec.backends[].customMetrics[].name

`string` · required

Metric name as reported by the backend in ORCA load reports (e.g. a
named utilization gauge). Must match what the backend actually emits.

- rule: {"required":true,"string":{"maxLen":"256"}}

### spec.backends[].customMetrics[].dryRun

`bool`

Report the metric without acting on it — the safe first step while
validating that backends emit sane values.

### spec.backends[].customMetrics[].maxUtilization

`double` · optional (explicit presence)

Target utilization (0.0-1.0) for this metric, above which the balancer
shifts new requests away. GCP's default is 0.8.

- default: `0.8`
- rule: {"double":{"lte":1,"gte":0}}

### spec.sessionAffinity

`string` · optional (explicit presence)

How requests from the same client stick to the same backend (default
NONE — every request is balanced independently). Cookie-based modes
(GENERATED_COOKIE, HTTP_COOKIE, STRONG_COOKIE_AFFINITY) need an
HTTP-family protocol; CLIENT_IP modes hash on network attributes.
Session affinity is best-effort, not a guarantee — backends going
unhealthy still break affinity. Not applicable when protocol is UDP.
Mutable.

- default: `NONE`
- rule: session_affinity must be one of NONE, CLIENT_IP, CLIENT_IP_PORT_PROTO, CLIENT_IP_PROTO, GENERATED_COOKIE, HEADER_FIELD, HTTP_COOKIE, or STRONG_COOKIE_AFFINITY

### spec.affinityCookieTtlSec

`int32`

Lifetime in seconds of the cookie GCP generates for GENERATED_COOKIE
session affinity (0, the default, makes it a non-persistent session
cookie; max 86400 = 1 day). Only meaningful with GENERATED_COOKIE.
Mutable.

- rule: affinity_cookie_ttl_sec must be between 0 and 86400 seconds (1 day)

### spec.strongSessionAffinityCookie

`GcpBackendServiceStrongSessionAffinityCookie`

The cookie GCP uses for STRONG_COOKIE_AFFINITY — stronger stickiness
than GENERATED_COOKIE because the cookie encodes the exact backend
endpoint. Only valid with session_affinity STRONG_COOKIE_AFFINITY, and
optional there: choosing that mode is the whole statement, and the
modules send GCP the cookie configuration it requires (GCP's generated
cookie name, whole-site path, session lifetime) when this block is
absent. Declare it only to customize the cookie.

### spec.strongSessionAffinityCookie.name

`string`

Cookie name the load balancer sets and matches. Empty uses GCP's
generated default name.

- rule: {"string":{"maxLen":"256"}}

### spec.strongSessionAffinityCookie.path

`string`

Path attribute of the cookie — limit affinity to a URL subtree (e.g.
/app). Empty applies to the whole site.

- rule: {"string":{"maxLen":"1024"}}

### spec.strongSessionAffinityCookie.ttl

`GcpBackendServiceDuration`

Cookie lifetime. Zero/unset makes it a non-persistent session cookie
that vanishes when the browser closes.

### spec.strongSessionAffinityCookie.ttl.seconds

`int64`

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.strongSessionAffinityCookie.ttl.nanos

`int32`

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.localityLbPolicy

`string`

The load balancing algorithm used within each backend group once the
group is chosen (GCP default ROUND_ROBIN). LEAST_REQUEST and the
hash-based policies (RING_HASH, MAGLEV) matter for uneven request
costs and soft session affinity; WEIGHTED_ROUND_ROBIN balances on
backend-reported custom metrics. Only ROUND_ROBIN and RING_HASH are
supported for proxyless gRPC. Mutable.

- rule: locality_lb_policy must be one of ROUND_ROBIN, LEAST_REQUEST, RING_HASH, RANDOM, ORIGINAL_DESTINATION, MAGLEV, WEIGHTED_MAGLEV, or WEIGHTED_ROUND_ROBIN

### spec.localityLbPolicies

`[]GcpBackendServiceLocalityLbPolicyConfig`

Ordered list of locality LB policies for Traffic Director deployments
that need a custom (xDS-configured) policy with built-in fallbacks.
Each entry is either a built-in policy name or a custom policy plus its
opaque configuration; Traffic Director uses the first one it supports.
Overrides locality_lb_policy when set. Mutable.

### spec.localityLbPolicies[].policy

`GcpBackendServiceLocalityLbPolicy`

A built-in locality policy by name. The WEIGHTED_* policies are not
valid inside this list — use the top-level locality_lb_policy for
those.

### spec.localityLbPolicies[].policy.name

`string` · required

The built-in policy name.

- rule: policy name must be one of ROUND_ROBIN, LEAST_REQUEST, RING_HASH, RANDOM, ORIGINAL_DESTINATION, or MAGLEV
- rule: {"required":true}

### spec.localityLbPolicies[].customPolicy

`GcpBackendServiceLocalityLbCustomPolicy`

A custom policy implemented in the xDS client (Envoy/gRPC), selected
by name with an opaque configuration string. Traffic Director falls
back to the next entry if the client does not recognize it.

### spec.localityLbPolicies[].customPolicy.name

`string` · required

Identifier of the custom policy as registered in the xDS client (e.g.
an Envoy load balancing extension name).

- rule: {"required":true,"string":{"maxLen":"256"}}

### spec.localityLbPolicies[].customPolicy.data

`string`

Opaque configuration handed to the custom policy, in whatever format
the policy implementation expects (commonly JSON).

- rule: {"string":{"maxLen":"4096"}}

### spec.consistentHash

`GcpBackendServiceConsistentHash`

Parameters for consistent-hash load balancing — soft session affinity
where a backend's share of the hash ring survives other backends
joining or leaving. Only applies with load_balancing_scheme
INTERNAL_SELF_MANAGED and locality_lb_policy MAGLEV or RING_HASH.

### spec.consistentHash.httpCookie

`GcpBackendServiceConsistentHashHttpCookie`

Hash on an HTTP cookie, generating it when absent — soft session
affinity for clients that keep cookies. Only applies when
session_affinity is HTTP_COOKIE.

### spec.consistentHash.httpCookie.name

`string`

Cookie name to hash on (and to generate when absent).

- rule: {"string":{"maxLen":"256"}}

### spec.consistentHash.httpCookie.path

`string`

Path attribute set when the cookie is generated.

- rule: {"string":{"maxLen":"1024"}}

### spec.consistentHash.httpCookie.ttl

`GcpBackendServiceDuration`

Lifetime of the generated cookie. Zero/unset makes it a session
cookie.

### spec.consistentHash.httpCookie.ttl.seconds

`int64`

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.consistentHash.httpCookie.ttl.nanos

`int32`

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.consistentHash.httpHeaderName

`string`

Hash on the value of this request header. Only applies when
session_affinity is HEADER_FIELD.

- rule: {"string":{"maxLen":"256"}}

### spec.consistentHash.minimumRingSize

`int64` · optional (explicit presence)

Minimum number of virtual nodes on the hash ring (default 1024).
Larger rings spread load more evenly across backends at slightly
higher memory cost; must be at least the number of backend hosts.

- default: `1024`
- rule: {"int64":{"gt":"0"}}

### spec.enableCdn

`bool`

Cache responses at Google's edge with Cloud CDN. Off by default:
without it every request is proxied to a backend. Only valid on
external schemes (EXTERNAL, EXTERNAL_MANAGED) — Cloud CDN does not
front internal load balancers. Turning it on activates cdn_policy (or
sensible CDN defaults when cdn_policy is omitted). Mutable.

### spec.cdnPolicy

`GcpBackendServiceCdnPolicy`

How Cloud CDN caches responses from these backends. Only meaningful
with enable_cdn — GCP ignores the policy while CDN is off.

- rule: with cache_mode USE_ORIGIN_HEADERS the origin's headers control lifetimes — remove client_ttl, default_ttl, and max_ttl (GCP would silently ignore them)
- rule: with cache_mode FORCE_CACHE_ALL every response is cached for default_ttl — remove max_ttl (GCP would silently ignore it)

### spec.cdnPolicy.cacheMode

`string`

What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
content types and honors origin cache headers for the rest;
USE_ORIGIN_HEADERS caches only what the backends explicitly mark
cacheable (TTL fields must be unset — the origin controls lifetimes);
FORCE_CACHE_ALL caches everything, ignoring origin headers (never
combine with private or per-user content; max_ttl must be unset).

- rule: cache_mode must be CACHE_ALL_STATIC, USE_ORIGIN_HEADERS, or FORCE_CACHE_ALL

### spec.cdnPolicy.clientTtl

`int32`

Seconds a response may be cached by browsers and other downstream
caches (sets the max-age clients see; GCP default 3600, max 86400).
Keep it shorter than default_ttl so edge caches revalidate before
clients do.

- rule: client_ttl must be between 0 and 86400 seconds (1 day)

### spec.cdnPolicy.defaultTtl

`int32`

Seconds the edge caches a response when the origin sets no caching
headers (GCP default 3600, max 31622400 = 1 year). The workhorse TTL
for CACHE_ALL_STATIC and FORCE_CACHE_ALL.

- rule: default_ttl must be between 0 and 31622400 seconds (1 year)

### spec.cdnPolicy.maxTtl

`int32`

Upper bound in seconds on any cache lifetime, capping even origin
headers that ask for longer (GCP default 86400, max 31622400). Not
allowed with USE_ORIGIN_HEADERS or FORCE_CACHE_ALL cache modes.

- rule: max_ttl must be between 0 and 31622400 seconds (1 year)

### spec.cdnPolicy.negativeCaching

`bool`

Cache error responses (404s, redirects) at the edge so failing paths
do not hammer the backends. Pair with negative_caching_policy to set
per-status TTLs; without it GCP applies default lifetimes.

### spec.cdnPolicy.negativeCachingPolicy

`[]GcpBackendServiceNegativeCachingPolicy`

Per-status-code TTLs for negative caching. Only effective with
negative_caching enabled. Codes limited by GCP to 300, 301, 308, 404,
405, 410, 421, 451, and 501.

### spec.cdnPolicy.negativeCachingPolicy[].code

`int32` · required

The HTTP status code to cache. GCP supports 300, 301, 308, 404, 405,
410, 421, 451, and 501.

- rule: code must be one of 300, 301, 308, 404, 405, 410, 421, 451, or 501 — the status codes Cloud CDN can negative-cache
- rule: {"required":true}

### spec.cdnPolicy.negativeCachingPolicy[].ttl

`int32`

Seconds responses with this status are cached at the edge
(0 to 1800 = 30 minutes).

- rule: ttl must be between 0 and 1800 seconds (30 minutes)

### spec.cdnPolicy.serveWhileStale

`int32`

Seconds the edge may keep serving a stale response while it
revalidates with the origin in the background (max 86400; 0 disables).
Smooths over brief backend outages for content that tolerates slight
staleness.

- rule: serve_while_stale must be between 0 and 86400 seconds (1 day)

### spec.cdnPolicy.requestCoalescing

`bool`

Collapse concurrent cache-miss requests for the same object into one
origin fetch. Protects the backends from thundering herds on cache
expiry of popular objects.

### spec.cdnPolicy.signedUrlCacheMaxAgeSec

`int32` · optional (explicit presence)

Seconds a response to a SIGNED request stays fresh in the cache before
revalidation (GCP default 3600, max 86400). Only meaningful with
signed URLs or cookies; the signature's own expiry still governs
access.

- default: `3600`
- rule: {"int32":{"lte":86400,"gte":0}}

### spec.cdnPolicy.cacheKeyPolicy

`GcpBackendServiceCdnCacheKeyPolicy`

What forms the cache key beyond the URL. The backend-service flavor is
richer than a backend bucket's: host, protocol, query handling, named
cookies, and headers can all join or leave the key. Leave unset for
GCP's default (host + protocol + full query string).

- rule: query_string_whitelist and query_string_blacklist are mutually exclusive — narrow the cache key with one or the other
- rule: query_string_whitelist/blacklist only apply when include_query_string is true — with the query string excluded there is nothing to filter

### spec.cdnPolicy.cacheKeyPolicy.includeHost

`bool`

Include the request host in the cache key. GCP's default is true —
turn it off only when several hosts genuinely serve identical content.

### spec.cdnPolicy.cacheKeyPolicy.includeProtocol

`bool`

Include the protocol (http/https) in the cache key. GCP's default is
true — turn it off only when both schemes serve identical bytes.

### spec.cdnPolicy.cacheKeyPolicy.includeQueryString

`bool`

Include the query string in the cache key. GCP's default is true.
When true, narrow it with query_string_whitelist or blacklist; when
false, the query string is ignored entirely (and the lists must be
unset).

### spec.cdnPolicy.cacheKeyPolicy.queryStringWhitelist

`[]string`

Query parameters included in the cache key, all others ignored.
Include only parameters that genuinely change the response so
equivalent requests share a cache entry. Mutually exclusive with
query_string_blacklist.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.cdnPolicy.cacheKeyPolicy.queryStringBlacklist

`[]string`

Query parameters excluded from the cache key, all others included —
for stripping tracking parameters (utm_*) that never change the
response. Mutually exclusive with query_string_whitelist.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.cdnPolicy.cacheKeyPolicy.includeHttpHeaders

`[]string`

Request headers whose values join the cache key — for backends that
vary responses by header (e.g. Accept for image format negotiation).
Each distinct value creates a separate cache entry, so keep this list
short.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.cdnPolicy.cacheKeyPolicy.includeNamedCookies

`[]string`

Cookie names whose values join the cache key — for backends that vary
cached content by cookie (e.g. an A/B bucket cookie). Each distinct
value creates a separate cache entry.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.cdnPolicy.bypassCacheOnRequestHeaders

`[]GcpBackendServiceBypassCacheOnRequestHeader`

Skip the cache entirely for requests carrying any of these headers
(at most 5) — an escape hatch for debugging or per-request freshness
(e.g. a Pragma: no-cache internal tooling header).

- rule: {"repeated":{"maxItems":"5"}}

### spec.cdnPolicy.bypassCacheOnRequestHeaders[].headerName

`string` · required

The header name to match (case-insensitive); any value triggers the
bypass.

- rule: {"required":true}

### spec.securityPolicy

`string | valueFrom`

Cloud Armor security policy evaluated on every request AFTER the CDN
cache (protects the backends: WAF rules, rate limiting, geo/IP
blocking). Reference a GcpCloudArmorPolicy of type CLOUD_ARMOR —
edge policies are not valid here. Mutable.

- references: GcpCloudArmorPolicy (`status.outputs.policy_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudArmorPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_self_link}} -- a bare string does not parse

### spec.edgeSecurityPolicy

`string | valueFrom`

Cloud Armor EDGE security policy filtering requests BEFORE the CDN
cache (protects cached content: geo/IP blocking at the edge).
Reference a GcpCloudArmorPolicy of type CLOUD_ARMOR_EDGE — standard
backend policies are not valid here. Mutable.

- references: GcpCloudArmorPolicy (`status.outputs.policy_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudArmorPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_self_link}} -- a bare string does not parse

### spec.iap

`GcpBackendServiceIap`

Identity-Aware Proxy: authenticate every request against Google
identities before it reaches the backends — zero-trust access to
internal tools without a VPN. Requests arrive with IAP assertion
headers the backend can trust. HTTPS frontends only.

- rule: oauth2_client_id and oauth2_client_secret must be set together — or both left empty to use the Google-managed OAuth client

### spec.iap.enabled

`bool`

Turn IAP enforcement on. When enabled, every request must carry a
valid Google identity; unauthenticated requests get a login redirect.

### spec.iap.oauth2ClientId

`string`

OAuth2 client ID of a custom IAP client. Leave both id and secret
empty to use the Google-managed OAuth client.

- rule: {"string":{"maxLen":"256"}}

### spec.iap.oauth2ClientSecret

`string` · sensitive

OAuth2 client secret paired with oauth2_client_id. Handled as a
secret: never stored in plaintext in the control plane, never exposed
in outputs (GCP itself only ever returns its SHA-256 after creation).

- rule: {"string":{"maxLen":"256"}}

### spec.logConfig

`GcpBackendServiceLogConfig`

Request logging to Cloud Logging for this backend service. Off by
default. Sampling keeps log volume (and cost) proportional on
high-traffic services.

- rule: optional_fields only applies with optional_mode CUSTOM
- rule: optional_mode configures log entries and only applies with enable true
- rule: request_headers and response_headers name headers to log and only apply with enable true

### spec.logConfig.enable

`bool`

Write request logs to Cloud Logging. Off by default.

### spec.logConfig.sampleRate

`double` · optional (explicit presence)

Fraction of requests logged, 0.0-1.0 (GCP default 1.0 = everything).
Sample aggressively on high-QPS services — full logging is a real
cost line.

- default: `1.0`
- rule: {"double":{"lte":1,"gte":0}}

### spec.logConfig.optionalMode

`string`

Which optional fields join each log entry: INCLUDE_ALL_OPTIONAL,
EXCLUDE_ALL_OPTIONAL (the GCP default), or CUSTOM (name them in
optional_fields).

- rule: optional_mode must be INCLUDE_ALL_OPTIONAL, EXCLUDE_ALL_OPTIONAL, or CUSTOM

### spec.logConfig.optionalFields

`[]string`

Names of the optional log fields to include with optional_mode CUSTOM
(e.g. tls.protocol, orca_load_report).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.logConfig.requestHeaders

`[]string`

HTTP request headers whose values join each log entry (e.g.
"X-Request-Id", "User-Agent") — for tracing a request across services
without instrumenting the backend. Requires enable and an HTTP-family
protocol (HTTP, HTTPS, HTTP2, GRPC). Header names are case-insensitive
in HTTP; each entry is one name.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.logConfig.responseHeaders

`[]string`

HTTP response headers whose values join each log entry (e.g.
"Content-Type", a backend's own "X-Cache" or "X-Served-By"). Same
preconditions as request_headers.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.customRequestHeaders

`[]string`

Headers the load balancer ADDS to requests before forwarding them to
the backends, in "Header-Name: value" form. Values may use variables
like {client_ip} or {tls_version}. Typical uses: passing the client's
geo data or TLS parameters to the application. Mutable.

- rule: {"repeated":{"maxItems":"25","items":{"string":{"pattern":"^[^:]+:.*$"}}}}

### spec.customResponseHeaders

`[]string`

Headers the load balancer ADDS to responses before returning them to
clients, in "Header-Name: value" form. Values may use variables like
{cdn_cache_status}. Typical uses: security headers
(Strict-Transport-Security) and cache observability. Mutable.

- rule: {"repeated":{"maxItems":"25","items":{"string":{"pattern":"^[^:]+:.*$"}}}}

### spec.compressionMode

`string`

Whether the load balancer compresses responses (gzip/brotli) for
clients that ask for it. AUTOMATIC compresses compressible content
types; DISABLED (the GCP default when unset) never compresses.
Compression is applied by the load balancer — backends keep serving
uncompressed responses. Mutable.

- rule: compression_mode must be AUTOMATIC or DISABLED

### spec.circuitBreakers

`GcpBackendServiceCircuitBreakers`

Connection-volume limits protecting backends from overload — the
service-mesh circuit breaker. Only applies with load_balancing_scheme
INTERNAL_SELF_MANAGED (Traffic Director).

### spec.circuitBreakers.maxConnections

`int32` · optional (explicit presence)

Max concurrent connections to the whole backend service (GCP default
1024).

- default: `1024`
- rule: {"int32":{"gt":0}}

### spec.circuitBreakers.maxPendingRequests

`int32` · optional (explicit presence)

Max requests queued waiting for a connection (GCP default 1024).

- default: `1024`
- rule: {"int32":{"gt":0}}

### spec.circuitBreakers.maxRequests

`int32` · optional (explicit presence)

Max concurrent requests to the whole backend service (GCP default
1024).

- default: `1024`
- rule: {"int32":{"gt":0}}

### spec.circuitBreakers.maxRequestsPerConnection

`int32`

Max requests per connection — 1 disables HTTP keep-alive. Unset means
unlimited.

- rule: {"int32":{"gte":0}}

### spec.circuitBreakers.maxRetries

`int32` · optional (explicit presence)

Max concurrent retries across the backend service (GCP default 3).
Retries amplify load during incidents — keep this bounded.

- default: `3`
- rule: {"int32":{"gt":0}}

### spec.outlierDetection

`GcpBackendServiceOutlierDetection`

Passive health checking: eject backends that keep erroring from the
load balancing pool for a cooling-off period, without waiting for the
active health check to fail. Only applies with load_balancing_scheme
INTERNAL_SELF_MANAGED or EXTERNAL_MANAGED.

### spec.outlierDetection.baseEjectionTime

`GcpBackendServiceDuration`

Base duration a host stays ejected; actual ejection time is this
multiplied by the number of times the host has been ejected (GCP
default 30s).

### spec.outlierDetection.baseEjectionTime.seconds

`int64`

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.outlierDetection.baseEjectionTime.nanos

`int32`

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.outlierDetection.consecutiveErrors

`int32`

Consecutive 5xx responses (or connection errors) before ejection (GCP
default 5).

- rule: {"int32":{"gte":0}}

### spec.outlierDetection.consecutiveGatewayFailure

`int32`

Consecutive gateway-class failures (502/503/504) before ejection (GCP
default 5). Catches infrastructure failures faster than
consecutive_errors on mixed error streams.

- rule: {"int32":{"gte":0}}

### spec.outlierDetection.enforcingConsecutiveErrors

`int32`

Percentage chance (0-100) that a host is ACTUALLY ejected when
consecutive_errors trips (GCP default 100). Lower values ease the
policy in gradually.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.outlierDetection.enforcingConsecutiveGatewayFailure

`int32`

Percentage chance (0-100) of ejection when consecutive_gateway_failure
trips (GCP default 0 — off unless raised).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.outlierDetection.enforcingSuccessRate

`int32`

Percentage chance (0-100) of ejection when a host's success rate falls
statistically below the pool (GCP default 100).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.outlierDetection.interval

`GcpBackendServiceDuration`

How often ejection sweeps run (GCP default 1s).

### spec.outlierDetection.interval.seconds

`int64`

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.outlierDetection.interval.nanos

`int32`

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.outlierDetection.maxEjectionPercent

`int32`

Max percentage (0-100) of the pool that may be ejected at once (GCP
default 10) — the safety valve that keeps outlier detection from
draining the whole service.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.outlierDetection.successRateMinimumHosts

`int32`

Minimum number of hosts in the pool before success-rate ejection
activates (GCP default 5) — below it the statistics are meaningless.

- rule: {"int32":{"gte":0}}

### spec.outlierDetection.successRateRequestVolume

`int32`

Minimum requests a host must have received in the interval for its
success rate to count (GCP default 100).

- rule: {"int32":{"gte":0}}

### spec.outlierDetection.successRateStdevFactor

`int32`

How many standard deviations below the pool mean a host's success
rate must fall to be ejected, multiplied by 1000 (GCP default 1900 =
1.9 stdev). Lower is more aggressive.

- rule: {"int32":{"gte":0}}

### spec.maxStreamDuration

`GcpBackendServiceDuration`

Default maximum duration for streams to this service, computed from
stream start until the response is completely processed (including
retries). Unset means no timeout limit. Can be overridden per-route in
the URL map. Only allowed with load_balancing_scheme
INTERNAL_SELF_MANAGED.

### spec.maxStreamDuration.seconds

`int64`

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.maxStreamDuration.nanos

`int32`

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.securitySettings

`GcpBackendServiceSecuritySettings`

Backend authentication and TLS settings for Traffic Director
(client TLS policy, SAN validation) and for AWS-hosted internet-NEG
origins (Signature Version 4 request signing).

### spec.securitySettings.clientTlsPolicy

`string`

Self-link of a networksecurity ClientTlsPolicy describing how the
load balancer authenticates itself to the backends (mTLS). Traffic
Director only. Plain URL — the policy is a Network Security resource
outside the compute family.

- rule: {"string":{"maxLen":"2048"}}

### spec.securitySettings.subjectAltNames

`[]string`

Subject Alternative Names the backend's server certificate must
present — pins backend identity for Traffic Director mTLS.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.securitySettings.awsV4Authentication

`GcpBackendServiceAwsV4Authentication`

Sign origin requests with AWS Signature Version 4 — for internet-NEG
backends fronting private S3 buckets or other SigV4-authenticated AWS
origins.

### spec.securitySettings.awsV4Authentication.accessKeyId

`string`

AWS access key ID — the username-like identifier of the key pair (not
itself a secret).

- rule: {"string":{"maxLen":"256"}}

### spec.securitySettings.awsV4Authentication.accessKey

`string` · sensitive

AWS secret access key paired with access_key_id. Handled as a secret:
never stored in plaintext in the control plane, and GCP never returns
it on reads.

- rule: {"string":{"maxLen":"256"}}

### spec.securitySettings.awsV4Authentication.accessKeyVersion

`string`

Optional version identifier for the key, echoed in logs to trace
which credential signed a request during rotation.

- rule: {"string":{"maxLen":"256"}}

### spec.securitySettings.awsV4Authentication.originRegion

`string`

AWS region of the origin (e.g. us-east-1) — part of the SigV4 signing
scope.

- rule: {"string":{"maxLen":"64"}}

### spec.tlsSettings

`GcpBackendServiceTlsSettings`

TLS parameters for the load balancer's connections TO the backends:
which server certificate authentication to apply and what SNI to send.
Only valid when protocol is SSL, HTTPS, or HTTP2.

### spec.tlsSettings.authenticationConfig

`string`

Self-link of a networksecurity BackendAuthenticationConfig that
validates the backend's server certificate (trust anchor + client
cert). Plain URL — the config is a Network Security resource outside
the compute family.

- rule: {"string":{"maxLen":"2048"}}

### spec.tlsSettings.sni

`string`

Server Name Indication sent in the TLS handshake to the backends —
for origins that route or select certificates by SNI.

- rule: {"string":{"maxLen":"253"}}

### spec.tlsSettings.subjectAltNames

`[]GcpBackendServiceTlsSubjectAltName`

Subject Alternative Names the backend certificate must match, each a
DNS name or a URI. GCP allows at most 5.

- rule: {"repeated":{"maxItems":"5"}}

### spec.tlsSettings.subjectAltNames[].dnsName

`string`

A DNS-name SAN (e.g. origin.example.com).

### spec.tlsSettings.subjectAltNames[].uniformResourceIdentifier

`string`

A URI SAN (e.g. spiffe://cluster/ns/prod/sa/web).

### spec.ipAddressSelectionPolicy

`string`

Whether the load balancer prefers IPv4 or IPv6 addresses when
connecting to dual-stack backends. Unset uses GCP's default (IPv4).
Mutable.

- rule: ip_address_selection_policy must be one of IPV4_ONLY, PREFER_IPV6, or IPV6_ONLY

### spec.externalManagedMigrationState

`string`

Canary state for migrating this backend service from the classic
EXTERNAL scheme to EXTERNAL_MANAGED without recreating it: PREPARE
first, then optionally TEST_BY_PERCENTAGE, then TEST_ALL_TRAFFIC —
after which load_balancing_scheme can be flipped to EXTERNAL_MANAGED.
Only meaningful while the scheme is still EXTERNAL. Mutable.

- rule: external_managed_migration_state must be one of PREPARE, TEST_BY_PERCENTAGE, or TEST_ALL_TRAFFIC

### spec.externalManagedMigrationTestingPercentage

`double`

Fraction of traffic (0-100) sent to the envoy-based global external
ALB during a TEST_BY_PERCENTAGE canary migration. Only meaningful with
external_managed_migration_state TEST_BY_PERCENTAGE. Mutable.

- rule: {"double":{"lte":100,"gte":0}}

### spec.customMetrics

`[]GcpBackendServiceCustomMetric`

Custom metrics the WEIGHTED_ROUND_ROBIN locality policy balances on,
reported by the backends via the Open Request Cost Aggregation (ORCA)
protocol. Only meaningful with locality_lb_policy
WEIGHTED_ROUND_ROBIN.

### spec.customMetrics[].name

`string` · required

Metric name as reported by the backends in ORCA load reports.

- rule: {"required":true,"string":{"maxLen":"256"}}

### spec.customMetrics[].dryRun

`bool`

Report the metric without acting on it — the safe first step while
validating that backends emit sane values.

### spec.serviceLbPolicy

`string`

Self-link of a networkservices ServiceLbPolicy attaching advanced
traffic-distribution features (e.g. auto-capacity failover) to this
backend service. Plain URL — the service LB policy is a Network
Services resource outside the compute family. Global backend services
only. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.signedUrlKeys

`[]GcpBackendServiceSignedUrlKey`

Keys for signing Cloud CDN signed URLs and signed cookies — the
mechanism for serving private content from the cache with expiring,
tamper-proof links. GCP allows at most 3 keys per backend service so
one can be rotated while another stays live. Each key's material is a
secret; rotate by adding a new key, re-signing URLs, then removing the
old one.

- rule: {"repeated":{"maxItems":"3"}}

### spec.signedUrlKeys[].name

`string` · required

Name of the key, referenced by the key_name parameter of signed URLs.
Must be 1-63 characters: lowercase letters, digits, and hyphens; must
start with a letter and end with a letter or digit. Immutable:
renaming replaces the key, invalidating URLs signed with the old name.

- rule: {"required":true,"string":{"pattern":"^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.signedUrlKeys[].keyValue

`string` · required · sensitive

The 128-bit signing key, base64url-encoded (RFC 4648 §5) — generate
one with: head -c 16 /dev/urandom | base64 | tr '+/' '-_'. 22
characters of base64url, with or without the trailing == padding.
Anyone holding this value can mint valid signed URLs, so it is
handled as a secret. Immutable per key name: rotating means adding a
new key and removing the old. The base64url shape is taught here
rather than enforced by a validation rule, because sensitive fields
hold a managed-secret reference on consuming platforms and a
content-shape rule would reject every reference.

- rule: {"required":true}

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the backend service for org-policy and
IAM conditions. Keys in the form "tagKeys/{id}", values
"tagValues/{id}". Create-time only: changing them later replaces the
backend service.

### spec.deletionPolicy

`string`

Deletion policy for the backend service AND its signed-URL keys — one
switch governs both objects this kind manages:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- both are deleted (GCP refuses to delete the backend
               service while a URL map or forwarding rule still
               references it); the backends it pointed at — instance
               groups, NEGs — are untouched
  "PREVENT" -- destroy FAILS; protects the routing target of a live
               load balancer
  "ABANDON" -- both are removed from management but keep serving in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `cdn_requires_external_scheme`: Cloud CDN can only be enabled on external backend services (scheme EXTERNAL or EXTERNAL_MANAGED) — it does not front internal load balancers
- `circuit_breakers_scheme`: circuit_breakers only applies to Traffic Director backend services — set load_balancing_scheme INTERNAL_SELF_MANAGED or remove it
- `max_stream_duration_scheme`: max_stream_duration only applies to Traffic Director backend services — set load_balancing_scheme INTERNAL_SELF_MANAGED or remove it
- `outlier_detection_scheme`: outlier_detection only applies with load_balancing_scheme INTERNAL_SELF_MANAGED or EXTERNAL_MANAGED
- `consistent_hash_coherence`: consistent_hash requires load_balancing_scheme INTERNAL_SELF_MANAGED and locality_lb_policy MAGLEV or RING_HASH
- `strong_cookie_affinity_coherence`: strong_session_affinity_cookie customizes the cookie of session_affinity STRONG_COOKIE_AFFINITY and is not valid with any other affinity mode
- `affinity_cookie_ttl_requires_generated_cookie`: affinity_cookie_ttl_sec only applies with session_affinity GENERATED_COOKIE
- `udp_forbids_session_affinity`: session affinity is not applicable when protocol is UDP
- `tls_settings_protocol`: tls_settings only applies when protocol is SSL, HTTPS, or HTTP2 — the load balancer must speak TLS to the backends
- `custom_metrics_locality_policy`: top-level custom_metrics only apply with locality_lb_policy WEIGHTED_ROUND_ROBIN
- `migration_percentage_requires_state`: external_managed_migration_testing_percentage only applies with external_managed_migration_state TEST_BY_PERCENTAGE
- `migration_requires_external_scheme`: external_managed_migration_state drives the EXTERNAL → EXTERNAL_MANAGED canary and only applies while load_balancing_scheme is EXTERNAL (or unset, which defaults to EXTERNAL)
- `backend_preference_not_external`: backend preference cannot be set when load_balancing_scheme is EXTERNAL (the default) — use EXTERNAL_MANAGED or an internal scheme

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBackendService, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the backend service. This is the value URL maps reference as a default service or path-rule target — the composition handle for routing traffic to these backends. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/backendServices/{name} |
| `status.outputs.backend_service_name` | `string` | Name of the backend service as it exists in GCP. |
| `status.outputs.generated_id` | `string` | Server-assigned numeric ID of the backend service. |
| `status.outputs.fingerprint` | `string` | Server-computed fingerprint of the backend service. Used for optimistic concurrency control when updating the service outside of IaC. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.healthCheck` | GcpHealthCheck | `status.outputs.self_link` |
| `spec.backends[].group` | GcpRegionNetworkEndpointGroup | `status.outputs.self_link` |
| `spec.securityPolicy` | GcpCloudArmorPolicy | `status.outputs.policy_self_link` |
| `spec.edgeSecurityPolicy` | GcpCloudArmorPolicy | `status.outputs.policy_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpUrlMap | `spec.defaultRouteAction.weightedBackendServices[].backendService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.defaultRouteAction.requestMirrorPolicy.backendService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].backendService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].defaultRouteAction.requestMirrorPolicy.backendService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].backendService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].pathRules[].routeAction.requestMirrorPolicy.backendService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].routeRules[].service` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].backendService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].routeRules[].routeAction.requestMirrorPolicy.backendService` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)

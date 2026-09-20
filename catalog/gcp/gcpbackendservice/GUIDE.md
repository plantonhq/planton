# GcpBackendService Guide

The judgment this guide protects: the backend service is the routing
target every other LB piece exists to reach, and its power dials
(balancing modes, affinity, draining) interact — change one at a time,
watch the metrics, then the next.

## One kind, two scopes

`region` empty builds the GLOBAL backend service — the backend of the
global external ALB, the cross-region internal ALB, and Traffic Director.
`region` set builds the REGIONAL one — the backend of the regional external
ALB (`EXTERNAL_MANAGED`), the regional internal ALB (`INTERNAL_MANAGED`),
and the internal and external passthrough Network Load Balancers
(`INTERNAL`, `EXTERNAL`), which a regional `GcpGlobalForwardingRule` names
directly through `backendService` with no proxy in between. The regional
service adds the passthrough levers — `network`, `backends[].failover` with
`failoverPolicy`, `connectionTrackingPolicy`, `haPolicy`, and zonal
affinity — and lacks the global edge features: compression, custom headers,
`edgeSecurityPolicy`, `serviceLbPolicy`, the migration canary, backend
`preference`, `localityLbPolicies`, `maxStreamDuration`, `securitySettings`,
`signedUrlKeys`, and Cloud CDN's advanced knobs (`requestCoalescing`,
header-driven bypass, header-keyed cache keys, per-code negative TTLs). The
spec rejects each on the wrong scope. Two things to know: the regional API
carries `enableCdn` and `cdnPolicy` although regional ALBs have no Cloud
CDN — leave them off; and an unset `loadBalancingScheme` is `EXTERNAL` on
both scopes (both engines send it), so name `INTERNAL` outright for an
internal passthrough NLB. A regional ALB needs a regional `GcpHealthCheck`;
a regional service attaches only a regional Cloud Armor policy. `region`
is immutable.

## Scheme first, everything else second

`loadBalancingScheme` is immutable and decides which protocols, affinity
modes, locality policies, and extras are even legal — the spec's pairing
CELs encode those rules so violations fail pre-deploy instead of at the
API. Getting the scheme wrong means a recreate; getting anything else
wrong is usually an in-place fix.

## Balancing mode is per backend, capacity is the throttle

Each backend group carries its own `balancingMode` (instance groups
default UTILIZATION, NEGs must use RATE; CUSTOM_METRICS and IN_FLIGHT
serve ORCA-reporting and queue-depth backends). `capacityScaler` is the
operational lever: 0 drains a backend without removing it — use that for
maintenance, never deletion, so session affinity and health history
survive.

## Health check is singular by design

GCP accepts at most ONE health check per backend service; the spec
models the reference singular where the provider inherits the API's list
shape. Share one health check across many services — it is its own
composable node, which is exactly why it is not created here.

## Logging headers for request tracing

`logConfig` enables per-request logs with sampling. `requestHeaders` and
`responseHeaders` name the HTTP headers whose values join each log entry
(for example `X-Request-Id` to trace a request across services without
instrumenting the backend, or a backend's own `X-Cache` on the response).
Both need `enable: true` and an HTTP-family protocol (HTTP, HTTPS, HTTP2,
GRPC); each entry is one header name.

## Signed-URL key rotation is add-then-remove

Same contract as the backend bucket: at most 3 keys, each immutable, so
rotation is add new → re-sign → remove old. Key material is secret in
both engines' state and never surfaces in outputs.

## Teardown discipline

One `deletionPolicy` governs the backend service AND its signed-URL
keys. GCP refuses to delete a backend service a URL map or forwarding
rule still references, so `DELETE` fails safely mid-chain; `PREVENT`
also covers the window after those are gone. `ABANDON` leaves the
service (and the backends it points at) serving unmanaged — the backends
themselves always belong to their own kinds.

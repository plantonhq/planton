# GcpBackendService Guide

The judgment this guide protects: the backend service is the routing
target every other LB piece exists to reach, and its power dials
(balancing modes, affinity, draining) interact — change one at a time,
watch the metrics, then the next.

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

## Logging headers wait on the SDK

`logConfig` enables per-request logs with sampling; the newer
request/response header capture lists are recorded exclusions — the
Pulumi bridge at the pinned SDK does not carry them, and this catalog
never models intent only one engine honors (re-evaluate at the next
pulumi-gcp bump; the reason lives in the parity manifest).

## Signed-URL key rotation is add-then-remove

Same contract as the backend bucket: at most 3 keys, each immutable, so
rotation is add new → re-sign → remove old. Key material is secret in
Pulumi state and sensitive in OpenTofu's (whose state file is encrypted
whole when given a key), and never surfaces in outputs.

## Teardown discipline

One `deletionPolicy` governs the backend service AND its signed-URL
keys. GCP refuses to delete a backend service a URL map or forwarding
rule still references, so `DELETE` fails safely mid-chain; `PREVENT`
also covers the window after those are gone. `ABANDON` leaves the
service (and the backends it points at) serving unmanaged — the backends
themselves always belong to their own kinds.

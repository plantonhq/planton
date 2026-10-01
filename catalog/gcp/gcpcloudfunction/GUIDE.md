# GcpCloudFunction Guide

The judgment this guide protects: a Gen 2 function is a Cloud Run
service with a build pipeline attached. The serving decisions (identity,
networking, scaling) outlive any single deploy — get those right and
redeploys stay boring.

## Direct VPC vs connector: pick by operational appetite

Direct VPC egress (`directVpcNetworkInterface`) attaches instances
straight to a subnet — no connector to size, pay for, or saturate. The
trade: instances draw IPs from the subnet, so the subnet's free range is
the function's real scaling ceiling; size it for `maxInstanceCount`,
not for today's traffic. The Serverless VPC Access connector remains the
right choice when many functions share one egress point or the subnet
is not yours to consume. The two paths are mutually exclusive
(enforced pre-deploy), and the egress mode (`directVpcEgress:
ALL_TRAFFIC`) is what enables static egress IPs via Cloud NAT.

One teardown consequence to plan around: a direct-VPC function leaves a
`serverless-ipv4-*` address reservation in its subnetwork after the
function is deleted. The reservation belongs to Google's serverless
service agent — you cannot delete it — and Google garbage-collects it on
an hours scale (live-verified: it survives the service delete and blocks
subnet deletion). Deleting a direct-VPC function therefore does NOT free
its subnet immediately; plan subnet/VPC decommissioning a few hours
behind the function's, or keep direct-VPC functions on subnets with an
independent lifecycle.

## Identity: one function, one service account

The default compute service account is broad by default; give each
production function its own `serviceAccountEmail` holding exactly what
the code touches. Secrets go in `secretEnvironmentVariables` or `secretVolumes`, never in `environmentVariables`. A `secretEnvironmentVariables` entry either names a secret you already manage or carries a `value`. For a `value`, the module creates one Secret Manager secret per variable, stores the value as a pinned version, and grants the runtime identity `secretmanager.secretAccessor` on that secret alone. The function never sees the value as plain text configuration. For a named secret, granting access is yours.

## Cold starts are a spend decision

`scaling.minInstanceCount: 1` eliminates cold starts for the price of
one always-on instance — right for latency-sensitive HTTP endpoints,
waste for event handlers that tolerate a second of warmup. Concurrency
above 1 cuts instance count for I/O-bound code but requires ≥ 1 CPU and
thread-safe code.

## Event functions: idempotency is not optional

`retryPolicy: RETRY_POLICY_RETRY` is at-least-once delivery — the
handler WILL eventually see duplicates. Write the handler idempotent
first, then enable retry; the reverse order pages you at 3am with a
double-charged customer.

## Public invocation is a one-line blast radius

`allowUnauthenticated: true` grants `run.invoker` to `allUsers` on the
underlying Cloud Run service. It is the right line for a public webhook
and the wrong default for everything else — prefer granting invoker to
the specific caller identity (Scheduler SA, Pub/Sub push SA, another
service's SA).

## Destroy semantics

`deletionPolicy: DELETE` (default) removes the function and its
triggers stop firing; queued events are the source system's problem, so
drain or pause producers first. `PREVENT` suits functions other systems
invoke by URL. `ABANDON` keeps the function serving AND consuming
events unmanaged — someone must still own its errors.

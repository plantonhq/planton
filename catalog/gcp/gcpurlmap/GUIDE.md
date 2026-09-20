# GcpUrlMap Guide

The judgment this guide protects: the URL map is the one resource where a
routing mistake is instantly global — every host, every path, every client.
Its levers are all MUTABLE (in-place, zero-downtime), which is exactly why
routing self-tests and a deliberate destroy stance matter more here than on
any resource behind it.

## One kind, two scopes

`region` empty builds the GLOBAL URL map — the routing brain of the global
external ALB, the cross-region internal ALB, and Traffic Director. `region`
set builds the REGIONAL map — the routing brain of the regional external
ALB and the regional internal ALB. A regional map routes only to regional
`GcpBackendService` blocks in its own region (never to a backend bucket,
which is global-only) and is referenced only by regional target proxies.
Four surfaces exist only on the global map and are rejected when `region`
is set: route-scoped `cachePolicy` (regional ALBs have no Cloud CDN),
custom error response policies, `maxStreamDuration` (except in a path
matcher's default route action, which both scopes carry), and the
header-driven routing-test fields (`headers`, `expectedOutputUrl`,
`expectedRedirectResponseCode` — a regional test names its `service` and
nothing else). One knob exists only on the regional map:
`pathTemplateRewrite` in a path matcher's default route action. `region` is
immutable — a map never moves between scopes.

## Route evaluation order

Host rules pick a path matcher; inside the matcher, route rules run first
(by ascending priority number), then path rules (longest prefix), then the
matcher's default; traffic matching no host rule falls to the URL map's
top-level default. Two consequences: a route rule can shadow a path rule
you thought was live, and the top-level default is the only route many
requests ever see — keep it boring, and put experiments behind a matcher.

## Overall timeout vs per-try timeout

`timeout` bounds the whole exchange INCLUDING retries; `retryPolicy.
perTryTimeout` bounds one attempt. Retries spend the overall budget — they
never extend it — so a 10s timeout with a 5s per-try and 3 retries means at
most one retry in practice. Size the overall timeout to what the CLIENT
tolerates, the per-try to the backend's honest p99, and let the quotient
decide how many retries are real. Unset means the backend service's own
timeout governs.

`maxStreamDuration` is the third clock, for streams that outlive a
request/response exchange — and it is Traffic Director-only, live-verified:
the API rejects it unless the backend service's load-balancing scheme is
INTERNAL_SELF_MANAGED. On URL maps serving external application load
balancers, leave it unset.

## URL-map cachePolicy vs backend cdnPolicy

Two layers, one CDN: the backend service's `cdnPolicy` is the fleet-wide
stance; a route action's `cachePolicy` overrides it for matching traffic
only. Caching happens only when the target backend has CDN enabled — on a
CDN-less backend the route policy is silently inert, which is the first
thing to check when cached routes miss. One field needs explicit intent:
set `requestCoalescing: true` when you use `cachePolicy` and want
coalescing — inside a written cache policy an unset boolean is sent as
false, which turns coalescing OFF even though CDN's own default is on.

## Mirroring and fault injection are production-affecting

`requestMirrorPolicy` duplicates real traffic: the mirror backend must be
sized for the full mirrored load, and its side effects (writes!) are real —
mirror read-only routes, or point at a stack whose writes are sandboxed.
`faultInjectionPolicy` returns real errors and real delays to real
clients; scope it to a route rule matched by a test header, never a
default route action, and treat any nonzero percentage on a production
map as an incident waiting to be paged.

## Weighted canary with header attribution

Per-backend `headerAction` on a weighted split stamps which arm served the
request (e.g. `x-canary: v2` on the 10% backend) — dashboards and clients
can then attribute errors to the arm, which is the difference between a
canary you can read and one you argue about. Weight 0 drains a backend
without removing it: the rollback lever stays in the manifest.

## Destroy stance

`deletionPolicy: PREVENT` is the right default posture for a map that
other teams' DNS and proxies point at — a destroy fails at the map instead
of orphaning every target proxy above it. ABANDON removes it from state
but keeps it serving (unmanaged — someone must remember it exists). Leave
DELETE for maps whose whole stack tears down together, e2e fixtures
included.

## On the diagram

The URL map sits between its backends and the proxy layer: it consumes
`GcpBackendService`/`GcpBackendBucket` self links (default target, rule
targets, weighted arms, mirror backend — each a visible edge), and
`GcpTargetHttpProxy`/`GcpTargetHttpsProxy` consume its `self_link` above.
A mirror reference renders as one more backend edge — the shadow stack is
visible topology, not a hidden side channel.

## Pairs well with

- `GcpBackendService` — every routing target, weighted arm, and mirror.
- `GcpBackendBucket` — static assets and custom-error-page origins.
- `GcpTargetHttpsProxy` / `GcpTargetHttpProxy` — the self_link consumers.
- `GcpGlobalAddress` + `GcpGlobalForwardingRule` — the front door the
  proxies attach to.

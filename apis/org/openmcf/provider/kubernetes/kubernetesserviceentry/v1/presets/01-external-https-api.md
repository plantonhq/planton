# Reach an External HTTPS API

The canonical ServiceEntry: register an external service (a SaaS API, a partner
endpoint) so mesh workloads can call it as a first-class destination, with TLS routed
by SNI and the host resolved via DNS. This is the pattern you use to allow controlled
egress to a specific external API instead of opening blanket egress.

## When to Use

- You call an external HTTPS API (Stripe, Twilio, a partner service) from inside the
  mesh and want it routable, observable, and policy-addressable by its hostname.
- Your mesh runs in `REGISTRY_ONLY` outbound mode and you must explicitly register the
  destinations workloads are allowed to reach.

## Key Configuration Choices

- **`spec.hosts`** -- the external hostname(s). Used as the TLS SNI and the DNS name to
  resolve. A bare `*` is not allowed (use a specific host or a suffix wildcard).
- **`location: MESH_EXTERNAL`** -- the service lives outside the mesh; mTLS and mesh
  policy are not applied to it.
- **`resolution: DNS`** -- istiod resolves the host via DNS; no static endpoints are
  needed. Use `DNS_ROUND_ROBIN` for large web-scale endpoints that change frequently.
- **`ports`** -- expose `443` as `protocol: TLS` so Istio routes by SNI without
  terminating the connection (use `HTTPS`/`HTTP2` if you want Istio to originate TLS).

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the calling workloads have sidecars or are in the ambient mesh
  (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace whose workloads call the external API (e.g. `payments`). |
| `<external-host>` | The external hostname (e.g. `api.stripe.com`). |

To also configure how the mesh talks to this host (load balancing, outlier detection,
TLS origination), pair this with a `DestinationRule` targeting the same host.

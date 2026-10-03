# Overview

The AwsLbListener API resource provisions an ELBv2 listener: the
port/protocol entry point on a load balancer, carrying the TLS material and
the default action taken when no listener rule matches a request.

## Why We Created This API Resource

A listener is a first-class node in the routing graph. One load balancer
carries many listeners (80 and 443 at minimum on most ALBs), each listener
owns its own certificates and default behavior, and per-service listener
rules attach to a specific listener as services deploy. Modeling it as its
own component -- instead of folding it into the load balancer -- lets you:

- **Keep the load balancer stable**: adding a port, rotating a certificate,
  or changing the default action edits one listener, never the shared
  `AwsAlb`/`AwsNlb`.
- **Give rules an anchor**: `AwsLbListenerRule` resources reference the
  listener's `listener_arn` output, so per-service routing composes without
  coordination.
- **Serve both load balancer families with one kind**: ALB listeners
  (HTTP/HTTPS) take the full action set; NLB listeners
  (TCP/UDP/TCP_UDP/TLS/QUIC/TCP_QUIC) forward only -- exactly the split AWS
  enforces.

## Key Features

### TLS

- **Default certificate plus SNI**: the default certificate serves clients
  that match no SNI hostname; additional certificates (typically
  `AwsCertManagerCert` references) serve multiple domains from one listener.
- **Security policy selection**: pin the TLS versions and cipher suites the
  listener negotiates (e.g. `ELBSecurityPolicy-TLS13-1-2-2021-06`).
- **Mutual TLS** (ALB HTTPS): `passthrough` forwards the client certificate
  to targets; `verify` validates it against an ELBv2 trust store.
- **ALPN** (NLB TLS): advertise HTTP/1.1 or HTTP/2 during the handshake.

### Default Actions

- **Forward**: one target group, or up to five weighted groups with optional
  group-level stickiness -- the blue/green and canary primitive.
- **Redirect**: HTTP_301/HTTP_302 with per-kind overrides; the canonical
  HTTP-to-HTTPS redirect is two fields.
- **Fixed response**: serve a canned status/body straight from the load
  balancer -- the classic 404 default under rule-based routing.
- **Authentication**: `authenticate-cognito` (an `AwsCognitoUserPool`
  reference), `authenticate-oidc` (any OIDC provider), or `jwt-validation`
  (stateless bearer-token checks) chained before the terminal action.

### HTTP Header Controls (ALB)

- **Request-header injection**: surface TLS version, cipher, and mTLS client
  certificate details to targets under configurable header names.
- **Response-header overrides**: set CORS and browser security headers (HSTS,
  CSP, X-Frame-Options) uniformly at the edge.

## Benefits

- **Composability**: the listener references its load balancer, certificates,
  and target groups through `valueFrom`, and rules reference the listener --
  the routing graph is explicit end to end.
- **Safe TLS operations**: certificates and policies update in place; rotating
  a certificate never replaces the listener or drops rules.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `listener_arn`: ARN of the listener (what listener rules attach through)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

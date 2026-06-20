---
title: "Validate JWTs Across a Namespace"
description: "The canonical RequestAuthentication: accept JWTs from a single trusted issuer for every workload in a namespace. A request that carries a valid token gets an authenticated principal; a request with..."
type: "preset"
rank: "01"
presetSlug: "01-namespace-jwt-validation"
componentSlug: "request-authentication"
componentTitle: "Request Authentication"
provider: "kubernetes"
icon: "package"
order: 1
---

# Validate JWTs Across a Namespace

The canonical RequestAuthentication: accept JWTs from a single trusted issuer for
every workload in a namespace. A request that carries a valid token gets an
authenticated principal; a request with an invalid token is rejected; a request
with no token is still allowed (pair this with an AuthorizationPolicy to require a
principal).

## When to Use

- You have one identity provider (Auth0, Okta, Google, Cognito, Keycloak, ...) and
  want its tokens recognized mesh-side for all workloads in a namespace.
- You want JWT identities available to authorization decisions without changing
  application code.

## Key Configuration Choices

- **`spec.namespace`** -- the namespace the policy governs. With no `selector` or
  `target_refs`, it applies to every workload in the namespace.
- **`jwt_rules[].issuer`** -- must exactly match the token's `iss` claim.
- **`jwt_rules[].jwks_uri`** -- where istiod fetches the signing keys. Omit it only
  when the keys are discoverable via OpenID Discovery from the issuer.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the namespace's workloads have sidecars or are in the
  ambient mesh (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace whose workloads should validate JWTs (e.g. `finance`). |
| `<issuer-url>` | The token issuer, matching the `iss` claim (e.g. `https://accounts.example.com`). |
| `<jwks-url>` | The issuer's JWKS endpoint (e.g. `https://accounts.example.com/.well-known/jwks.json`). |

To require a token (not just validate one when present), add an
`AuthorizationPolicy` that demands `requestPrincipals`.

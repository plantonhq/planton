# Require an Authenticated JWT Principal

The canonical AuthorizationPolicy: allow a request only if it carries a valid JWT
identity (`request_principals: ["*"]` matches any issuer/subject). Anonymous requests --
those with no token, or a token that failed validation -- are denied for every workload
in the namespace.

## When to Use

- You front a namespace's workloads with a RequestAuthentication that validates JWTs and
  now want to *require* a token, not just validate one when present.
- You want a simple "must be logged in" gate without changing application code.

## How It Works

This is an ALLOW policy. With ALLOW, a request is permitted only if it matches a rule;
the single rule here matches any request that has an authenticated request principal, so
everything else is denied. Pair it with a `KubernetesRequestAuthentication` for the same
namespace/workloads so istiod knows how to validate the tokens.

## Key Configuration Choices

- **`spec.namespace`** -- the namespace the policy governs. With no `selector` or
  `target_refs`, it applies to every workload in the namespace.
- **`action: ALLOW`** -- the default; stated explicitly for clarity.
- **`rules[].from[].source.request_principals: ["*"]`** -- require any valid JWT
  principal. Narrow it to specific `<issuer>/<subject>` values to restrict further.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the namespace's workloads have sidecars or are in the ambient
  mesh (`KubernetesIstio`).
- A `KubernetesRequestAuthentication` validates the JWTs (otherwise no request principal
  is ever set and all requests are denied).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace whose workloads should require an authenticated principal (e.g. `finance`). |

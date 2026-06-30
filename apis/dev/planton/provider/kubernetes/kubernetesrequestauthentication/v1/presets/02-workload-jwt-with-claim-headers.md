# Workload JWT Validation with a Custom Header and Claim Forwarding

Validate JWTs for a single selected workload, extract the token from a custom
header, restrict it to a specific audience, and forward a verified claim to the
backend as an HTTP header the application can read.

## When to Use

- A specific service expects its identity provider's tokens in a non-standard
  header (e.g. `x-jwt-assertion` instead of `Authorization`).
- The application wants a verified claim (group, tenant, role) surfaced as a plain
  request header rather than parsing the JWT itself.
- You want a tighter, workload-scoped policy that does not affect the rest of the
  namespace.

## Key Configuration Choices

- **`spec.selector.match_labels`** -- targets just this workload's pods.
- **`jwt_rules[].audiences`** -- the token's `aud` must contain this value, so
  tokens minted for other services are rejected.
- **`jwt_rules[].from_headers`** -- extract the token from `x-jwt-assertion` after
  stripping the `Bearer ` prefix (note the trailing space).
- **`jwt_rules[].output_claim_to_headers`** -- copy the verified `groups` claim into
  the `x-jwt-group` header for the backend. The header name allows only
  `[-_A-Za-z0-9]`.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the workload has a sidecar or is in the ambient mesh
  (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace the workload runs in (e.g. `finance`). |
| `<workload>` | Value of the `app` label selecting the workload's pods (e.g. `finance`). |
| `<issuer-url>` | The token issuer, matching the `iss` claim. |
| `<jwks-url>` | The issuer's JWKS endpoint. |
| `<audience>` | The required `aud` claim value for this workload (e.g. `finance-api.example.com`). |

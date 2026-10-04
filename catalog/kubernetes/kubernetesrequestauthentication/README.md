# Kubernetes Request Authentication

Provision an Istio `RequestAuthentication` -- a namespaced policy that defines which
JSON Web Tokens (JWTs) are accepted on the workloads it selects. Use it to validate
end-user / caller tokens from one or more issuers at the mesh layer: a valid token's
identity becomes available to authorization policies, an invalid token is rejected,
and (by itself) a request with no token is still allowed.

## What Gets Created

- A namespaced `security.istio.io/v1` `RequestAuthentication` custom resource.
- A set of `jwt_rules` (issuer, key source, token locations, claim forwarding), and
  an optional workload `selector` **or** `target_refs`, scoped to this policy's
  namespace.

## Scope

A RequestAuthentication's reach depends on its attachment and namespace:

- **Workload-specific** -- with a `selector`, it applies only to matching pods in
  its namespace.
- **Attached** -- with `target_refs`, it binds to specific Gateway / Service /
  ServiceEntry resources (required for waypoint proxies).
- **Namespace-wide** -- with neither, it applies to every workload in its namespace.
- **Mesh-wide** -- placed in the Istio root namespace with no selector/target_refs,
  it applies across the mesh.

At most one of `selector` and `target_refs` may be set.

## RequestAuthentication vs the other security policies

- **RequestAuthentication** validates *request*-level credentials (end-user JWTs).
- **PeerAuthentication** controls *peer* (service-to-service) mTLS.
- **AuthorizationPolicy** allows/denies requests, and can require a JWT principal.

RequestAuthentication alone does not require a token; pair it with an
AuthorizationPolicy that demands `requestPrincipals` to reject anonymous requests.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to enforce the policy.
  The resource applies with only the CRDs present, but enforcement requires istiod
  and sidecar (or ambient) data-plane proxies.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRequestAuthentication
metadata:
  name: jwt-auth
spec:
  namespace:
    value: finance
  jwt_rules:
    - issuer: https://accounts.example.com
      jwks_uri: https://accounts.example.com/.well-known/jwks.json
```

```bash
planton apply -f requestauthentication.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the policy is created in (and the scope of a selector-less policy). |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `selector.match_labels` | map | Pod labels selecting target workloads. Omit to cover the whole namespace. Mutually exclusive with `target_refs`. |
| `target_refs` | list | Resources (Gateway / Service / ServiceEntry) the policy binds to. Required for waypoints. Mutually exclusive with `selector`. Each `name` is a foreign key defaulting to a `KubernetesGateway` reference; pass literals with `value:`. |
| `jwt_rules` | list | JWT validation rules (see below). When empty, the policy is a no-op. |

### JWT rule fields

| Field | Type | Description |
|-------|------|-------------|
| `issuer` | string | Must match the token's `iss` claim. May be omitted only when `jwks_uri` is set -- at least one of the two is required, so the verifier can locate the signing keys. An issuer-less rule validates any issuer against the fixed key set. |
| `audiences` | list | Accepted `aud` values. When empty, the workload's service name is accepted. |
| `jwks_uri` | string | URL of the issuer's JWKS. Mutually exclusive with `jwks`. Must be `http(s)://`. |
| `jwks` | string | Inline JWKS. Mutually exclusive with `jwks_uri`. |
| `from_headers` | list | Header locations (`name`, `prefix`) to extract the token from. |
| `from_params` | list | Query parameter names to extract the token from. |
| `from_cookies` | list | Cookie names to extract the token from. |
| `output_payload_to_header` | string | Header to emit the base64 JWT payload to. |
| `forward_original_token` | bool | Forward the original token upstream (default false). |
| `output_claim_to_headers` | list | Copy verified claims to headers (`header`, `claim`). Header names allow `[-_A-Za-z0-9]`. |
| `timeout` | string | Max JWKS-fetch time, a duration like `5s` (default 5s, minimum 1ms). |
| `space_delimited_claims` | list | Claim names whose string values are treated as space-delimited lists (the OAuth2 `scope` convention, e.g. `"read write admin"`) by authorization policies and claim-to-header operations. Up to 64. |

## Examples

### Namespace-wide JWT validation

```yaml
spec:
  namespace:
    value: finance
  jwt_rules:
    - issuer: https://accounts.example.com
      jwks_uri: https://accounts.example.com/.well-known/jwks.json
```

### Per-workload, custom header, forwarded claim

```yaml
spec:
  namespace:
    value: finance
  selector:
    match_labels:
      app: finance
  jwt_rules:
    - issuer: https://accounts.example.com
      jwks_uri: https://accounts.example.com/.well-known/jwks.json
      audiences:
        - finance-api.example.com
      from_headers:
        - name: x-jwt-assertion
          prefix: "Bearer "
      output_claim_to_headers:
        - header: x-jwt-group
          claim: groups
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace`
output via `valueFrom`. Each `target_refs[].name` is also a foreign key: it defaults
to a `KubernetesGateway` reference (`status.outputs.gateway_name`), so wiring it with
`valueFrom` gives the chart a real dependency edge from the policy to the gateway it
protects:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: edge-ns
      fieldPath: spec.name
  target_refs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name:
        valueFrom:
          kind: KubernetesGateway
          name: edge-gateway
          fieldPath: status.outputs.gateway_name
  jwt_rules:
    - issuer: https://accounts.example.com
      jwks_uri: https://accounts.example.com/.well-known/jwks.json
```

When the target is a Service, a ServiceEntry, or any resource not managed as a
Planton kind, pass the literal name with `value:`. `selector.match_labels` is a plain
label match, not a foreign key -- istiod resolves it at runtime, so it creates no
automatic DAG edge to the workloads it selects. To order this policy relative to those
workloads in an Infra Chart, declare the dependency on `metadata.relationships`:

```yaml
metadata:
  name: finance-jwt
  relationships:
    - kind: KubernetesDeployment
      name: finance
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: finance-ns
      fieldPath: spec.name
  selector:
    match_labels:
      app: finance
  jwt_rules:
    - issuer: https://accounts.example.com
      jwks_uri: https://accounts.example.com/.well-known/jwks.json
```

See `GUIDE.md` for the full composability rationale.

## Outputs

| Output | Description |
|--------|-------------|
| `request_authentication_name` | Name of the created RequestAuthentication (equals metadata.name). |
| `namespace` | Namespace the RequestAuthentication was created in. |

## Related Kinds

- [Kubernetes Peer Authentication](../kubernetespeerauthentication)
- [Kubernetes Istio](../kubernetesistio)
- [Kubernetes Istio Base CRDs](../kubernetesistiobasecrds)
- [Kubernetes Namespace](../kubernetesnamespace)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

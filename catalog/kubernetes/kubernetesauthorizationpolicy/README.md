# Kubernetes Authorization Policy

Provision an Istio `AuthorizationPolicy` -- a namespaced policy that enforces access
control on the workloads it selects. It decides whether to **ALLOW**, **DENY**, or
**AUDIT** a request based on a list of rules, or delegates the decision to an external
authorizer with the **CUSTOM** action. Rules match on request source (peer/JWT
identity, namespace, IP), operation (host, port, method, path), and arbitrary Istio
attribute conditions.

## What Gets Created

- A namespaced `security.istio.io/v1` `AuthorizationPolicy` custom resource.
- An `action` (default ALLOW) applied to requests matched by `rules`, with an optional
  workload `selector` **or** `target_refs`, scoped to this policy's namespace.

## How the action works

When multiple policies apply to a workload, Istio evaluates them in a fixed order:
CUSTOM first, then DENY, then ALLOW.

- **ALLOW** (default) -- allow a request only if it matches a rule. An empty `rules`
  list therefore denies everything (a deny-all policy).
- **DENY** -- deny a request if it matches any rule. Always scope DENY rules to a port.
- **AUDIT** -- mark matching requests for auditing; does not change allow/deny.
- **CUSTOM** -- delegate to the external authorizer named in `provider` (which must be
  declared in the mesh's MeshConfig).

## Scope

An AuthorizationPolicy's reach depends on its attachment and namespace:

- **Workload-specific** -- with a `selector`, it applies only to matching pods in its
  namespace.
- **Attached** -- with `target_refs`, it binds to specific Gateway / Service /
  ServiceEntry resources (required for waypoint proxies).
- **Namespace-wide** -- with neither, it applies to every workload in its namespace.
- **Mesh-wide** -- placed in the Istio root namespace with no selector/target_refs.

At most one of `selector` and `target_refs` may be set.

## AuthorizationPolicy vs the other security policies

- **AuthorizationPolicy** allows/denies/audits requests (this component).
- **RequestAuthentication** validates *request*-level credentials (end-user JWTs) but
  does not require one. Pair it with an AuthorizationPolicy that demands
  `request_principals` to reject anonymous requests.
- **PeerAuthentication** controls *peer* (service-to-service) mTLS.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to enforce the policy.
  The resource applies with only the CRDs present, but enforcement requires istiod and
  sidecar (or ambient) data-plane proxies.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAuthorizationPolicy
metadata:
  name: require-jwt
spec:
  namespace:
    value: finance
  action: ALLOW
  rules:
    - from:
        - source:
            request_principals:
              - "*"
```

```bash
planton apply -f authorizationpolicy.yaml
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
| `rules` | list | Rules to match the request (see below). With ALLOW and no rules, all requests are denied. |
| `action` | string | `ALLOW` (default), `DENY`, `AUDIT`, or `CUSTOM`. |
| `provider.name` | string | The MeshConfig extension provider to delegate to; used with the CUSTOM action. |

### Rule fields

A request matches a rule when at least one `from`, at least one `to`, and all `when`
match. A rule that should match every request says so: `matchAll: true` (the empty
rule on the Istio wire); a rule that names neither a matcher nor `matchAll` is refused,
so matching everything is always written down.

| Field | Type | Description |
|-------|------|-------------|
| `from[].source` | object | Request source match (identities, namespaces, IPs). See source fields. |
| `to[].operation` | object | Request operation match (hosts, ports, methods, paths). |
| `when[]` | list | Additional conditions (`key`, `values`, `not_values`). |
| `matchAll` | bool | Match every request; cannot be combined with `from`, `to`, or `when`. Never reaches the cluster -- the rule is emitted with no matchers, Istio's own spelling of "everything". |

### Source fields (all optional lists; each supports `*`/prefix/suffix match)

`principals`, `not_principals`, `request_principals`, `not_request_principals`,
`namespaces`, `not_namespaces`, `service_accounts`, `not_service_accounts`,
`ip_blocks`, `not_ip_blocks`, `remote_ip_blocks`, `not_remote_ip_blocks`,
`trust_domains`, `not_trust_domains`.

`service_accounts` / `not_service_accounts` cannot be combined with `principals` or
`namespaces` (positive or negative), allow no wildcards, and are capped at 16 entries
of up to 320 characters each.

`trust_domains` / `not_trust_domains` match any workload whose peer identity belongs
to one of the listed trust domains (the mesh's identity root, e.g. `cluster.local`)
without enumerating principals; they require mTLS.

### Operation fields (all optional lists)

`hosts`, `not_hosts`, `ports`, `not_ports`, `methods`, `not_methods`, `paths`,
`not_paths`.

## Examples

### Require an authenticated JWT principal (namespace-wide)

```yaml
spec:
  namespace:
    value: finance
  action: ALLOW
  rules:
    - from:
        - source:
            request_principals:
              - "*"
```

### Deny POST on a port for a selected workload

```yaml
spec:
  namespace:
    value: finance
  selector:
    match_labels:
      app: httpbin
  action: DENY
  rules:
    - to:
        - operation:
            methods: ["POST"]
            ports: ["8080"]
```

### CUSTOM external authorization on an ingress path

```yaml
spec:
  namespace:
    value: istio-system
  selector:
    match_labels:
      app: istio-ingressgateway
  action: CUSTOM
  provider:
    name: my-custom-authz
  rules:
    - to:
        - operation:
            paths: ["/admin/*"]
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace` output
via `valueFrom`. Each `target_refs[].name` is also a foreign key: it defaults to a
`KubernetesGateway` reference (`status.outputs.gateway_name`), so wiring it with
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
  action: ALLOW
  rules:
    - from:
        - source:
            request_principals:
              - "*"
```

When the target is a Service, a ServiceEntry, or any resource not managed as a Planton
kind, pass the literal name with `value:`. `selector.match_labels` is a plain label
match, not a foreign key -- istiod resolves it at runtime, so it creates no automatic
DAG edge to the workloads it selects. To order this policy relative to those workloads
in an infra chart, declare the dependency on `metadata.relationships`:

```yaml
metadata:
  name: httpbin-authz
  relationships:
    - kind: KubernetesDeployment
      name: httpbin
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: finance-ns
      fieldPath: spec.name
  selector:
    match_labels:
      app: httpbin
  action: ALLOW
  rules:
    - from:
        - source:
            request_principals:
              - "*"
```

See `GUIDE.md` for the full composability rationale.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `authorization_policy_name` | Name of the created AuthorizationPolicy (equals metadata.name). |
| `namespace` | Namespace the AuthorizationPolicy was created in. |

## Related Components

- [Kubernetes Request Authentication](../kubernetesrequestauthentication)
- [Kubernetes Peer Authentication](../kubernetespeerauthentication)
- [Kubernetes Istio](../kubernetesistio)
- [Kubernetes Istio Base CRDs](../kubernetesistiobasecrds)
- [Kubernetes Namespace](../kubernetesnamespace)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

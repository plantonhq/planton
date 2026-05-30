# Allow Routes to Reference Backend Services in Another Namespace

Authorize HTTP and gRPC routes in an application/frontend namespace to forward
traffic to backend Services that live in a different namespace. By default a
route's cross-namespace `backendRefs` reference is denied; this grant, placed in
the backend's namespace, authorizes it.

## When to Use

- Your `KubernetesHttpRoute` / `KubernetesGrpcRoute` lives in one namespace but
  targets backend Services in another (a common multi-team layout).
- You want a single grant to cover both route kinds from the same source
  namespace.

## Key Configuration Choices

- **`spec.namespace`** -- the namespace the backend Services live in (the "to"
  side). The grant must be created here.
- **`from`** -- one entry per trusted route kind (`HTTPRoute`, `GRPCRoute`) in the
  route namespace. Entries combine with OR. Add `TLSRoute`/`TCPRoute` if those
  route kinds also target this namespace.
- **`to`** -- `kind: Service` with `group: ""` (Service is a core kind). Omit
  `name` to allow all Services, or set it to restrict the grant.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- The target (backend) namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<backend-namespace>` | Namespace where the backend Services live. |
| `<route-namespace>` | Namespace where the routes live. |

Set `spec.namespace.value` to your backend namespace, or replace it with a
`valueFrom` reference to a `KubernetesNamespace`.

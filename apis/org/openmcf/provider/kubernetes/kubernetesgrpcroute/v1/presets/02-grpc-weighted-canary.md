# gRPC Weighted Canary

Split gRPC traffic for a service across two backends by weight -- the standard
progressive-delivery pattern. Here 90% of calls go to the stable backend and 10%
to the canary; adjust the weights to shift traffic.

## When to Use

- You are rolling out a new version of a gRPC service and want to send a small
  percentage of calls to it first.
- You want weighted traffic splitting without an external progressive-delivery
  controller.

## Key Configuration Choices

- **`backendRefs[].weight`** -- relative weight; each backend receives
  `weight / (sum of weights)` of the traffic. `90` and `10` yield a 90/10 split.
- **`method.service`** -- scopes the split to one gRPC service; omit `matches` to
  split all traffic on the route.
- **`backendRefs[].port`** -- required when the backend is a core Service.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- The `Gateway` referenced in `parentRefs` exists (`KubernetesGateway`) and its
  listener accepts HTTP/2.
- The target namespace exists (`KubernetesNamespace`).
- Both backend gRPC Services exist in the route's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<gateway-name>` | Name of the `KubernetesGateway` this route attaches to. |
| `<api-hostname>` | Public hostname this route serves, e.g. `api.example.com`. |
| `<grpc-service>` | Fully-qualified gRPC service, e.g. `helloworld.Greeter`. |
| `<stable-service-name>` | Name of the stable backend Kubernetes Service. |
| `<canary-service-name>` | Name of the canary backend Kubernetes Service. |

Set `spec.namespace.value` to your namespace, or replace it with a `valueFrom`
reference to a `KubernetesNamespace`.

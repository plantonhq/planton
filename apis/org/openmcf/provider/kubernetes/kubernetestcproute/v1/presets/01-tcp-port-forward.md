# TCP Port Forwarding

The most common TCPRoute: forward all connections arriving on a Gateway's TCP
listener to a backend Service. A TCP route has no matching -- the listener's port
selects the traffic, and the route forwards it. This is the standard pattern for
exposing a non-HTTP TCP service (a database, a message broker, a custom protocol)
through a Gateway.

## When to Use

- You expose a raw TCP service (Postgres, Redis, Kafka, a custom protocol) behind
  a Gateway.
- You route purely by listener port -- there is no application-layer matching.

## Key Configuration Choices

- **`parentRefs`** -- attaches the route to the Gateway by name; add `sectionName` to target the TCP listener (the listener's port determines which connections arrive).
- **`backendRefs[].port`** -- the backend Service port that receives the forwarded connection.

## Prerequisites

- The Gateway API **experimental-channel** CRDs are installed
  (`KubernetesGatewayApiCrds` with `install_channel: experimental`). TCPRoute is
  not part of the standard channel.
- The `Gateway` referenced in `parentRefs` exists (`KubernetesGateway`) with a
  listener of protocol `TCP`.
- The target namespace exists (`KubernetesNamespace`).
- The backend Service exists in the route's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<gateway-name>` | Name of the `KubernetesGateway` this route attaches to. |
| `<service-name>` | Name of the backend Kubernetes Service. |
| `<service-port>` | Backend Service port, e.g. `5432` for Postgres. |

Set `spec.namespace.value` to your namespace, or replace it with a `valueFrom`
reference to a `KubernetesNamespace`.

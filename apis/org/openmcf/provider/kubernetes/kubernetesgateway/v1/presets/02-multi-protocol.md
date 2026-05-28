# Multi-Protocol Gateway

A single Gateway exposing three listeners on distinct ports: cleartext HTTP
(often used to redirect to HTTPS), HTTPS with TLS termination, and a raw TCP
listener for a non-HTTP workload such as a database. This demonstrates how one
Gateway can front several protocols, each attaching its own Route kind.

## When to Use

- You need HTTP, HTTPS, and TCP entry points behind one Gateway.
- You run a mix of HTTP services and a TCP workload (for example Postgres).
- You want a single address/load balancer shared across protocols.

## Key Configuration Choices

- **Three listeners with unique name + port + protocol** -- required for listeners to be distinct.
- **http (port 80, HTTP)** -- cleartext entry; commonly paired with an HTTPRoute that redirects to HTTPS.
- **https (port 443, HTTPS, Terminate)** -- TLS termination with a certificate Secret.
- **postgres (port 5432, TCP)** -- raw TCP forwarding; attach a `TCPRoute` to reach the backend.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- A `GatewayClass` named `istio` (or your controller) exists, and the controller supports TCP listeners.
- The target namespace exists (`KubernetesNamespace`).
- The referenced TLS Secret exists in the Gateway's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<app-hostname>` | The public hostname the HTTPS listener serves, e.g. `app.example.com`. |
| `<tls-secret-name>` | Name of the `kubernetes.io/tls` Secret holding the certificate and key. |

Note: not every controller supports TCP listeners. Remove the `postgres`
listener if your controller is HTTP-only, or consult its documentation.

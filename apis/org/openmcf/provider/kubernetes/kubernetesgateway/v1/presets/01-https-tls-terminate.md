# HTTPS Gateway with TLS Termination

A single HTTPS listener that terminates TLS at the Gateway using a certificate
stored in a Kubernetes Secret (typically created by a cert-manager
`KubernetesCertificate`). This is the most common production ingress pattern:
the Gateway owns the public endpoint and certificate, and `HTTPRoute`s in the
same namespace attach to it for host/path routing.

## When to Use

- You terminate TLS at the edge and route cleartext to backends.
- Your TLS certificate is provisioned into a `kubernetes.io/tls` Secret.
- You want HTTPRoutes from the Gateway's own namespace to attach.

## Key Configuration Choices

- **listeners[0].protocol** (`HTTPS`) -- terminates TLS at the Gateway.
- **listeners[0].tls.mode** (`Terminate`) -- the Gateway decrypts and requires a certificate.
- **listeners[0].tls.certificateRefs[0].name** -- the TLS Secret holding the cert/key.
- **listeners[0].hostname** -- restricts the listener to a single virtual host (SNI/Host).
- **allowedRoutes.namespaces.from** (`Same`) -- only same-namespace Routes may attach.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- A `GatewayClass` named `istio` (or your controller) exists (`KubernetesGatewayClass`).
- The target namespace exists (`KubernetesNamespace`).
- The referenced TLS Secret exists in the Gateway's namespace
  (for example created by a cert-manager `KubernetesCertificate`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<app-hostname>` | The public hostname this listener serves, e.g. `app.example.com`. |
| `<tls-secret-name>` | Name of the `kubernetes.io/tls` Secret holding the certificate and key. |

Set `spec.namespace.value` and `spec.gatewayClassName.value` to your namespace
and GatewayClass, or replace them with `valueFrom` references to a
`KubernetesNamespace` and `KubernetesGatewayClass`.

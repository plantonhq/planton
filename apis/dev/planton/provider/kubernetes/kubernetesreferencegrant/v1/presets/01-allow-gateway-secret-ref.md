# Allow a Gateway to Reference TLS Secrets in Another Namespace

The most common ReferenceGrant: a Gateway terminates TLS using a certificate
Secret that lives in a different namespace (typically the cert-manager namespace).
By default that cross-namespace `certificateRefs` reference is denied; this grant,
placed in the Secret's namespace, authorizes it.

## When to Use

- Your `KubernetesGateway` lives in an ingress namespace (e.g. `istio-ingress`)
  but its TLS Secret is produced by cert-manager in another namespace.
- You want explicit, auditable authorization rather than co-locating the Secret
  with the Gateway.

## Key Configuration Choices

- **`spec.namespace`** -- the namespace the Secret lives in (the "to" side). The
  grant must be created here.
- **`from`** -- the Gateway's namespace and kind (`Gateway`,
  `gateway.networking.k8s.io`) that is trusted to reference in.
- **`to`** -- `kind: Secret` with `group: ""` (Secret is a core kind). Omit
  `name` to allow all Secrets, or set it to restrict the grant to one Secret.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- The target (Secret) namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<cert-namespace>` | Namespace where the TLS Secret lives (e.g. `cert-manager`). |
| `<gateway-namespace>` | Namespace where the `KubernetesGateway` lives (e.g. `istio-ingress`). |

Set `spec.namespace.value` to your Secret's namespace, or replace it with a
`valueFrom` reference to a `KubernetesNamespace`.

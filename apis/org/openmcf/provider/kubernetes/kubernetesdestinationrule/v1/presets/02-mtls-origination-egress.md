# mTLS Origination to an Egress Host

Configure the sidecar to originate mutual TLS to an external service, presenting client
certificates loaded from a Kubernetes secret. This is how you let in-mesh workloads talk to
an external database or partner API that requires client-certificate authentication, without
each application managing TLS itself.

## When to Use

- Mesh workloads connect to an external endpoint (a managed database, a partner API) that
  requires mutual TLS, and you want istiod/Envoy to present the client certs.
- You keep the client credential in a Kubernetes secret and want Istio to load it by name
  rather than mounting files into every pod.

## Key Configuration Choices

- **`host`** -- the external host (typically also registered via a `ServiceEntry`).
- **`workload_selector.match_labels`** -- required for `credential_name` to take effect at
  sidecars: it scopes the rule to the client workloads that hold the credential.
- **`traffic_policy.tls.mode: MUTUAL`** -- originate mutual TLS, presenting client certs.
- **`traffic_policy.tls.credential_name`** -- the secret holding the client cert/key/CA. Use
  `SIMPLE` instead of `MUTUAL` (and drop the credential) for one-way TLS origination, or
  `ISTIO_MUTUAL` to use Istio's own workload certificates.
- **`traffic_policy.tls.sni`** -- the SNI to present, when it differs from the host.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the client workloads have sidecars (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).
- The TLS secret named by `credential_name` exists in the proxy's namespace
  (`KubernetesSecret` / `KubernetesCertificate`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace of the client workloads (e.g. `egress`). |
| `<external-host>` | The external host to originate TLS to (e.g. `external-db.example.com`). |
| `<client-app-label>` | The label value selecting the client workloads (e.g. `db-client`). |
| `<credential-secret>` | The secret holding the client TLS certs (e.g. `db-client-cert`). |

Wire the secret dependency on `metadata.relationships` (`uses` -> KubernetesSecret) so the
infra chart creates it before this rule. See the component README's "Composing in Infra
Charts" section.

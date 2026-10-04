# Internal CA via ConfigMap

The most common BackendTLSPolicy: the gateway originates TLS to an
internal backend whose serving certificate is signed by YOUR OWN CA — the
private-PKI posture behind most cert-manager installations. The trust
anchor is a same-namespace ConfigMap carrying the PEM CA bundle in a key
named `ca.crt` (the Core-supported shape, and exactly what a cert-manager
CA chain materializes). Both the target Service and the CA ConfigMap are
wired as foreign keys, so the policy deploys after the resources it
depends on in one Infra Chart.

## When to Use

- Backends behind a Gateway serve TLS with certificates issued by an
  internal CA (cert-manager `CA` or `SelfSigned`-rooted issuers)
- You want end-to-end encryption: TLS terminates at the gateway AND is
  re-originated — verified — to the backend hop
- The backend Service and CA bundle ConfigMap are Planton-managed (use
  `value:` for either name when they are not)

## Key Configuration Choices

- **Same-namespace everything** — upstream forbids cross-namespace
  targetRefs and CA references for this policy; create the policy in the
  backend's namespace
- **`targetRefs[0]`: `group: ""`, `kind: Service`** — the Core-supported
  target; `group` must be present-but-empty (the CRD rejects a missing
  key). No `sectionName`, so the policy covers every Service port.
- **`name` via `valueFrom`** — the Service name flows from the
  `KubernetesService` resource's `status.outputs.service_name`, giving the
  Infra Chart its dependency edge for free
- **`caCertificateRefs`: one ConfigMap** — Core support is exactly one
  ConfigMap with the bundle under `ca.crt`; the name flows from the
  `KubernetesConfigMap` resource's `status.outputs.configmap_name`
- **`hostname`** — does double duty: the SNI for the backend connection
  and (since `subjectAltNames` is unset) the identity the backend
  certificate must match — it must appear in the certificate's SANs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<backend-namespace>` | Namespace of the backend Service (the policy must live there too) | Your Infra Chart / cluster namespaces |
| `<backend-service-resource>` | Name of the `KubernetesService` resource the policy secures (inside `valueFrom.name`) | Your Infra Chart / `planton` resource listing |
| `<ca-bundle-configmap-resource>` | Name of the `KubernetesConfigMap` resource carrying the CA bundle under `ca.crt` (inside `valueFrom.name`) | Your Infra Chart / cert-manager CA chain outputs |
| `backend.internal.example.com` | SNI + certificate identity of the backend | The backend certificate's DNS SANs |

## Related Presets

- **02-public-ca-system** — backends serving publicly-issued certificates
  (system trust store, no bundle to manage)
- **03-spiffe-mtls-backend** — SPIFFE-identity backends where the
  certificate identity differs from the SNI hostname

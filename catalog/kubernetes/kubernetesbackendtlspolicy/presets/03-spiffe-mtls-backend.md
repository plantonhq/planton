# SPIFFE mTLS Backend

This preset handles the mesh-identity case: the backend's certificate does
not carry the DNS hostname the gateway dials — it carries a SPIFFE ID (the
URI-SAN identity pattern of SPIRE, Istio, and other mTLS meshes). The
policy pins the backend's expected SPIFFE ID through `subjectAltNames`,
trusts the mesh/workload CA bundle from a ConfigMap, and narrows the
attachment to one named Service port with `sectionName`.

## When to Use

- Backends behind the gateway run inside an mTLS mesh whose workload
  certificates carry SPIFFE URI SANs instead of DNS names
- You want the gateway to verify the backend's *workload identity*, not
  just its transport certificate
- Only one port of the Service serves TLS (the `https` port) — the rest
  should stay untouched by this policy

## Key Configuration Choices

- **`subjectAltNames` with one `URI` entry** — the backend certificate
  must prove the SPIFFE ID. The spec mirrors the CRD's pairing rules:
  `uri` is required (and `hostname` forbidden) for type `URI`. Upstream
  support for SANs is Extended — confirm your implementation.
- **`hostname` demoted to SNI-only** — with `subjectAltNames` set, the
  hostname only selects the backend certificate; add it as a `Hostname`
  SAN entry if it should also authenticate
- **`sectionName: https`** — the policy applies only to connections to the
  Service port named `https`; a name that does not exist on the Service
  makes the policy fail to attach (surfaced via ResolvedRefs)
- **CA bundle ConfigMap via `valueFrom`** — the mesh trust bundle (e.g.
  the SPIRE trust bundle) in a ConfigMap under `ca.crt`, wired as a
  foreign key to the `KubernetesConfigMap` resource
- **Foreign-keyed target Service** — deploys after the Service it secures

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<backend-namespace>` | Namespace of the backend Service (the policy must live there too) | Your Infra Chart / cluster namespaces |
| `<backend-service-resource>` | Name of the `KubernetesService` resource the policy secures (inside `valueFrom.name`) | Your Infra Chart / `planton` resource listing |
| `<trust-bundle-configmap-resource>` | Name of the `KubernetesConfigMap` resource carrying the mesh CA bundle under `ca.crt` (inside `valueFrom.name`) | Your mesh's trust-bundle distribution (SPIRE bundle publisher, trust-manager) |
| `backend.internal.example.com` | SNI sent to the backend (certificate selection only) | Your backend's serving config |
| `spiffe://cluster.example.com/ns/payments/sa/backend` | The backend workload's SPIFFE ID | Your mesh trust domain + the workload's namespace/service account |
| `https` (`sectionName`) | The named Service port serving TLS | The backend Service's port names |

## Related Presets

- **01-internal-ca-configmap** — the plain internal-CA case where the
  certificate identity IS the hostname
- **02-public-ca-system** — publicly-issued backend certificates

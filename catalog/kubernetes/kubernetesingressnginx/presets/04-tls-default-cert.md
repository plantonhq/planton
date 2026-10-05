# Cluster-Wide Default TLS Certificate (cert-manager Composition)

This preset installs the public entry controller with a cluster-wide
default TLS certificate wired from cert-manager: a KubernetesCertificate's
secret output flows into `defaultTlsCertificate` as a reference, and the
controller serves that certificate on any HTTPS request that matches no
Ingress TLS block (and on the default backend). With a wildcard certificate
this removes per-Ingress TLS boilerplate for every host under the domain.

## When to Use

- Clusters where most Ingress hosts live under one (wildcard-coverable)
  domain and per-Ingress TLS blocks are pure repetition
- A guaranteed-valid fallback certificate for hostnames that have no
  explicit TLS configuration
- Composing the controller with the cert-manager kinds
  (KubernetesCertManager + issuers + KubernetesCertificate) in one infra
  chart

## Key Configuration Choices

- **`defaultTlsCertificate.secretName` via `valueFrom`** — the cert-manager
  seam: the reference resolves to the KubernetesCertificate's
  `status.outputs.secret_name`, so the controller always points at the
  Secret cert-manager creates AND renews. No manual Secret management,
  no expiry drift
- **How it lands** — the chart exposes no first-class key for the default
  certificate; upstream's own documented mechanism is the
  `--default-ssl-certificate` controller flag, which the modules render
  through the chart's `extraArgs`
- **`namespace`** — where the certificate's Secret lives; leave it empty
  when the certificate is issued into the controller's installation
  namespace

Individual Ingresses can still carry their own TLS blocks — the default
only answers when no Ingress TLS matches.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<wildcard-certificate-name>` | Name of the KubernetesCertificate resource (e.g. a `*.example.com` wildcard) | Your Infra Chart / certificate manifests |
| `<certificate-secret-namespace>` | Namespace the certificate's Secret is issued into | The KubernetesCertificate's spec |

## Related Presets

- **01-aws-nlb-public** — the same entry posture without the default
  certificate
- **02-internal-only** — a second instance for private traffic (give it its
  own internal-domain default certificate)

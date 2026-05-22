# KubernetesIssuer

Creates a namespace-scoped cert-manager Issuer for CA or self-signed certificate signing. Each instance manages one Issuer in one namespace.

## What Gets Created

- **Issuer** -- cert-manager Issuer CR in the specified namespace (CA or SelfSigned mode)

## Prerequisites

- cert-manager installed on the cluster (via KubernetesCertManager)
- Target namespace must already exist
- For CA mode: a Secret with CA keypair (`tls.crt` + `tls.key`) in the same namespace

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIssuer
metadata:
  name: selfsigned-issuer
spec:
  namespace:
    value: cert-manager
  selfSigned: {}
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where the Issuer was created |
| `issuer_name` | Name of the Issuer (equals `metadata.name`) |

## Related Components

- **KubernetesCertManager** -- installs the cert-manager controller
- **KubernetesClusterIssuer** -- cluster-scoped ACME issuer for DNS-01 challenges
- **KubernetesCertificate** -- creates Certificates that reference this Issuer

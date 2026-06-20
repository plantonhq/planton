---
title: "KubernetesCertificate"
description: "KubernetesCertificate deployment documentation"
icon: "package"
order: 100
componentName: "kubernetescertificate"
---

# KubernetesCertificate

Creates a cert-manager Certificate for requesting signed TLS certificates from an Issuer or ClusterIssuer. Each instance manages one Certificate and its corresponding TLS Secret.

## What Gets Created

- **Certificate** -- cert-manager Certificate CR in the specified namespace
- **TLS Secret** -- Kubernetes Secret containing the signed certificate and private key (created by cert-manager)

## Prerequisites

- cert-manager installed on the cluster (via KubernetesCertManager)
- A configured Issuer or ClusterIssuer to sign the certificate

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertificate
metadata:
  name: my-app-cert
spec:
  namespace:
    value: my-app
  dnsNames:
    - app.example.com
  secretName: my-app-tls
  issuerRef:
    clusterIssuer:
      name:
        value: example.com
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where the Certificate was created |
| `certificate_name` | Name of the Certificate resource |
| `secret_name` | TLS Secret name for Gateway/Ingress/CA Issuer consumption |

## Related Components

- **KubernetesClusterIssuer** -- creates ACME ClusterIssuers for public TLS
- **KubernetesIssuer** -- creates namespace-scoped Issuers (SelfSigned, CA)
- **KubernetesCertManager** -- installs the cert-manager controller

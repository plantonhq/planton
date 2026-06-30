# KubernetesCertificate

> Declarative cert-manager Certificate management for any Kubernetes cluster

## Overview

KubernetesCertificate creates a cert-manager [Certificate](https://cert-manager.io/docs/concepts/certificate/) resource that requests a signed TLS certificate from a specified Issuer or ClusterIssuer. The resulting certificate and private key are stored in a Kubernetes Secret, ready for consumption by Ingress controllers, Gateway APIs, or internal CA bootstrap chains.

This component supports two issuer types via a proto oneof:
- **ClusterIssuer** -- cluster-scoped, typically ACME / Let's Encrypt (most common for public TLS)
- **Issuer** -- namespace-scoped, typically for internal CA or self-signed certificate workflows

## Prerequisites

- **cert-manager must be installed** on the target cluster (via KubernetesCertManager or manually)
- A configured **Issuer or ClusterIssuer** to sign the certificate (via KubernetesClusterIssuer, KubernetesIssuer, or manually)

## Quick Start

### Public TLS Certificate (via ClusterIssuer)

```yaml
apiVersion: kubernetes.planton.dev/v1
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

### Self-Signed Root CA Certificate (CA Bootstrap)

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesCertificate
metadata:
  name: my-root-ca
spec:
  namespace:
    value: cert-manager
  dnsNames:
    - my-root-ca
  secretName: my-root-ca-secret
  isCa: true
  issuerRef:
    issuer:
      name:
        value: selfsigned-issuer
  durationConfig:
    duration: "87600h"
    renewBefore: "2160h"
  privateKey:
    algorithm: rsa
    size: 4096
    encoding: pkcs1
    rotationPolicy: never
```

### Deploy

```bash
planton pulumi up --manifest certificate.yaml --stack org/project/env
```

## How It Works

1. The module creates a single cert-manager Certificate CR in the specified namespace
2. cert-manager's controller watches for Certificate resources and communicates with the configured Issuer to obtain a signed certificate
3. The signed certificate and private key are stored in the Kubernetes Secret named by `secret_name`
4. Consumers (Ingress, Gateway, CA Issuer) reference the Secret directly

## CA Bootstrap Workflow

KubernetesCertificate supports internal PKI bootstrap by setting `is_ca: true`:

```
Step 1: KubernetesIssuer (SelfSigned type)
   Creates a self-signed Issuer in a namespace
        ↓
Step 2: KubernetesCertificate (is_ca=true, issuerRef → SelfSigned Issuer)
   Requests a self-signed CA certificate, stored in a Secret
        ↓
Step 3: KubernetesIssuer (CA type, ca_secret_name → Step 2's Secret)
   Creates a CA Issuer backed by the root CA certificate
        ↓
Step 4: KubernetesCertificate (issuerRef → CA Issuer)
   Requests leaf certificates signed by the internal CA
```

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `namespace` | StringValueOrRef | Yes | Namespace for the Certificate resource |
| `dnsNames` | string[] | Yes | DNS Subject Alternative Names (at least one) |
| `secretName` | string | Yes | Secret name for the signed certificate and key |
| `issuerRef.clusterIssuer.name` | StringValueOrRef | One of | ClusterIssuer name (cluster-scoped) |
| `issuerRef.issuer.name` | StringValueOrRef | One of | Issuer name (namespace-scoped) |
| `isCa` | bool | No | Issue as a CA certificate (default: false) |
| `durationConfig.duration` | string | No | Certificate lifetime (default: 2160h = 90 days) |
| `durationConfig.renewBefore` | string | No | Renew this long before expiry (default: 360h = 15 days) |
| `privateKey.algorithm` | enum | No | RSA, ECDSA, or Ed25519 (default: RSA) |
| `privateKey.size` | int | No | Key size in bits (default: 2048) |
| `privateKey.encoding` | enum | No | PKCS1 or PKCS8 (default: PKCS1) |
| `privateKey.rotationPolicy` | enum | No | Always or Never (default: Always) |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where the Certificate was created |
| `certificate_name` | Name of the Certificate resource |
| `secret_name` | TLS Secret name (use for Gateway certificateRefs, Ingress tls.secretName, or CA Issuer ca_secret_name) |

## Related Components

- **KubernetesClusterIssuer** -- creates ACME ClusterIssuers for public TLS
- **KubernetesIssuer** -- creates namespace-scoped Issuers (SelfSigned, CA)
- **KubernetesCertManager** -- installs the cert-manager controller (prerequisite)

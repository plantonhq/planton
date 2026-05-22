# KubernetesIssuer

> Namespace-scoped cert-manager Issuer for CA and self-signed certificate signing

## Overview

KubernetesIssuer creates a cert-manager [Issuer](https://cert-manager.io/docs/concepts/issuer/) in a specific namespace for CA or self-signed certificate signing. Unlike ClusterIssuer (which is cluster-scoped and uses ACME DNS-01 challenges), an Issuer is namespace-scoped and supports simpler signing modes.

This component supports two issuer types:

- **CA** -- signs certificates using a CA keypair stored in a Kubernetes Secret (typically created by a KubernetesCertificate with `is_ca=true`)
- **SelfSigned** -- issues self-signed certificates, commonly used to bootstrap a CA chain or for development/testing

## Prerequisites

- **cert-manager must be installed** on the target cluster (via KubernetesCertManager or manually)
- The target namespace must already exist on the cluster
- For CA mode: a Kubernetes Secret containing `tls.crt` and `tls.key` must exist in the same namespace

## Quick Start

### Self-Signed Issuer (simplest)

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

### CA Issuer

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIssuer
metadata:
  name: my-ca-issuer
spec:
  namespace:
    value: my-namespace
  ca:
    caSecretName:
      value: my-ca-keypair
```

### Deploy

```bash
openmcf pulumi up --manifest issuer.yaml --stack org/project/env
```

## How It Works

1. The module reads the `issuer_type` oneof from the spec to determine CA or SelfSigned mode
2. Creates a single namespace-scoped cert-manager Issuer CR in the specified namespace
3. For CA mode, the Issuer references a Kubernetes Secret containing the CA keypair (`tls.crt` + `tls.key`)
4. For SelfSigned mode, no external dependencies are needed

## Naming Convention

The Issuer Kubernetes resource is named after `metadata.name`. This differs from KubernetesClusterIssuer (which uses `dns_domain` as the name) because namespace-scoped Issuers don't follow the ingress-domain convention.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `namespace` | StringValueOrRef | Yes | Namespace where the Issuer will be created |
| `ca.caSecretName` | StringValueOrRef | If CA | Secret containing CA certificate and private key |
| `selfSigned` | object | If SelfSigned | Empty object to select self-signed mode |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where the Issuer was created |
| `issuer_name` | Name of the created Issuer (equals `metadata.name`) |

## Related Components

- **KubernetesCertManager** -- installs the cert-manager controller (prerequisite)
- **KubernetesClusterIssuer** -- cluster-scoped ACME issuer for DNS-01 challenges
- **KubernetesCertificate** -- creates Certificate resources that reference this Issuer

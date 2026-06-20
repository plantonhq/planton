---
title: "Standard TLS Certificate via ClusterIssuer"
description: "This preset creates a standard TLS certificate signed by a ClusterIssuer (typically ACME / Let's Encrypt). This is the most common production pattern for public-facing services."
type: "preset"
rank: "01"
presetSlug: "01-cluster-issuer"
componentSlug: "kubernetescertificate"
componentTitle: "KubernetesCertificate"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard TLS Certificate via ClusterIssuer

This preset creates a standard TLS certificate signed by a ClusterIssuer (typically ACME / Let's Encrypt). This is the most common production pattern for public-facing services.

## When to Use

- You need a TLS certificate for a public-facing service
- You have a ClusterIssuer configured (via KubernetesClusterIssuer)
- You want automated certificate issuance and renewal

## Key Configuration Choices

- **ClusterIssuer reference** -- cluster-scoped issuer, typically configured for ACME DNS-01 challenges
- **Default duration** -- 90 days (Let's Encrypt standard), auto-renewed 15 days before expiry
- **Default private key** -- RSA 2048, PKCS1 encoding, rotated on every renewal

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Namespace where the Certificate resource will be created | Your application's namespace |
| `<your-hostname.example.com>` | DNS hostname for the certificate | Your application's ingress hostname |
| `<your-hostname>-tls` | Secret name for the signed certificate | Convention: hostname with `-tls` suffix |
| `<your-cluster-issuer>` | Name of the ClusterIssuer | KubernetesClusterIssuer's `cluster_issuer_name` output |

## Related Presets

- **02-wildcard** -- Use for wildcard certificates covering all subdomains
- **03-root-ca-bootstrap** -- Use for internal PKI with self-signed root CA

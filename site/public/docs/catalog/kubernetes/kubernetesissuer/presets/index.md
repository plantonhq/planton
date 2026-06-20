---
title: "Presets"
description: "Ready-to-deploy configuration presets for KubernetesIssuer"
type: "preset-list"
componentSlug: "kubernetesissuer"
componentTitle: "KubernetesIssuer"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-self-signed"
    rank: "01"
    title: "Self-Signed Issuer"
    excerpt: "This preset creates a self-signed Issuer, the simplest issuer type. Self-signed Issuers require no external dependencies and are commonly used to bootstrap a CA chain or for development/testing..."
  - slug: "02-ca"
    rank: "02"
    title: "CA Issuer"
    excerpt: "This preset creates a CA Issuer that signs certificates using a CA keypair stored in a Kubernetes Secret. The Secret must contain `tls.crt` (CA certificate) and `tls.key` (CA private key), and must..."
---

# KubernetesIssuer Presets

Ready-to-deploy configuration presets for KubernetesIssuer. Each preset is a complete manifest you can copy, customize, and deploy.

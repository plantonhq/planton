---
title: "Presets"
description: "Ready-to-deploy configuration presets for KubernetesCertificate"
type: "preset-list"
componentSlug: "kubernetescertificate"
componentTitle: "KubernetesCertificate"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-cluster-issuer"
    rank: "01"
    title: "Standard TLS Certificate via ClusterIssuer"
    excerpt: "This preset creates a standard TLS certificate signed by a ClusterIssuer (typically ACME / Let's Encrypt). This is the most common production pattern for public-facing services."
  - slug: "02-wildcard"
    rank: "02"
    title: "Wildcard TLS Certificate"
    excerpt: "This preset creates a wildcard certificate (`*.example.com`) via a ClusterIssuer. Wildcard certificates cover all subdomains under a domain, reducing the number of certificates needed for..."
  - slug: "03-root-ca-bootstrap"
    rank: "03"
    title: "Self-Signed Root CA Certificate (CA Bootstrap)"
    excerpt: "This preset creates a self-signed root CA certificate for bootstrapping an internal PKI. The resulting certificate Secret becomes the signing key for a CA Issuer, which can then issue leaf..."
---

# KubernetesCertificate Presets

Ready-to-deploy configuration presets for KubernetesCertificate. Each preset is a complete manifest you can copy, customize, and deploy.

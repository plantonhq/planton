---
title: "Presets"
description: "Ready-to-deploy configuration presets for Secret"
type: "preset-list"
componentSlug: "secret"
componentTitle: "Secret"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-opaque"
    rank: "01"
    title: "Opaque Secret"
    excerpt: "This preset creates an opaque Kubernetes Secret with arbitrary key-value data. The most common secret type, used for storing credentials, API keys, connection strings, and other sensitive data."
  - slug: "02-tls"
    rank: "02"
    title: "TLS Secret"
    excerpt: "This preset creates a Kubernetes TLS secret containing a certificate and private key pair. Used by ingress controllers, service meshes, and any workload that needs TLS termination."
  - slug: "03-docker-registry"
    rank: "03"
    title: "Docker Registry Secret"
    excerpt: "This preset creates a Docker registry authentication secret (`kubernetes.io/dockerconfigjson`) for pulling images from private container registries. Referenced by pods via `imagePullSecrets`."
---

# Secret Presets

Ready-to-deploy configuration presets for Secret. Each preset is a complete manifest you can copy, customize, and deploy.

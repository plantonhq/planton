---
title: "Presets"
description: "Ready-to-deploy configuration presets for Gateway"
type: "preset-list"
componentSlug: "gateway"
componentTitle: "Gateway"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-https-tls-terminate"
    rank: "01"
    title: "HTTPS Gateway with TLS Termination"
    excerpt: "A single HTTPS listener that terminates TLS at the Gateway using a certificate stored in a Kubernetes Secret (typically created by a cert-manager `KubernetesCertificate`). This is the most common..."
  - slug: "02-multi-protocol"
    rank: "02"
    title: "Multi-Protocol Gateway"
    excerpt: "A single Gateway exposing three listeners on distinct ports: cleartext HTTP (often used to redirect to HTTPS), HTTPS with TLS termination, and a raw TCP listener for a non-HTTP workload such as a..."
---

# Gateway Presets

Ready-to-deploy configuration presets for Gateway. Each preset is a complete manifest you can copy, customize, and deploy.

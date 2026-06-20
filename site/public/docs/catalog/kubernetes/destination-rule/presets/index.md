---
title: "Presets"
description: "Ready-to-deploy configuration presets for Destination Rule"
type: "preset-list"
componentSlug: "destination-rule"
componentTitle: "Destination Rule"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-circuit-breaking-outlier-detection"
    rank: "01"
    title: "Circuit Breaking & Outlier Detection"
    excerpt: "The canonical DestinationRule: protect a service by capping connection-pool size and ejecting hosts that keep returning errors. This is how you stop a struggling backend from taking down its callers..."
  - slug: "02-mtls-origination-egress"
    rank: "02"
    title: "mTLS Origination to an Egress Host"
    excerpt: "Configure the sidecar to originate mutual TLS to an external service, presenting client certificates loaded from a Kubernetes secret. This is how you let in-mesh workloads talk to an external..."
---

# Destination Rule Presets

Ready-to-deploy configuration presets for Destination Rule. Each preset is a complete manifest you can copy, customize, and deploy.

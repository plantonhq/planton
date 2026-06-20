---
title: "Presets"
description: "Ready-to-deploy configuration presets for Peer Authentication"
type: "preset-list"
componentSlug: "peer-authentication"
componentTitle: "Peer Authentication"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-namespace-strict-mtls"
    rank: "01"
    title: "Require Strict mTLS Across a Namespace"
    excerpt: "The canonical PeerAuthentication: require mutual TLS for every workload in a namespace. With no selector, the policy is the namespace default, so all in-mesh traffic to those workloads must arrive..."
  - slug: "02-workload-strict-with-plaintext-port"
    rank: "02"
    title: "Strict mTLS for One Workload, with a Plaintext Port"
    excerpt: "Require mTLS for a single selected workload, while exempting one port that must stay plaintext -- for example a health-check, metrics-scrape, or legacy port that a non-mesh client probes directly."
---

# Peer Authentication Presets

Ready-to-deploy configuration presets for Peer Authentication. Each preset is a complete manifest you can copy, customize, and deploy.

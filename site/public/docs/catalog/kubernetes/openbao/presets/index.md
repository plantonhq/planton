---
title: "Presets"
description: "Ready-to-deploy configuration presets for OpenBao"
type: "preset-list"
componentSlug: "openbao"
componentTitle: "OpenBao"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-dev-mode"
    rank: "01"
    title: "OpenBao Dev Mode"
    excerpt: "This preset deploys OpenBao (open-source Vault fork) in a simple configuration with the UI enabled and ingress access. Suitable for development and testing secrets management workflows."
  - slug: "02-production-ha"
    rank: "02"
    title: "Production OpenBao with HA"
    excerpt: "This preset deploys OpenBao in high-availability mode with 3 replicas, TLS encryption, and the sidecar injector for automatic secrets injection into pods."
---

# OpenBao Presets

Ready-to-deploy configuration presets for OpenBao. Each preset is a complete manifest you can copy, customize, and deploy.

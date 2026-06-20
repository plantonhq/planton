---
title: "Presets"
description: "Ready-to-deploy configuration presets for TLS Route"
type: "preset-list"
componentSlug: "tls-route"
componentTitle: "TLS Route"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-tls-passthrough-sni"
    rank: "01"
    title: "TLS Passthrough by SNI"
    excerpt: "The most common TLSRoute: match a TLS connection by its SNI hostname and forward it, unmodified (passthrough), to a single backend Service. The backend terminates TLS itself -- the Gateway never sees..."
  - slug: "02-tls-weighted-backends"
    rank: "02"
    title: "TLS Weighted Backends"
    excerpt: "A single TLSRoute rule that splits passthrough TLS connections across two backends by weight -- the building block for a canary or blue/green rollout of a TLS-terminating service. Because a TLSRoute..."
---

# TLS Route Presets

Ready-to-deploy configuration presets for TLS Route. Each preset is a complete manifest you can copy, customize, and deploy.

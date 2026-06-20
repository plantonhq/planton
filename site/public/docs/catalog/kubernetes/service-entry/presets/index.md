---
title: "Presets"
description: "Ready-to-deploy configuration presets for Service Entry"
type: "preset-list"
componentSlug: "service-entry"
componentTitle: "Service Entry"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-external-https-api"
    rank: "01"
    title: "Reach an External HTTPS API"
    excerpt: "The canonical ServiceEntry: register an external service (a SaaS API, a partner endpoint) so mesh workloads can call it as a first-class destination, with TLS routed by SNI and the host resolved via..."
  - slug: "02-static-mesh-internal-endpoints"
    rank: "02"
    title: "Bring Static Endpoints Into the Mesh"
    excerpt: "Register a service that has a fixed set of backing IPs (a VM-hosted database, a legacy service, an appliance) as a MESH_INTERNAL destination with STATIC resolution. Mesh workloads then reach it by..."
---

# Service Entry Presets

Ready-to-deploy configuration presets for Service Entry. Each preset is a complete manifest you can copy, customize, and deploy.

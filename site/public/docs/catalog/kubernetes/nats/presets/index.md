---
title: "Presets"
description: "Ready-to-deploy configuration presets for NATS"
type: "preset-list"
componentSlug: "nats"
componentTitle: "NATS"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-instance"
    rank: "01"
    title: "Single Instance NATS with JetStream"
    excerpt: "This preset deploys a single-node NATS server with JetStream enabled and the NATS Box diagnostic tool. JetStream provides persistent messaging, key-value store, and object store capabilities."
  - slug: "02-clustered"
    rank: "02"
    title: "Clustered NATS with JetStream"
    excerpt: "This preset deploys a 3-node NATS cluster with JetStream enabled for production messaging. The cluster provides high availability with automatic leader election for JetStream streams."
---

# NATS Presets

Ready-to-deploy configuration presets for NATS. Each preset is a complete manifest you can copy, customize, and deploy.

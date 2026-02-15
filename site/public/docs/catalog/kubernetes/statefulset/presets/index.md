---
title: "Presets"
description: "Ready-to-deploy configuration presets for StatefulSet"
type: "preset-list"
componentSlug: "statefulset"
componentTitle: "StatefulSet"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard StatefulSet"
    excerpt: "This preset deploys a single-replica StatefulSet without persistent volumes. Suitable for stateful applications that need stable pod identities and network names but do not require persistent storage."
  - slug: "02-with-persistent-volumes"
    rank: "02"
    title: "StatefulSet with Persistent Volumes"
    excerpt: "This preset deploys a 3-replica StatefulSet with persistent volume claims, a pod disruption budget, and data volume mounts. Each replica gets its own 10Gi persistent volume."
---

# StatefulSet Presets

Ready-to-deploy configuration presets for StatefulSet. Each preset is a complete manifest you can copy, customize, and deploy.

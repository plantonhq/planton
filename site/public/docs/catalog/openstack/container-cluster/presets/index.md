---
title: "Presets"
description: "Ready-to-deploy configuration presets for Container Cluster"
type: "preset-list"
componentSlug: "container-cluster"
componentTitle: "Container Cluster"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-dev-single-master"
    rank: "01"
    title: "Dev Cluster (Single Master)"
    excerpt: "This preset creates a minimal Magnum container cluster with a single master and one worker node. The cluster configuration (COE, networking, flavors) is defined by the referenced cluster template...."
  - slug: "02-ha-multi-master"
    rank: "02"
    title: "HA Cluster (Multi-Master)"
    excerpt: "This preset creates a production Magnum container cluster with 3 master nodes and 3 worker nodes. The 3-master configuration provides etcd quorum and control plane high availability. Flavors are..."
---

# Container Cluster Presets

Ready-to-deploy configuration presets for Container Cluster. Each preset is a complete manifest you can copy, customize, and deploy.

---
title: "Presets"
description: "Ready-to-deploy configuration presets for Dataproc Cluster"
type: "preset-list"
componentSlug: "dataproc-cluster"
componentTitle: "Dataproc Cluster"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-dev-jupyter"
    rank: "01"
    title: "Dev Jupyter"
    excerpt: "A lightweight development cluster with Jupyter Notebook for interactive data exploration and prototyping Spark jobs."
  - slug: "02-ha-production"
    rank: "02"
    title: "HA Production"
    excerpt: "A high-availability Dataproc cluster designed for production Spark workloads with 3 masters, SSD storage, CMEK encryption, and private networking."
  - slug: "03-cost-optimized-batch"
    rank: "03"
    title: "Cost-Optimized Batch"
    excerpt: "An ephemeral Dataproc cluster optimized for batch Spark jobs using Spot VMs for secondary workers, with aggressive auto-delete for cost control."
---

# Dataproc Cluster Presets

Ready-to-deploy configuration presets for Dataproc Cluster. Each preset is a complete manifest you can copy, customize, and deploy.

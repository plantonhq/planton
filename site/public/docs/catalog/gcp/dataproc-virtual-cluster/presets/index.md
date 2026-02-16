---
title: "Presets"
description: "Ready-to-deploy configuration presets for Dataproc Virtual Cluster"
type: "preset-list"
componentSlug: "dataproc-virtual-cluster"
componentTitle: "Dataproc Virtual Cluster"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-basic-spark-on-gke"
    rank: "01"
    title: "Basic Spark on GKE"
    excerpt: "A minimal Dataproc on GKE virtual cluster for running Spark workloads on an existing GKE cluster. Single node pool handles all workload types."
  - slug: "02-production-multi-pool"
    rank: "02"
    title: "Production Multi-Pool"
    excerpt: "A production-grade Dataproc on GKE virtual cluster with role-separated node pools, Spot executor autoscaling, and dedicated staging bucket."
  - slug: "03-metastore-integrated"
    rank: "03"
    title: "Metastore Integrated"
    excerpt: "A Dataproc on GKE virtual cluster with Hive Metastore integration and Spark History Server for catalog-based Spark SQL workloads and job history browsing."
---

# Dataproc Virtual Cluster Presets

Ready-to-deploy configuration presets for Dataproc Virtual Cluster. Each preset is a complete manifest you can copy, customize, and deploy.

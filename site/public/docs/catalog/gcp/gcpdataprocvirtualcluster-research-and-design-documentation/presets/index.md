---
title: "Presets"
description: "Ready-to-deploy configuration presets for GcpDataprocVirtualCluster - Research and Design Documentation"
type: "preset-list"
componentSlug: "gcpdataprocvirtualcluster-research-and-design-documentation"
componentTitle: "GcpDataprocVirtualCluster - Research and Design Documentation"
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

# GcpDataprocVirtualCluster - Research and Design Documentation Presets

Ready-to-deploy configuration presets for GcpDataprocVirtualCluster - Research and Design Documentation. Each preset is a complete manifest you can copy, customize, and deploy.

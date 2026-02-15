---
title: "Basic Spark on GKE"
description: "A minimal Dataproc on GKE virtual cluster for running Spark workloads on an existing GKE cluster. Single node pool handles all workload types."
type: "preset"
rank: "01"
presetSlug: "01-basic-spark-on-gke"
componentSlug: "gcpdataprocvirtualcluster-research-and-design-documentation"
componentTitle: "GcpDataprocVirtualCluster - Research and Design Documentation"
provider: "gcp"
icon: "package"
order: 1
---

# Basic Spark on GKE

A minimal Dataproc on GKE virtual cluster for running Spark workloads on an existing GKE cluster. Single node pool handles all workload types.

## When to Use

- Getting started with Dataproc on GKE
- Development and testing of Spark jobs on Kubernetes
- Small-scale Spark workloads that don't need role separation
- Quick proof-of-concept for Spark-on-GKE migration

## Key Configuration Choices

- **Single node pool with DEFAULT role**: All Spark pods (driver, executor, controller) run on one pool
- **Spark 3.5**: Latest stable Spark version for Dataproc on GKE
- **No auxiliary services**: No Metastore or Spark History Server
- **No staging bucket**: GCP auto-creates a default staging bucket
- **No namespace override**: Dataproc creates a namespace from the cluster name

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |
| `<your-gke-cluster-resource-id>` | Fully qualified GKE cluster ID (`projects/{project}/locations/{location}/clusters/{name}`) | GCP Console > Kubernetes Engine > Clusters |
| `<your-gke-node-pool-name>` | Name of an existing GKE node pool | GCP Console > Kubernetes Engine > Clusters > Node Pools |

## Prerequisites

1. A running GKE cluster in the same project and region
2. At least one GKE node pool with sufficient resources for Spark pods
3. Dataproc service agent has `roles/container.developer` on the GKE cluster
4. Workload Identity bindings for Spark pod authentication

## Related Presets

- **02-production-multi-pool**: Production setup with separate driver/executor pools and autoscaling
- **03-metastore-integrated**: Virtual cluster with Hive Metastore and Spark History Server

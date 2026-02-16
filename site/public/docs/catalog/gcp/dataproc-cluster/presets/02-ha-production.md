---
title: "HA Production"
description: "A high-availability Dataproc cluster designed for production Spark workloads with 3 masters, SSD storage, CMEK encryption, and private networking."
type: "preset"
rank: "02"
presetSlug: "02-ha-production"
componentSlug: "dataproc-cluster"
componentTitle: "Dataproc Cluster"
provider: "gcp"
icon: "package"
order: 2
---

# HA Production

A high-availability Dataproc cluster designed for production Spark workloads with 3 masters, SSD storage, CMEK encryption, and private networking.

## When to Use

- Production ETL pipelines that must not fail
- Long-running Spark Streaming applications
- Mission-critical data processing with SLA requirements
- Environments requiring CMEK encryption and private networking

## Key Configuration Choices

- **3 masters**: High-availability HDFS/YARN with automatic failover
- **5 workers (min 3)**: Baseline capacity with autoscaling floor
- **n2-standard-8**: Production-grade compute with 32 GB memory per node
- **SSD boot disks + 2 local SSDs**: Fast shuffle and spill performance
- **Internal IP only**: No public internet exposure
- **CMEK encryption**: Customer-managed keys for persistent disk encryption
- **1-hour graceful decommission**: Running jobs complete before scale-down
- **Component Gateway**: Authenticated web UI access to Spark/YARN

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |
| `<your-subnetwork-self-link>` | VPC subnetwork self-link | GCP Console > VPC > Subnets |
| `<your-dataproc-sa-email>` | Dataproc service account email | GCP Console > IAM > Service Accounts |
| `<your-kms-key-resource-name>` | Cloud KMS crypto key resource name | GCP Console > Security > Key Management |

## Related Presets

- **01-dev-jupyter**: Lightweight cluster for development
- **03-cost-optimized-batch**: Spot instances for batch processing

---
title: "Dev Jupyter"
description: "A lightweight development cluster with Jupyter Notebook for interactive data exploration and prototyping Spark jobs."
type: "preset"
rank: "01"
presetSlug: "01-dev-jupyter"
componentSlug: "dataproc-cluster"
componentTitle: "Dataproc Cluster"
provider: "gcp"
icon: "package"
order: 1
---

# Dev Jupyter

A lightweight development cluster with Jupyter Notebook for interactive data exploration and prototyping Spark jobs.

## When to Use

- Interactive Spark development and debugging
- Data exploration with Jupyter notebooks
- Prototyping ML pipelines before production deployment
- Learning and experimentation with Spark/Hadoop

## Key Configuration Choices

- **1 master + 2 workers**: Minimal cluster for development
- **e2-standard-4**: Cost-effective machine type for dev workloads
- **Jupyter enabled**: Interactive notebook access via Component Gateway
- **Component Gateway enabled**: Web UI access to Spark UI, YARN, HDFS NameNode
- **30-minute idle auto-delete**: Prevents runaway costs from forgotten clusters

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |

## Related Presets

- **02-ha-production**: High-availability cluster for production workloads
- **03-cost-optimized-batch**: Spot instances for cost-effective batch processing

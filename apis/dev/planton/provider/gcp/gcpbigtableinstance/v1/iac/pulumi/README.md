# GcpBigtableInstance Pulumi Module

This Pulumi module provisions a Google Cloud Bigtable instance with one or more clusters. Cloud Bigtable is a fully managed, wide-column NoSQL database designed for large analytical and operational workloads.

## Architecture

The module creates a single GCP resource:

1. **bigtable.Instance** — The Bigtable instance with embedded cluster configurations, labels, deletion protection, and optional CMEK encryption

The instance resource includes all cluster definitions inline (Bigtable's API bundles clusters within the instance resource). The module uses the `pulumi-gcp` (Google) provider with a **local backend** for state storage.

## Prerequisites

- GCP project with Cloud Bigtable API enabled
- Cloud KMS keys (if using CMEK encryption) with appropriate IAM permissions for the Bigtable service account

## Structure

```
iac/pulumi/
├── main.go              # Entry point: loads stack input, calls module.Resources
├── Pulumi.yaml          # Project definition
└── module/
    ├── main.go              # Resources(): orchestrates provider and bigtableInstance
    ├── locals.go            # Label construction, context extraction from stack input
    ├── bigtable_instance.go # bigtable.NewInstance with clusters, scaling, CMEK, labels
    └── outputs.go           # Export constants (instance_id, instance_name)
```

## Outputs

| Name | Description |
|------|-------------|
| `instance_id` | Fully qualified instance resource name |
| `instance_name` | Short instance name |

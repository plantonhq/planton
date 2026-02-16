# GcpBigtableInstance Pulumi Module Architecture

## Overview

This module provisions a Google Cloud Bigtable instance with one or more clusters using the `pulumi-gcp` provider. It creates a `bigtable.Instance` resource with embedded cluster configurations.

## File Organization

```
iac/pulumi/
├── main.go              # Entry point: loads stack input, calls module.Resources
├── Pulumi.yaml          # Project definition
└── module/
    ├── main.go              # Resources(): creates provider, calls bigtableInstance()
    ├── locals.go            # Label construction, context extraction from stack input
    ├── bigtable_instance.go # bigtable.NewInstance with clusters, scaling, CMEK, labels
    └── outputs.go           # Export constants (instance_id, instance_name)
```

## Control Flow

### main.go (entry point)

- Loads `GcpBigtableInstanceStackInput` from the Pulumi context
- Invokes `module.Resources(ctx, stackInput)` to provision resources

### module/main.go

- `initializeLocals()` — Builds `Locals` with GCP labels and target resource reference
- `pulumigoogleprovider.Get()` — Configures the Google provider
- `bigtableInstance()` — Creates the Bigtable instance with all clusters

### locals.go

- **Label construction** — Derives GCP labels from metadata: `openmcf-resource`, `openmcf-resource-name`, `openmcf-resource-kind`, plus optional `openmcf-organization`, `openmcf-environment`, `openmcf-resource-id`
- **Context extraction** — Extracts `GcpBigtableInstance` target and provider config from stack input

### bigtable_instance.go

Creates `bigtable.NewInstance` with:

- **Instance config** — Name, project, display name, deletion protection, force destroy
- **Labels** — GCP labels from locals for resource organization and cost allocation
- **Clusters** — Iterates over `spec.clusters` to build `InstanceClusterArray`:
  - **cluster_id** — Unique identifier within the instance
  - **zone** — GCP zone placement
  - **num_nodes** — Fixed node count (optional, mutually exclusive with autoscaling)
  - **storage_type** — SSD or HDD (optional, defaults to SSD)
  - **kms_key_name** — CMEK encryption key (optional)
  - **node_scaling_factor** — 1X or 2X scaling increments (optional)
  - **autoscaling_config** — Dynamic scaling with min/max nodes, CPU target, storage target (optional)

### outputs.go

Exports the following stack outputs:

- `instance_id` — Fully qualified instance resource name
- `instance_name` — Short instance name

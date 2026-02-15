# GcpAlloydbCluster Pulumi Module Architecture

## Overview

This module provisions a Google Cloud AlloyDB cluster with a bundled primary instance using the `pulumi-gcp` provider. It creates an `alloydb.Cluster` followed by an `alloydb.Instance` of type `PRIMARY`.

## File Organization

```
iac/pulumi/
├── main.go              # Entry point: loads stack input, calls module.Resources
├── Pulumi.yaml          # Project definition
└── module/
    ├── main.go          # Resources(): creates provider, calls cluster() then primaryInstance()
    ├── locals.go        # Label construction, context extraction from stack input
    ├── cluster.go       # alloydb.NewCluster with network_config, backup policies, encryption, maintenance
    ├── instance.go      # alloydb.NewInstance for PRIMARY type with machine config, query insights, client connection
    └── outputs.go       # Export constants (cluster_id, cluster_name, primary_instance_ip, etc.)
```

## Control Flow

### main.go (entry point)

- Loads `GcpAlloydbClusterStackInput` from the Pulumi context
- Invokes `module.Resources(ctx, stackInput)` to provision resources

### module/main.go

- `initializeLocals()` — Builds `Locals` with GCP labels and target resource reference
- `pulumigoogleprovider.Get()` — Configures the Google provider
- `cluster()` — Creates the AlloyDB cluster
- `primaryInstance()` — Creates the primary instance (depends on cluster)

### locals.go

- **Label construction** — Derives GCP labels from metadata: `openmcf-resource`, `openmcf-resource-name`, `openmcf-resource-kind`, plus optional `openmcf-organization`, `openmcf-environment`, `openmcf-resource-id`
- **Context extraction** — Extracts `GcpAlloydbCluster` target and provider config from stack input

### cluster.go

Creates `alloydb.NewCluster` with:

- **network_config** — VPC network (required) and optional `allocated_ip_range`
- **automated_backup_policy** — Enabled flag, backup window, location, quantity-based or time-based retention, weekly schedule, optional backup CMEK
- **continuous_backup_config** — Enabled flag, recovery window days, optional continuous backup CMEK
- **encryption_config** — Cluster-level CMEK for data at rest (optional)
- **maintenance_update_policy** — Maintenance window (day, start hour)
- **initial_user** — Optional initial superuser (password, user)
- **deletion_protection** — Boolean flag
- **database_version** — PostgreSQL version (e.g., POSTGRES_15)
- **display_name** — Human-readable name

### instance.go

Creates `alloydb.NewInstance` for PRIMARY type with:

- **machine_config** — Either `cpu_count` or `machine_type` (mutually exclusive)
- **availability_type** — ZONAL or REGIONAL
- **database_flags** — PostgreSQL server parameters
- **query_insights_config** — Query plans per minute, query string length, record application tags, record client address
- **client_connection_config** — `require_connectors` (enforce Auth Proxy / Language Connectors), `ssl_mode` (ENCRYPTED_ONLY or ALLOW_UNENCRYPTED_AND_ENCRYPTED)
- **display_name** — Human-readable name

The instance is created with `DependsOn(createdCluster)` so it is provisioned after the cluster.

### outputs.go

Exports the following stack outputs:

- `cluster_id` — Cluster resource name
- `cluster_name` — Short cluster name
- `primary_instance_ip` — Private IP of the primary instance
- `primary_instance_name` — Primary instance resource name
- `database_version` — PostgreSQL version
- `state` — Cluster state

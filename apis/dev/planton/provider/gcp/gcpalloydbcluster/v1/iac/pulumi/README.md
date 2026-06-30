# GcpAlloydbCluster Pulumi Module

This Pulumi module provisions a Google Cloud AlloyDB cluster with a bundled primary instance. AlloyDB is a fully managed, PostgreSQL-compatible database service.

## Architecture

The module creates two GCP resources in sequence:

1. **alloydb.Cluster** — Logical container with network configuration, backup policies, encryption, and maintenance settings
2. **alloydb.Instance** — Primary compute instance (type `PRIMARY`) that serves queries

Both resources use the `pulumi-gcp` (Google) provider. The module expects a **local backend** for state storage.

## Prerequisites

- VPC network with [Private Service Access](https://cloud.google.com/vpc/docs/private-service-access) configured
- GCP project with AlloyDB API enabled

## Structure

```
iac/pulumi/
├── main.go              # Entry point: loads stack input, calls module.Resources
├── Pulumi.yaml          # Project definition
└── module/
    ├── main.go          # Resources(): orchestrates provider, cluster, then primaryInstance
    ├── locals.go        # Label construction, context extraction from stack input
    ├── cluster.go       # alloydb.NewCluster with network, backup, encryption, maintenance
    ├── instance.go      # alloydb.NewInstance for PRIMARY type
    └── outputs.go       # Export constants (cluster_id, primary_instance_ip, etc.)
```

## Outputs

| Name | Description |
|------|-------------|
| `cluster_id` | Fully qualified cluster resource name |
| `cluster_name` | Short cluster name |
| `primary_instance_ip` | Private IP of the primary instance |
| `primary_instance_name` | Fully qualified primary instance resource name |
| `database_version` | PostgreSQL version (e.g., POSTGRES_15) |
| `state` | Cluster state (CREATING, READY, etc.) |

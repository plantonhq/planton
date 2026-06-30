# GcpSpannerInstance: Research and Design Documentation

## Cloud Spanner Overview

Google Cloud Spanner is a fully managed, globally distributed, strongly consistent relational database service. It combines the benefits of relational database structure (schemas, SQL, ACID transactions) with non-relational horizontal scale. Spanner is designed for workloads that need:

- **Strong consistency** across globally distributed data
- **Horizontal scalability** without manual sharding
- **High availability** (99.999% SLA with multi-region configurations)
- **Relational semantics** (SQL, schemas, foreign keys, secondary indexes)

### Spanner's Architecture: Instances, Databases, and Tables

Spanner uses a three-level hierarchy:

1. **Instance** -- The unit of compute and storage capacity allocation. An instance defines where your data lives (via the instance configuration) and how much compute is available (via nodes or processing units). This is what this component provisions.

2. **Database** -- Lives within an instance. Databases have their own schemas, IAM policies, and backup configurations. Multiple databases can share an instance's compute capacity.

3. **Tables/Indexes** -- Live within a database. Schema management (DDL) is typically handled by application migration tools, not IaC.

### Why Instance and Database Are Separate Components

Spanner instances and databases have fundamentally different lifecycles:

- **Instances** are infrastructure: they define capacity and geographic placement. They are rarely changed after creation. They are shared across multiple databases.
- **Databases** are application-level: they hold schemas, data, and application-specific configurations. They are created and destroyed as applications evolve.

Bundling them would force users to redeploy their entire database whenever they want to adjust instance capacity (or vice versa). The split also matches the Terraform and Pulumi resource model.

## Deployment Landscape

### Methods of Provisioning Cloud Spanner

| Method | Strengths | Weaknesses |
|---|---|---|
| GCP Console | Visual, discoverable | Not repeatable, no version control |
| gcloud CLI | Scriptable, quick | Imperative, state drift |
| Terraform (google_spanner_instance) | Declarative, state management | HCL syntax, provider version coupling |
| Pulumi (spanner.Instance) | Declarative, real programming languages | Smaller community than Terraform |
| Planton (GcpSpannerInstance) | Multi-cloud consistency, infra-chart composability | Newer ecosystem |

### Planton's Value Add

Planton's GcpSpannerInstance component provides:

1. **Cross-resource references** via `StringValueOrRef` -- the project_id can reference a GcpProject output, enabling infra-chart composition
2. **Validation before deployment** -- CEL rules catch mutual exclusion violations (nodes vs processing_units vs autoscaling) and FREE_INSTANCE restrictions before any API calls
3. **Consistent metadata** -- Framework labels, KRM envelope, and output contracts that match every other Planton component
4. **Preset configurations** -- Ready-to-deploy templates for common scenarios (free, regional production, autoscaling)

## Capacity Model Deep Dive

### Nodes vs Processing Units

Historically, Spanner capacity was measured in **nodes**. Each node provides approximately:
- 10,000 QPS for reads
- 2,000 QPS for writes
- 10 TB of storage

Processing units (PUs) were introduced for finer-grained sizing:
- 1 node = 1,000 processing units
- For values < 1,000, PUs must be in multiples of 100
- For values >= 1,000, PUs must be in multiples of 1,000

### Autoscaling

Spanner autoscaling adjusts compute capacity automatically based on utilization targets:

- **High-priority CPU utilization target** -- Recommended: 65%. Triggers scale-up when exceeded.
- **Storage utilization target** -- Recommended: 80%. Prevents running out of storage.
- **Min/max bounds** -- Prevents runaway scaling. Use same unit for both (nodes or PUs).

When autoscaling is enabled, `num_nodes` and `processing_units` become read-only outputs reflecting the current allocation.

### Free Instances

Free instances provide zero-cost Spanner for development:
- ~10 GB storage
- Equivalent to ~100 processing units
- Cannot set edition, capacity fields, or AUTOMATIC backup schedule
- One free instance per GCP billing account

## Edition Model

Spanner editions control available features and SLA:

| Feature | STANDARD | ENTERPRISE | ENTERPRISE_PLUS |
|---|---|---|---|
| Regional SLA | 99.99% | 99.99% | 99.99% |
| Multi-region SLA | N/A | 99.999% | 99.999% |
| Granular instance sizing | Limited | Full | Full |
| Advanced security | Basic | Enhanced | Full |
| Pricing | Lowest | Medium | Highest |

## Instance Configuration Reference

### Regional Configurations

Format: `regional-{region}` (e.g., `regional-us-central1`, `regional-europe-west1`)

All replicas in a single GCP region. Lower latency, lower cost, 99.99% SLA.

### Multi-Region Configurations

| Config | Regions | Use Case |
|---|---|---|
| `nam-eur-asia1` | North America + Europe + Asia | Global coverage |
| `nam6` | 3 US regions | US-only multi-region |
| `nam7` | North America regions | North American coverage |
| `eur3` | European regions | EU-only multi-region |

Multi-region configs replicate data across regions for disaster recovery and lower read latency globally.

## 80/20 Scoping Rationale

### What We Include

- **Core capacity fields** -- num_nodes, processing_units, autoscaling_config (covers 100% of capacity allocation patterns)
- **Instance type** -- PROVISIONED vs FREE_INSTANCE (covers development and production)
- **Edition** -- STANDARD, ENTERPRISE, ENTERPRISE_PLUS (significant pricing and feature implications)
- **Default backup schedule** -- Controls automatic backup behavior for new databases
- **Force destroy** -- Safety control for instance teardown with backups

### What We Deliberately Exclude

| Feature | Reason |
|---|---|
| Asymmetric autoscaling | Advanced per-replica scaling for multi-region. Deep nesting (5 levels), niche use case. Can add in v2. |
| User-defined labels | Framework labels are applied automatically. User labels are a cross-cutting enhancement for all components. |
| Instance-level IAM bindings | IAM is typically managed at the database or project level, not instance level. |
| Replication type/topology | Fully determined by the `config` field. Not user-configurable. |

### Downstream Composition

GcpSpannerInstance outputs `instance_name` specifically for GcpSpannerDatabase to reference via `StringValueOrRef`. In the `gcp-spanner-application` infra chart:

```
GcpProject (project_id)
  └── GcpSpannerInstance (instance_name)
        └── GcpSpannerDatabase (database)
```

## Immutability Notes

The following fields are **ForceNew** (changing them requires recreating the instance):
- `instance_name`
- `config`
- `project_id`

Capacity fields (num_nodes, processing_units, autoscaling_config) are mutable -- you can scale up/down without recreation.

## Provider Version Note

The Terraform module requires Google provider `~> 6.0` (not `~> 5.0`). Provider v5.x does not support the `instance_type`, `edition`, and `default_backup_schedule_type` fields. The Pulumi module uses the standard `pulumi-gcp/sdk/v9` which supports all fields.

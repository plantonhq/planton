# GcpSpannerInstance

Provisions a [Google Cloud Spanner](https://cloud.google.com/spanner) instance -- the unit of compute and storage capacity that hosts one or more Spanner databases.

## What It Does

Cloud Spanner is a fully managed, globally distributed, strongly consistent relational database. This component creates and manages a Spanner **instance**, which defines where your databases live and how much compute capacity is allocated to them.

A Spanner instance is **not** a database. It is the container of databases -- think of it as the "server" that hosts your databases. Capacity is allocated at the instance level and shared across all databases within it.

## When to Use

- You need a relational database that scales horizontally without sharding your application
- Your workload requires strong consistency across regions (globally distributed)
- You need 99.999% availability for mission-critical data
- You want a managed database that handles replication, failover, and scaling automatically

## Key Configuration

### Instance Configuration (`config`)

The `config` field determines where your data is replicated:

- **Regional** (e.g., `regional-us-central1`) -- All replicas in one region. Lower latency, lower cost. 99.99% SLA.
- **Multi-region** (e.g., `nam-eur-asia1`, `nam6`) -- Replicas across multiple regions. Higher availability (99.999% SLA with ENTERPRISE_PLUS), higher cost, higher write latency.

This field is **immutable** -- changing it requires recreating the instance.

### Capacity

Exactly one of three capacity methods must be chosen:

| Method | When to Use |
|---|---|
| `num_nodes` | Simple allocation. 1 node = ~10,000 QPS reads. Good for predictable workloads. |
| `processing_units` | Finer-grained. 1 node = 1000 PUs. Good for smaller workloads or precise sizing. |
| `autoscaling_config` | Automatic scaling. Set min/max bounds and CPU/storage targets. Best for variable workloads. |

### Editions

| Edition | SLA | Features |
|---|---|---|
| STANDARD | 99.99% regional | Cost-optimized, good for most workloads |
| ENTERPRISE | 99.99% regional | Granular instance sizing, advanced security |
| ENTERPRISE_PLUS | 99.999% multi-region | Highest availability, advanced compliance |

### Free Instances

Set `instance_type: FREE_INSTANCE` for a zero-cost development instance with ~10 GB storage. Free instances cannot set capacity, edition, or automatic backups.

## Outputs

| Output | Description |
|---|---|
| `instance_id` | Fully qualified path (`projects/{project}/instances/{name}`) |
| `instance_name` | Short name (used by GcpSpannerDatabase to reference this instance) |
| `state` | CREATING or READY |

## Relationships

- **Depends on**: GcpProject (project_id)
- **Referenced by**: GcpSpannerDatabase (instance field)

## Deployment

```shell
planton apply -f spanner-instance.yaml
```

For copy-paste ready manifests, see [examples.md](examples.md).

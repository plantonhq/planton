# Production PostgreSQL HA

This preset creates a production-grade PostgreSQL database cluster with three nodes for high availability, VPC isolation for secure private access, automatic storage growth, and a Sunday-night maintenance window, on PostgreSQL 16 with 2 vCPU / 4 GB nodes. Suitable for mission-critical applications requiring automatic failover.

## When to Use

- Production applications requiring database high availability
- Workloads needing automatic failover when a primary node fails
- Environments where database traffic must stay within a private VPC

## Key Configuration Choices

- **Three nodes** (`nodeCount: 3`) -- primary plus two standby nodes for HA. DigitalOcean provides automatic failover within the cluster.
- **PostgreSQL 16** (`engine: pg`, `engineVersion: "16"`) -- latest stable major version with extended support.
- **VPC placement** (`vpc.valueFrom`) -- references a `DigitalOceanVpc` resource named `my-vpc`; rename it to your VPC resource, or replace the block with `value: <uuid>` for an unmanaged VPC. Keeps database traffic off the public internet.
- **Storage autoscale** (`storageAutoscale`) -- grows the disk automatically at 80% usage, applied by the Terraform provisioner (the Pulumi bridge does not support it yet and rejects it loudly).
- **Maintenance window** (`maintenanceWindow`) -- pins automatic updates to Sunday 02:00 UTC instead of a DigitalOcean-chosen slot.
- **Node size** (`sizeSlug: db-s-2vcpu-4gb`) -- general-purpose sizing; scale up for heavier workloads.

## Related Presets

- **02-postgresql-dev** -- Use instead for dev/test where HA and VPC are unnecessary
- **03-valkey** -- Use for caching workloads instead of relational data

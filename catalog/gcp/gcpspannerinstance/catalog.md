# GCP Spanner Instance

Deploys a Cloud Spanner instance with configurable compute capacity (fixed nodes, processing units, or managed autoscaling), edition selection (Standard, Enterprise, Enterprise Plus), regional or multi-region placement, and automatic backup scheduling for new databases. The instance is the unit of compute and storage allocation: it pins a geographic topology and a capacity envelope that every database on it shares. The instance name, configuration, and project are immutable — everything else, including capacity and edition, updates in place with no downtime.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Spanner API enablement** -- the module enables `spanner.googleapis.com` on the target project before creating the instance, so a fresh project works without manual API setup
- **Spanner Instance** -- a managed instance in the specified GCP project with the chosen instance configuration (regional or multi-region), edition, and compute capacity allocation
- **Compute Capacity** -- allocated via one of three mutually exclusive modes: fixed nodes (each node = 1000 processing units, roughly 10,000 QPS of reads), processing units (finer-grained, multiples of 100 below 1000), or autoscaling with CPU and storage utilization targets and optional per-replica overrides
- **Automatic Backup Scheduling** -- configured only when `defaultBackupScheduleType` is set to `AUTOMATIC`; GCP attaches a default backup schedule to each new database created on this instance
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Spanner instance will be created. Provide the project ID directly, reference a GcpProject Infra Component via ValueFromRef, or omit `projectId` to use the provider connection's default project. The module enables the Cloud Spanner API itself — no manual API activation is needed.
- **Billing enabled** on the project. Spanner has no always-free provisioned tier: the smallest billable footprint is 100 processing units, and a `FREE_INSTANCE` is limited to one per billing account.

## Deploy

### Console

Open the deployment store, find **GCP Spanner Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Free Instance** preset in the [Presets](#presets) tab for a zero-cost development instance, or the **Regional Production** preset for production workloads.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSpannerInstance
metadata:
  name: app-spanner
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-prod-12345
  instanceName: app-spanner-prod
  config: regional-us-central1
  displayName: App Spanner Production
  numNodes: 1
  edition: ENTERPRISE
```

```shell
planton apply -f spanner-instance.yaml
```

This creates a 1-node Enterprise instance in `us-central1` with no automatic backup scheduling; databases are created separately with GcpSpannerDatabase. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the instance to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the Spanner instance with the resolved project ID.

## Key Configuration

These are the most important decisions when configuring a Spanner instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Instance configuration** -- The `config` field is the one decision you cannot walk back: it pins geographic placement and replication topology for every database that will ever live on the instance. Regional configs (e.g., `regional-us-central1`) give the lowest write latency and cost; multi-region configs (e.g., `nam6`, `nam-eur-asia1`) buy availability across regional failures — and the 99.999% SLA on ENTERPRISE_PLUS — at higher cost and higher write latency, since the write quorum spans regions. Changing it means a new instance and a data migration.

**Capacity mode** -- Three mutually exclusive options: `numNodes` for coarse allocation (1 node = 1000 processing units, roughly 10,000 QPS of reads or 2,000 QPS of writes), `processingUnits` for finer control (multiples of 100 below 1000, multiples of 1000 above), or `autoscalingConfig` with min/max bounds and utilization targets. Capacity changes are online with no downtime, so err small and grow — 100-500 processing units is a sane start for anything unproven. If no capacity field is set on a PROVISIONED instance, GCP defaults to 1 node.

**Autoscaling bounds** -- The autoscaler operates strictly inside `autoscalingLimits`: an undersized max is a ceiling you hit during your first traffic spike, an oversized min is a bill you pay every idle hour. Google recommends a 65% high-priority CPU target for regional configurations and about 45% for multi-region, where failover headroom matters. On multi-region instances, `asymmetricAutoscalingOptions` scale a read-heavy replica region independently instead of sizing every region for the hottest one (per-replica processing-unit bounds must be multiples of 1000).

**Edition** -- `STANDARD` is cost-optimized for single-region basics. `ENTERPRISE` unlocks granular sizing, asymmetric autoscaling, and incremental backups. `ENTERPRISE_PLUS` adds the 99.999% multi-region SLA and advanced compliance controls. Upgrades apply in place; downgrades require first disabling every higher-edition feature — treat a downgrade as a project, not a field edit. Cannot be set for FREE_INSTANCE.

**Instance type** -- `PROVISIONED` (default) requires explicit capacity. `FREE_INSTANCE` provides one zero-cost instance per billing account with about 10 GB of storage; it must not set capacity, edition, or automatic backups (the spec rejects all three pre-deploy). The upgrade to PROVISIONED works in place — the reverse does not exist, so the free slot is spent once.

**Destroy semantics** -- Two layered guards. `forceDestroy: false` (the default) makes destroy fail while any database on the instance holds a backup, so the last restore point never rides along with a stack teardown. `deletionPolicy` governs the instance itself: `PREVENT` for the instance a whole topology depends on, `ABANDON` to leave it running (and billing) in GCP outside Planton management. The two compose — a DELETE policy still stops on backups unless `forceDestroy` is armed.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_id` | Fully qualified instance ID (`projects/{p}/instances/{i}`) | IAM bindings, API calls, monitoring dashboards |
| `instance_name` | Short instance name | GcpSpannerDatabase and GcpSpannerBackupSchedule `instance` fields, client library connections |
| `state` | Instance state (`CREATING` or `READY`) | Deployment validation, health checks |
| `config` | The instance configuration the instance was created with (e.g. `regional-us-central1`, `nam6`) | Choosing the matching CMEK key shape (regional vs one-key-per-region) for databases and backup schedules |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Free instance** -- Zero-cost instance for development and testing with about 10 GB of storage and limited throughput. One per billing account, and the application code is unchanged when you later move to provisioned capacity. Start from the **Free Instance** preset.

**Regional production** -- Single-node Enterprise instance with automatic backup scheduling for new databases. Suitable for production workloads with predictable capacity in a single region — significantly cheaper than any multi-region topology. Start from the **Regional Production** preset.

**Autoscaling multi-region** -- Enterprise instance on a multi-region config with managed autoscaling and an asymmetric override that scales one read-heavy replica region independently. The shape for variable traffic where fixed capacity means paying for the peak. Start from the **Autoscaling Production (Multi-Region)** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the Spanner instance is created
- [**GCP Spanner Database**](/infra-catalog/gcp-spanner-database) -- databases live on this instance and reference it by `instance_name`
- [**GCP Spanner Backup Schedule**](/infra-catalog/gcp-spanner-backup-schedule) -- explicit backup schedules for databases on this instance

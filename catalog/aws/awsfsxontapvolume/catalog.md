# AWS FSx ONTAP Volume

Deploys a data volume within an FSx for NetApp ONTAP Storage Virtual Machine, with configurable tiering policies, SnapLock WORM compliance, FlexGroup distribution, and storage efficiency features. The volume is what clients actually mount: it joins the SVM namespace at its junction path, and its tiering policy is the ONTAP family's cost lever, moving cold data to the elastic capacity pool per volume rather than per file system.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ONTAP Volume** -- a data container within the specified SVM, mounted at the configured junction path, with the chosen volume style (FlexVol or FlexGroup) and security style
- **Tiering Policy** -- created only when `tieringPolicy` is provided; controls automatic data movement between primary SSD and capacity pool storage based on access patterns
- **SnapLock Configuration** -- created only when `snaplockConfiguration` is provided; enables WORM (Write Once Read Many) immutable storage with configurable retention periods, autocommit, and privileged delete settings
- **Aggregate Configuration** -- created only when `aggregateConfiguration` is provided and `volumeStyle` is FLEXGROUP; distributes the volume across multiple aggregates for parallel throughput
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An FSx ONTAP Storage Virtual Machine** -- the volume's parent SVM must be provisioned first. Provide the SVM ID directly or reference an AwsFsxOntapStorageVirtualMachine Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS FSx ONTAP Volume**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **General Purpose ONTAP Volume** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapVolume
metadata:
  name: app-data
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  storageVirtualMachineId:
    value: "svm-0123456789abcdef0"
  name: vol_data
  sizeInMegabytes: 102400
  junctionPath: /data
  storageEfficiencyEnabled: true
```

```shell
planton apply -f fsx-ontap-volume.yaml
```

This creates a 100 GB FlexVol volume mounted at `/data` with ONTAP storage efficiency (deduplication, compression, compaction) enabled. No tiering, SnapLock, or FlexGroup is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the volume to an SVM deployed in the same InfraPipeline:

```yaml
spec:
  storageVirtualMachineId:
    valueFrom:
      kind: AwsFsxOntapStorageVirtualMachine
      name: app-svm
      fieldPath: status.outputs.svm_id
```

The InfraPipeline resolves the dependency graph, deploys the SVM first, then provisions the volume with the resolved SVM ID.

## Key Configuration

These are the most important decisions when configuring an ONTAP volume. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Volume size** -- Exactly one of two sizing arms: `sizeInMegabytes` (the everyday arm, minimum 20 MB) or `sizeInBytes` (byte-precise, up to ~20 PiB -- the only arm that reaches past int32 megabytes). ONTAP supports thin provisioning, so logical size can exceed physical capacity. Size grows in place.

**Volume style** -- Defaults to FLEXVOL (single aggregate). Choose FLEXGROUP for high-throughput workloads (data lakes, genomics, media rendering) that benefit from striping across multiple aggregates. FlexGroup requires `aggregateConfiguration`. ForceNew -- cannot change after creation.

**Tiering policy** -- Controls cost optimization by moving infrequently accessed data to capacity pool storage. AUTO tiers data not accessed for the configured cooling period. SNAPSHOT_ONLY tiers only snapshot data. NONE keeps all data on SSD. ALL stores everything on capacity pool for lowest cost.

**SnapLock** -- Enables WORM immutable storage for regulatory compliance. COMPLIANCE mode is irrevocable -- no one can delete files before retention expires. ENTERPRISE mode allows privileged deletion. ForceNew for `snaplockType` -- choosing the wrong mode requires volume recreation.

**Junction path** -- The mount point in the SVM namespace (e.g., `/data`). Without a junction path, the volume exists but is not accessible via NFS/SMB. Must start with `/` and be unique within the SVM.

**Storage efficiency** -- Enable `storageEfficiencyEnabled` for most workloads. ONTAP deduplication, compression, and compaction reduce physical storage consumption. Disable only for pre-compressed or encrypted data where the CPU overhead provides no benefit.

**Final backup tags** -- When `skipFinalBackup` is off, `finalBackupTags` are applied to the backup taken on deletion, so cost-allocation and retention markers survive the volume itself.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsFsxOntapStorageVirtualMachine** | `storageVirtualMachineId` | `status.outputs.svm_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `volume_id` | Volume identifier (e.g., fsvol-0123456789abcdef0) | CloudWatch metrics, AWS API references |
| `arn` | Amazon Resource Name of the volume | IAM policies for resource-level permissions |
| `uuid` | ONTAP UUID for SnapMirror and REST API operations | Replication relationships, cross-cluster identification |
| `file_system_id` | Parent file system ID | Cross-referencing with file system resources |
| `flexcache_endpoint_type` | FlexCache role (NONE, ORIGIN, or CACHE) | FlexCache topology identification |
| `ontap_volume_type` | Volume type confirmation (RW or DP) | Operational metadata |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**General purpose volume** -- 100 GB FlexVol with UNIX security style, storage efficiency enabled, and AUTO tiering with a 31-day cooling period. The standard configuration for NFS application data volumes. Start from the **General Purpose ONTAP Volume** preset.

**Compliance SnapLock volume** -- 500 GB volume with SnapLock COMPLIANCE for immutable record retention (SEC 17a-4, HIPAA, FINRA). 5-year default retention, 1-day autocommit, SNAPSHOT_ONLY tiering. Start from the **Compliance SnapLock Volume** preset.

**High-performance FlexGroup volume** -- 1 TB FlexGroup distributed across 2 aggregates with 8 constituents each for parallel I/O. No tiering -- all data on SSD for consistent latency. Designed for data lakes, genomics, and media workflows. Start from the **High-Performance FlexGroup Volume** preset.

## Works With

- [**AWS FSx ONTAP Storage Virtual Machine**](/infra-catalog/aws-fsx-ontap-storage-virtual-machine) -- provides the parent SVM that hosts this volume's protocol endpoints and namespace
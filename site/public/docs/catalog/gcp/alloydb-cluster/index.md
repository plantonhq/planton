---
title: "AlloyDB Cluster"
description: "AlloyDB Cluster deployment documentation"
icon: "package"
order: 100
componentName: "gcpalloydbcluster"
---

# GCP AlloyDB Cluster

Deploys an AlloyDB cluster with a bundled primary instance, automated and continuous backup policies, optional CMEK encryption at three levels (data, backups, PITR), and VPC-based private networking via Private Service Access. The primary instance is bundled because a cluster without one cannot serve queries.

## What Gets Created

When you deploy a GcpAlloydbCluster resource, Planton provisions:

- **AlloyDB Cluster** — a `google_alloydb_cluster` resource in the specified region with network configuration, backup policies, encryption, and maintenance settings
- **Primary Instance** — a `google_alloydb_instance` of type `PRIMARY` attached to the cluster, with configurable machine size, availability type, query insights, and client connection settings
- **Automated Backup Policy** — created only when `automatedBackupPolicy` is specified, configures periodic snapshot backups with retention and schedule settings
- **Continuous Backup** — enabled by default with 14-day PITR window, configurable via `continuousBackupConfig`
- **CMEK Encryption** — created only when `kmsKeyName` fields are specified, encrypts cluster data, backup snapshots, and continuous backup data independently

## Prerequisites

- **GCP credentials** configured via service account key or Planton provider config
- **A VPC network** with Private Service Access configured (the network must have a private services connection to `servicenetworking.googleapis.com`)
- **A GCP project** with the AlloyDB API enabled (`alloydb.googleapis.com`)
- **A Cloud KMS key** if enabling CMEK encryption (must be in the same region as the cluster)

## Quick Start

Create a file `alloydb.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpAlloydbCluster
metadata:
  name: my-alloydb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpAlloydbCluster.my-alloydb
spec:
  projectId:
    value: my-gcp-project
  clusterName: my-alloydb-cluster
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/default
  primaryInstance:
    instanceId: my-primary
    cpuCount: 2
    availabilityType: ZONAL
```

Deploy:

```shell
planton apply -f alloydb.yaml
```

This creates an AlloyDB cluster with a 2-CPU primary instance in `us-central1`, connected to the default VPC with GCP-managed encryption and default backup policies.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the cluster is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `clusterName` | `string` | Name of the AlloyDB cluster. Becomes the GCP resource ID. Immutable after creation. | Pattern: `^[a-z][a-z0-9-]{0,61}[a-z0-9]$`, 2-63 chars |
| `location` | `string` | GCP region (e.g., `us-central1`). Immutable after creation. | Required |
| `network` | `StringValueOrRef` | VPC network with Private Service Access. Can reference a GcpVpc resource via `valueFrom`. Immutable after creation. | Required |
| `primaryInstance.instanceId` | `string` | Name of the primary instance. Immutable after creation. | Pattern: `^[a-z][a-z0-9-]{0,61}[a-z0-9]$`, 2-63 chars |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `allocatedIpRange` | `string` | — | Named IP range for Private Service Access. Use when IP ranges are pre-planned. |
| `databaseVersion` | `string` | Latest | PostgreSQL version: `POSTGRES_14`, `POSTGRES_15`, or `POSTGRES_16`. |
| `displayName` | `string` | — | Human-readable name for the cluster. |
| `initialUser.password` | `string` | — | Initial superuser password. Min 8 characters. Required when `initialUser` is set. |
| `initialUser.user` | `string` | `postgres` | Initial superuser name. |
| `automatedBackupPolicy.enabled` | `bool` | `true` | Enable or disable automated periodic backups. |
| `automatedBackupPolicy.backupWindow` | `string` | `3600s` | Backup window duration in seconds (e.g., `3600s`). |
| `automatedBackupPolicy.location` | `string` | Same as cluster | Region where backups are stored. |
| `automatedBackupPolicy.quantityBasedRetentionCount` | `int` | — | Number of backups to retain. Mutually exclusive with `timeBasedRetentionPeriod`. |
| `automatedBackupPolicy.timeBasedRetentionPeriod` | `string` | — | Retention duration (e.g., `1209600s` for 14 days). Mutually exclusive with `quantityBasedRetentionCount`. |
| `automatedBackupPolicy.weeklySchedule.daysOfWeek` | `string[]` | Daily | Days to run backups: `MONDAY` through `SUNDAY`. |
| `automatedBackupPolicy.weeklySchedule.startHour` | `int` | — | UTC hour (0-23) to start backups. |
| `automatedBackupPolicy.encryptionKmsKeyName` | `StringValueOrRef` | Google-managed | KMS key for backup encryption. Can reference GcpKmsKey via `valueFrom`. |
| `continuousBackupConfig.enabled` | `bool` | `true` | Enable continuous backup for PITR. |
| `continuousBackupConfig.recoveryWindowDays` | `int` | `14` | PITR recovery window in days (1-35). |
| `continuousBackupConfig.encryptionKmsKeyName` | `StringValueOrRef` | Google-managed | KMS key for continuous backup encryption. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | KMS key for cluster data-at-rest encryption. Immutable. Can reference GcpKmsKey via `valueFrom`. |
| `maintenanceWindow.day` | `string` | GCP-selected | Day of week: `MONDAY` through `SUNDAY`. |
| `maintenanceWindow.startHour` | `int` | GCP-selected | UTC hour (0-23) for maintenance start. |
| `deletionProtection` | `bool` | `true` | Prevents accidental cluster destruction. |
| `primaryInstance.cpuCount` | `int` | — | Number of CPUs. Mutually exclusive with `machineType`. |
| `primaryInstance.machineType` | `string` | — | Machine type (e.g., `n2-highmem-4`). Mutually exclusive with `cpuCount`. |
| `primaryInstance.availabilityType` | `string` | GCP default | `ZONAL` or `REGIONAL`. REGIONAL provides automatic failover. |
| `primaryInstance.databaseFlags` | `map<string,string>` | `{}` | PostgreSQL configuration flags. |
| `primaryInstance.displayName` | `string` | — | Human-readable name for the instance. |
| `primaryInstance.queryInsightsConfig.queryPlansPerMinute` | `int` | `5` | Plans captured per minute (0-20). |
| `primaryInstance.queryInsightsConfig.queryStringLength` | `int` | `1024` | Max query string length (256-4500). |
| `primaryInstance.queryInsightsConfig.recordApplicationTags` | `bool` | `true` | Record application tags in insights. |
| `primaryInstance.queryInsightsConfig.recordClientAddress` | `bool` | `true` | Record client IP in insights. |
| `primaryInstance.requireConnectors` | `bool` | `false` | Force AlloyDB Auth Proxy for all connections. |
| `primaryInstance.sslMode` | `string` | — | `ENCRYPTED_ONLY` or `ALLOW_UNENCRYPTED_AND_ENCRYPTED`. |

## Examples

### Production HA Cluster

A production-ready cluster with REGIONAL availability, initial user, and deletion protection:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpAlloydbCluster
metadata:
  name: prod-alloydb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: infra
    pulumi.planton.dev/stack.name: prod.GcpAlloydbCluster.prod-alloydb
spec:
  projectId:
    value: prod-project
  clusterName: prod-alloydb-cluster
  location: us-central1
  network:
    value: projects/prod-project/global/networks/prod-vpc
  databaseVersion: POSTGRES_16
  deletionProtection: true
  initialUser:
    password: "change-me-immediately"
    user: dbadmin
  primaryInstance:
    instanceId: prod-primary
    cpuCount: 4
    availabilityType: REGIONAL
    sslMode: ENCRYPTED_ONLY
```

### Enterprise CMEK Encryption

Full CMEK coverage with separate keys for data, backups, and PITR:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpAlloydbCluster
metadata:
  name: enterprise-alloydb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: infra
    pulumi.planton.dev/stack.name: prod.GcpAlloydbCluster.enterprise-alloydb
spec:
  projectId:
    value: enterprise-project
  clusterName: enterprise-alloydb
  location: us-east1
  network:
    valueFrom:
      kind: GcpVpc
      name: enterprise-vpc
  databaseVersion: POSTGRES_16
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: alloydb-data-key
  automatedBackupPolicy:
    enabled: true
    timeBasedRetentionPeriod: "2592000s"
    encryptionKmsKeyName:
      valueFrom:
        kind: GcpKmsKey
        name: alloydb-backup-key
  continuousBackupConfig:
    enabled: true
    recoveryWindowDays: 21
    encryptionKmsKeyName:
      valueFrom:
        kind: GcpKmsKey
        name: alloydb-pitr-key
  primaryInstance:
    instanceId: enterprise-primary
    cpuCount: 8
    availabilityType: REGIONAL
    requireConnectors: true
    sslMode: ENCRYPTED_ONLY
    queryInsightsConfig:
      queryPlansPerMinute: 10
      queryStringLength: 4096
      recordApplicationTags: true
      recordClientAddress: true
```

### Custom Backup Policy

Quantity-based retention with a weekly schedule:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpAlloydbCluster
metadata:
  name: backup-alloydb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: infra
    pulumi.planton.dev/stack.name: staging.GcpAlloydbCluster.backup-alloydb
spec:
  projectId:
    value: staging-project
  clusterName: backup-alloydb
  location: europe-west1
  network:
    value: projects/staging-project/global/networks/staging-vpc
  automatedBackupPolicy:
    enabled: true
    quantityBasedRetentionCount: 10
    backupWindow: "7200s"
    weeklySchedule:
      daysOfWeek:
        - MONDAY
        - WEDNESDAY
        - FRIDAY
      startHour: 3
  maintenanceWindow:
    day: SUNDAY
    startHour: 4
  primaryInstance:
    instanceId: backup-primary
    cpuCount: 4
    availabilityType: REGIONAL
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | Fully qualified cluster resource name: `projects/{p}/locations/{l}/clusters/{c}` |
| `cluster_name` | `string` | Short cluster name (same as `clusterName` input) |
| `primary_instance_ip` | `string` | Private IP of the primary instance. Connect on port 5432. |
| `primary_instance_name` | `string` | Fully qualified instance resource name. Used for Auth Proxy connections. |
| `database_version` | `string` | Computed PostgreSQL version (e.g., `POSTGRES_15`) |
| `state` | `string` | Cluster state: `READY`, `CREATING`, `MAINTENANCE`, etc. |

## Related Components

- [GcpVpc](/docs/catalog/gcp/vpc) — VPC network with Private Service Access for cluster connectivity
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — CMEK encryption keys for data, backups, and continuous backups
- [GcpKmsKeyRing](/docs/catalog/gcp/kms-key-ring) — Key ring containing KMS keys (must be in the same region as the cluster)
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — Service account for application-level IAM access to AlloyDB
- [GcpFirewallRule](/docs/catalog/gcp/firewall-rule) — Network-level access control for AlloyDB connectivity

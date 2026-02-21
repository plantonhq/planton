# OCI Block Volume

Deploys an Oracle Cloud Infrastructure Block Volume with configurable performance tiers (VPUs/GB), optional autotune policies for automatic performance adjustment, cross-region replicas for disaster recovery, and an optional backup policy assignment for scheduled backups.

## What Gets Created

When you deploy an OciBlockVolume resource, OpenMCF provisions:

- **Block Volume** — an `oci_core_volume` in the specified compartment and availability domain with configurable size (50-32768 GB), performance tier (VPUs/GB), optional KMS encryption, autotune policies, cross-region replicas, and SCSI persistent reservation support.
- **Backup Policy Assignment** — created only when `backupPolicyId` is set. An `oci_core_volume_backup_policy_assignment` that links the volume to an Oracle-defined (Gold, Silver, Bronze) or custom backup policy for scheduled backups.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the volume will be created — either a literal value or a reference to an OciCompartment resource
- **An availability domain** name within the target region (e.g., `"Uocm:US-ASHBURN-AD-1"`) — the volume and any attached compute instance must be in the same AD
- **A KMS key OCID** (optional) if using customer-managed encryption
- **A backup policy OCID** (optional) if assigning a scheduled backup policy — retrieve Oracle-defined policy OCIDs via `oci bv volume-backup-policy list`

## Quick Start

Create a file `block-volume.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: my-volume
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciBlockVolume.my-volume
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 50
```

Deploy:

```shell
openmcf apply -f block-volume.yaml
```

This creates a 50 GB block volume with Balanced performance (10 VPUs/GB) and Oracle-managed encryption. The volume OCID is exported as a stack output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the volume will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `availabilityDomain` | `string` | Availability domain for volume placement (e.g., `"Uocm:US-ASHBURN-AD-1"`). The volume and any attached compute instance must share the same AD. Changing this forces recreation. | Min length 1 |
| `sizeInGbs` | `int32` | Size of the volume in gigabytes. Must be specified explicitly to prevent accidental creation at OCI's 1 TB default. | >= 50, max 32768 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | metadata name | Display name for the volume in the OCI Console. |
| `vpusPerGb` | `int32` | `10` (Balanced) | Volume Performance Units per GB. Controls IOPS and throughput. Values: `0` (Lower Cost, 2 IOPS/GB), `10` (Balanced, 60 IOPS/GB), `20` (Higher Performance, 75 IOPS/GB), `30`-`120` (Ultra High Performance, increments of 10). |
| `kmsKeyId` | `StringValueOrRef` | — | OCID of a KMS master encryption key. When unset, Oracle-managed keys are used. Can reference an OciKmsKey resource via `valueFrom`. |
| `isReservationsEnabled` | `bool` | `false` | Enables SCSI persistent reservation support. Required for shared-storage clustering (e.g., Oracle RAC). |
| `autotunePolicies` | `AutotunePolicy[]` | — | Autotune policies for automatic performance adjustment. See below. |
| `blockVolumeReplicas` | `BlockVolumeReplica[]` | — | Cross-region replicas for disaster recovery. See below. |
| `backupPolicyId` | `StringValueOrRef` | — | OCID of a backup policy (Oracle-defined: Gold, Silver, Bronze, or custom). When set, creates a backup policy assignment sub-resource. |
| `xrcKmsKeyId` | `StringValueOrRef` | — | OCID of a KMS key for encrypting cross-region volume backups. Only relevant when a backup policy with cross-region copy is assigned. Changing this forces recreation. |

### AutotunePolicy

| Field | Type | Description |
|-------|------|-------------|
| `autotuneType` | `enum` | Policy type. Values: `detached_volume` (reduce VPUs to 0 when detached, restore on re-attach), `performance_based` (dynamically adjust VPUs based on workload). |
| `maxVpusPerGb` | `int32` | Maximum VPUs/GB for `performance_based` autotune. Required when `autotuneType` is `performance_based`. Must be > 0. |

### BlockVolumeReplica

| Field | Type | Description |
|-------|------|-------------|
| `availabilityDomain` | `string` | Availability domain for the replica (e.g., `"Uocm:US-PHOENIX-AD-1"`). Can be in a different region from the source volume. |
| `displayName` | `string` | Display name for the replica. When omitted, OCI generates one. |
| `xrrKmsKeyId` | `StringValueOrRef` | OCID of a KMS key for encrypting the replica. Can reference an OciKmsKey resource via `valueFrom`. |

## Examples

### Minimal Volume

A 50 GB volume with default Balanced performance — suitable for development or low-IOPS workloads:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: dev-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciBlockVolume.dev-data
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 50
```

### High-Performance Database Volume

A 500 GB volume with Higher Performance tier and KMS encryption for a database workload:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: db-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBlockVolume.db-data
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  displayName: "db-data-prod"
  sizeInGbs: 500
  vpusPerGb: 20
  kmsKeyId:
    value: "ocid1.key.oc1..example"
```

### Autotune with Backup Policy

A volume with detached-volume autotune (reduces cost when not attached) and a Gold backup policy for daily backups with cross-region copy:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: app-storage
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBlockVolume.app-storage
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 200
  vpusPerGb: 10
  autotunePolicies:
    - autotuneType: detached_volume
  backupPolicyId:
    value: "ocid1.volumebackuppolicy.oc1..gold-example"
  xrcKmsKeyId:
    value: "ocid1.key.oc1..xrc-example"
```

### Cross-Region Replica with Performance Autotune

A production volume with performance-based autotune and a cross-region replica for disaster recovery:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciBlockVolume
metadata:
  name: critical-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciBlockVolume.critical-data
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  sizeInGbs: 1024
  vpusPerGb: 20
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  isReservationsEnabled: false
  autotunePolicies:
    - autotuneType: performance_based
      maxVpusPerGb: 60
  blockVolumeReplicas:
    - availabilityDomain: "Uocm:US-PHOENIX-AD-1"
      displayName: "critical-data-dr-phx"
      xrrKmsKeyId:
        value: "ocid1.key.oc1..phx-example"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `volume_id` | `string` | OCID of the created block volume |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciKmsKey](/docs/catalog/oci/ocikmskey) — provides encryption keys referenced by `kmsKeyId`, `xrcKmsKeyId`, and replica `xrrKmsKeyId`
- [OciComputeInstance](/docs/catalog/oci/ocicomputeinstance) — attaches block volumes for persistent storage
- [OciKmsVault](/docs/catalog/oci/ocikmsvault) — contains the KMS keys used for volume encryption

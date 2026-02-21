---
title: "PostgreSQL DB System"
description: "PostgreSQL DB System deployment documentation"
icon: "package"
order: 100
componentName: "ocipostgresqldbsystem"
---

# OCI PostgreSQL DB System

Deploys an Oracle Cloud Infrastructure PostgreSQL Database System — a fully managed PostgreSQL service with configurable compute shapes, flexible OCPU/memory sizing, regional or AD-local storage durability, read replicas, and built-in backup policies.

## What Gets Created

When you deploy an OciPostgresqlDbSystem resource, OpenMCF provisions:

- **PostgreSQL DB System** — an `oci_psql_db_system` resource in the specified compartment running the chosen PostgreSQL major version on dedicated compute shapes. The system includes a primary (read-write) endpoint and optional read replicas when `instanceCount` is 2 or more.
- **Storage Backend** — OCI-optimized storage with a choice between regionally durable (multi-AD replication) or AD-local placement. IOPS performance tier is configurable.
- **Backup Policy** — automatic backups on a daily, weekly, or monthly schedule with configurable retention, or disabled entirely via the `none` kind.
- **Freeform Tags** — automatically applied tags capturing the resource kind, resource ID, organization, and environment from metadata.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the DB System will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** in a VCN where the DB System instances will be placed — either a literal value or a reference to an OciSubnet resource
- **A PostgreSQL major version** supported in OCI (e.g. "14", "15", "16")
- **A compute shape** available for PostgreSQL DB Systems (e.g. "VM.Standard.E4.Flex")

## Quick Start

Create a file `postgresql.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: my-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciPostgresqlDbSystem.my-postgres
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 2
  instanceMemorySizeInGbs: 16
  instanceCount: 1
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  storageDetails:
    isRegionallyDurable: true
  credentials:
    username: postgres
    passwordDetails:
      passwordType: plain_text
      password: "change-me-immediately"
```

Deploy:

```shell
openmcf apply -f postgresql.yaml
```

This creates a single-node PostgreSQL 16 DB System on a flexible shape with 2 OCPUs and 16 GB memory, regionally durable storage, and a plain-text admin password. The DB System ID, primary endpoint IP, and admin username are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the DB System will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `dbVersion` | `string` | PostgreSQL major version (e.g. "14", "15", "16"). Minor versions are managed by OCI. Changing this forces recreation. | Non-empty |
| `shape` | `string` | Compute shape for DB System instances. The provider auto-prefixes "PostgreSQL." if not present. For flexible shapes, set `instanceOcpuCount` and `instanceMemorySizeInGbs`. Example: "VM.Standard.E4.Flex". | Non-empty |
| `networkDetails` | `NetworkDetails` | Network placement configuration. See [NetworkDetails](#networkdetails) below. | Required |
| `storageDetails` | `StorageDetails` | Storage backend configuration. See [StorageDetails](#storagedetails) below. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. |
| `instanceOcpuCount` | `int32` | — | Number of OCPUs allocated to each instance. Used with flexible shapes. Updatable. |
| `instanceMemorySizeInGbs` | `int32` | — | Memory in GB allocated to each instance. Used with flexible shapes. Updatable. |
| `instanceCount` | `int32` | — | Number of database instances. 1 = standalone; 2+ = primary with read replicas. |
| `credentials` | `Credentials` | Provider defaults | Initial database admin credentials. Immutable after creation. See [Credentials](#credentials). |
| `managementPolicy` | `ManagementPolicy` | — | Backup schedule and maintenance window. See [ManagementPolicy](#managementpolicy). |
| `configId` | `StringValueOrRef` | Shape default | OCID of a PostgreSQL configuration (server parameters like `shared_buffers`, `max_connections`). |
| `description` | `string` | — | User-provided description of the DB System. |
| `instancesDetails` | `InstanceDetails[]` | — | Per-instance config (display name, description, private IP). List size must match `instanceCount`. Immutable after creation. See [InstanceDetails](#instancedetails). |

### NetworkDetails

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `subnetId` | `StringValueOrRef` | — | **Required.** OCID of the subnet for DB System placement. Can reference an OciSubnet via `valueFrom`. Changing this forces recreation. |
| `nsgIds` | `StringValueOrRef[]` | — | OCIDs of network security groups applied to instances. Can reference OciSecurityGroup resources. |
| `isReaderEndpointEnabled` | `bool` | — | When `true`, creates a reader endpoint for distributing read queries across replicas. |
| `primaryDbEndpointPrivateIp` | `string` | Auto-assigned | Specific private IP for the primary (read-write) endpoint. Changing this forces recreation. |

### StorageDetails

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isRegionallyDurable` | `bool` | — | **Required.** When `true`, data is replicated across multiple ADs. When `false`, `availabilityDomain` must be specified. Changing this forces recreation. |
| `availabilityDomain` | `string` | — | AD for single-AD storage. Required when `isRegionallyDurable` is `false`. Example: "Uocm:PHX-AD-1". Changing this forces recreation. |
| `iops` | `int64` | — | Guaranteed IOPS for the storage tier. See OCI documentation for supported values per shape. Updatable. |

### Credentials

| Field | Type | Description |
|-------|------|-------------|
| `username` | `string` | **Required.** Administrator username. Changing this forces recreation. |
| `passwordDetails` | `PasswordDetails` | **Required.** Password configuration. See [PasswordDetails](#passworddetails). |

### PasswordDetails

| Field | Type | Description |
|-------|------|-------------|
| `passwordType` | `PasswordType` | Discriminator: `plain_text` or `vault_secret`. |
| `password` | `string` | Plain-text password. Required when `passwordType` is `plain_text`. Not returned by the API after creation. |
| `secretId` | `StringValueOrRef` | OCID of the OCI Vault secret. Required when `passwordType` is `vault_secret`. |
| `secretVersion` | `string` | Vault secret version. When omitted, the latest version is used. |

### ManagementPolicy

| Field | Type | Description |
|-------|------|-------------|
| `backupPolicy` | `BackupPolicy` | Backup schedule configuration. See [BackupPolicy](#backuppolicy). |
| `maintenanceWindowStart` | `string` | Maintenance window start in UTC. Format: "{day-of-week} {time-of-day}" (e.g. "tue 02:00:00"). |

### BackupPolicy

| Field | Type | Description |
|-------|------|-------------|
| `kind` | `BackupKind` | Schedule frequency: `daily`, `weekly`, `monthly`, or `none`. |
| `backupStart` | `string` | Hour (UTC) when the backup starts. Required for `daily`, `weekly`, and `monthly`. |
| `retentionDays` | `int32` | Days to retain backups after the DB System is deleted. |
| `daysOfTheMonth` | `int32[]` | Days of the month (1-28) for monthly backups. Max 28 items. |
| `daysOfTheWeek` | `string[]` | Days of the week (e.g. "MONDAY", "FRIDAY") for weekly backups. |

### InstanceDetails

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | `string` | Display name for this instance node. |
| `description` | `string` | Description of this instance node. |
| `privateIp` | `string` | Specific private IP within the subnet. When omitted, OCI auto-assigns. |

## Examples

### Minimal Standalone Instance

A single-node PostgreSQL 16 instance with regionally durable storage — suitable for development:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: dev-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciPostgresqlDbSystem.dev-postgres
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 1
  instanceMemorySizeInGbs: 8
  instanceCount: 1
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  storageDetails:
    isRegionallyDurable: true
  credentials:
    username: postgres
    passwordDetails:
      passwordType: plain_text
      password: "dev-password-change-me"
```

### Production with Read Replicas and Vault Secret

A multi-node PostgreSQL system with Vault-managed credentials, NSG-secured networking, a reader endpoint, and daily backups retained for 30 days:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: prod-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-org
    pulumi.openmcf.org/project: acme-data
    pulumi.openmcf.org/stack.name: prod.OciPostgresqlDbSystem.prod-postgres
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 4
  instanceMemorySizeInGbs: 32
  instanceCount: 2
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
    nsgIds:
      - value: "ocid1.networksecuritygroup.oc1..example"
    isReaderEndpointEnabled: true
  storageDetails:
    isRegionallyDurable: true
  credentials:
    username: postgres
    passwordDetails:
      passwordType: vault_secret
      secretId:
        value: "ocid1.vaultsecret.oc1..example"
  managementPolicy:
    backupPolicy:
      kind: daily
      backupStart: "03:00"
      retentionDays: 30
    maintenanceWindowStart: sun 04:00:00
```

### Single-AD Development Instance

A cost-optimized single-AD setup with plain-text password and weekly backups — suitable for development or testing environments:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: test-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: test.OciPostgresqlDbSystem.test-postgres
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 1
  instanceMemorySizeInGbs: 8
  instanceCount: 1
  networkDetails:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  storageDetails:
    isRegionallyDurable: false
    availabilityDomain: "Uocm:PHX-AD-1"
  credentials:
    username: postgres
    passwordDetails:
      passwordType: plain_text
      password: "test-password"
  managementPolicy:
    backupPolicy:
      kind: weekly
      backupStart: "02:00"
      retentionDays: 7
      daysOfTheWeek:
        - SUNDAY
```

### Using Foreign Key References

Reference OpenMCF-managed compartment and subnet resources instead of hardcoding OCIDs:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPostgresqlDbSystem
metadata:
  name: ref-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciPostgresqlDbSystem.ref-postgres
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  dbVersion: "16"
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 2
  instanceMemorySizeInGbs: 16
  instanceCount: 1
  networkDetails:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: db-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: db-nsg
          fieldPath: status.outputs.networkSecurityGroupId
  storageDetails:
    isRegionallyDurable: true
  credentials:
    username: postgres
    passwordDetails:
      passwordType: vault_secret
      secretId:
        value: "ocid1.vaultsecret.oc1..example"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `dbSystemId` | `string` | OCID of the PostgreSQL DB System |
| `primaryDbEndpointPrivateIp` | `string` | Private IP address of the primary (read-write) endpoint |
| `adminUsername` | `string` | Administrator username (computed after creation) |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciVcn](/docs/catalog/oci/vcn) — the virtual cloud network containing the subnet where the DB System is placed
- [OciSubnet](/docs/catalog/oci/subnet) — provides the subnet referenced by `networkDetails.subnetId` via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — provides network security groups referenced by `networkDetails.nsgIds` via `valueFrom`

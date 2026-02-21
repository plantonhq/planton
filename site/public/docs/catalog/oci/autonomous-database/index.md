---
title: "Autonomous Database"
description: "Autonomous Database deployment documentation"
icon: "package"
order: 100
componentName: "ociautonomousdatabase"
---

# OCI Autonomous Database

Deploys an Oracle Cloud Infrastructure Autonomous Database — a fully managed, self-driving database service supporting OLTP (ATP), data warehouse (ADW), JSON (AJD), APEX, and lakehouse workloads. The component handles compute and storage sizing, networking, encryption, Data Guard, and backup retention through a single manifest.

## What Gets Created

When you deploy an OciAutonomousDatabase resource, OpenMCF provisions:

- **Autonomous Database** — an `oci_database_autonomous_database` resource in the specified compartment. The database type is determined by `dbWorkload` (OLTP, DW, AJD, APEX, or LH). Freeform tags are applied automatically from metadata labels, environment, and organization.
- **Connection Strings** — three prioritized connection strings (high, medium, low) are exported as stack outputs for use by application workloads.
- **Private Endpoint** (conditional) — when `subnetId` is set, the database is provisioned with a private endpoint in the specified subnet, disabling public secure access. NSGs and IP access lists can further restrict connectivity.
- **Customer-Managed Encryption** (conditional) — when `kmsKeyId` and `vaultId` are set, Transparent Data Encryption uses the specified KMS key instead of Oracle-managed keys.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the autonomous database will be created — either a literal value or a reference to an OciCompartment resource
- **An admin password** or a **Vault secret OCID** containing the password — one of the two is required for the database administrator account

## Quick Start

Create a file `adb.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: my-adb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciAutonomousDatabase.my-adb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "myatp"
  dbWorkload: oltp
  computeModel: ecpu
  computeCount: 2
  dataStorageSizeInTbs: 1
  adminPassword: "ExamplePass#2026"
```

Deploy:

```shell
openmcf apply -f adb.yaml
```

This creates a serverless Autonomous Transaction Processing database with 2 ECPUs and 1 TB of storage. The database OCID, connection strings, and service console URL are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the autonomous database will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `dbName` | `string` | The database name. Must be alphanumeric, begin with a letter, and be unique within the tenancy. Cannot be changed after creation. | 1–30 characters, pattern `^[a-zA-Z][a-zA-Z0-9]*$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. Falls back to `metadata.name` if not provided. |
| `dbWorkload` | `enum` | `db_workload_unspecified` | Workload type: `oltp` (ATP), `dw` (ADW), `ajd` (JSON Database), `apex` (APEX), `lh` (Lakehouse). Determines optimizer behavior and available features. |
| `dbVersion` | `string` | latest | Oracle Database version (e.g. `"19c"`, `"23ai"`, `"26ai"`). When omitted, the latest available version is used. |
| `databaseEdition` | `enum` | `database_edition_unspecified` | `standard_edition` or `enterprise_edition`. Enterprise includes partitioning, compression, and advanced security. |
| `licenseModel` | `enum` | `license_model_unspecified` | `bring_your_own_license` or `license_included`. AJD and APEX workloads always use LICENSE_INCLUDED regardless of this setting. |
| `characterSet` | `string` | `AL32UTF8` | Character set for the database. Cannot be changed after creation. |
| `ncharacterSet` | `string` | `AL16UTF16` | National character set. Valid values: `AL16UTF16`, `UTF8`. Cannot be changed after creation. |
| `computeModel` | `enum` | `compute_model_unspecified` | `ecpu` (recommended) or `ocpu` (legacy). |
| `computeCount` | `float` | — | Number of compute units (ECPUs or OCPUs). Minimum 2 ECPUs for ECPU model. |
| `dataStorageSizeInTbs` | `int32` | — | Maximum storage in terabytes. For serverless deployments. Mutually exclusive with `dataStorageSizeInGb`. |
| `dataStorageSizeInGb` | `int32` | — | Maximum storage in gigabytes. For dedicated Exadata deployments needing finer granularity. Mutually exclusive with `dataStorageSizeInTbs`. |
| `isAutoScalingEnabled` | `bool` | — | When `true`, CPU auto-scaling allows up to 3x the provisioned compute count during demand spikes. |
| `isAutoScalingForStorageEnabled` | `bool` | — | When `true`, storage auto-scaling automatically expands storage when usage reaches the threshold. |
| `adminPassword` | `string` | — | Administrator password. 12–30 characters, must contain uppercase, lowercase, and numeric. Cannot contain "admin" or double-quote. Mutually exclusive with `secretId`. |
| `secretId` | `StringValueOrRef` | — | OCID of a Vault secret containing the admin password. Use instead of `adminPassword` for production. Mutually exclusive with `adminPassword`. |
| `secretVersionNumber` | `int32` | latest | Version of the Vault secret. Only applicable when `secretId` is set. |
| `subnetId` | `StringValueOrRef` | — | OCID of the subnet for private endpoint access. When set, public access is disabled. Can reference an OciSubnet resource. |
| `nsgIds` | `StringValueOrRef[]` | — | OCIDs of network security groups for the private endpoint. Maximum 5. Only applicable when `subnetId` is set. |
| `privateEndpointLabel` | `string` | — | DNS label prefix for the private endpoint FQDN. |
| `privateEndpointIp` | `string` | — | Specific private IP for the endpoint within the subnet. Auto-assigned when omitted. |
| `whitelistedIps` | `string[]` | — | Client IP access control list. Each entry can be an IP address, CIDR block, or VCN OCID. |
| `isMtlsConnectionRequired` | `bool` | — | When `true`, only mutual TLS connections are allowed. When `false`, both TLS and mTLS are accepted. |
| `isAccessControlEnabled` | `bool` | — | When `true`, enables database-level access control. For Exadata Cloud@Customer deployments. |
| `kmsKeyId` | `StringValueOrRef` | — | OCID of the KMS key for Transparent Data Encryption. When omitted, Oracle-managed encryption is used. |
| `vaultId` | `StringValueOrRef` | — | OCID of the OCI Vault containing the KMS key. Required when `kmsKeyId` is set. |
| `isDedicated` | `bool` | — | When `true`, uses dedicated Exadata infrastructure (requires `autonomousContainerDatabaseId`). Cannot be changed after creation. |
| `isFreeTier` | `bool` | — | When `true`, provisions an Always Free database with limited compute and storage. Reclaimed after extended inactivity. |
| `isDevTier` | `bool` | — | When `true`, provisions a Developer tier database at reduced cost for development and testing. |
| `autonomousContainerDatabaseId` | `StringValueOrRef` | — | OCID of the autonomous container database for dedicated deployments. Required when `isDedicated` is `true`. |
| `backupRetentionPeriodInDays` | `int32` | service default | Number of days to retain automatic backups. |
| `isLocalDataGuardEnabled` | `bool` | — | When `true`, enables local Autonomous Data Guard. A standby is provisioned in a different availability domain within the same region. |
| `autonomousMaintenanceScheduleType` | `enum` | `maintenance_schedule_type_unspecified` | `early` (patches sooner) or `regular` (standard Oracle schedule). |
| `customerContacts` | `CustomerContact[]` | — | Contact email addresses for operational notifications. Each entry has an `email` field. |

## Examples

### Basic ATP Database

A serverless Autonomous Transaction Processing database for a development workload:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: dev-atp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciAutonomousDatabase.dev-atp
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "devatp"
  dbWorkload: oltp
  computeModel: ecpu
  computeCount: 2
  dataStorageSizeInTbs: 1
  adminPassword: "DevPass#2026abc"
```

### ADW for Analytics

An Autonomous Data Warehouse with BYOL licensing and auto-scaling for analytic workloads:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: analytics-adw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciAutonomousDatabase.analytics-adw
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "analyticsadw"
  dbWorkload: dw
  computeModel: ecpu
  computeCount: 4
  dataStorageSizeInTbs: 2
  licenseModel: bring_your_own_license
  isAutoScalingEnabled: true
  adminPassword: "AnalyticsPass#2026"
```

### Production ATP with Private Endpoint and Data Guard

A production-grade ATP database with private networking, Vault-managed credentials, customer-managed encryption, Data Guard, and maintenance scheduling:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: prod-atp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: prod-databases
    pulumi.openmcf.org/stack.name: prod.OciAutonomousDatabase.prod-atp
  env: prod
  org: acme
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  dbName: "prodatp"
  displayName: "Production ATP"
  dbWorkload: oltp
  dbVersion: "23ai"
  databaseEdition: enterprise_edition
  computeModel: ecpu
  computeCount: 8
  dataStorageSizeInTbs: 5
  licenseModel: bring_your_own_license
  isAutoScalingEnabled: true
  isAutoScalingForStorageEnabled: true
  secretId:
    value: "ocid1.vaultsecret.oc1..example"
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: prod-db-subnet
      fieldPath: status.outputs.subnetId
  nsgIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: prod-db-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  privateEndpointLabel: "prodatp"
  isMtlsConnectionRequired: true
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  vaultId:
    value: "ocid1.vault.oc1..example"
  isLocalDataGuardEnabled: true
  backupRetentionPeriodInDays: 30
  autonomousMaintenanceScheduleType: regular
  customerContacts:
    - email: "dba-team@acme.com"
    - email: "oncall@acme.com"
```

### Always Free Tier

An Always Free ATP database for experimentation — no cost, limited resources:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciAutonomousDatabase
metadata:
  name: free-atp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciAutonomousDatabase.free-atp
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  dbName: "freeatp"
  dbWorkload: oltp
  isFreeTier: true
  adminPassword: "FreePass#2026abc"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `autonomousDatabaseId` | `string` | OCID of the autonomous database |
| `connectionStringHigh` | `string` | High-priority connection string for latency-sensitive workloads |
| `connectionStringMedium` | `string` | Medium-priority connection string for typical application workloads |
| `connectionStringLow` | `string` | Low-priority connection string for batch and background workloads |
| `serviceConsoleUrl` | `string` | URL of the OCI Service Console for this database |
| `privateEndpoint` | `string` | Private endpoint IP address. Empty when the database is not configured with a private endpoint. |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/subnet) — provides the subnet for private endpoint access via `subnetId`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — provides NSGs for the private endpoint via `nsgIds`
- [OciVcn](/docs/catalog/oci/vcn) — the VCN containing the subnet used for private endpoint access

---
title: "DB System"
description: "DB System deployment documentation"
icon: "package"
order: 100
componentName: "ocidbsystem"
---

# OCI DB System

Deploys an Oracle Cloud Infrastructure Database System — a managed Oracle Database running on Virtual Machine or Bare Metal infrastructure. The component provisions the underlying compute and storage, a DB Home containing the Oracle Database software, and an initial database instance as an inseparable unit.

## What Gets Created

When you deploy an OciDbSystem resource, OpenMCF provisions:

- **Database DB System** — an `oci_database_db_system` resource in the specified compartment and subnet. Configures the compute shape, CPU core count, storage layout, SSH access, and optional features like KMS encryption, NSG attachment, and maintenance windows.
- **DB Home** — an Oracle Database Home containing the database software at a specific version (e.g., 19.0.0.0) or from a custom database software image. Created as a nested resource within the DB System.
- **Initial Database** — an Oracle Database instance within the DB Home, configured with the specified admin password, database name, character sets, optional pluggable database, and optional automatic backup schedule.
- **Freeform Tags** — applied automatically from metadata (resource kind, resource ID, organization, environment, and any custom labels).

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the DB System will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** for the DB System's primary VNIC — either a literal value or a reference to an OciSubnet resource
- **An SSH public key** for administrative access to the DB System nodes
- **An availability domain** name (e.g., `Uocm:PHX-AD-1`)
- **A compute shape** appropriate for your workload (e.g., `VM.Standard2.4`, `VM.Standard.E4.Flex`, `BM.DenseIO2.52`)

## Quick Start

Create a file `db-system.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: my-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDbSystem.my-db
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard2.2"
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "mydb"
  dbHome:
    dbVersion: "19.0.0.0"
    database:
      adminPassword: "BEstr0ng#2024"
      dbName: "MYDB"
```

Deploy:

```shell
openmcf apply -f db-system.yaml
```

This creates a single-node VM DB System running Oracle Database 19c with default settings. The DB System OCID, DB Home OCID, database OCID, and listener port are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the DB System will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `availabilityDomain` | `string` | Availability domain where the DB System will be placed. Example: `Uocm:PHX-AD-1`. Changing this forces recreation. | Non-empty |
| `shape` | `string` | Compute shape for the DB System nodes. Determines CPU architecture, core count range, and memory. Examples: `VM.Standard2.4`, `VM.Standard.E4.Flex`, `BM.DenseIO2.52`. | Non-empty |
| `subnetId` | `StringValueOrRef` | OCID of the subnet where the DB System will be placed. Can reference an OciSubnet resource via `valueFrom`. Changing this forces recreation. | Required |
| `sshPublicKeys` | `string[]` | SSH public keys in OpenSSH format for administrative access to the DB System nodes. | Minimum 1 item |
| `hostname` | `string` | Hostname for the DB System. Must be unique within the subnet. Combined with the domain to form the FQDN. Changing this forces recreation. | Non-empty |
| `dbHome` | `DbHome` | DB Home configuration containing the Oracle Database software version and initial database. See nested fields below. | Required |
| `dbHome.dbVersion` | `string` | Oracle Database version (e.g., `19.0.0.0`, `21.0.0.0`). Mutually exclusive with `dbHome.databaseSoftwareImageId`. One of the two is required. Changing this forces recreation. | — |
| `dbHome.database` | `Database` | Initial database to create within the DB Home. | Required |
| `dbHome.database.adminPassword` | `string` | Administrator password for SYS and SYSTEM users. Must be 2–30 characters, contain at least one uppercase, one lowercase, and one numeric character. Cannot contain double-quote. Not returned by the API after creation. | Min length 2 |
| `dbHome.database.dbName` | `string` | Database name. Alphanumeric, must begin with a letter. At most 8 characters for single-node or 30 characters otherwise. Changing this forces recreation. | 1–30 chars, `^[a-zA-Z][a-zA-Z0-9]*$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name for the DB System shown in the OCI Console. |
| `cpuCoreCount` | `int32` | Shape default | Number of CPU cores (OCPUs for VM, total cores for BM). |
| `databaseEdition` | `enum` | — | Oracle Database edition: `standard_edition`, `enterprise_edition`, `enterprise_edition_high_performance`, `enterprise_edition_extreme_performance`. Changing this forces recreation. |
| `licenseModel` | `enum` | — | Licensing model: `bring_your_own_license` or `license_included`. |
| `dataStorageSizeInGb` | `int32` | — | Initial data storage size in GB. Required for VM DB Systems. Typical values: 256, 512, 1024, 2048, 4096, etc. |
| `dataStoragePercentage` | `int32` | — | Percentage allocated to data vs. recovery: 40 or 80. BM only. Changing this forces recreation. |
| `diskRedundancy` | `enum` | — | Disk mirroring level: `normal` (2-way) or `high` (3-way). BM only. Changing this forces recreation. |
| `nodeCount` | `int32` | — | Number of nodes: 1 for single-node, 2 for 2-node RAC cluster. Changing this forces recreation. |
| `domain` | `string` | Subnet domain | Network domain name. Changing this forces recreation. |
| `clusterName` | `string` | — | RAC cluster name. Maximum 11 characters. Changing this forces recreation. |
| `faultDomains` | `string[]` | — | Fault domains for distributing RAC nodes. Changing this forces recreation. |
| `nsgIds` | `StringValueOrRef[]` | — | OCIDs of network security groups for the DB System VNIC. Can reference OciSecurityGroup resources. |
| `backupSubnetId` | `StringValueOrRef` | — | OCID of the backup subnet. Required for RAC. Can reference an OciSubnet resource. Changing this forces recreation. |
| `backupNetworkNsgIds` | `StringValueOrRef[]` | — | OCIDs of network security groups for the backup VNIC. Can reference OciSecurityGroup resources. |
| `kmsKeyId` | `StringValueOrRef` | — | OCID of a KMS key for encrypting DB System data at rest. |
| `kmsKeyVersionId` | `string` | Current version | OCID of a specific KMS key version. |
| `timeZone` | `string` | — | Time zone for the DB System (e.g., `UTC`, `US/Pacific`). Changing this forces recreation. |
| `sparseDiskgroup` | `bool` | — | Enable sparse disk group on BM DB Systems. BM only. Changing this forces recreation. |
| `storageVolumePerformanceMode` | `enum` | — | Storage I/O mode: `balanced` or `high_performance`. Changing this forces recreation. |
| `privateIp` | `string` | Auto-assigned | Specific private IP for the DB System. Changing this forces recreation. |
| `dataCollectionOptions` | `object` | — | Diagnostic telemetry settings. See sub-fields below. |
| `dataCollectionOptions.isDiagnosticsEventsEnabled` | `bool` | — | Enable diagnostic event collection. |
| `dataCollectionOptions.isHealthMonitoringEnabled` | `bool` | — | Enable health monitoring. |
| `dataCollectionOptions.isIncidentLogsEnabled` | `bool` | — | Enable incident log collection. |
| `dbSystemOptions` | `object` | — | Storage management settings. |
| `dbSystemOptions.storageManagement` | `enum` | — | Storage strategy: `asm` (most shapes) or `lvm` (single-node VM only). Changing this forces recreation. |
| `maintenanceWindowDetails` | `object` | — | Maintenance window scheduling. When `preference` is `no_preference`, OCI selects automatically. |
| `maintenanceWindowDetails.preference` | `enum` | — | Scheduling preference: `no_preference` or `custom_preference`. |
| `maintenanceWindowDetails.patchingMode` | `enum` | — | Patching strategy: `rolling` (zero downtime for RAC) or `nonrolling`. |
| `maintenanceWindowDetails.leadTimeInWeeks` | `int32` | — | Weeks of advance notice before maintenance. |
| `maintenanceWindowDetails.months` | `string[]` | — | Months when maintenance is allowed (e.g., `["JANUARY", "JULY"]`). |
| `maintenanceWindowDetails.weeksOfMonth` | `int32[]` | — | Weeks of the month (1–4) when maintenance is allowed. |
| `maintenanceWindowDetails.daysOfWeek` | `string[]` | — | Days when maintenance is allowed (e.g., `["MONDAY"]`). |
| `maintenanceWindowDetails.hoursOfDay` | `int32[]` | — | Hours (0–23) when maintenance may start. |
| `maintenanceWindowDetails.customActionTimeoutInMins` | `int32` | — | Custom patching timeout in minutes (0–120). |
| `maintenanceWindowDetails.isCustomActionTimeoutEnabled` | `bool` | — | Enable custom action timeout. |
| `maintenanceWindowDetails.isMonthlyPatchingEnabled` | `bool` | — | Enable monthly patching. |
| `dbHome.displayName` | `string` | — | Human-readable name for the DB Home. |
| `dbHome.databaseSoftwareImageId` | `StringValueOrRef` | — | OCID of a custom database software image. Mutually exclusive with `dbHome.dbVersion`. Changing this forces recreation. |
| `dbHome.database.characterSet` | `string` | `AL32UTF8` | Character set for the database. Changing this forces recreation. |
| `dbHome.database.ncharacterSet` | `string` | `AL16UTF16` | National character set: `AL16UTF16` or `UTF8`. Changing this forces recreation. |
| `dbHome.database.pdbName` | `string` | — | Pluggable database name. Alphanumeric, begins with a letter. Changing this forces recreation. |
| `dbHome.database.dbDomain` | `string` | DB System domain | Database domain. Changing this forces recreation. |
| `dbHome.database.kmsKeyId` | `StringValueOrRef` | — | OCID of KMS key for TDE. When omitted, Oracle-managed encryption is used. Changing this forces recreation. |
| `dbHome.database.kmsKeyVersionId` | `string` | — | OCID of KMS key version for TDE. Changing this forces recreation. |
| `dbHome.database.vaultId` | `StringValueOrRef` | — | OCID of the OCI Vault for TDE. Required when `kmsKeyId` is set. Changing this forces recreation. |
| `dbHome.database.dbBackupConfig` | `object` | — | Automatic backup configuration. |
| `dbHome.database.dbBackupConfig.autoBackupEnabled` | `bool` | — | Enable automatic backups. |
| `dbHome.database.dbBackupConfig.autoBackupWindow` | `string` | — | Preferred backup window: `SLOT_ONE` through `SLOT_TWELVE` (2-hour UTC windows starting 00:00). |
| `dbHome.database.dbBackupConfig.recoveryWindowInDays` | `int32` | — | Backup retention period in days (1–60). |

## Examples

### Minimal Single-Node VM

A single-node VM DB System with Oracle 19c and default settings:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: dev-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDbSystem.dev-db
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard2.2"
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "devdb"
  dbHome:
    dbVersion: "19.0.0.0"
    database:
      adminPassword: "BEstr0ng#2024"
      dbName: "DEVDB"
```

### Enterprise Edition with Storage and Backups

An Enterprise Edition DB System with explicit storage sizing, BYOL licensing, and automatic daily backups:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: staging-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciDbSystem.staging-db
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard.E4.Flex"
  cpuCoreCount: 4
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "stagingdb"
  databaseEdition: enterprise_edition
  licenseModel: bring_your_own_license
  dataStorageSizeInGb: 1024
  dbHome:
    dbVersion: "19.0.0.0"
    displayName: "staging-home"
    database:
      adminPassword: "BEstr0ng#2024"
      dbName: "STAGDB"
      characterSet: "AL32UTF8"
      pdbName: "STAGPDB"
      dbBackupConfig:
        autoBackupEnabled: true
        autoBackupWindow: "SLOT_TWO"
        recoveryWindowInDays: 30
```

### 2-Node RAC with NSGs and Maintenance Window

A 2-node RAC cluster distributed across fault domains, with network security groups and a quarterly maintenance window:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: prod-rac
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDbSystem.prod-rac
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard2.8"
  cpuCoreCount: 8
  subnetId:
    value: "ocid1.subnet.oc1.phx.example"
  sshPublicKeys:
    - "ssh-rsa AAAA...key-one"
    - "ssh-rsa AAAA...key-two"
  hostname: "prodrac"
  databaseEdition: enterprise_edition_extreme_performance
  licenseModel: license_included
  dataStorageSizeInGb: 4096
  nodeCount: 2
  clusterName: "prodclstr"
  faultDomains:
    - "FAULT-DOMAIN-1"
    - "FAULT-DOMAIN-2"
  nsgIds:
    - value: "ocid1.networksecuritygroup.oc1.phx.example"
  backupSubnetId:
    value: "ocid1.subnet.oc1.phx.backup-example"
  backupNetworkNsgIds:
    - value: "ocid1.networksecuritygroup.oc1.phx.backup-nsg"
  maintenanceWindowDetails:
    preference: custom_preference
    patchingMode: rolling
    leadTimeInWeeks: 2
    months:
      - "JANUARY"
      - "APRIL"
      - "JULY"
      - "OCTOBER"
    weeksOfMonth:
      - 2
    daysOfWeek:
      - "SUNDAY"
    hoursOfDay:
      - 4
  dbHome:
    dbVersion: "19.0.0.0"
    displayName: "prod-rac-home"
    database:
      adminPassword: "BEstr0ng#Prod2024"
      dbName: "PRODDB"
      characterSet: "AL32UTF8"
      ncharacterSet: "AL16UTF16"
      pdbName: "PRODPDB"
      dbBackupConfig:
        autoBackupEnabled: true
        autoBackupWindow: "SLOT_FOUR"
        recoveryWindowInDays: 60
```

### Using Foreign Key References

Reference OpenMCF-managed compartment and subnet resources instead of hardcoding OCIDs:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDbSystem
metadata:
  name: ref-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDbSystem.ref-db
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "VM.Standard2.4"
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: db-subnet
      fieldPath: status.outputs.subnetId
  sshPublicKeys:
    - "ssh-rsa AAAA...your-key"
  hostname: "refdb"
  nsgIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: db-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  dbHome:
    dbVersion: "19.0.0.0"
    database:
      adminPassword: "BEstr0ng#2024"
      dbName: "REFDB"
      pdbName: "REFPDB"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `dbSystemId` | `string` | OCID of the DB System |
| `dbHomeId` | `string` | OCID of the first DB Home created with the DB System |
| `databaseId` | `string` | OCID of the initial database created within the first DB Home |
| `listenerPort` | `int32` | TCP port on which the database listener accepts connections |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciVcn](/docs/catalog/oci/vcn) — creates the VCN containing the subnet used by this DB System
- [OciSubnet](/docs/catalog/oci/subnet) — provides the subnet referenced by `subnetId` and `backupSubnetId` via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — manages network security rules applied via `nsgIds` and `backupNetworkNsgIds`

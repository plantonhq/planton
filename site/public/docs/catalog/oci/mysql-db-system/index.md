---
title: "MySQL DB System"
description: "MySQL DB System deployment documentation"
icon: "package"
order: 100
componentName: "ocimysqldbsystem"
---

# OCI MySQL DB System

Deploys an Oracle Cloud Infrastructure MySQL HeatWave Database System — a fully managed MySQL database service with optional High Availability across fault domains, automated backups, point-in-time recovery, and read-scaling endpoints. The component manages the DB System resource itself; HeatWave cluster and replication channels are separate OCI resources with independent lifecycles.

## What Gets Created

When you deploy an OciMysqlDbSystem resource, Planton provisions:

- **MySQL DB System** — an `oci_mysql_mysql_db_system` resource in the specified compartment and subnet, placed in a given availability domain on a chosen compute shape. OCI automatically creates a primary read/write endpoint with a private IP address.
- **High Availability replicas** — when `isHighlyAvailable` is `true`, three instances are provisioned across different fault domains with automatic failover. Standby instances are not directly accessible.
- **Automatic backups** — when `backupPolicy` is configured, daily backups run within a 30-minute window with configurable retention. Point-in-time recovery can be enabled via the nested `pitrPolicy`.
- **Read endpoint** — when `readEndpoint` is configured and enabled, a separate DNS endpoint distributes read queries across HA replicas for read scaling.
- **Database Console** — when `databaseConsole` is configured and enabled, a web-based MySQL management UI is available on the specified port.
- **REST API service** — when `rest` is configured, the MySQL Router REST API is exposed on the specified port.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the DB System will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** in an existing VCN — either a literal value or a reference to an OciSubnet resource
- **An availability domain** name (e.g., `Uocm:PHX-AD-1`)
- **A compute shape** name (e.g., `MySQL.VM.Standard.E4.1.8GB`)

## Quick Start

Create a file `mysql-db.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciMysqlDbSystem
metadata:
  name: my-mysql
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciMysqlDbSystem.my-mysql
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.1.8GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "admin"
  adminPassword: "Ex4mpl3!Pass"
```

Deploy:

```shell
planton apply -f mysql-db.yaml
```

This creates a single-instance MySQL DB System with Oracle-managed encryption and the default MySQL configuration for the selected shape. The DB System ID, endpoint hostname, IP address, and port are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the MySQL DB System will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `availabilityDomain` | `string` | Availability domain for the primary endpoint (e.g., `Uocm:PHX-AD-1`). Changing this forces recreation. | Min length 1 |
| `shapeName` | `string` | Compute shape for the DB System. Determines CPU, memory, and network bandwidth (e.g., `MySQL.VM.Standard.E4.1.8GB`). | Min length 1 |
| `subnetId` | `StringValueOrRef` | OCID of the subnet where the DB System will be placed. Can reference an OciSubnet resource via `valueFrom`. Changing this forces recreation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. Falls back to `metadata.name` if not provided. |
| `adminUsername` | `string` | — | Administrative username for the database. Changing this forces recreation. |
| `adminPassword` | `string` | — | Administrative password. Must be 8-32 characters with at least one numeric, one lowercase, one uppercase, and one special character. Changing this forces recreation. |
| `mysqlVersion` | `string` | latest | MySQL version identifier (e.g., `8.0.36`, `9.1.0`). When omitted, the latest available version is used. Changing this forces recreation. |
| `configurationId` | `StringValueOrRef` | — | OCID of a MySQL Configuration defining server variable settings. When omitted, the default configuration for the selected shape is used. |
| `isHighlyAvailable` | `bool` | — | When `true`, provisions three instances across different fault domains with automatic failover. |
| `hostnameLabel` | `string` | — | Hostname for the primary endpoint. Combined with the subnet's DNS domain to form the FQDN. |
| `ipAddress` | `string` | — | Specific private IP for the primary endpoint. When omitted, OCI auto-assigns. Changing this forces recreation. |
| `faultDomain` | `string` | — | Fault domain for the primary endpoint (e.g., `FAULT-DOMAIN-1`). Changing this forces recreation. |
| `port` | `int32` | `3306` | TCP port for the MySQL protocol. Changing this forces recreation. |
| `portX` | `int32` | `33060` | TCP port for the X Protocol (MySQL Shell, connectors). Changing this forces recreation. |
| `description` | `string` | — | User-provided description of the DB System. |
| `crashRecovery` | `string` | — | Controls InnoDB crash recovery. Values: `ENABLED`, `DISABLED`. Disabling improves write performance but risks data loss. |
| `databaseManagement` | `string` | — | Enables monitoring via OCI Database Management service. Values: `ENABLED`, `DISABLED`. |
| `nsgIds` | `StringValueOrRef[]` | — | OCIDs of network security groups for the DB System VNIC. Can reference OciSecurityGroup resources. |
| `dataStorage` | `DataStorage` | — | Data storage configuration. See [DataStorage](#datastorage). |
| `backupPolicy` | `BackupPolicy` | — | Automatic backup configuration. See [BackupPolicy](#backuppolicy). |
| `maintenance` | `Maintenance` | — | Maintenance window configuration. See [Maintenance](#maintenance). |
| `deletionPolicy` | `DeletionPolicy` | — | Deletion safety configuration. See [DeletionPolicy](#deletionpolicy). |
| `encryptData` | `EncryptData` | — | Data-at-rest encryption configuration. See [EncryptData](#encryptdata). |
| `secureConnections` | `SecureConnections` | — | TLS certificate configuration for client connections. See [SecureConnections](#secureconnections). |
| `customerContacts` | `CustomerContact[]` | — | Email addresses for operational notifications. Maximum 10 contacts. See [CustomerContact](#customercontact). |
| `readEndpoint` | `ReadEndpoint` | — | Read-only endpoint for read scaling. See [ReadEndpoint](#readendpoint). |
| `databaseConsole` | `DatabaseConsole` | — | Web-based MySQL management console. See [DatabaseConsole](#databaseconsole). |
| `rest` | `Rest` | — | MySQL REST API service configuration. See [Rest](#rest). |

### DataStorage

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dataStorageSizeInGb` | `int32` | — | Initial data volume size in gigabytes. Minimum depends on shape (typically 50 GB). |
| `isAutoExpandStorageEnabled` | `bool` | — | When `true`, storage automatically expands when usage nears the limit. |
| `maxStorageSizeInGbs` | `int32` | — | Maximum storage size in GB for auto-expansion. Range: 32768-131072 depending on initial size. Only effective when `isAutoExpandStorageEnabled` is `true`. |

### BackupPolicy

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isEnabled` | `bool` | — | Whether automatic backups are enabled. |
| `retentionInDays` | `int32` | — | Number of days to retain automatic backups. |
| `windowStartTime` | `string` | — | Start of the 30-minute daily backup window in RFC 3339 time format (e.g., `03:00`). When omitted, OCI selects the window. |
| `pitrPolicy` | `PitrPolicy` | — | Point-in-time recovery configuration. See [PitrPolicy](#pitrpolicy). |

### PitrPolicy

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isEnabled` | `bool` | — | Whether point-in-time recovery is enabled. Requires automatic backups to be enabled. |

### Maintenance

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `windowStartTime` | `string` | — | Start of the maintenance window. Format: `{day-of-week} {time-of-day}` (e.g., `mon 10:00`). Required when maintenance is configured. |
| `maintenanceScheduleType` | `enum` | — | When maintenance patches are applied. Values: `early` (receive patches earlier), `regular` (standard Oracle schedule). |
| `versionPreference` | `enum` | — | Version selected during upgrades. Values: `oldest`, `second_newest`, `newest`. |
| `versionTrackPreference` | `enum` | — | MySQL release stream to follow. Values: `long_term_support`, `innovation`, `follow` (OCI-recommended). |

### DeletionPolicy

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `automaticBackupRetention` | `string` | — | What to do with automatic backups on deletion. Values: `DELETE`, `RETAIN`. |
| `finalBackup` | `string` | — | Whether to create a final backup before deletion. Values: `REQUIRE_FINAL_BACKUP`, `SKIP_FINAL_BACKUP`. |
| `isDeleteProtected` | `bool` | — | When `true`, the DB System cannot be deleted until this is set to `false`. |

### EncryptData

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `keyGenerationType` | `enum` | — | Encryption key strategy. Values: `system` (Oracle-managed), `byok` (Bring Your Own Key — requires `keyId`). |
| `keyId` | `StringValueOrRef` | — | OCID of the customer-managed encryption key. Required when `keyGenerationType` is `byok`. |

### SecureConnections

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `certificateGenerationType` | `enum` | — | TLS certificate strategy. Values: `system_cert` (Oracle-managed), `byoc` (Bring Your Own Certificate — requires `certificateId`). |
| `certificateId` | `StringValueOrRef` | — | OCID of the customer-managed certificate. Required when `certificateGenerationType` is `byoc`. |

### CustomerContact

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `email` | `string` | — | Email address for operational notifications (maintenance windows, critical alerts). |

### ReadEndpoint

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isEnabled` | `bool` | — | Whether the read endpoint is enabled. |
| `excludeIps` | `string[]` | — | IP addresses to exclude from serving read requests. |
| `readEndpointHostnameLabel` | `string` | — | Hostname for the read endpoint. Combined with the subnet's DNS domain to form the FQDN. |
| `readEndpointIpAddress` | `string` | — | Specific private IP for the read endpoint. When omitted, OCI auto-assigns. |

### DatabaseConsole

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | `enum` | — | Whether the console is active. Values: `enabled`, `disabled`. |
| `port` | `int32` | — | Port for the database console. Valid values: 443 or 1024-65535. |

### Rest

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `configuration` | `string` | — | REST API configuration mode. |
| `port` | `int32` | — | Port for the REST API service. Valid values: 443 or 1024-65535. |

## Examples

### Minimal Development Instance

A single-instance MySQL DB System with defaults — suitable for development or testing:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciMysqlDbSystem
metadata:
  name: dev-mysql
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciMysqlDbSystem.dev-mysql
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.1.8GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "admin"
  adminPassword: "Ex4mpl3!Pass"
```

### High Availability with Backups

HA enabled with daily backups, point-in-time recovery, and a weekly maintenance window:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciMysqlDbSystem
metadata:
  name: ha-mysql
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OciMysqlDbSystem.ha-mysql
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.4.64GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "admin"
  adminPassword: "Pr0d$ecure!99"
  mysqlVersion: "8.0.36"
  isHighlyAvailable: true
  dataStorage:
    dataStorageSizeInGb: 200
    isAutoExpandStorageEnabled: true
    maxStorageSizeInGbs: 32768
  backupPolicy:
    isEnabled: true
    retentionInDays: 14
    windowStartTime: "03:00"
    pitrPolicy:
      isEnabled: true
  maintenance:
    windowStartTime: "sun 04:00"
    maintenanceScheduleType: regular
    versionPreference: second_newest
    versionTrackPreference: long_term_support
```

### Production with BYOK Encryption, Deletion Protection, and Read Endpoint

Full production configuration with customer-managed encryption, deletion safeguards, read scaling, NSG attachment, and customer contact notifications:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciMysqlDbSystem
metadata:
  name: prod-mysql
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciMysqlDbSystem.prod-mysql
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.8.128GB"
  subnetId:
    value: "ocid1.subnet.oc1.phx..example"
  adminUsername: "dbadmin"
  adminPassword: "Pr0d!Str0ng#42"
  mysqlVersion: "8.0.36"
  isHighlyAvailable: true
  hostnameLabel: "prod-mysql"
  faultDomain: "FAULT-DOMAIN-1"
  port: 3306
  portX: 33060
  description: "Production MySQL for order processing"
  crashRecovery: "ENABLED"
  databaseManagement: "ENABLED"
  nsgIds:
    - value: "ocid1.networksecuritygroup.oc1.phx..example"
  dataStorage:
    dataStorageSizeInGb: 500
    isAutoExpandStorageEnabled: true
    maxStorageSizeInGbs: 65536
  backupPolicy:
    isEnabled: true
    retentionInDays: 30
    windowStartTime: "02:00"
    pitrPolicy:
      isEnabled: true
  maintenance:
    windowStartTime: "sun 05:00"
    maintenanceScheduleType: regular
    versionPreference: oldest
    versionTrackPreference: long_term_support
  deletionPolicy:
    automaticBackupRetention: "RETAIN"
    finalBackup: "REQUIRE_FINAL_BACKUP"
    isDeleteProtected: true
  encryptData:
    keyGenerationType: byok
    keyId:
      value: "ocid1.key.oc1.phx..example"
  secureConnections:
    certificateGenerationType: system_cert
  customerContacts:
    - email: "dba-team@example.com"
    - email: "oncall@example.com"
  readEndpoint:
    isEnabled: true
    readEndpointHostnameLabel: "prod-mysql-ro"
  databaseConsole:
    status: enabled
    port: 443

```

### Using Foreign Key References

Reference Planton-managed compartment and subnet instead of hardcoding OCIDs:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciMysqlDbSystem
metadata:
  name: ref-mysql
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciMysqlDbSystem.ref-mysql
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:PHX-AD-1"
  shapeName: "MySQL.VM.Standard.E4.4.64GB"
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: db-subnet
      fieldPath: status.outputs.subnetId
  adminUsername: "admin"
  adminPassword: "R3fPass!word1"
  isHighlyAvailable: true
  nsgIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: mysql-nsg
        fieldPath: status.outputs.networkSecurityGroupId
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `db_system_id` | `string` | OCID of the MySQL DB System |
| `endpoint_hostname` | `string` | Hostname of the primary (read/write) endpoint |
| `endpoint_ip_address` | `string` | Private IP address of the primary (read/write) endpoint |
| `endpoint_port` | `string` | TCP port of the primary (read/write) endpoint |

## Related Components

- [OciVcn](/docs/catalog/oci/vcn) — provides the VCN containing the subnet where the DB System is placed
- [OciSubnet](/docs/catalog/oci/subnet) — provides the subnet referenced by `subnetId`
- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — manages network security rules attached via `nsgIds`

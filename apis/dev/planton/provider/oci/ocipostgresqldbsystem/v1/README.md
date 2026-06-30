# OciPostgresqlDbSystem

Planton component for deploying an Oracle Cloud Infrastructure PostgreSQL Database System — a fully managed PostgreSQL service running on dedicated compute shapes with configurable storage durability, flexible sizing, and built-in backup policies.

## Purpose

This component wraps the OCI PostgreSQL DB System (`oci_psql_db_system`) into a declarative YAML resource. It provisions a managed PostgreSQL instance in a specified compartment and subnet, handling compute shape selection, storage backend configuration, credential management, backup scheduling, and maintenance windows through a single manifest.

## Key Features

- **Flexible compute shapes** — supports fixed and flexible shapes with per-instance OCPU and memory configuration
- **Read replicas** — set `instanceCount` to 2 or more to add read replicas behind a reader endpoint
- **Dual storage durability modes** — regionally durable (multi-AD replication) or AD-local placement for cost-optimized development
- **Configurable IOPS** — set a guaranteed IOPS tier for the storage backend
- **Credential management** — supports plain-text passwords for development and OCI Vault secret references for production
- **Backup policies** — daily, weekly, monthly, or disabled backup schedules with configurable retention
- **Maintenance windows** — specify when OCI can apply patches and updates
- **Custom PostgreSQL configuration** — apply server parameter tuning via `configId` referencing an OCI PostgreSQL Configuration resource
- **Per-instance pinning** — assign display names, descriptions, and private IPs to individual nodes
- **Foreign key references** — reference compartments, subnets, NSGs, and Vault secrets from other Planton resources via `valueFrom`
- **Automatic tagging** — freeform tags applied from resource metadata (kind, ID, organization, environment, custom labels)

## Critical Constraints

- **Credentials are immutable** — the entire `credentials` block (username and password details) cannot be changed after creation. Modifications force recreation of the DB System.
- **Storage durability is immutable** — changing `isRegionallyDurable` or `availabilityDomain` forces recreation.
- **Subnet is immutable** — changing the subnet OCID forces recreation.
- **Only fresh creation** — restoring from backup (`source` block) is not supported in v1.
- **No defined_tags or system_tags** — tagging is limited to freeform tags derived from metadata.
- **Storage system type is fixed** — hardcoded to `OCI_OPTIMIZED_STORAGE` (the only valid value).
- **Per-instance details are immutable** — `instancesDetails` entries cannot be modified after creation.

## Use Cases

### Development / Testing

Single-node, single-AD instance with plain-text password, small shape, and short backup retention:

```yaml
spec:
  shape: VM.Standard.E4.Flex
  instanceOcpuCount: 1
  instanceMemorySizeInGbs: 8
  instanceCount: 1
  storageDetails:
    isRegionallyDurable: false
    availabilityDomain: "Uocm:PHX-AD-1"
  credentials:
    username: postgres
    passwordDetails:
      passwordType: plain_text
      password: "dev-password"
```

### Production

Multi-node with Vault-managed credentials, regionally durable storage, NSG-secured networking, reader endpoint, and daily backups:

```yaml
spec:
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
```

## Production Features

### Backup Configuration

The `managementPolicy.backupPolicy` block supports four schedule kinds:

| Kind | Required Fields | Description |
|------|----------------|-------------|
| `daily` | `backupStart` | Daily backup at the specified hour (UTC) |
| `weekly` | `backupStart`, `daysOfTheWeek` | Backup on specified weekdays |
| `monthly` | `backupStart`, `daysOfTheMonth` | Backup on specified days of the month (1-28) |
| `none` | — | Backups disabled |

### Vault Secret Integration

For production credential management, use `vault_secret` password type with a reference to an OCI Vault secret:

```yaml
credentials:
  username: postgres
  passwordDetails:
    passwordType: vault_secret
    secretId:
      value: "ocid1.vaultsecret.oc1..example"
    secretVersion: "1"
```

When `secretVersion` is omitted, the latest version is used.

### Stack Outputs

| Output | Description |
|--------|-------------|
| `dbSystemId` | OCID of the PostgreSQL DB System |
| `primaryDbEndpointPrivateIp` | Private IP of the primary (read-write) endpoint |
| `adminUsername` | Administrator username |

These outputs can be referenced by downstream components via `valueFrom`.

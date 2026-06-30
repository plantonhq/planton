# OCI Autonomous Database

## Overview

The OciAutonomousDatabase component deploys an Oracle Cloud Infrastructure Autonomous Database instance. It wraps the `oci_database_autonomous_database` Pulumi resource and exposes the full set of configuration options through a declarative YAML manifest.

Autonomous Database is OCI's fully managed database service. It handles patching, tuning, scaling, and backups without manual intervention. A single component definition covers all five workload types — ATP (OLTP), ADW (Data Warehouse), AJD (JSON), APEX, and Lakehouse — selected via the `dbWorkload` field.

## Purpose

This component exists to give teams a single, version-controlled way to provision and manage Autonomous Databases across environments. Instead of clicking through the OCI Console or writing raw Pulumi programs, operators declare the desired state in a manifest and let Planton handle provisioning, tagging, and output wiring.

## Key Features

- **All workload types in one component** — ATP, ADW, AJD, APEX, and Lakehouse are configured through the same manifest. The `dbWorkload` field selects the optimizer profile and feature set.
- **Compute model choice** — supports both ECPU (current recommended model) and OCPU (legacy). Auto-scaling can independently scale compute up to 3x provisioned capacity.
- **Storage flexibility** — specify capacity in terabytes (serverless) or gigabytes (dedicated Exadata). Storage auto-scaling is available as a separate toggle.
- **Private endpoint networking** — optionally place the database behind a private endpoint in a VCN subnet, with NSG and IP allowlist controls.
- **Vault-managed credentials** — reference an OCI Vault secret for the admin password instead of embedding it in the manifest.
- **Customer-managed encryption** — bring your own KMS key and vault for Transparent Data Encryption.
- **Data Guard** — enable local Autonomous Data Guard for automatic standby provisioning in a different availability domain.
- **Foreign key references** — `compartmentId`, `subnetId`, `nsgIds`, `kmsKeyId`, `vaultId`, `secretId`, and `autonomousContainerDatabaseId` all accept `valueFrom` references to other Planton-managed resources.
- **Automatic tagging** — freeform tags are applied from metadata labels, organization, environment, resource kind, and resource ID.

## How ADB Differs from Other Managed Databases

OCI offers several managed database services. Here is how Autonomous Database compares to others available in the Planton catalog:

| Aspect | Autonomous Database | OCI DB System | OCI MySQL DB System | OCI PostgreSQL DB System |
|--------|-------------------|---------------|---------------------|--------------------------|
| Engine | Oracle | Oracle | MySQL | PostgreSQL |
| Management | Fully autonomous (self-patching, self-tuning) | Customer-managed patching and tuning | Managed by OCI with manual config | Managed by OCI with manual config |
| Workload types | OLTP, DW, JSON, APEX, Lakehouse | General purpose | OLTP | OLTP |
| Scaling | Auto-scaling for compute and storage | Manual scaling | Manual scaling | Manual scaling |
| Infrastructure | Shared (serverless) or dedicated Exadata | Bare metal or VM | VM | VM |

Use Autonomous Database when you want OCI to handle operational tasks automatically. Use DB System when you need full control over the Oracle instance. Use MySQL/PostgreSQL DB System when the application requires those engines.

## Critical Constraints

- **`dbName` is immutable** — cannot be changed after creation.
- **`characterSet` and `ncharacterSet` are immutable** — set them correctly at creation time.
- **`isDedicated` is immutable** — switching between shared and dedicated infrastructure requires recreating the database.
- **Storage fields are mutually exclusive** — set exactly one of `dataStorageSizeInTbs` or `dataStorageSizeInGb`, not both.
- **Credential fields are mutually exclusive** — set exactly one of `adminPassword` or `secretId`, not both.
- **Always Free databases** have fixed compute and storage limits, cannot scale, and are reclaimed after extended inactivity.
- **AJD and APEX workloads** always use LICENSE_INCLUDED regardless of the `licenseModel` setting.
- **NSGs require a subnet** — `nsgIds` is only applicable when `subnetId` is set.

## Use Cases

### Development and Testing
Use `isFreeTier: true` or `isDevTier: true` to provision low-cost databases for development. Combine with `dbWorkload: oltp` for application testing or `dbWorkload: dw` for query development.

### Transaction Processing
Set `dbWorkload: oltp` with ECPU compute and auto-scaling enabled. Place the database behind a private endpoint in the application VCN for low-latency access.

### Analytics and Reporting
Set `dbWorkload: dw` with higher compute counts and BYOL licensing. Auto-scaling handles burst analytic queries without over-provisioning.

### JSON Document Store
Set `dbWorkload: ajd` for applications that primarily store and query JSON documents. AJD includes SODA (Simple Oracle Document Access) APIs.

### Low-Code Applications
Set `dbWorkload: apex` for Oracle APEX application development. The database includes a built-in APEX workspace.

## Production Features

### High Availability
Enable `isLocalDataGuardEnabled` to provision an automatic standby in a different availability domain. Failover is automatic and transparent to applications using the connection strings.

### Security
- Use `secretId` to reference Vault-managed credentials instead of plaintext passwords.
- Set `kmsKeyId` and `vaultId` for customer-managed TDE encryption.
- Enable `isMtlsConnectionRequired` to enforce mutual TLS for all connections.
- Use `subnetId` and `nsgIds` for network isolation.
- Use `whitelistedIps` for IP-based access control.

### Operations
- Set `autonomousMaintenanceScheduleType` to control when patches are applied (`early` or `regular`).
- Add `customerContacts` for email notifications about maintenance windows and critical alerts.
- Configure `backupRetentionPeriodInDays` to control automatic backup retention.

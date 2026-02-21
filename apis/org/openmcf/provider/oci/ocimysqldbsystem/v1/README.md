# OciMysqlDbSystem

## Overview

OciMysqlDbSystem deploys an Oracle Cloud Infrastructure MySQL HeatWave Database System through OpenMCF. It creates a single `oci_mysql_mysql_db_system` resource with configurable High Availability, automated backups, point-in-time recovery, read endpoints, encryption, and maintenance windows.

The component manages the DB System resource itself. HeatWave cluster attachment and replication channels are separate OCI resources with independent lifecycles and are not included in this component.

## Purpose

Provide a declarative, version-controlled way to provision MySQL HeatWave DB Systems on OCI. A single YAML manifest defines the compute shape, networking placement, HA mode, storage, backup policy, encryption, and operational controls. OpenMCF translates the manifest into Pulumi IaC and manages the full create/update lifecycle.

## Key Features

- **Single-resource focus** — creates one `oci_mysql_mysql_db_system` resource; no implicit sidecar resources
- **High Availability** — optional three-instance HA across fault domains with automatic failover via `isHighlyAvailable`
- **Automated backups** — configurable daily backup window, retention period, and point-in-time recovery through `backupPolicy`
- **Read scaling** — optional read-only endpoint via `readEndpoint` distributes queries across HA replicas
- **BYOK encryption** — supports both Oracle-managed and customer-managed (Vault) encryption keys through `encryptData`
- **BYOC TLS** — supports both Oracle-managed and customer-managed TLS certificates through `secureConnections`
- **Foreign key references** — `compartmentId`, `subnetId`, `configurationId`, `nsgIds`, `keyId`, and `certificateId` accept both literal OCIDs and cross-resource references via `valueFrom`
- **Deletion protection** — `deletionPolicy` controls backup retention on delete, final backup creation, and delete-protection toggle
- **Freeform tags** — automatically applies resource kind, resource ID, organization, and environment as OCI freeform tags

## Critical Constraints

- **Fresh creation only** — the `source` block (BACKUP, PITR, IMPORTURL) is excluded in v1; only new DB System creation is supported
- **No HeatWave cluster** — HeatWave in-memory analytics cluster attachment is a separate OCI resource and not managed here
- **No replication channels** — MySQL replication channels are separate OCI resources
- **No operational lifecycle controls** — `shutdown_type`, `state`, `access_mode`, and `database_mode` are excluded
- **No ZPR security attributes** — `security_attributes` are not exposed
- **No cross-region backup copy** — `backup_policy.copy_policies` is excluded
- **Recreation-triggering fields** — changing `availabilityDomain`, `subnetId`, `adminUsername`, `adminPassword`, `mysqlVersion`, `ipAddress`, `faultDomain`, `port`, or `portX` forces resource recreation

## Use Cases

1. **Development database** — minimal single-instance DB System with defaults for local application development
2. **Staging environment** — HA-enabled DB System with backups and a maintenance window that mirrors production
3. **Production workload** — full configuration with BYOK encryption, deletion protection, read endpoints, NSG attachment, PITR, and customer contact notifications
4. **Multi-resource composition** — reference OpenMCF-managed compartments, subnets, and NSGs via `valueFrom` to compose a complete stack from independent manifests

## Production Features

| Feature | Spec Field | Notes |
|---------|-----------|-------|
| High Availability | `isHighlyAvailable` | Three instances across fault domains |
| Automated Backups | `backupPolicy` | Daily window, retention, PITR |
| Read Scaling | `readEndpoint` | Separate read-only DNS endpoint |
| BYOK Encryption | `encryptData` | Customer-managed Vault keys |
| BYOC TLS | `secureConnections` | Customer-managed certificates |
| Deletion Protection | `deletionPolicy.isDeleteProtected` | Prevents accidental deletion |
| Final Backup on Delete | `deletionPolicy.finalBackup` | `REQUIRE_FINAL_BACKUP` or `SKIP_FINAL_BACKUP` |
| NSG Attachment | `nsgIds` | Network security group VNIC rules |
| Crash Recovery Control | `crashRecovery` | Toggle InnoDB crash recovery |
| Database Management | `databaseManagement` | OCI Database Management monitoring |
| Database Console | `databaseConsole` | Web-based MySQL management UI |
| REST API | `rest` | MySQL Router REST API |
| Customer Contacts | `customerContacts` | Up to 10 notification emails |
| Auto-Expand Storage | `dataStorage.isAutoExpandStorageEnabled` | Automatic volume growth |

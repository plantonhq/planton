---
title: "High Availability"
description: "This preset creates a production MySQL HeatWave DB System with High Availability enabled. Three instances are provisioned across fault domains with automatic failover, PITR-enabled backups with..."
type: "preset"
rank: "01"
presetSlug: "01-high-availability"
componentSlug: "mysql-db-system"
componentTitle: "MySQL DB System"
provider: "oci"
icon: "package"
order: 1
---

# High Availability

This preset creates a production MySQL HeatWave DB System with High Availability enabled. Three instances are provisioned across fault domains with automatic failover, PITR-enabled backups with 30-day retention, auto-expanding storage, and delete protection. This is the standard pattern for production MySQL workloads on OCI.

## When to Use

- Production application databases requiring automatic failover on instance or fault domain failure
- Workloads where recovery point objective (RPO) near zero is needed via point-in-time recovery
- MySQL databases backing SaaS applications, e-commerce platforms, or CMS systems
- Environments requiring delete protection to prevent accidental data loss

## Key Configuration Choices

- **High Availability** (`isHighlyAvailable: true`) -- provisions three MySQL instances across fault domains. The primary handles read/write traffic; standbys receive synchronous replication and automatically promote on failure. Failover is transparent to applications.
- **4 OCPUs, 64 GB RAM** (`shapeName: MySQL.VM.Standard.E4.4.64GB`) -- provides a solid production baseline for buffer pool, query processing, and connection handling. Scale up by selecting a larger shape.
- **Auto-expanding storage** (`isAutoExpandStorageEnabled: true`) -- starts at 200 GB and grows automatically up to 32 TB, eliminating storage-related outages.
- **PITR enabled** (`pitrPolicy.isEnabled: true`) -- allows restoring the database to any point within the backup retention window, providing fine-grained recovery beyond full-backup granularity.
- **30-day backup retention** with required final backup -- ensures recovery capability for the last 30 days and forces a final backup on deletion to prevent data loss.
- **Delete protection** (`isDeleteProtected: true`) -- the DB System cannot be deleted until this flag is explicitly set to false, preventing accidental destruction.
- **Crash recovery enabled** (`crashRecovery: ENABLED`) -- InnoDB redo logs, double write buffer, and binary log syncing ensure data durability on unexpected failure.
- **LTS version track** (`versionTrackPreference: long_term_support`) -- follows MySQL Long-Term Support releases for maximum stability.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the DB System | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<availability-domain>` | AD for the primary endpoint (e.g., `Uocm:PHX-AD-1`) | OCI Console > Compute > Availability Domains |
| `<private-subnet-ocid>` | OCID of the private subnet for the DB System | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<admin-password>` | Admin password (8-32 chars, must include numeric, lowercase, uppercase, and special character) | Generate a strong password |
| `<db-nsg-ocid>` | OCID of the NSG allowing MySQL traffic (port 3306) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |

## Related Presets

- **02-standalone-development** -- Use instead for cost-optimized non-production environments without HA

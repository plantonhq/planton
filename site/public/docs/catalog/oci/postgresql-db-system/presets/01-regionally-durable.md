---
title: "Regionally Durable"
description: "This preset creates a production PostgreSQL DB System with regionally durable storage (replicated across availability domains), a read replica for scaling read queries, reader endpoint for automatic..."
type: "preset"
rank: "01"
presetSlug: "01-regionally-durable"
componentSlug: "postgresql-db-system"
componentTitle: "PostgreSQL DB System"
provider: "oci"
icon: "package"
order: 1
---

# Regionally Durable

This preset creates a production PostgreSQL DB System with regionally durable storage (replicated across availability domains), a read replica for scaling read queries, reader endpoint for automatic read distribution, and daily backups with 30-day retention. The admin password is sourced from OCI Vault for production-grade secret management.

## When to Use

- Production PostgreSQL databases requiring storage-level durability across availability domains
- Applications that benefit from read replicas (read-heavy workloads, reporting queries, analytics dashboards)
- Environments requiring Vault-managed credentials instead of plain-text passwords in manifests
- Any PostgreSQL workload where AD-level failure must not cause data loss

## Key Configuration Choices

- **Regionally durable storage** (`isRegionallyDurable: true`) -- data is replicated across multiple availability domains. If an entire AD becomes unavailable, the data survives. This is the recommended setting for production.
- **2 instances** (`instanceCount: 2`) -- one primary (read-write) and one read replica. The reader endpoint distributes read queries across replicas. Add more replicas by increasing `instanceCount`.
- **Reader endpoint enabled** (`isReaderEndpointEnabled: true`) -- creates a separate DNS endpoint that load-balances read queries across replicas, offloading the primary for writes.
- **PostgreSQL 16** (`dbVersion: "16"`) -- the latest major version with performance improvements and logical replication enhancements.
- **4 OCPUs, 32 GB RAM** per instance -- provides a solid production baseline. Flex shapes allow independent scaling of CPU and memory.
- **Vault secret for password** (`passwordType: vault_secret`) -- the admin password is stored in OCI Vault and referenced by OCID, keeping credentials out of the manifest.
- **Daily backups with 30-day retention** -- provides daily recovery points for the last month.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the DB System | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<private-subnet-ocid>` | OCID of the private subnet for the DB System | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<db-nsg-ocid>` | OCID of the NSG allowing PostgreSQL traffic (port 5432) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |
| `<vault-secret-ocid>` | OCID of the Vault secret containing the admin password | OCI Console > Security > Vault > Secrets, or `OciVaultSecret` outputs |

## Related Presets

- **02-standalone-development** -- Use instead for cost-optimized non-production environments with AD-local storage

# AzureMssqlDatabase

An Azure SQL Database on an AzureMssqlServer logical server. In Azure
SQL's model the DATABASE is the unit of compute and billing: it carries
its own SKU (DTU, vCore, serverless, or Hyperscale), storage ceiling,
availability posture, backup policy, and encryption -- or joins an
AzureMssqlElasticPool to share pooled compute instead.

## When to Use

Use AzureMssqlDatabase when you need:

- **Any database on an Azure SQL logical server** -- one resource per
  database, composed against the server's `server_id` output
- **Serverless economics** -- per-second billing with auto-pause
  (`GP_S_`/`HS_S_` SKUs)
- **Hyperscale** -- elastic storage to 100 TB with readable replicas
- **Pool membership** -- `sku_name: ElasticPool` + `elastic_pool_id`
- **Copies, secondaries, and restores** -- a copy/DR/restored database
  is another AzureMssqlDatabase whose `create_mode` and source reference
  the original
- **Real backup postures** -- short-term PITR (1-35 days) plus
  long-term weekly/monthly/yearly retention
- **Database-scoped CMK** -- override the server's TDE key per database

## Key Configuration

### SKUs (`sku_name`)

| Family | Examples | Notes |
|--------|----------|-------|
| DTU | `Basic`, `S0`-`S12`, `P1`-`P15` | Bundled compute/IO/storage |
| vCore | `GP_Gen5_2`, `BC_Gen5_4` | Independent compute/storage; Hybrid Benefit |
| Serverless | `GP_S_Gen5_1`, `HS_S_Gen5_2` | Auto-pause + per-second billing |
| Hyperscale | `HS_Gen5_2` | Elastic storage to 100 TB, readable replicas |
| Pooled | `ElasticPool` | Compute from the referenced pool |
| Warehouse/Stretch | `DW100c`, `DS100` | Analytics tiers |

Leaving/entering Hyperscale and changing `enclave_type` REPLACE the
database (ARM's contract).

### Lifecycle (`create_mode`)

| Mode | Source field |
|------|--------------|
| DEFAULT (unset) | -- |
| COPY / SECONDARY / ONLINE_SECONDARY / POINT_IN_TIME_RESTORE | `creation_source_database_id` (+ `restore_point_in_time` for PITR, `secondary_type` for secondaries) |
| RECOVERY | `recover_database_id` / `recovery_point_id` |
| RESTORE | `restore_dropped_database_id` |
| RESTORE_LONG_TERM_RETENTION_BACKUP | `restore_long_term_retention_backup_id` |

### Backups

- `short_term_retention_policy` -- the PITR window (1-35 days) +
  differential cadence (12/24h)
- `long_term_retention_policy` -- weekly/monthly/yearly horizons as
  ISO-8601 durations (`P5W`, `P12M`, `P5Y`)
- `storage_account_type` -- where backups replicate (geo / geo-zone /
  zone / local)

## Outputs

| Output | Description |
|--------|-------------|
| `database_id` | ARM ID -- the seam copy/secondary/restore databases reference |
| `database_name` | The `Database=` segment of connection strings against the server's fqdn |

## Related Resources

- **AzureMssqlServer** -- the parent logical server (`server_id`)
- **AzureMssqlElasticPool** -- pooled compute (`elastic_pool_id`)
- **AzureUserAssignedIdentity** -- the database-scoped CMK unwrap identity
- **AzureKeyVaultKey** -- the database-scoped TDE key (versioned `key_id`)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

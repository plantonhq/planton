# AzureMssqlElasticPool

An Azure SQL elastic pool on an AzureMssqlServer logical server: a
shared-compute container that member databases (AzureMssqlDatabase with
`sku_name: ElasticPool` + `elastic_pool_id`) draw from instead of
carrying their own SKU -- the right economics for many small databases
with non-overlapping usage peaks (the SaaS tenant-per-database pattern).

## When to Use

Use AzureMssqlElasticPool when you need:

- **Shared compute for a database fleet** -- pay for the pool's peak,
  not the sum of per-database peaks
- **Per-database governance** -- guarantee every tenant a floor
  (`min_capacity`) and cap noisy neighbors (`max_capacity`)
- **One maintenance window for the fleet** -- member databases inherit
  the pool's

## Key Configuration

### SKUs (`sku_name` + `capacity`)

| Family | Names | Capacity unit |
|--------|-------|---------------|
| DTU | `BasicPool`, `StandardPool`, `PremiumPool` | eDTUs (e.g. 50-4000) |
| General Purpose | `GP_Gen5`, `GP_Fsv2`, `GP_DC` | vCores |
| Business Critical | `BC_Gen5`, `BC_DC` | vCores |
| Hyperscale | `HS_Gen5`, `HS_PRMS`, `HS_MOPRMS` | vCores |

The service tier and hardware family ARM wants alongside the name are
derived from it on both engines -- a name/tier/family mismatch is
unrepresentable.

### Sizing

- `per_database_settings` (required) -- `min_capacity` is guaranteed per
  database (0 lets the pool oversubscribe), `max_capacity` caps any one
  database
- `max_size_gb` XOR `max_size_bytes` -- the pool's total storage cap
- `region` must match the parent server's region (ARM rejects a
  mismatch)

## Outputs

| Output | Description |
|--------|-------------|
| `elastic_pool_id` | ARM ID -- the seam `AzureMssqlDatabase.elastic_pool_id` references |
| `elastic_pool_name` | The pool's name |

## Related Resources

- **AzureMssqlServer** -- the parent logical server (`server_id`)
- **AzureMssqlDatabase** -- member databases joining via `elastic_pool_id`

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# AliCloudPolardbCluster Pulumi Examples

See [../../examples.md](../../examples.md) for complete YAML examples that drive this Pulumi module.

## Stack Input

The Pulumi module receives its configuration via `AliCloudPolardbClusterStackInput`, which contains:

- `target` -- the full `AliCloudPolardbCluster` resource (metadata + spec)
- `provider_config` -- optional Alibaba Cloud credential overrides

## Outputs

| Key | Description |
|-----|-------------|
| `cluster_id` | PolarDB cluster ID |
| `connection_string` | Primary endpoint connection string |
| `port` | Database service port |
| `database_ids` | Map of database names to IDs |

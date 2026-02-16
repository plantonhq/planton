# GcpMemorystoreInstance Pulumi Module Overview

## Module Architecture

The Pulumi module provisions a Google Cloud Memorystore instance (new-generation API)
using `memorystore.NewInstance` from `pulumi-gcp` SDK v9. It creates a single instance
with PSC endpoints, persistence, sharding, and CMEK, then extracts discovery endpoint
details from the nested PSC connection structure.

### File Organization

```
iac/pulumi/
├── main.go              # Entrypoint: loads stack input, calls module.Resources
├── Pulumi.yaml          # Project definition
└── module/
    ├── main.go                    # Resources(): orchestrates provider + resource creation
    ├── locals.go                  # initializeLocals(): derives labels from stack input
    ├── outputs.go                 # Output key constants matching stack_outputs.proto
    └── memorystore_instance.go    # memorystoreInstance(): creates the resource
```

### Control Flow

```
StackInput (GcpMemorystoreInstanceStackInput)
   ↓
initializeLocals() → Locals (labels, references)
   ↓
pulumigoogleprovider.Get() → GCP Provider
   ↓
memorystoreInstance(ctx, locals, gcpProvider)
   ├── Build InstanceArgs (conditional field setting)
   ├── memorystore.NewInstance()
   ├── ApplyT(Endpoints) → discovery_address, discovery_port
   ├── ApplyT(NodeConfigs) → node_size_gb
   └── ctx.Export() → 4 stack outputs
```

## Key Components

### Locals (locals.go)

Derives GCP labels from metadata: `openmcf-resource`, `openmcf-resource-name`,
`openmcf-resource-kind`, plus optional `openmcf-organization`, `openmcf-environment`,
and `openmcf-resource-id`.

### Memorystore Instance (memorystore_instance.go)

Creates a single `memorystore.Instance` with conditional field setting:
- **Required**: instance_id, location, shard_count, labels
- **Optional**: project, mode, node_type, engine_version, engine_configs,
  replica_count, psc_auto_connections, authorization_mode, transit_encryption_mode,
  kms_key, persistence_config, zone_distribution_config, maintenance_policy,
  automated_backup_config, deletion_protection_enabled

Discovery endpoint extraction uses `ApplyT` on `Endpoints` to walk the nested
`Endpoints → Connections → PscAutoConnection` structure, searching for
`CONNECTION_TYPE_DISCOVERY` with fallback to any available connection.

### Outputs (outputs.go)

Four outputs matching `GcpMemorystoreInstanceStackOutputs`:
- `discovery_address` — PSC discovery endpoint IP
- `discovery_port` — discovery endpoint port (typically 6379)
- `instance_uid` — server-generated unique identifier
- `node_size_gb` — memory per node in GB

## Design Decisions

1. **Single resource file** — Standalone resource with no sub-resources.
2. **Conditional field setting** — Optional fields only set when non-zero,
   letting GCP apply defaults.
3. **ApplyT for nested outputs** — PSC endpoint structure is deeply nested;
   ApplyT walks it with a discovery-type-first, fallback-to-any strategy.
4. **Labels from framework** — Standard OpenMCF labels for resource discovery
   and cost attribution.
5. **No defensive defaults** — OpenMCF middleware applies proto defaults upstream.

## Customization Guide

- **Engine configs**: Pass key-value pairs in `engine_configs` to tune Valkey
- **Scale horizontally**: Increase `shard_count` with `mode: CLUSTER`
- **Enable persistence**: Add `persistence_config` with RDB or AOF mode
- **Cross-project PSC**: Multiple `psc_auto_connections` entries with different project IDs

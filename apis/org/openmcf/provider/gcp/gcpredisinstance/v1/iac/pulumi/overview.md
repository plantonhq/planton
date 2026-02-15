# GcpRedisInstance Pulumi Module Overview

## Module Architecture

The Pulumi module provisions a Google Cloud Memorystore for Redis instance
using the `redis.NewInstance` resource from the `pulumi-gcp` SDK.

### File Organization

```
iac/pulumi/
├── main.go              # Entrypoint: loads stack input, calls module.Resources
├── Pulumi.yaml          # Project definition
└── module/
    ├── main.go          # Resources(): orchestrates provider setup and resource creation
    ├── locals.go        # initializeLocals(): derives labels and references from stack input
    ├── outputs.go       # Output key constants matching stack_outputs.proto
    └── redis_instance.go  # redisInstance(): creates the google_redis_instance resource
```

### Control Flow

```
StackInput (GcpRedisInstanceStackInput)
   ↓
initializeLocals() → Locals (labels, references)
   ↓
pulumigoogleprovider.Get() → GCP Provider
   ↓
redisInstance(ctx, locals, gcpProvider) → redis.NewInstance
   ↓
ctx.Export() → host, port, read_endpoint, read_endpoint_port, current_location_id, auth_string
```

## Key Components

### Locals (locals.go)

Derives GCP labels from metadata:
- `openmcf-resource: true`
- `openmcf-resource-name: <instance_name>`
- `openmcf-resource-kind: gcpredisinstance`
- `openmcf-organization: <org>` (if set)
- `openmcf-environment: <env>` (if set)
- `openmcf-resource-id: <id>` (if set)

### Redis Instance (redis_instance.go)

Creates a single `redis.Instance` resource with conditional field setting:
- Required: name, project, region, tier, memory_size_gb, labels
- Optional: redis_version, display_name, location_id, authorized_network,
  connect_mode, reserved_ip_range, auth_enabled, transit_encryption_mode,
  redis_configs, maintenance_policy, read_replicas_mode, replica_count,
  persistence_config, customer_managed_key, deletion_protection

All `StringValueOrRef` fields use `.GetValue()` directly.

### Outputs (outputs.go)

Six outputs exported matching `GcpRedisInstanceStackOutputs`:
- `host` — primary endpoint IP
- `port` — primary endpoint port
- `read_endpoint` — read replica endpoint (HA only)
- `read_endpoint_port` — read replica port (HA only)
- `current_location_id` — zone of primary
- `auth_string` — AUTH password (sensitive, only if auth_enabled)

## Design Decisions

1. **Single resource file** — Redis instance is a standalone resource with no
   sub-resources to manage, keeping the module flat and simple.

2. **Conditional field setting** — Optional fields are only set when non-zero,
   allowing GCP to apply its defaults (e.g., connect_mode defaults to
   DIRECT_PEERING, transit_encryption_mode defaults to DISABLED).

3. **Labels from framework** — Standard OpenMCF labels applied automatically,
   enabling consistent resource discovery and cost attribution.

4. **No defensive defaults** — OpenMCF middleware applies proto defaults before
   the module runs, so no `if x == "" { x = "default" }` patterns needed.

## Customization Guide

- **Add redis_configs**: Pass key-value pairs in spec to tune Redis behavior
  (e.g., maxmemory-policy, notify-keyspace-events)
- **Change networking**: Set `authorized_network` to a VPC and `connect_mode`
  to `PRIVATE_SERVICE_ACCESS` for Shared VPC environments
- **Enable persistence**: Add `persistence_config` with RDB mode for data
  durability across restarts

# OCI Redis Cluster — Design Notes

Internal design documentation for the OciRedisCluster component.

## Design Rationale

### Single-Resource Component

The component provisions exactly one OCI resource: `oci_redis_redis_cluster`. Unlike components such as OciVcn (which bundles gateways) or OciDbSystem (which bundles storage), the Redis cluster resource in OCI is self-contained. There are no child resources that need co-management, so the component is intentionally thin.

### Cluster Mode as an Enum

The `clusterMode` field is a proto enum (`cluster_mode_unspecified`, `nonsharded`, `sharded`) rather than a boolean. This matches the OCI API's string-based mode field and leaves room for future topologies without a breaking schema change. The Go module converts the enum to an uppercase string (`NONSHARDED`, `SHARDED`) for the Pulumi provider.

### CEL Validation for Shard Count

A CEL rule on `OciRedisClusterSpec` enforces that `shardCount > 0` when `clusterMode` is `sharded`. This catches misconfiguration at validation time rather than letting OCI return an opaque API error. The rule uses the enum's integer value (`!= 2` for non-sharded) to avoid string comparison.

### Display Name Defaulting

The Go module falls back `displayName` to `metadata.name` when the spec field is empty. This avoids requiring users to specify a name twice while still allowing an explicit override for OCI Console readability.

### Freeform Tag Propagation

Freeform tags are built from a fixed set of keys (`resource`, `resource_kind`, `resource_id`) plus optional `organization` and `environment` from metadata. All metadata labels are merged in. `defined_tags`, `system_tags`, and direct `freeform_tags` spec fields are excluded — tags are always derived, never user-specified.

### Foreign Key References

`compartmentId`, `subnetId`, `nsgIds`, and `configSetId` all use `StringValueOrRef` to support both literal OCIDs and cross-resource references via `valueFrom`. Default kind annotations on the proto fields define the expected source resource kind and field path, enabling tooling to validate references without runtime resolution.

## Trade-Offs

### No Config Set Management

OCI Cache Config Sets are separate resources with independent lifecycles. This component references them by OCID but does not create or manage them. This keeps the component focused on cluster lifecycle and avoids coupling two distinct resource lifecycles. The trade-off is that users must manage config sets through a separate component or out-of-band.

### No Tag Customization

Users cannot set arbitrary `freeform_tags` through the spec. All tags are derived from metadata. This prevents tag drift and ensures consistency but limits flexibility for teams that need custom tag keys beyond what metadata labels provide.

### Immutable Topology Fields

`clusterMode` and `subnetId` force recreation when changed. The component does not attempt to handle migration (e.g., data export/import between non-sharded and sharded clusters). Users must manage data migration externally if they need to change topology.

### No NSG Rule Management

`nsgIds` attaches existing NSGs to the cluster but does not create NSG rules. Security rules must be managed through the OciSecurityGroup component. This is consistent with OCI's model where NSGs and their rules are independent resources.

## What's Deferred to Future Versions

| Feature | Reason Deferred |
|---------|-----------------|
| `defined_tags` | Managed at the platform level; adding per-resource defined tags requires tag namespace coordination |
| `system_tags` | Read-only in OCI; set by the service, not by users |
| `security_attributes` | Oracle Zero-Trust Packet Routing is a specialized feature with limited adoption |
| Config Set component | Separate resource kind with its own lifecycle; planned as an independent component |
| Maintenance window | OCI Cache maintenance windows are managed at the cluster level but require additional scheduling fields |
| Cross-region replication | Not currently supported by the OCI Cache service |

## Module Structure

```
v1/
├── api.proto                  # Top-level OciRedisCluster message
├── spec.proto                 # OciRedisClusterSpec with ClusterMode enum
├── stack_outputs.proto        # Exported outputs (cluster ID, FQDNs, IP)
├── stack_input.proto          # Pulumi stack input wrapper
└── iac/pulumi/module/
    ├── main.go                # Entry point: initializes locals, provider, calls redisCluster()
    ├── locals.go              # Locals struct: display name defaulting, tag assembly
    ├── redis_cluster.go       # Creates oci_redis_redis_cluster, exports outputs
    └── outputs.go             # Output key constants
```

The module follows the standard Planton Pulumi module pattern: `main.go` orchestrates, `locals.go` prepares derived values, resource-specific files create infrastructure, and `outputs.go` defines constant keys for stack exports.

# AwsRedisElasticache Pulumi Module Architecture

## Module Structure

```
module/
├── main.go              # Entry point: provider setup, orchestration
├── locals.go            # Locals struct: tags, spec references
├── outputs.go           # Output key constants matching AwsRedisElasticacheStackOutputs
├── subnet_group.go      # Conditional ElastiCache subnet group creation
├── parameter_group.go   # Conditional custom parameter group creation
└── replication_group.go # Main replication group resource and output exports
```

## Data Flow

1. **main.go** receives `AwsRedisElasticacheStackInput` containing the target resource and provider config
2. **locals.go** constructs AWS tags from metadata (organization, environment, resource kind, resource ID)
3. **subnet_group.go** conditionally creates an `elasticache.SubnetGroup` when `subnet_ids` are provided; name is sanitized from the metadata ID
4. **parameter_group.go** conditionally creates an `elasticache.ParameterGroup` when `parameters` are provided along with `parameter_group_family`
5. **replication_group.go** creates the `elasticache.ReplicationGroup` with:
   - Engine selection (Redis or Valkey) and version
   - Topology: non-clustered (`num_cache_clusters`) or clustered (`num_node_groups` + `replicas_per_node_group`)
   - High availability (automatic failover, multi-AZ)
   - Encryption (at-rest, in-transit with mode, KMS)
   - Authentication (AUTH token or user group IDs)
   - Maintenance windows and snapshot configuration
   - Log delivery to CloudWatch Logs or Kinesis Firehose
   - Event notifications via SNS
6. **outputs.go** defines constants for stack output keys matching `AwsRedisElasticacheStackOutputs`

## Key Patterns

- **StringValueOrRef**: Uses `.GetValue()` to extract literal string values. The platform resolves `valueFrom` references before passing to the IaC module.
- **Conditional resources**: Subnet group and parameter group are only created when their inputs are provided. The replication group references them via Pulumi output wiring.
- **Topology modes**: The spec uses mutually exclusive `num_cache_clusters` vs `num_node_groups` fields. The module branches on which is set, mapping to the corresponding Pulumi args.
- **Zero means default**: Numeric fields left at 0 are not set in the Pulumi args, letting AWS apply its defaults (e.g., port 6379, no snapshots).

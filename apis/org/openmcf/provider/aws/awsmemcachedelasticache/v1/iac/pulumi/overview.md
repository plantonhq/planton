# AwsMemcachedElasticache Pulumi Module Architecture

## Module Structure

```
module/
├── main.go              # Entry point: provider setup, orchestration
├── locals.go            # Locals struct: tags, spec references
├── outputs.go           # Output key constants matching AwsMemcachedElasticacheStackOutputs
├── subnet_group.go      # Conditional ElastiCache subnet group creation
├── parameter_group.go   # Conditional custom parameter group creation
└── cluster.go           # Main Memcached cluster resource and output exports
```

## Data Flow

1. **main.go** receives `AwsMemcachedElasticacheStackInput` containing the target resource and provider config
2. **locals.go** constructs AWS tags from metadata (organization, environment, resource kind, resource ID)
3. **subnet_group.go** conditionally creates an `elasticache.SubnetGroup` when `subnet_ids` are provided; name is sanitized from the metadata ID
4. **parameter_group.go** conditionally creates an `elasticache.ParameterGroup` when `parameters` are provided along with `parameter_group_family`
5. **cluster.go** creates the `elasticache.Cluster` with engine="memcached" and:
   - Engine version and node type
   - Node count (1–40) with AZ mode (single-az or cross-az)
   - Transit encryption (requires engine 1.6.12+)
   - Security groups for network-level access control
   - Maintenance window and auto minor version upgrade
   - SNS notifications for cluster events
   - Preferred AZ placement for nodes
6. **outputs.go** defines constants for stack output keys matching `AwsMemcachedElasticacheStackOutputs`

## Key Patterns

- **StringValueOrRef**: Uses `.GetValue()` to extract literal string values. The platform resolves `valueFrom` references before passing to the IaC module.
- **Conditional resources**: Subnet group and parameter group are only created when their inputs are provided. The cluster references them via Pulumi output wiring.
- **Engine hardcoded**: Unlike AwsRedisElasticache which supports redis/valkey, this module always creates a Memcached cluster (engine="memcached" is set by the module, not the user).
- **No authentication**: Memcached has no AUTH mechanism. Security relies on VPC security groups.
- **No encryption at rest**: Memcached does not support at-rest encryption. Only transit encryption is available.

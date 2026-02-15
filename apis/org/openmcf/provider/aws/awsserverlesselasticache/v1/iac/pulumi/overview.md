# AwsServerlessElasticache Pulumi Module Architecture

## Module Structure

```
module/
├── main.go              # Entry point: provider setup, orchestration
├── locals.go            # Locals struct: tags, spec references
├── outputs.go           # Output key constants matching AwsServerlessElasticacheStackOutputs
└── serverless_cache.go  # ElastiCache Serverless cache resource and output exports
```

## Data Flow

1. **main.go** receives `AwsServerlessElasticacheStackInput` containing the target resource and provider config
2. **locals.go** constructs AWS tags from metadata (organization, environment, resource kind, resource ID)
3. **serverless_cache.go** creates the `elasticache.ServerlessCache` with:
   - Engine selection (redis, valkey, or memcached)
   - Cache usage limits: data storage (min/max GB) and ECPU (min/max per second)
   - VPC networking: subnet IDs and security group IDs
   - KMS encryption: customer-managed key (optional)
   - Snapshots: daily time and retention (Redis/Valkey only)
   - Authentication: user group ID (Redis/Valkey only)
4. **outputs.go** defines constants for stack output keys matching `AwsServerlessElasticacheStackOutputs`

## Key Patterns

- **StringValueOrRef**: Uses `.GetValue()` to extract literal string values. The platform resolves `valueFrom` references before passing to the IaC module.
- **Flattened limits**: The spec exposes flat fields (`data_storage_min_gb`, `ecpu_max`, etc.) that the module reconstructs into the nested `CacheUsageLimits` block expected by AWS. The `data_storage.unit` is hardcoded to "GB".
- **No sub-resources**: Unlike provisioned siblings (AwsRedisElasticache, AwsMemcachedElasticache), there are no subnet groups or parameter groups. Serverless caches accept raw subnet IDs and have no parameter tuning.
- **Engine-agnostic**: The module passes the engine field through to AWS. Engine-specific field guards are handled by CEL validations in the spec, not by IaC conditional logic.

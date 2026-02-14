# AzureRedisCache Pulumi Module Architecture

## Module Structure

```
iac/pulumi/
├── main.go          # Pulumi entrypoint (loads stack input, calls module)
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Build/test/run targets
├── README.md        # Module documentation
├── overview.md      # This file
├── debug.sh         # Debug and build verification script
└── module/
    ├── main.go      # Resource creation (cache + firewall rules)
    ├── locals.go    # Local variables, tags, family derivation
    └── outputs.go   # Stack output constant names
```

## Resource Graph

```
AzureProvider
  └── redis.Cache (main cache instance)
        ├── redis.FirewallRule[0] (depends on cache)
        ├── redis.FirewallRule[1]
        └── redis.FirewallRule[N]
```

## Key Design Decisions

1. **Family auto-derived**: "C" for Basic/Standard, "P" for Premium (in locals.go)
2. **Single redis_configuration block**: Only maxmemory_policy exposed in v1
3. **Firewall rules use RedisCacheName**: Not ServerID (different from PostgreSQL pattern)
4. **Patch schedules are cache-level**: Embedded in CacheArgs, not separate resources

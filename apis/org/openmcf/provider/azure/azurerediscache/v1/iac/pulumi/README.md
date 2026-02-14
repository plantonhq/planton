# AzureRedisCache Pulumi Module

This Pulumi module provisions an Azure Cache for Redis instance with optional
firewall rules and patch schedules.

## Resources Created

- `redis.Cache` -- The Redis cache instance
- `redis.FirewallRule` -- IP-based firewall rules (one per rule in spec)

## Key Implementation Details

### SKU Family Auto-Derivation

The Azure API requires a `family` field ("C" for Basic/Standard, "P" for Premium).
This is fully deterministic from `sku_name`, so the module derives it automatically
in `locals.go`. Users never interact with the family field.

### VNet Injection

When `subnet_id` is set (Premium only), the cache is deployed inside the subnet
with private IP addressing. The cache is not reachable from the public internet
regardless of `public_network_access_enabled`.

### Redis Configuration

The `redis_configuration` block is used for the `maxmemory_policy` setting.
Additional redis_configuration fields (persistence, memory tuning) are deferred
to v2.

## Usage

```bash
make build  # Build the module
make test   # Run tests
make run    # Run the Pulumi program
```

## Dependencies

- `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/redis`
- `github.com/pulumi/pulumi/sdk/v3/go/pulumi`

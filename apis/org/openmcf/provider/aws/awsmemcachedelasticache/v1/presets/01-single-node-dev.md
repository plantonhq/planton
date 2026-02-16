# Memcached Single Node (Development)

This preset creates a single-node Memcached 1.6.22 cluster. It is the fastest way to get a development or testing cache running.

## When to Use

- Local development and testing environments
- Low-traffic applications that don't need multi-node distribution
- Prototyping cache-backed features before scaling up
- CI/CD pipeline caches

## Key Configuration Choices

- **Single node** (`numCacheNodes: 1`) — minimal cost, no distribution overhead
- **cache.t3.micro** — smallest burstable instance; sufficient for development workloads
- **Memcached 1.6.22** — latest stable version with TLS support

## Placeholders to Replace

Rename `metadata.name` to match your use case (e.g., `dev-session-cache`, `ci-cache`).

## Common Additions

- Add `subnetIds` and `securityGroupIds` for VPC-based deployments
- Increase `numCacheNodes` to 3 with `azMode: cross-az` for production resilience
- Add `parameterGroupFamily: memcached1.6` with `parameters` to tune engine behavior
- Enable `transitEncryptionEnabled: true` for TLS (requires engine 1.6.12+)

## Related Presets

- **02-multi-node-cross-az** — multi-node cluster distributed across AZs
- **03-production-encrypted** — full production setup with encryption and notifications

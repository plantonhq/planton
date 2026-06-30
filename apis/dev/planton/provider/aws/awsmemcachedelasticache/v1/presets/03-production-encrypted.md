# Memcached Production Encrypted

This preset creates a production-ready Memcached cluster with TLS encryption, cross-AZ distribution, custom parameters, and a defined maintenance window.

## When to Use

- Production environments with compliance requirements for encryption in transit
- Applications handling sensitive cached data
- Teams that need predictable maintenance windows and automatic version upgrades

## Key Configuration Choices

- **3 nodes, cross-az** — distributed caching with AZ-failure resilience
- **cache.r7g.large** — memory-optimized instance with Graviton3 for cost-effective production workloads
- **Transit encryption** — TLS enabled for all client connections (requires engine 1.6.12+)
- **Custom parameter group** — `chunk_size: 96` for workloads with larger-than-default cache entries
- **Maintenance window** — Sunday 05:00–06:00 UTC (low-traffic period)
- **Auto minor version upgrade** — keeps the cluster on supported engine versions

## Placeholders to Replace

- `metadata.name` — your cache name
- Add `subnetIds` and `securityGroupIds` for your VPC
- Add `notificationTopicArn` for cluster event alerts via SNS
- Adjust `preferredAvailabilityZones` for your region

## Common Additions

- Add `subnetIds` referencing AwsVpc private subnets
- Add `securityGroupIds` referencing AwsSecurityGroup
- Add `notificationTopicArn` referencing AwsSnsTopic for alerts
- Increase `numCacheNodes` for higher capacity
- Add more `parameters` for workload-specific tuning

## Related Presets

- **01-single-node-dev** — minimal single-node for development
- **02-multi-node-cross-az** — multi-node without encryption (simpler setup)

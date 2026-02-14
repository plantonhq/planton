# Production HA Redis Cluster

This preset creates a 3-node Scaleway Redis cluster with TLS encryption and Private Network connectivity. The cluster provides automatic failover -- if the primary fails, a replica is promoted within seconds. This is the standard production configuration for caching, session storage, and real-time data.

## When to Use

- Production caching layers that cannot tolerate downtime
- Session stores for web applications requiring high availability
- Real-time data pipelines, leaderboards, or pub/sub messaging
- Any workload requiring encrypted Redis connections

## Key Configuration Choices

- **Redis 7.2** (`version: 7.2.5`) -- latest stable version
- **RED1-M node** (`nodeType: RED1-M`) -- medium-tier Redis node; upgrade to `RED1-L` or `RED1-XL` for higher memory and throughput
- **3-node cluster** (`clusterSize: 3`) -- one primary and two replicas for automatic failover and read scaling
- **TLS enabled** (`tlsEnabled: true`) -- all client connections are encrypted in transit
- **Private Network** (`privateNetworkId`) -- Redis is reachable only via private IPs; note that `privateNetworkId` and `aclRules` are mutually exclusive on Scaleway
- **No ACL rules** -- access is controlled by Private Network membership instead of IP-based ACLs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-redis-user>` | Redis username (max 63 characters) | Choose a username |
| `<your-redis-password>` | Redis password (min 8 characters) | Generate a strong password |
| `<your-private-network-id>` | UUID of the Private Network for Redis connectivity | Scaleway console or `ScalewayPrivateNetwork` status outputs |

## Related Presets

- **01-dev-standalone** -- Use instead for development with a single node and public access

# Development Standalone Redis

This preset creates a single-node Scaleway Redis cluster using the smallest available node type with public ACL access. It is the fastest path to a working Redis instance for development, caching, and session storage testing.

## When to Use

- Development and testing environments needing a quick Redis instance
- Session storage and caching for development applications
- Learning Redis on Scaleway

## Key Configuration Choices

- **Redis 7.2** (`version: 7.2.5`) -- latest stable version with improved performance and new commands
- **RED1-MICRO node** (`nodeType: RED1-MICRO`) -- smallest and most affordable Redis node
- **Single node** (`clusterSize: 1`) -- no replication; acceptable for non-critical environments
- **Public ACL** (`aclRules`) -- allows connections from any IP; restrict to specific CIDRs for staging environments
- **No Private Network** -- uses public endpoint with ACL-based access control; note that `aclRules` and `privateNetworkId` are mutually exclusive on Scaleway

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-redis-user>` | Redis username (max 63 characters) | Choose a username |
| `<your-redis-password>` | Redis password (min 8 characters) | Generate a strong password |

## Related Presets

- **02-production-ha** -- Use instead for production with Private Network connectivity and cluster replication

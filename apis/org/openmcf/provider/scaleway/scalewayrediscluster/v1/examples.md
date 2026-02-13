# ScalewayRedisCluster Examples

Copy-paste examples for common deployment patterns. Adjust `zone`, `nodeType`, `version`, and credentials for your environment.

## Example 1: Development Standalone (Minimal)

A single-node Redis instance for development with open ACL.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: dev-cache
  env: development
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-MICRO
  clusterSize: 1
  userName: admin
  password: dev-cache-password-123
  aclRules:
    - ip: "0.0.0.0/0"
      description: "Allow all (dev only)"
```

## Example 2: Production HA with Private Network

A two-node HA cluster with TLS and Private Network connectivity.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: prod-cache
  org: mycompany
  env: production
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-M
  clusterSize: 2
  tlsEnabled: true
  userName: cache_admin
  password: very-strong-random-password-here
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  settings:
    maxclients: "10000"
    tcp-keepalive: "120"
    maxmemory-policy: allkeys-lru
```

## Example 3: Cluster Mode (Sharding)

A three-node cluster for high-throughput workloads with data sharding.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: high-throughput-cache
  org: dataplatform
  env: production
spec:
  zone: nl-ams-1
  version: "7.2.5"
  nodeType: RED1-L
  clusterSize: 3
  tlsEnabled: true
  userName: shard_admin
  password: cluster-password-very-strong
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  settings:
    maxclients: "50000"
    tcp-keepalive: "60"
    maxmemory-policy: volatile-lfu
    timeout: "300"
```

## Example 4: Restricted Public Access (ACL Only)

A publicly accessible cluster locked down to specific IP ranges.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: restricted-cache
spec:
  zone: pl-waw-1
  version: "6.2.7"
  nodeType: RED1-S
  clusterSize: 1
  tlsEnabled: true
  userName: admin
  password: restricted-admin-password
  aclRules:
    - ip: "198.51.100.0/24"
      description: "Corporate network"
    - ip: "203.0.113.5/32"
      description: "CI/CD runner"
    - ip: "10.0.0.0/8"
      description: "Internal VPN range"
```

## Example 5: Session Store with Custom Settings

A Redis cluster configured as a session store with appropriate eviction and timeout settings.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: session-store
  org: webapp
  env: staging
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-S
  clusterSize: 2
  userName: session_svc
  password: session-store-password-strong
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  settings:
    maxmemory-policy: volatile-ttl
    timeout: "1800"
    tcp-keepalive: "120"
```

## Example 6: Infra Chart Pattern (valueFrom Reference)

Using `valueFrom` to wire the Private Network from an upstream resource in an infra chart template.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: "{{ values.env }}-cache"
  org: "{{ values.org }}"
  env: "{{ values.env }}"
spec:
  zone: "{{ values.zone }}"
  version: "{{ values.redis_version | default('7.2.5') }}"
  nodeType: "{{ values.redis_node_type | default('RED1-S') }}"
  clusterSize: "{{ values.redis_cluster_size | default(2) }}"
  tlsEnabled: "{{ values.redis_tls_enabled | default(true) }}"
  userName: "{{ values.redis_user_name | default('cache_admin') }}"
  password: "{{ values.redis_password }}"
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: "{{ values.env }}-network"
      fieldPath: status.outputs.private_network_id
  settings:
    maxclients: "{{ values.redis_maxclients | default('10000') }}"
    tcp-keepalive: "120"
    maxmemory-policy: "{{ values.redis_eviction_policy | default('allkeys-lru') }}"
```

## Example 7: Bare Cluster (No ACL, No Private Network)

A cluster with default Scaleway networking (public, all IPs allowed). Only for development and non-sensitive data.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayRedisCluster
metadata:
  name: open-cache
spec:
  zone: fr-par-1
  version: "7.2.5"
  nodeType: RED1-MICRO
  userName: admin
  password: bare-cache-password-123
```

## Configuration Patterns Summary

| Pattern | HA | TLS | Network | Use Case |
|---------|----|----|---------|----------|
| Dev standalone | No | No | ACL (open) | Local development |
| Restricted public | No | Yes | ACL (restricted) | External access |
| Production HA | Yes | Yes | Private Network | Live workloads |
| Cluster mode | Sharded | Yes | Private Network | High-throughput |
| Session store | Yes | No | Private Network | Web app sessions |
| Bare (defaults) | No | No | Public (open) | Quick prototyping |

## Deployment Checklist

1. Choose Redis version (`7.2.5` recommended for new deployments)
2. Select node type based on workload (dev vs production)
3. Set cluster size (1 for dev, 2 for HA, 3+ for sharding)
4. Plan TLS decision upfront (cannot change without recreation)
5. Choose networking: Private Network (recommended) or ACL
6. Configure eviction policy matching your caching strategy
7. Set strong, randomly generated password
8. Tune `tcp-keepalive` and `timeout` for your connection patterns

## Next Steps

After deploying a `ScalewayRedisCluster`:

- **Connect applications** using the endpoint IPs and port from stack outputs
- **Verify TLS** using `status.outputs.certificate` as the CA cert (if TLS enabled)
- **Monitor** via Scaleway console or integrate with your monitoring stack
- **Scale up** by changing `nodeType` (online migration, no downtime)
- **Scale out** by increasing `clusterSize` in Cluster mode (online, no downtime)

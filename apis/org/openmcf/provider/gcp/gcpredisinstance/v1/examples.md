# GCP Redis Instance Examples

This document provides YAML examples for deploying Memorystore for Redis via OpenMCF. Each example includes a use-case description and the manifest.

---

## Example 1: Basic Cache (BASIC, 1GB, Minimal)

**When to use:** Development, testing, or lightweight caching. Single-node instance with no HA, no auth, minimal configuration.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpRedisInstance
metadata:
  name: dev-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpRedisInstance.dev-cache
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: dev-cache
  region: us-central1
  tier: BASIC
  memorySizeGb: 1
```

---

## Example 2: HA Production (STANDARD_HA, 5GB, Auth, TLS, Maintenance Window)

**When to use:** Production workloads requiring high availability, authentication, encrypted connections, and controlled maintenance.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpRedisInstance
metadata:
  name: prod-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpRedisInstance.prod-cache
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-cache
  region: us-central1
  tier: STANDARD_HA
  memorySizeGb: 5
  displayName: Production Redis cache
  authEnabled: true
  transitEncryptionMode: SERVER_AUTHENTICATION
  authorizedNetwork:
    value: projects/my-gcp-project-123/global/networks/prod-vpc
  maintenanceWindow:
    day: SUNDAY
    hour: 3
  deletionProtection: true
```

---

## Example 3: HA with Read Replicas (STANDARD_HA, 10GB, 3 Replicas)

**When to use:** Read-heavy workloads that benefit from scaling read throughput across multiple replicas.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpRedisInstance
metadata:
  name: prod-cache-read-scale
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpRedisInstance.prod-cache-read-scale
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-cache-read-scale
  region: us-central1
  tier: STANDARD_HA
  memorySizeGb: 10
  readReplicasMode: READ_REPLICAS_ENABLED
  replicaCount: 3
  authEnabled: true
  authorizedNetwork:
    value: projects/my-gcp-project-123/global/networks/prod-vpc
```

---

## Example 4: Persistence Enabled (STANDARD_HA with RDB Snapshots)

**When to use:** When you need durability across restarts or failovers. RDB snapshots are written periodically to disk.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpRedisInstance
metadata:
  name: prod-cache-persistent
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpRedisInstance.prod-cache-persistent
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-cache-persistent
  region: us-central1
  tier: STANDARD_HA
  memorySizeGb: 5
  persistenceConfig:
    persistenceMode: RDB
    rdbSnapshotPeriod: SIX_HOURS
  authEnabled: true
  authorizedNetwork:
    value: projects/my-gcp-project-123/global/networks/prod-vpc
```

---

## Example 5: CMEK Encrypted (Customer-Managed Key)

**When to use:** Compliance or governance requirements for customer-managed encryption at rest.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpRedisInstance
metadata:
  name: prod-cache-cmek
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpRedisInstance.prod-cache-cmek
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-cache-cmek
  region: us-central1
  tier: STANDARD_HA
  memorySizeGb: 5
  authEnabled: true
  transitEncryptionMode: SERVER_AUTHENTICATION
  customerManagedKey:
    value: projects/my-gcp-project-123/locations/us-central1/keyRings/redis-keys/cryptoKeys/redis-cmek
  authorizedNetwork:
    value: projects/my-gcp-project-123/global/networks/prod-vpc
  deletionProtection: true
```

---

## Example 6: Full Production (Everything Together)

**When to use:** Maximum production configuration: HA, read replicas, persistence, CMEK, auth, TLS, maintenance window, and deletion protection.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpRedisInstance
metadata:
  name: prod-cache-full
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpRedisInstance.prod-cache-full
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: prod-cache-full
  region: us-central1
  tier: STANDARD_HA
  memorySizeGb: 10
  redisVersion: REDIS_7_0
  displayName: Full production Redis instance
  authorizedNetwork:
    value: projects/my-gcp-project-123/global/networks/prod-vpc
  connectMode: DIRECT_PEERING
  authEnabled: true
  transitEncryptionMode: SERVER_AUTHENTICATION
  redisConfigs:
    maxmemory-policy: allkeys-lru
  maintenanceWindow:
    day: SUNDAY
    hour: 2
  readReplicasMode: READ_REPLICAS_ENABLED
  replicaCount: 2
  persistenceConfig:
    persistenceMode: RDB
    rdbSnapshotPeriod: SIX_HOURS
  customerManagedKey:
    value: projects/my-gcp-project-123/locations/us-central1/keyRings/prod-keys/cryptoKeys/redis-cmek
  deletionProtection: true
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).

# GCP Memorystore Instance Examples

This document provides YAML examples for deploying Memorystore instances via OpenMCF. Each example includes a use-case description and a copy-paste ready manifest.

---

## Example 1: Minimal Dev Instance

**When to use:** Development, testing, or experimentation. Smallest possible footprint with a standalone (non-clustered) instance.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: dev-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpMemorystoreInstance.dev-cache
spec:
  projectId:
    value: "<gcp-project-id>"
  instanceName: dev-cache
  location: us-central1
  shardCount: 1
  mode: CLUSTER_DISABLED
  nodeType: SHARED_CORE_NANO
  pscAutoConnections:
    - network:
        value: "projects/<gcp-project-id>/global/networks/<vpc-name>"
      projectId:
        value: "<gcp-project-id>"
```

---

## Example 2: Production HA Cluster

**When to use:** Production workloads requiring sharded data distribution, read replicas for failover, TLS encryption, and a controlled maintenance window.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: prod-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.prod-cache
spec:
  projectId:
    value: "<gcp-project-id>"
  instanceName: prod-cache
  location: us-central1
  shardCount: 3
  mode: CLUSTER
  nodeType: HIGHMEM_MEDIUM
  replicaCount: 1
  transitEncryptionMode: SERVER_AUTHENTICATION
  maintenancePolicy:
    weeklyMaintenanceWindow:
      day: SUNDAY
      hour: 3
  deletionProtectionEnabled: true
  pscAutoConnections:
    - network:
        value: "projects/<gcp-project-id>/global/networks/<vpc-name>"
      projectId:
        value: "<gcp-project-id>"
```

---

## Example 3: AOF Persistent Instance

**When to use:** Workloads that need strong durability guarantees. AOF logs every write operation and flushes once per second, minimizing data loss on unexpected restarts.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: persistent-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.persistent-cache
spec:
  projectId:
    value: "<gcp-project-id>"
  instanceName: persistent-cache
  location: us-central1
  shardCount: 1
  mode: CLUSTER_DISABLED
  nodeType: STANDARD_SMALL
  persistenceConfig:
    mode: AOF
    aofConfig:
      appendFsync: EVERY_SEC
  pscAutoConnections:
    - network:
        value: "projects/<gcp-project-id>/global/networks/<vpc-name>"
      projectId:
        value: "<gcp-project-id>"
```

---

## Example 4: CMEK Encrypted Instance

**When to use:** Compliance or governance requirements mandating customer-managed encryption keys for data at rest.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: cmek-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.cmek-cache
spec:
  projectId:
    value: "<gcp-project-id>"
  instanceName: cmek-cache
  location: us-central1
  shardCount: 1
  mode: CLUSTER_DISABLED
  nodeType: HIGHMEM_MEDIUM
  kmsKey:
    value: "projects/<gcp-project-id>/locations/us-central1/keyRings/<keyring-name>/cryptoKeys/<key-name>"
  transitEncryptionMode: SERVER_AUTHENTICATION
  authorizationMode: IAM_AUTH
  deletionProtectionEnabled: true
  pscAutoConnections:
    - network:
        value: "projects/<gcp-project-id>/global/networks/<vpc-name>"
      projectId:
        value: "<gcp-project-id>"
```

---

## Example 5: Multi-Zone Production with Automated Backups

**When to use:** Production instances that need multi-zone high availability and daily automated backups with 35-day retention.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: prod-ha-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.prod-ha-cache
spec:
  projectId:
    value: "<gcp-project-id>"
  instanceName: prod-ha-cache
  location: us-central1
  shardCount: 2
  mode: CLUSTER
  nodeType: HIGHMEM_MEDIUM
  replicaCount: 1
  zoneDistributionConfig:
    mode: MULTI_ZONE
  automatedBackupConfig:
    startHour: 2
    retention: "3024000s"
  maintenancePolicy:
    weeklyMaintenanceWindow:
      day: SATURDAY
      hour: 4
  deletionProtectionEnabled: true
  pscAutoConnections:
    - network:
        value: "projects/<gcp-project-id>/global/networks/<vpc-name>"
      projectId:
        value: "<gcp-project-id>"
```

---

## Example 6: Full Configuration (All Fields)

**When to use:** Reference manifest showing every available field. Use as a starting point and remove fields you do not need.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: full-config-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.full-config-cache
spec:
  projectId:
    value: "<gcp-project-id>"
  instanceName: full-config-cache
  location: us-central1
  shardCount: 3
  mode: CLUSTER
  nodeType: HIGHMEM_XLARGE
  engineVersion: VALKEY_8_0
  engineConfigs:
    maxmemory-policy: allkeys-lru
    notify-keyspace-events: ""
  replicaCount: 2
  authorizationMode: IAM_AUTH
  transitEncryptionMode: SERVER_AUTHENTICATION
  kmsKey:
    value: "projects/<gcp-project-id>/locations/us-central1/keyRings/<keyring-name>/cryptoKeys/<key-name>"
  persistenceConfig:
    mode: RDB
    rdbConfig:
      rdbSnapshotPeriod: SIX_HOURS
      rdbSnapshotStartTime: "2026-03-01T02:00:00Z"
  zoneDistributionConfig:
    mode: MULTI_ZONE
  maintenancePolicy:
    weeklyMaintenanceWindow:
      day: SUNDAY
      hour: 2
  automatedBackupConfig:
    startHour: 3
    retention: "2592000s"
  deletionProtectionEnabled: true
  pscAutoConnections:
    - network:
        value: "projects/<gcp-project-id>/global/networks/<vpc-name>"
      projectId:
        value: "<gcp-project-id>"
    - network:
        value: "projects/<second-project-id>/global/networks/<second-vpc-name>"
      projectId:
        value: "<second-project-id>"
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).

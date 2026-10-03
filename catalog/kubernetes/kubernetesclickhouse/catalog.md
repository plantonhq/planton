# ClickHouse

Deploys a ClickHouse cluster — the open-source columnar OLAP database — declared as a `ClickHouseInstallation` (CHI) custom resource and reconciled by the Altinity ClickHouse Operator. The operator renders every shard×replica host as its own single-pod StatefulSet with generated configuration mounted from ConfigMaps, and manages the full lifecycle: in-place version rolls, topology changes, storage edits, user provisioning, and the cluster-wide client Service. Topology is `shards × replicas`: shards split the data for parallel query processing (Distributed-engine tables fan queries out across shards); replicas within a shard hold full copies of the same data through ReplicatedMergeTree. Replication and `ON CLUSTER` DDL both need a coordination service — leave `coordination` unset and a managed 3-node ClickHouse Keeper deploys automatically exactly when the topology needs one.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ClickHouseInstallation custom resource** — one single-pod StatefulSet per shard×replica host, the cluster-wide client Service `clickhouse-<name>` (ClusterIP), per-cluster and per-host Services, and the generated ConfigMaps, all reconciled by the operator
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist
- **Auth Secret** — `<name>-clickhouse-auth`, one key per declared user; passwords never appear in the CHI (the operator injects them via `secretKeyRef`)
- **ClickHouseKeeperInstallation** — the managed coordination ensemble (`<name>-keeper`), created when `coordination` is unset and the topology needs coordination, or when `coordination.type: managed_keeper` is declared

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.
- **Altinity ClickHouse Operator** — an **Altinity ClickHouse Operator** resource must be running and **watching the target namespace**. Its default watch scope is its OWN namespace only — widen its `watchNamespaces` or deploy the cluster beside it. Deploy the operator first.

### Kubernetes Cluster

- **A StorageClass** for the data volumes. Kubernetes cannot shrink PVCs, and expanding requires a class that allows it — size `diskSize` for growth. The managed Keeper's volumes ride the same mechanics.

## Deploy

### Console

Open the deployment store, find **ClickHouse**, and click **Deploy**. The creation wizard walks you through namespace placement, the pinned server version and cluster name, shards × replicas, the coordination quorum, storage, settings profiles and quotas, users, engine configuration, host resources, placement, and the service face with the pause switch. Start from the **Dev minimal preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClickHouse
metadata:
  name: dev-clickhouse
  org: acme-corp
  env: dev
spec:
  namespace:
    value: dev-clickhouse
  createNamespace: true
  version: "25.3"
  diskSize: 20Gi
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: "2"
      memory: 4Gi
  users:
    - name: dev
      password:
        value: change-me-dev-password
      grants:
        - GRANT SELECT, INSERT, CREATE, DROP ON *.*
```

```shell
planton apply -f clickhouse.yaml
```

This creates the smallest declarable ClickHouse that actually serves: one host, a PVC, a pinned server version, and one named user. No Keeper deploys because a 1×1 topology needs no coordination — which also means no replication: a lost volume loses the data. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose the cluster behind its namespace with a reference:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: analytics-namespace
      fieldPath: spec.name
  createNamespace: false
  version: "25.3"
  diskSize: 100Gi
```

The InfraPipeline creates the namespace first, then deploys the cluster into it.

## Key Configuration

These are the most important decisions when configuring a ClickHouse cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Always pin `version`** — the operator's built-in fallback is the `latest` tag, which makes cluster upgrades happen implicitly on pod restarts. Pin an LTS line like `"25.3"`; version bumps are day-2 edits the operator rolls in place.

**Replicas buy durability, shards buy capacity** — the 1×1 default means no redundancy: a lost volume loses that shard's data. Production runs 2–3 replicas per shard with `spreadReplicasAcrossNodes: true` (co-located replicas make replication pointless against node loss). Scale `shards` only when one shard can no longer carry the dataset or the write rate — rebalancing existing data after a shard change is a manual migration.

**Coordination absence is the recommendation** — leave `coordination` unset and a managed 3-node ClickHouse Keeper (Raft-based, no JVM, upstream-recommended over ZooKeeper) deploys exactly when the topology needs one. Point at an existing Keeper or ZooKeeper ensemble for shared coordination infrastructure; `type: none` is legal only for single-replica topologies.

**Storage grows but never shrinks** — `diskSize` is per host, mounted at `/var/lib/clickhouse`. `retainVolumesOnDelete: true` keeps every PVC through deletion (a re-created cluster with the same name re-attaches the data) — and retained PVCs are never garbage-collected; deleting them becomes a manual act.

**Users are Secret-native, and networks are a trap** — every declared password lands in the auth Secret, never the CHI. Empty `networks` is NOT any-network: the operator fences the user to the cluster's own pods, and ClickHouse reports the rejection as a wrong-password error (port-forwarded smoke tests pass while in-cluster clients fail — verified live). Declare networks explicitly for every user a workload connects as. Grants constrain, never widen: a user with NO grants gets ClickHouse's unrestricted config-file default, and once grants exist, `ON CLUSTER` DDL additionally needs `GRANT CLUSTER ON *.*`. Rotating the Secret alone does not re-render config — bump any spec field to roll a rotation out.

**Verify the cluster before distributed DDL** — the in-server cluster definition can lag the installation reaching Completed; `ON CLUSTER` DDL in that window silently executes on a subset of hosts and still returns success. Confirm `system.clusters` reports every host after a deploy or topology change. Keep `metadata.name` within 48 characters (with the default `clusterName`); a longer `clusterName` shrinks that budget one-for-one.

**Pause without losing data** — `stopped: true` scales every host StatefulSet to zero and keeps every PVC; flipping it back brings the same data up. The declarative off switch for expensive dev/staging clusters.

## Outputs and Dependencies

### What This Kind Consumes

| Field | References | Purpose |
|-------|-----------|---------|
| `spec.namespace` | KubernetesNamespace (`spec.name`) | Where the cluster runs |
| `spec.storage_class` | KubernetesStorageClass (`status.outputs.storage_class_name`) | Data volume class |
| `spec.coordination.keeper.storage_class` | KubernetesStorageClass (`status.outputs.storage_class_name`) | Managed Keeper volume class |
| `spec.users[].password` | Organization secrets or other resources' outputs | User credentials (never plaintext in the CHI) |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Application deployment manifests |
| `chi_name` | Name of the ClickHouseInstallation (= metadata.name) | Operational tooling |
| `cluster_name` | Logical cluster name — the `ON CLUSTER` / `remote_servers` target | Distributed DDL, Distributed-table definitions |
| `service_name` | The cluster-wide client Service (`clickhouse-<name>`) | Ingress/Gateway composition |
| `tcp_endpoint` | In-cluster native-protocol endpoint (port 9000) | clickhouse-client, native drivers |
| `http_endpoint` | In-cluster HTTP endpoint (port 8123) | curl, JDBC/ODBC over HTTP, SigNoz and other consumers |
| `auth_secret_name` | The users' password Secret (one key per user); empty when no users are declared | Application pod env via `secretKeyRef` |
| `keeper_name` | The managed ClickHouseKeeperInstallation; empty when coordination is external or none | Operational tooling |
| `keeper_service_name` | The managed Keeper's client Service; empty when coordination is external or none | Diagnostics |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the HTTP interface | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev minimal** — one host, a PVC, a pinned version, one named user; no Keeper, no replication. The real ClickHouse SQL surface for developers and CI without production ceremony. Start from the **Dev minimal preset**.

**Production replicated** — one shard carried by three replicas in ReplicatedMergeTree lockstep, a three-node managed Keeper, replicas forced onto different nodes, retained volumes, and a least-privilege user split (a readonly analyst under a profile and quota; a narrow-granted ingest user). Start from the **Production replicated preset**.

**Sharded analytics** — four shards × two replicas (eight hosts a Distributed table queries in parallel), a named cluster for `ON CLUSTER` DDL, and an ETL user carrying the `GRANT CLUSTER` teaching. Start from the **Sharded analytics preset**.

## Works With

- **Kubernetes Altinity Operator** — the engine that reconciles this cluster; deploy it first and make sure it is watching the target namespace.
- **Kubernetes Namespace** — referenced placement; the InfraPipeline orders namespace-first.
- **Kubernetes Storage Class** — SSD-backed classes for data and Keeper volumes.
- **Kubernetes SigNoz** — composes this kind as its external ClickHouse storage over the exported endpoints and auth Secret.
- **Kubernetes Ingress / Gateway API kinds** — external exposure over the exported service handle (all generated Services stay ClusterIP by design).

# Kubernetes ClickHouse

## When NOT to Use This

**The operator must already be on the cluster — and watching this
namespace.** This component declares a ClickHouse cluster;
KubernetesAltinityOperator installs the ENGINE that reconciles it. The
Altinity chart's default posture watches ONLY the operator's own
namespace — a cluster declared anywhere else is silently ignored: no
error, no pods, nothing. Deploy the operator first with
`watch_namespaces` covering this namespace (or `[".*"]`), clusters
after.

Also not the right kind when:

- **You want the operator itself** — installing and configuring the
  Altinity ClickHouse operator is KubernetesAltinityOperator; this
  component is one cluster it manages.
- **You want a managed cloud service** — ClickHouse Cloud and the
  cloud providers' managed offerings run the database for you; this
  component is for running ClickHouse ON the Kubernetes cluster
  itself.
- **You need an OLTP database** — ClickHouse is a columnar OLAP
  engine: brilliant at scanning billions of rows, wrong for
  row-by-row transactional workloads. Use the PostgreSQL or MySQL
  kinds for those.
- **You want HTTP exposure baked in** — every generated Service is
  ClusterIP (the operator's own default). External reachability
  composes from first-class kinds (KubernetesIngress, Gateway API
  kinds) over the exported `service_name`, with
  `service_annotations` for LB/mesh recipes — never embedded here.

## Overview

**KubernetesClickHouse** declares a ClickHouse cluster — the columnar
OLAP database built for analytical queries over billions of rows — as
a `ClickHouseInstallation` (CHI) custom resource reconciled by the
Altinity ClickHouse operator. The operator renders every host as its
own single-pod StatefulSet with generated ClickHouse configuration
mounted from ConfigMaps, and manages rolling restarts, PDBs, and
Services around them.

**Topology**: `shards` × `replicas` hosts. Shards split the data for
parallel query processing (Distributed-engine tables fan queries out
across all shards); replicas within a shard hold copies of the same
data, kept in sync through ReplicatedMergeTree. Replication REQUIRES
a coordination service — the spec enforces it: `replicas > 1` with
coordination type `none` is rejected at validation, never discovered
at runtime.

**The coordination design** (the decision most deployments never have
to make): leave `coordination` unset and the module deploys a managed
ClickHouse Keeper — a `ClickHouseKeeperInstallation` reconciled by
the same operator — automatically whenever the topology needs one
(`replicas > 1` or `shards > 1`), and none otherwise. Set it
explicitly to size the managed Keeper (quorum of 1, 3, or 5), point
at an existing Keeper or ZooKeeper ensemble, or opt out with `none`
(single-replica only; a multi-shard cluster without coordination
loses `ON CLUSTER` DDL but Distributed queries keep working). The CHI
references the managed Keeper through the CRD's native keeper
reference — the operator resolves the endpoints itself.

**The users truth**: passwords never land in the CHI. The module
writes them into a Kubernetes Secret (`<name>-clickhouse-auth`, one
key per user) and the operator injects them via
`valueFrom.secretKeyRef`. KNOW THIS (upstream-documented):
secret-sourced passwords reach ClickHouse through pod environment
variables, so rotating the Secret alone does not re-render config — a
spec change triggers the re-reconcile that rolls the rotation out.
The built-in `default` user stays operator-managed: passwordless but
network-restricted to the cluster's own pods — create named users for
every real client.

**Key design points:**

- **Server configuration is layered, not escape-hatched** — typed
  fields cover topology, storage, users, and placement; ClickHouse's
  own configuration vocabulary rides the CHI's path-keyed maps
  (`settings`, `files`, per-user `settings`, `profiles`, `quotas` —
  `/`-separated XML paths exactly as the upstream CRD defines them).
  Those maps ARE the upstream's native model at this layer.
- **Storage is per host** — a data PVC (`disk_size`, optional
  `storage_class` reference) mounted at /var/lib/clickhouse, plus an
  optional separate log volume (`log_disk_size`).
  `retain_volumes_on_delete` maps to the operator's PVC reclaim
  policy: the upstream default DELETES volumes with the resource;
  retain makes a re-created same-name cluster re-attach its data.
- **Operational verbs are spec fields** — `stopped` (the CHI stop
  verb: hosts scale to zero, PVCs stay), `pdb_max_unavailable`
  (operator-managed PDB), `spread_replicas_across_nodes` (the
  operator's ShardAntiAffinity: replicas of the same shard never
  co-locate on a node), `auto_inter_node_secret` (operator-generated
  shared secret securing distributed queries, rendered only for
  multi-host topologies).
- **Always pin `version`** — the operator's own fallback is the
  `latest` tag, which turns pod restarts into implicit upgrades; the
  spec makes the pin required (e.g. "25.3", an LTS line). `image`
  overrides the repository for mirrors.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the cluster — literal or a
  KubernetesNamespace reference; the operator must watch it
- **`spec.version`**: the ClickHouse server version to run (a
  `clickhouse/clickhouse-server` tag, e.g. "25.3") — always pinned,
  never `latest`
- **`spec.disk_size`**: the data volume for EACH host (e.g. "100Gi")
  — PVCs cannot shrink, plan for growth

### Common

- **`spec.shards` / `spec.replicas`**: the topology (defaults 1×1) —
  replicas buy durability, shards buy capacity; production runs 2–3
  replicas
- **`spec.coordination`**: unset = managed Keeper appears exactly
  when the topology needs it; `managed_keeper` sizes it (1/3/5
  quorum, resources, disk); `external_keeper` / `external_zookeeper`
  point at existing ensembles (`external.nodes`, optional `root` and
  `identity`); `none` opts out (single-replica only)
- **`spec.users`**: named users with Secret-delivered passwords
  (literal or a reference to another resource's output), optional
  `profile`, `quota`, `networks`, SQL `grants`, `access_management`,
  and per-user `settings`. A user without `grants` gets ClickHouse's
  unrestricted default access; `ON CLUSTER` DDL additionally requires
  `GRANT CLUSTER ON *.*` (both verified live)
- **`spec.profiles` / `spec.quotas`**: named settings bundles users
  reference by name (path-keyed, e.g. `readonly: "1"` or
  `interval/duration: "3600"`)
- **`spec.settings` / `spec.files`**: server-level settings
  (`/`-separated XML paths, e.g. `max_concurrent_queries: "200"`) and
  raw config-file drop-ins (name → content; `{common}` / `{users}` /
  `{hosts}` prefixes choose the target directory)
- **`spec.cluster_name`**: the `remote_servers` entry that
  `ON CLUSTER` DDL targets (default "main") — capped at 15 characters
  by the upstream CRD because it becomes a segment of every generated
  child name; keep `metadata.name` within 48 characters with the
  default
- **`spec.resources`**: CPU/memory per host — ClickHouse is
  memory-hungry under analytical load; give production hosts at
  least 4Gi
- **`spec.storage_class` / `spec.log_disk_size`**: the data volumes'
  StorageClass (literal or KubernetesStorageClass reference) and an
  optional separate log volume
- **`spec.retain_volumes_on_delete`**: keep PVCs when the resource is
  deleted — the production data-safety switch; retained PVCs are
  never garbage-collected
- **`spec.spread_replicas_across_nodes`**: replicas of the same shard
  never share a Kubernetes node — off by default so single-node dev
  clusters schedule; turn on in production
- **`spec.stopped`**: scale every host to zero keeping all data — the
  declarative pause switch for expensive dev/staging clusters
- **`spec.pdb_max_unavailable` / `spec.node_selector` /
  `spec.tolerations`**: disruption budget and pod placement
- **`spec.service_annotations`**: annotations for the cluster-wide
  client Service (internal LB / service-mesh recipes) — the Service
  stays ClusterIP
- **`spec.image` / `spec.image_pull_secrets`**: the air-gap /
  private-mirror path
- **`spec.auto_inter_node_secret`**: operator-generated shared secret
  for distributed queries (default true; disable only below
  ClickHouse 20.10)

## Environment Injection

This component calls no cloud APIs; managed-Kubernetes integration
rides the Service annotations and ClickHouse's own storage
configuration.

| Cloud / posture | Where | Mechanism |
|---|---|---|
| Internal / external LB recipes | `service_annotations` | Cloud-controller annotations on the cluster-wide client Service |
| S3 disks / tiered storage, declared keys | `settings` / `files` | ClickHouse-native storage configuration (XML drop-ins) carrying the access keys |
| S3, keyless (EKS) | `settings` / `files` only | `use_environment_credentials` with IRSA-bound identity on the nodes |
| GCS, keyless (GKE) | `settings` / `files` only | Workload Identity on the nodes |

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `chi_name` | Name of the ClickHouseInstallation resource (= `metadata.name`) |
| `cluster_name` | Logical cluster name — the `ON CLUSTER` / `remote_servers` target |
| `service_name` | The cluster-wide client Service covering all hosts (`clickhouse-<name>`, ClusterIP) |
| `tcp_endpoint` | In-cluster native-protocol endpoint, port 9000 (clickhouse-client, drivers) |
| `http_endpoint` | In-cluster HTTP interface endpoint, port 8123 (curl, JDBC/ODBC over HTTP) |
| `auth_secret_name` | The module-managed `<name>-clickhouse-auth` Secret (one key per user); empty when no users are declared |
| `keeper_name` | The managed ClickHouseKeeperInstallation (`<name>-keeper`); empty when coordination is external or none |
| `keeper_service_name` | The managed Keeper's client Service (`keeper-<name>-keeper`); empty when coordination is external or none |
| `port_forward_command` | Port-forward command for workstation access when no exposure is composed |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`); **`storage_class`**
  (data and Keeper volumes) references a KubernetesStorageClass; a
  user's **`password`** accepts a reference to another resource's
  output — generated credentials flow in without ever being written
  down.
- **Applications consume the outputs**: `tcp_endpoint` for
  clickhouse-client and native drivers, `http_endpoint` for
  HTTP-speaking clients, `auth_secret_name` as env-from references —
  credentials ride the Secret, never the manifest.
- **Exposure composes, never embeds**: a KubernetesIngress or Gateway
  API route targets `service_name`; `service_annotations` carries the
  internal-LB recipes.
- **The operator is a cluster prerequisite**, not a reference: deploy
  KubernetesAltinityOperator first, watching this namespace.

## Examples

### Development (single host, no Keeper)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClickHouse
metadata:
  name: dev-clickhouse
spec:
  namespace:
    value: dev-clickhouse
  create_namespace: true
  version: "25.3"
  disk_size: 20Gi
  resources:
    requests: { cpu: 500m, memory: 1Gi }
    limits: { cpu: "2", memory: 4Gi }
  users:
    - name: dev
      password:
        value: change-me-dev-password
      grants:
        - GRANT SELECT, INSERT, CREATE, DROP ON *.*
```

One shard, one replica: no Keeper is deployed and there is no
replication — a lost volume loses the data. That trade-off is the
point of a dev cluster.

### Production (replicated, managed Keeper, retained volumes)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClickHouse
metadata:
  name: prod-clickhouse
spec:
  namespace:
    value: clickhouse
  create_namespace: true
  version: "25.3"
  replicas: 3
  disk_size: 200Gi
  resources:
    requests: { cpu: "2", memory: 8Gi }
    limits: { cpu: "4", memory: 16Gi }
  coordination:
    type: managed_keeper
    keeper:
      replicas: 3
      disk_size: 10Gi
  spread_replicas_across_nodes: true
  retain_volumes_on_delete: true
  pdb_max_unavailable: 1
  users:
    - name: analyst
      password:
        value: change-me-analyst-password
      profile: readonly
      quota: default
    - name: ingest
      password:
        value: change-me-ingest-password
      grants:
        - GRANT SELECT, INSERT ON analytics.*
  profiles:
    - name: readonly
      settings:
        readonly: "2"
        max_memory_usage: "8000000000"
  quotas:
    - name: default
      settings:
        interval/duration: "3600"
        interval/queries: "50000"
  settings:
    max_concurrent_queries: "200"
```

Three full copies of the data in ReplicatedMergeTree lockstep,
coordinated by a three-node Keeper, replicas forced onto different
nodes, volumes that outlive the resource. Scale `shards` only when a
single shard can no longer carry the dataset or the write rate —
replicas buy durability, shards buy capacity.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

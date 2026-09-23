# Neo4j

Deploys a Neo4j graph database — the standard engine for knowledge graphs, GraphRAG and agent-memory architectures — on any Kubernetes cluster from the official `neo4j` Helm chart (helm.neo4j.com/neo4j). One release is one Neo4j server: a StatefulSet of exactly one pod, the chart's own grain. Community edition (the default) is single-instance by license; Enterprise clustering is built by installing multiple KubernetesNeo4j resources that share a `clusterName` — each member a first-class resource, never a replicas knob.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Kubernetes Namespace** — created (with the standard Planton governance labels) only when `createNamespace` is `true`; otherwise the namespace must already exist
- **Neo4j Helm Release** — the official chart at the pinned `chartVersion`, which creates:
  - a StatefulSet with a single Neo4j pod, your CPU/memory resources, and the memory split rendered into `neo4j.conf`
  - the always-created **default Service** (named after the resource) carrying bolt 7687, http 7474 and https 7473 — what in-cluster clients use, and what the exported endpoints point at
  - the **exposure Service** `<name>-lb-neo4j`, ClusterIP by this component's deliberate default (the chart's own default is LoadBalancer)
  - a PersistentVolumeClaim for the data volume (10Gi on the cluster's default StorageClass unless configured)
- **Auth Secret** — the module materializes the admin password (declared, or generated when `auth` is empty) as the `<name>-auth` Secret (key `NEO4J_AUTH`, the chart's contract, plus a bare `password` key); the password never lands in rendered Helm values
- **ServiceMonitor** — only when `serviceMonitorEnabled` is `true` (requires the Prometheus Operator CRDs on the cluster)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A StorageClass** capable of dynamic PV provisioning for the data volume. Graph workloads reward fast (SSD-backed) storage — page-cache misses become disk reads.
- **Prometheus Operator CRDs** — only when enabling the ServiceMonitor; without them the install fails on the unknown resource type.
- **An existing `NEO4J_AUTH` Secret** — only when using `auth.existingSecret`; the chart reads it at template time, so it must exist before the first install.

## Deploy

### Console

Open the deployment store, find **Neo4j**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev single instance preset** in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNeo4j
metadata:
  name: knowledge-graph
  org: acme-corp
  env: prod
spec:
  namespace:
    value: neo4j
  createNamespace: true
  auth:
    existingSecret: neo4j-auth
  resources:
    requests:
      cpu: "2"
      memory: 8Gi
    limits:
      cpu: "4"
      memory: 8Gi
  memory:
    heapInitial: 3G
    heapMax: 3G
    pageCache: 3G
  dataVolume:
    size: 100Gi
```

```shell
planton apply -f neo4j.yaml
```

This creates one production-shaped Neo4j server: credentials referenced from a pre-existing Secret, an explicit heap/page-cache split sized to the 8Gi container, and a 100Gi data volume. Create the `neo4j-auth` Secret (key `NEO4J_AUTH`, value `neo4j/<password>`) before applying. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the server to a namespace managed by another Cloud Resource:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: neo4j-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the Neo4j server into it.

## Key Configuration

These are the most important decisions when configuring Neo4j. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Edition and clustering** — `edition` defaults to `community` (GPLv3, single-instance **by license** — no replicas, no failover; availability is the StatefulSet rescheduling the pod and the PVC surviving it). `enterprise` requires `acceptLicenseAgreement: true` and a valid commercial license, and unlocks clustering: multiple KubernetesNeo4j resources sharing a `clusterName` form one cluster. `clusterName` on community fails validation.

**Admin credentials are minted, never required** — leave `auth` empty and the module generates the password and materializes it as the `<name>-auth` Secret exactly as it would a declared one; `auth_secret_name` and `password_secret` tell workloads where to read it. The `auth` block takes at most one arm to bring your own: `password` (a managed-secret reference, materialized the same way) or `existingSecret` (a Secret already carrying `NEO4J_AUTH: neo4j/<password>`, existing before the install).

**Sizing** — the chart's own minimum is **500m CPU / 2Gi memory, and it rejects installs below that floor**. The `memory` block renders an explicit heap/page-cache split into `neo4j.conf`; empty lets Neo4j auto-compute from the container memory (usually right — set it explicitly on shared or memory-tight nodes). Keep initial heap = max heap; give the page cache what remains after heap and OS overhead.

**Data volume** — `dataVolume.size` defaults to 10Gi on the cluster's default StorageClass; `dataVolume.storageClass` accepts a literal class name or a KubernetesStorageClass reference. Growing a PVC later depends on the class allowing expansion.

**Exposure** — `service.type` defaults to ClusterIP, a deliberate override of the chart's LoadBalancer default: exposure composes from first-class kinds (KubernetesIngress, Gateway API) over the exported service handle. For a direct cloud load balancer, set `type: LoadBalancer` and ride the provider recipe on `service.annotations`. This block shapes only the exposure Service `<name>-lb-neo4j`; in-cluster clients use the default Service — the endpoints in the stack outputs.

**TLS** — `ssl.bolt` / `ssl.https` each reference an existing certificate Secret (a KubernetesCertificate reference resolves to its Secret name). Note the key-name bridge: the chart expects `private.key`/`public.crt` while cert-manager Secrets carry `tls.key`/`tls.crt` — see the component docs for the bridge.

**Escape hatch** — `helmValues` merges LAST over everything the typed fields render (Helm `-f` semantics, identical on both engines). It is for the chart surface beyond the typed fields — log4j XML, additional volumes/mounts, LDAP secrets, the operations sidecar, probes, PDB, per-service splits, the `NEO4J_PLUGINS` env key that activates APOC — never a substitute for them, and never a place for secrets.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `dataVolume.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesCertificate** | `ssl.bolt.secret`, `ssl.https.secret` | `status.outputs.secret_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the server runs in | Application deployment manifests |
| `release_name` | Helm release name (= metadata.name) | Operational tooling |
| `service_name` | Name of the main Neo4j Service (bolt/http/https ports) | Ingress/Gateway backends, NetworkPolicies |
| `bolt_endpoint` | In-cluster bolt endpoint drivers connect to (e.g. `neo4j://main.graph.svc.cluster.local:7687`) | Application driver connection strings |
| `http_endpoint` | In-cluster HTTP API endpoint (e.g. `http://main.graph.svc.cluster.local:7474`) | Neo4j Browser access and REST calls |
| `auth_secret_name` | Secret holding the admin credentials in the chart's contract (the module-materialized `<name>-auth`, declared or generated, or the referenced existing Secret); always set | Workload env wiring — read the password from the Secret, never an env literal |
| `password_secret` | Secret name + key of the admin user's bare password (the module-materialized `<name>-auth`, key `password`); unset when `existingSecret` is declared | Workloads that take a password rather than the `neo4j/<password>` pair |
| `port_forward_command` | Command to port-forward bolt to a developer laptop | Local development |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single instance** — the smallest declarable Neo4j: community edition, a declared password, resources at the chart's own floor (500m / 2Gi — there is no smaller Neo4j), a 10Gi volume. Start from the **Dev single instance preset**.

**Production single server** — an existing-Secret credential, an explicit memory split sized to the container, 100Gi on fast storage, transaction-timeout and telemetry settings in `config`, node-selector scheduling. Start from the **Production preset**.

**GraphRAG with APOC** — the APOC procedure library activated at startup (it ships inside the official image), file/trigger arms enabled in `apocConfig`, the procedure sandbox opened exactly to `apoc.*`, a page-cache-weighted memory split, and heap dumps on OOM. Start from the **GraphRAG APOC preset**.

## Works With

- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) — provides the namespace for the server
- [**Kubernetes StorageClass**](/cloud-catalog/kubernetes-storage-class) — fast storage for the data volume
- [**Cert Manager Certificate**](/cloud-catalog/kubernetes-certificate) — cert-manager-issued TLS Secrets for the bolt/https listeners (mind the key-name bridge)
- [**Kubernetes Ingress**](/cloud-catalog/kubernetes-ingress) — composes exposure over the exported service handle instead of a direct LoadBalancer

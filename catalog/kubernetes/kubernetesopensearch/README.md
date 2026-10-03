# Kubernetes OpenSearch

## When NOT to Use This

**The operator must already be on the cluster.** This component
declares a search cluster; KubernetesOpenSearchOperator installs the
ENGINE that reconciles it. The default operator posture watches ALL
namespaces — but an operator fenced with `watch_namespace` silently
ignores clusters anywhere else. Deploy the operator first, clusters
after.

Also not the right kind when:

- **You want the operator itself** — installing and configuring the
  OpenSearch Kubernetes Operator is KubernetesOpenSearchOperator; this
  component is one cluster it manages.
- **You want a managed cloud search service** — use the host cloud
  provider's managed search kinds; this kind is for running
  OpenSearch ON the Kubernetes cluster itself.
- **You expect production credentials out of the box** — without a
  custom `security.config`, the bootstrapped admin credentials are the
  image's well-known DEMO credentials (see the security truth below).
  Fine inside a private cluster for development; bring a custom
  security config or rotate immediately for anything real.
- **You want HTTP exposure baked in** — the cluster and Dashboards
  Services are ClusterIP by design. External reachability composes
  from first-class kinds (KubernetesIngress, Gateway API kinds) over
  the exported service handles, or via the Dashboards service type for
  a quick LoadBalancer — never embedded here.

## Overview

**KubernetesOpenSearch** declares an OpenSearch cluster — the
Apache-2.0 search and analytics engine, a drop-in replacement for the
Elasticsearch APIs at the 7.10 fork line, with its own 2.x/3.x feature
line since — as an `OpenSearchCluster` custom resource reconciled by
the OpenSearch Kubernetes Operator. The licensing is the point:
Apache-2.0 end to end, so existing Elasticsearch clients and tooling
speak to it unchanged while the engine itself stays open.

The operator manages the full cluster lifecycle: node StatefulSets per
pool, cluster bootstrap, TLS (generated or provided), the security
plugin's admin bootstrap, safe rolling upgrades with drain ordering,
and an optional OpenSearch Dashboards deployment — the console is a
SECTION of the same custom resource, not a separate component.

**Topology**: `node_pools` is the cluster's shape — every pool declares
its roles (`cluster_manager`, `data`, `ingest`, `ml`,
`remote_cluster_client`, `search`; underscore forms — a dashed
"cluster-manager" is NOT a role and fails only at node startup, never
at apply). The smallest real cluster is one pool with all roles and 3
replicas; a 1-replica all-roles pool works for development. Storage is
per-pool: PVC-backed by default (survives pod loss), emptyDir for
throwaway data.

**The security truth** (from the operator source, not aspiration):

- The HTTP API serves **https in EVERY posture** — even without a
  `security` block, the image's demo security configuration is active
  and the operator itself always talks https to the cluster.
- The operator does **not** generate a random admin password at this
  release: the bootstrapped credentials in the `<name>-admin-password`
  Secret are the image's **well-known demo admin credentials**.
- Production guidance: bring a custom `security.config` (your own
  internal_users.yml and admin credentials — all three secrets are
  typically required), or rotate the admin password through the
  security API immediately after install. Clients read credentials
  from the Secret named in the outputs; no credential ever
  appears in this spec unless you bring your own security config.

**Key design points:**

- **Declare `security` with generated TLS** (the recommended default):
  the operator issues a CA and per-layer certificates — per-node
  transport certificates by component default (the stronger posture;
  the operator's own default is a shared certificate). Provided
  certificates ride the cert-manager seam (`secret` referencing a
  KubernetesCertificate) with `nodes_dn`/`admin_dn` required.
- **Dashboards fold in** — `dashboards.enabled` deploys the console
  alongside the cluster; its version defaults to the cluster's (they
  must match), TLS and a LoadBalancer service type are knobs, and its
  Service stays ClusterIP by default.
- **Snapshot repositories + keystore are the backup surface** —
  repositories register on the cluster (`type`/`settings` pass through
  to the snapshot API); credentials belong in the KEYSTORE (Secrets
  loaded into every node before startup), never in settings. Keyless
  paths — instance or workload identity on the nodes — skip the
  keystore entirely.
- **Plugins install at startup** — `plugins_list` downloads from the
  internet at pod start unless the plugin ships in the image; a
  failing install crash-loops the pod. Air-gapped clusters bake
  plugins into a custom image.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the cluster — literal or a
  KubernetesNamespace reference; the operator must watch it
- **`spec.version`**: the OpenSearch version to run (a published
  `opensearchproject/opensearch` image tag, e.g. "2.19.4") — check the
  operator's compatibility table before pinning a major line
- **`spec.node_pools`**: at least one pool; every cluster needs at
  least one pool carrying `cluster_manager` and one carrying `data`
  (one pool can carry both); each pool requires `name`, `replicas`,
  and `roles`

### Common

- **`spec.security`**: TLS posture and security-plugin bootstrap —
  empty = operator-generated CA and certificates with the demo admin
  bootstrap; `config` is the bring-your-own arm (custom realms, OIDC,
  seeded roles)
- **`spec.dashboards`**: the web console — `enabled`, `replicas`,
  `tls`, `base_path` (behind a path-rewriting proxy), `service` (type
  and LoadBalancer source ranges)
- **`spec.node_pools[].jvm` / `resources` / `disk_size`**: size heap
  to roughly half the container memory; disk defaults to the
  operator's 30Gi
- **`spec.node_pools[].persistence`**: PVC on the default (or a
  referenced) StorageClass, or emptyDir for throwaway data
- **`spec.node_pools[].pdb`**: per-pool PodDisruptionBudget (one of
  `min_available` / `max_unavailable`)
- **`spec.additional_config`**: extra opensearch.yml entries for ALL
  pools (a pool's own `additional_config` wins) — cluster-formation,
  network and TLS keys are operator-owned
- **`spec.keystore`**: Secrets whose entries land in the OpenSearch
  keystore on every node (`key_mappings` renames) — the safe home for
  snapshot-repository credentials
- **`spec.snapshot_repositories`**: name + type + settings (e.g. type
  `s3` with bucket/region needs the `repository-s3` plugin in
  `plugins_list` and credentials in the keystore or node identity)
- **`spec.monitoring`**: the Aiven prometheus-exporter plugin plus a
  ServiceMonitor — requires the Prometheus Operator CRDs; enabling it
  without them fails reconciliation
- **`spec.service_annotations`**: annotations for the operator-created
  Services — the cloud-controller injection surface (internal LB
  recipes ride here)
- **`spec.set_vm_max_map_count`**: the privileged init container that
  raises `vm.max_map_count` (default true — most distributions ship a
  lower kernel default and OpenSearch crash-loops without it)
- **`spec.drain_data_nodes`**: drain before stopping during rolling
  operations — for large data nodes where replica recovery is slower
  than draining
- **`spec.additional_volumes` / `spec.image` /
  `spec.image_pull_secrets` / `spec.bootstrap`**: projected
  Secret/ConfigMap volumes, the air-gap image path, and bootstrap-pod
  tuning

## Environment Injection

This component calls no cloud APIs; managed-Kubernetes integration
rides the Service annotations and the keystore.

| Cloud / posture | Where | Mechanism |
|---|---|---|
| Internal / external LB recipes | `service_annotations` | Cloud-controller annotations on the operator-created Services |
| S3 snapshots, declared keys | `keystore` + `snapshot_repositories` | `repository-s3` plugin; access/secret keys loaded into the node keystore from a Secret |
| S3 snapshots, keyless (EKS) | `snapshot_repositories` only | IRSA-bound identity on the nodes — no keystore entry |
| GCS snapshots, declared key | `keystore` + `snapshot_repositories` | `repository-gcs` plugin; the service-account key loaded via the keystore |
| GCS snapshots, keyless (GKE) | `snapshot_repositories` only | Workload Identity on the nodes |

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `cluster_name` | Name of the OpenSearchCluster resource (= `metadata.name`) |
| `service_name` | The cluster's main Service (all nodes; = `metadata.name` — the module pins the operator's serviceName to it) |
| `http_endpoint` | In-cluster HTTP API endpoint — ALWAYS https (the operator serves TLS on the HTTP layer in every posture) |
| `admin_credentials_secret_name` | The operator-generated `<name>-admin-password` Secret (fields `username`/`password`) — the DEMO credentials unless rotated; empty when a custom security config replaces the bootstrap |
| `dashboards_service_name` | The Dashboards Service (`<name>-dashboards`); empty when Dashboards are not enabled |
| `dashboards_endpoint` | In-cluster Dashboards endpoint (port 5601); empty when not enabled |
| `port_forward_command` | Port-forward command for workstation access when no exposure is composed |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`); a pool's
  **`persistence.pvc.storage_class`** references a
  KubernetesStorageClass; TLS **`secret`** fields reference a
  KubernetesCertificate's output secret — the cert-manager seam.
- **Applications consume the outputs**: `http_endpoint` as the
  connection URL (https — trust the operator-generated CA or the
  provided certificate chain), `admin_credentials_secret_name` as
  env-from references — credentials ride the Secret, never the
  manifest.
- **Exposure composes, never embeds**: a KubernetesIngress or Gateway
  API route targets `service_name` (the API) or
  `dashboards_service_name` (the console).
- **The operator is a cluster prerequisite**, not a reference: deploy
  KubernetesOpenSearchOperator first, watching this namespace.

## Examples

### Development (minimal two-node, demo security)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearch
metadata:
  name: dev-search
spec:
  namespace:
    value: dev-search
  create_namespace: true
  version: "2.19.4"
  node_pools:
    - name: all
      replicas: 2
      roles: [cluster_manager, data, ingest]
      jvm: "-Xmx512M -Xms512M"
```

### Production (3-node core, generated TLS, Dashboards, S3 snapshots)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearch
metadata:
  name: main
spec:
  namespace:
    value: search
  version: "2.19.4"
  node_pools:
    - name: core
      replicas: 3
      roles: [cluster_manager, data]
      disk_size: 100Gi
      jvm: "-Xmx2G -Xms2G"
      resources:
        requests: { cpu: "1", memory: 4Gi }
        limits: { cpu: "2", memory: 4Gi }
      pdb:
        enable: true
        max_unavailable: "1"
  security:
    transport_tls: {} # operator-generated CA, per-node certificates
    http_tls: {}
  dashboards:
    enabled: true
  plugins_list:
    - repository-s3
  keystore:
    - secret:
        value: s3-snapshot-keys # keys: s3.client.default.access_key / secret_key
  snapshot_repositories:
    - name: nightly
      type: s3
      settings:
        bucket: my-snapshots
        region: us-west-2
```

Rotate the admin password through the security API immediately after
install, or replace the bootstrap with a custom `security.config`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

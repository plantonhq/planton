# Kubernetes Solr

## When NOT to Use This

**The operator must already be on the cluster.** This component
declares a SolrCloud cluster; KubernetesSolrOperator installs the
ENGINE that reconciles it — and, for provided ZooKeeper ensembles, the
bundled zookeeper-operator that provisions them. Deploy the operator
first, clusters after.

Also not the right kind when:

- **You want the operator itself** — installing and configuring the
  Apache Solr Operator is KubernetesSolrOperator; this kind is
  one SolrCloud cluster it manages.
- **You want a managed cloud search service** — use the host cloud
  provider's managed search kinds; this kind is for running
  SolrCloud ON the Kubernetes cluster itself.
- **You expect durable storage by default** — the operator's default
  is EPHEMERAL (emptyDir): data is LOST when a pod leaves its node.
  Declare `storage.persistent` for anything beyond throwaway
  experiments.
- **You expect an open cluster to be safe** — empty `security` means
  no authentication at all; anyone with network reach can read, write,
  and administer. Development only.

## Overview

**KubernetesSolr** declares an Apache SolrCloud cluster as a
`SolrCloud` custom resource reconciled by the Apache Solr Operator.
The operator manages the node StatefulSet, ZooKeeper wiring, managed
rolling updates that keep shards available, scale operations that move
replicas off departing pods, TLS, basic-auth security bootstrap, and
registered backup repositories.

**The naming contract**: every object the operator creates derives
from `metadata.name` — the StatefulSet (`<name>-solrcloud`), the
common Service fronting all nodes (`<name>-solrcloud-common`), the
generated basic-auth Secret (`<name>-solrcloud-basic-auth`), and the
provided ZooKeeper client service
(`<name>-solrcloud-zookeeper-client`).

**Key design points:**

- **ZooKeeper wiring is a oneof** — empty means a provided 3-node
  ensemble managed by the zookeeper-operator the
  KubernetesSolrOperator chart bundles (replicas, persistence, and
  chroot are tunable); `zookeeper.external` connects to an ensemble
  you already run instead.
- **Storage honesty** — ephemeral (emptyDir) is the operator default
  and data does not survive pod eviction. `storage.persistent`
  declares PVC-backed volumes with a `reclaim_policy` defaulting to
  Retain (data outlives the resource).
- **Basic-auth bootstrap** — `security.authentication_type: basic` has
  the operator bootstrap security.json with `admin`, `solr`, and
  `k8s-oper` users carrying random passwords. The admin and solr
  credentials (plus the bootstrapped security.json) land in
  `<name>-solrcloud-security-bootstrap`; the operator's own `k8s-oper`
  credentials land in `<name>-solrcloud-basic-auth` (the exported
  handle). THE ROTATION CONTRACT: if you later rotate a password
  through Solr's security API, update the Secret too — the operator
  locks itself out otherwise (the upstream contract). No credential
  ever appears in this spec.
- **Backup repositories register at the cluster** — `s3` (declared
  keys from Secrets, or an empty credentials block for the keyless
  IRSA path), `gcs` (a service-account key Secret, or keyless GKE
  Workload Identity), or `volume` (an existing PVC that MUST be
  multi-writer: ReadWriteMany or NFS). Backups themselves are
  operational verbs — SolrBackup resources or Solr API calls — not
  declared here.
- **Managed updates keep shards available** — the operator's Managed
  strategy restarts pods in parallel while bounding both
  `max_pods_unavailable` (default 25%) and
  `max_shard_replicas_unavailable` (default 1); scale-downs vacate
  replicas off departing pods and scale-ups populate new ones (both
  default true).
- **Two exposure paths** — the `external` block models the operator's
  OWN Ingress/ExternalDNS exposure with per-node addressability (what
  SolrJ/CloudSolrClient needs to reach individual nodes from outside);
  composing a KubernetesIngress or Gateway API route over the common
  service handle is equally valid and keeps exposure a first-class
  graph node.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the cluster — literal or a
  KubernetesNamespace reference; the operator must watch it
- **`spec.version`**: the Solr version — the image tag of the official
  `solr` image (e.g. "9.10.0"); check the operator's compatibility
  matrix before pinning a new major line

### Common

- **`spec.replicas`**: Solr node count (default 3; one works for
  development, production shard/replica placement needs 3+)
- **`spec.zookeeper`**: provided (replicas default 3 — the quorum
  minimum; persistence size and StorageClass; chroot) or external
  (connection string + chroot)
- **`spec.storage`**: `persistent` (size required; StorageClass;
  `reclaim_policy` Retain/Delete) or `ephemeral` (optional size limit)
- **`spec.java_mem`**: JVM heap (operator default "-Xms300m -Xmx300m"
  — size to roughly half the container memory)
- **`spec.resources`**: empty = no requests/limits — set them for any
  real deployment, paired with `java_mem`
- **`spec.security`**: `authentication_type: basic` plus optional
  `basic_auth_secret` (bring your own kubernetes.io/basic-auth
  credentials for the operator), `probes_require_auth`, and
  `bootstrap_security_json` (bring-your-own security.json — applied
  once, never updated)
- **`spec.tls`**: keystore-based server TLS — a PKCS#12 keystore
  Secret + keystore password Secret (e.g. produced by a
  KubernetesCertificate with a pkcs12 keystore output), optional
  truststore, and `client_auth` (None/Want/Need — Need requires the
  operator's own mTLS identity from KubernetesSolrOperator, or probes
  and reconciliation fail)
- **`spec.backup_repositories`**: named repositories with exactly one
  backend arm (s3/gcs/volume); modules a repository needs load
  automatically
- **`spec.update_strategy`**: Managed (default), StatefulSet, or
  Manual; Managed budgets and an optional cron `restart_schedule`
- **`spec.external`**: the operator's own exposure — `method`
  (Ingress/ExternalDNS), `domain_name` (nodes become
  `<ns>-<name>-solrcloud-*.<domain>`), `use_external_address`,
  `hide_common`/`hide_nodes`
- **`spec.solr_modules` / `spec.additional_libs` / `spec.solr_opts` /
  `spec.gc_tune` / `spec.log_level` / `spec.pod_port`**: runtime
  tuning (pod port default 8983; the common Service exposes the
  operator's own default — 80, or 443 with TLS)
- **`spec.node_selector` / `spec.tolerations`**: Solr pod scheduling
- **`spec.availability.pdb_enabled`**: the cluster-wide
  PodDisruptionBudget (operator default true)

## Environment Injection

Backup repositories are where cloud identity rides.

| Cloud / posture | Where | Mechanism |
|---|---|---|
| S3, declared keys | `backup_repositories[].s3.credentials` | Access key ID + secret access key from existing Secrets |
| S3, keyless (EKS) | `backup_repositories[].s3` with empty credentials | The nodes' ambient identity — an IRSA-bound ServiceAccount on the Solr pods |
| S3-compatible (MinIO, ...) | `s3.endpoint` | Full endpoint URL; `region` still required (any value works) |
| GCS, declared key | `backup_repositories[].gcs.gcs_credential_secret` | A Google service-account key JSON from a Secret |
| GCS, keyless (GKE) | `backup_repositories[].gcs` without a credential | Workload Identity on the Solr pods |
| Shared volume | `backup_repositories[].volume` | An existing ReadWriteMany PVC mounted to every node |

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `cluster_name` | Name of the SolrCloud resource (= `metadata.name`) |
| `common_service_name` | The common Service fronting all nodes (`<name>-solrcloud-common`) |
| `internal_endpoint` | In-cluster base URL through the common service (http on 80, https on 443 with TLS) |
| `basic_auth_secret_name` | The operator-generated basic-auth Secret (`<name>-solrcloud-basic-auth`, fields `username`/`password` — the `k8s-oper` user); empty when security is disabled or a user-provided `basic_auth_secret` is in play |
| `zookeeper_connection_string` | The ensemble the cluster actually uses (host:port, plus the chroot when it diverges from "/") |
| `port_forward_command` | Port-forward command for workstation access when no exposure is composed |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`); storage and ZooKeeper
  **`storage_class`** fields reference a KubernetesStorageClass; the
  TLS keystore Secrets are the cert-manager seam (a
  KubernetesCertificate with a pkcs12 keystore output).
- **Applications consume the outputs**: `internal_endpoint` as the
  base URL, `basic_auth_secret_name` as env-from references —
  credentials ride operator-managed Secrets, never the manifest. The
  admin and solr user passwords live in
  `<name>-solrcloud-security-bootstrap`.
- **Exposure composes or embeds — your call**: a KubernetesIngress
  over `common_service_name` for simple HTTP access, or the `external`
  block when clients need per-node addressability (CloudSolrClient
  outside the cluster).
- **The operator is a cluster prerequisite**, not a reference: deploy
  KubernetesSolrOperator first (with its bundled zookeeper-operator
  for provided ensembles).

## Examples

### Development (single node, ephemeral)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolr
metadata:
  name: dev-solr
spec:
  namespace:
    value: dev-solr
  create_namespace: true
  version: "9.10.0"
  replicas: 1
  zookeeper:
    provided:
      replicas: 1
```

### Production (persistent, basic auth, S3 backups, sized JVM)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolr
metadata:
  name: main
spec:
  namespace:
    value: search
  version: "9.10.0"
  replicas: 3
  storage:
    persistent:
      size: 50Gi
  java_mem: "-Xms2g -Xmx2g"
  resources:
    requests: { cpu: "1", memory: 4Gi }
    limits: { cpu: "2", memory: 4Gi }
  security:
    authentication_type: basic
  backup_repositories:
    - name: nightly
      s3:
        region: us-west-2
        bucket: my-solr-backups
        # empty credentials = the nodes' ambient identity (IRSA)
  update_strategy:
    method: Managed
    max_shard_replicas_unavailable: "1"
```

### External ZooKeeper

```yaml
spec:
  zookeeper:
    external:
      connection_string: "zk-0.zk:2181,zk-1.zk:2181,zk-2.zk:2181"
      chroot: /solr
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

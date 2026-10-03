# Apache Solr

Deploys an Apache Solr SolrCloud cluster — the open-source search platform built on Lucene. The cluster is declared as a `SolrCloud` custom resource reconciled by the Apache Solr Operator, which manages the full lifecycle: the node StatefulSet, ZooKeeper wiring, shard-aware managed rolling updates, scale operations that move replicas off departing pods, TLS, the basic-auth security bootstrap, and registered backup repositories.

Every SolrCloud needs a ZooKeeper ensemble for its collection topology. The default (an empty `zookeeper` block) provisions a 3-node ensemble through the zookeeper-operator the Apache Solr Operator chart bundles; an external ensemble you already run is one field away.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SolrCloud custom resource** — the node StatefulSet, the common Service fronting all nodes, per-node Services, and (by default) a bundled ZooKeeper ensemble, all reconciled by the operator
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist
- **Security bootstrap** — with `security.authenticationType: basic`, the operator bootstraps `security.json` plus the admin/solr/k8s-oper users and writes generated credentials to Secrets (no credential ever appears in the spec)
- **Backup repositories** — registered on the cluster at startup when declared; the Solr modules they need (`s3-repository`, `gcs-repository`) load automatically

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **Apache Solr Operator** — an **Apache Solr Operator** resource must be running and reconciling SolrCloud resources in the target namespace (its default watch scope is ALL namespaces). Deploy it first.
- **A StorageClass** for persistent volumes — required for `storage.persistent` and for the ZooKeeper ensemble's own volumes. The operator default storage is ephemeral emptyDir: data is LOST when a pod leaves its node, so declare persistent storage for anything beyond throwaway experiments.
- **An ingress controller or external-dns** — only if you use the operator's own `external` exposure block.

## Deploy

### Console

Open the deployment store, find **Apache Solr**, and click **Deploy**. The creation wizard walks you through namespace placement, the Solr version and node count, the ZooKeeper quorum, the storage data-safety decision, JVM tuning, node resources, the basic-auth security posture, keystore-model TLS, backup repositories, modules, lifecycle dials, and exposure. Start from the **Dev single node preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolr
metadata:
  name: dev-solr
  org: acme-corp
  env: dev
spec:
  namespace:
    value: dev-solr
  createNamespace: true
  replicas: 1
  version: 9.10.0
  zookeeper:
    provided:
      replicas: 1
  storage:
    ephemeral: {}
  resources:
    requests:
      cpu: 250m
      memory: 768Mi
    limits:
      cpu: "1"
      memory: 1Gi
```

```shell
planton apply -f solr.yaml
```

This creates the smallest declarable SolrCloud that actually serves: one Solr node, a single-member provided ZooKeeper, ephemeral storage, no authentication — the full SolrCloud API surface for development and CI. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose the cluster behind its namespace with a reference, and the InfraPipeline orders the deploys:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: search-namespace
      fieldPath: spec.name
  createNamespace: false
  version: 9.10.0
```

The InfraPipeline deploys the namespace first, then provisions the cluster into it.

## Key Configuration

These are the most important decisions when configuring an Apache Solr cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Storage is the data-safety decision** — the operator default is ephemeral emptyDir: indices are LOST whenever a pod leaves its node. Declare `storage.persistent` (with a size like `100Gi` and an SSD StorageClass) for real workloads. The `reclaimPolicy` default is Retain — the index volumes outlive the resource; deletion is not data loss unless you choose Delete.

**ZooKeeper is protected like data** — losing ZooKeeper state means losing collection topology. Production runs a 3-member quorum with its own persistent volumes (`zookeeper.provided.persistence`); the ensemble does NOT scale with Solr nodes. A 1-member ensemble is fine for dev only — any restart pauses Solr.

**Heap at half the container memory** — `javaMem` (operator default `-Xms300m -Xmx300m`) should be a fixed heap at roughly half the container memory from `resources`; the rest is OS page cache Lucene leans on. The production shape: 4Gi memory with `-Xms2g -Xmx2g`.

**Security empty means OPEN** — an untouched `security` block is an unauthenticated cluster (development only). `authenticationType: basic` bootstraps security.json plus the admin/solr/k8s-oper users with generated credentials in Secrets. If you later rotate a password through Solr's security API, update the Secret too — the operator locks itself out otherwise.

**Backups register up front** — `backupRepositories` are part of the SolrCloud spec, so adding one later is a rolling restart. S3/GCS credentials are keyless by absence: leave the credentials block empty on EKS with IRSA (or GKE with Workload Identity) and the client uses the pods' ambient identity. Volume repositories need an existing ReadWriteMany PVC.

**Exposure is deliberate** — in-cluster access rides the common Service (see the outputs). The `external` block is the operator's own Ingress/ExternalDNS exposure with per-node addressability (what CloudSolrClient needs from outside); composing a **Kubernetes Ingress** over the exported common-service handle is the equally valid simple-HTTP path.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `storage.persistent.storageClass` / `zookeeper.provided.persistence.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesSecret** | `security.basicAuthSecret` | `metadata.name` |

The TLS keystore fields (`tls.pkcs12Secret`, `tls.keystorePasswordSecret`, truststore variants), `security.bootstrapSecurityJson`, and backup-repository credentials are name+key references to EXISTING Secrets in the install namespace — not foreign keys.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Application deployment manifests |
| `cluster_name` | Name of the SolrCloud resource (= metadata.name) | Operational tooling, SolrBackup CRs |
| `common_service_name` | The common Service fronting all nodes (`<name>-solrcloud-common`) | Ingress/Gateway composition |
| `internal_endpoint` | In-cluster base URL (the common service listens on 80, or 443 with TLS) | Application connection strings |
| `basic_auth_secret_name` | The operator-generated basic-auth Secret (the read-only k8s-oper user; admin/solr passwords live in the sibling `<name>-solrcloud-security-bootstrap` Secret). Empty when security is disabled or a user-provided Secret is in play | Client authentication |
| `zookeeper_connection_string` | The ensemble the cluster uses (host:port/chroot) | Diagnostics, shared-ensemble planning |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the common service | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single node** — one Solr node, a single-member ZooKeeper, ephemeral storage, no auth: the real SolrCloud API for developers and CI without production ceremony. Start from the **Dev single node preset**.

**Production cloud** — three nodes on persistent SSD storage with Retain, a three-member ZooKeeper quorum with its own volumes, basic auth, Managed rolling updates bounded to one pod / one shard replica at a time, and the cluster-wide PodDisruptionBudget. Start from the **Production cloud preset**.

**S3 backups** — the production shape plus an S3 backup repository registered from day one, with the keyless-credential teaching (delete the credentials block under IRSA). Start from the **S3 backups preset**.

## Works With

- [**Apache Solr Operator**](/infra-catalog/kubernetes-solr-operator) — the engine that reconciles this cluster; deploy it first (its `mtls` block is also the prerequisite for `tls.clientAuth: Need`).
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first.
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — SSD-backed classes for index and ensemble volumes.
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — the natural PKCS#12 keystore producer for `spec.tls` (enable its pkcs12 keystore output).
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — bring-your-own credentials, bootstrap security.json, static backup credentials.
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) and [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) — simple HTTP exposure over the exported common service.

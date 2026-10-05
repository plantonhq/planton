# External Secret

Declares ONE secret sync: the External Secrets Operator reads the referenced entries from a store's backend (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao, and more) and materializes them as a Kubernetes Secret in this namespace, refreshing on the configured interval. Workloads consume the materialized Secret exactly like any other (env valueFrom, volume mounts) — the external system stays the single source of truth and the cluster never stores the value anywhere else. The store connection is a separate first-class resource (Kubernetes Secret Store / Kubernetes Cluster Secret Store); this resource picks WHAT to sync. Requires the External Secrets Operator on the cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ExternalSecret** -- the ESO custom resource declaring the sync (store reference, entries/pulls, target, refresh lifecycle)
- **Kubernetes Secret** (materialized by the operator, not the module) -- named `target.name` (defaulting to this resource's name), refreshed on the interval, consumed by workloads

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster.
- **External Secrets Operator** -- deployed on the cluster.
- **A store** -- a Kubernetes Secret Store in the same namespace, or a Kubernetes Cluster Secret Store whose access fence allows this namespace.

### Backend Side

- The entries you reference must exist in the backend, and the STORE's identity must be authorized to read them — whatever that identity can read, this resource can sync.

## Deploy

### Console

Open the deployment store, find **External Secret**, and click **Deploy**. The creation wizard walks you through the namespace, the store selector, the sync declarations (explicit entries and bulk pulls), the target Secret's policies and template, and the refresh lifecycle. Start from the **Explicit Keys (Application Credentials)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesExternalSecret
metadata:
  name: app-database
  org: acme-corp
  env: prod
spec:
  namespace:
    value: team-a
  storeRef:
    secretStore:
      name:
        value: team-a-vault
  data:
    - secretKey: username
      remoteRef:
        key: prod/app/database
        property: username
    - secretKey: password
      remoteRef:
        key: prod/app/database
        property: password
```

```shell
planton apply -f external-secret.yaml
```

This syncs two fields of one structured backend entry into a Kubernetes Secret named `app-database`, refreshed hourly (the upstream default). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace and the store to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: team-a-namespace
      fieldPath: spec.name
  storeRef:
    secretStore:
      name:
        valueFrom:
          name: team-a-vault
          fieldPath: status.outputs.store_name
```

The InfraPipeline deploys the namespace and the store first, then declares the sync against the resolved store name.

## Key Configuration

These are the most important decisions when configuring the sync. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Explicit entries vs bulk pulls** -- a `data` entry names exactly one backend key (or one property within it) and one Secret key: reviewable, impossible to over-sync — prefer it for application credentials. A `dataFrom` pull extracts ALL properties of one structured entry, or finds every entry matching a name pattern/tags: for JSON documents of related credentials and fleet patterns. When both produce the same key, the explicit entry wins.

**The store's grain is the arm you choose** -- `storeRef.secretStore` names a namespaced SecretStore in this secret's namespace, `storeRef.clusterSecretStore` a cluster-scoped ClusterSecretStore. Each arm references its own kind (KubernetesSecretStore or KubernetesClusterSecretStore), so a reference inherits the store's `store_name` output and draws the dependency edge to the right store.

**Key rewrites** -- pulls can reshape their key names in order (`^prod/app/(.*)$` → `$1` strips the path prefix). Names only; values are never touched.

**Deletion policy Delete prunes** -- with `target.deletionPolicy: Delete`, a key removed from the backend disappears from the cluster on the next refresh — workloads reading it break in sync with the backend. Retain (the default) never surprises.

**Immutable pairs with CreatedOnce** -- an immutable target Secret cannot be updated, so every refresh that would change data FAILS. Pair `target.immutable` with `refreshPolicy: CreatedOnce` (and `refreshInterval: 0s`) — the immutable-bootstrap-secret pattern.

**The template reshapes values** -- Go templates over the synced keys turn raw credentials into what workloads want: a connection string, a `kubernetes.io/dockerconfigjson`, a `kubernetes.io/tls` pair. With the default Replace policy, templated data REPLACES the synced keys — switch to Merge to keep both.

**Refresh is not a pod restart** -- env-var consumers keep the old value until they restart; volume mounts update in place. Pair rotation with a reloader when workloads must follow immediately.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesSecretStore** | `storeRef.secretStore.name` | `status.outputs.store_name` |
| **KubernetesClusterSecretStore** | `storeRef.clusterSecretStore.name` | `status.outputs.store_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `secret_name` | Name of the materialized Kubernetes Secret | Workload env `valueFrom.secretKeyRef` / volume `secretName` — wire references to THIS, never to the ExternalSecret |
| `external_secret_name` | Name of the ExternalSecret resource | Debugging the sync (`kubectl describe externalsecret`) |
| `namespace` | Namespace both live in | Placing consumers |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application credentials** -- two named fields from one structured entry, refreshed hourly. Start from the **Explicit Keys (Application Credentials)** preset.

**Whole JSON document** -- one `dataFrom.extract` pulls every property, with a rewrite stripping the path prefix. Start from the **Extract a JSON Document (Bulk Pull with Rewrite)** preset.

**Registry pull secret** -- synced registry credentials templated into a `kubernetes.io/dockerconfigjson` Secret for `imagePullSecrets`. Start from the **Docker Registry Pull Secret (Template)** preset.

## Works With

- [**Secret Store**](/infra-catalog/kubernetes-secret-store) -- the namespaced connection this sync reads through; deploy one first
- [**Cluster Secret Store**](/infra-catalog/kubernetes-cluster-secret-store) -- the cluster-scoped alternative when many namespaces share one backend connection
- [**External Secrets Operator**](/infra-catalog/kubernetes-external-secrets-operator) -- the machinery that does the syncing
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- the typical consumer: wire env and volumes to this resource's `secret_name` output instead of hardcoding; a registry pull secret rendered here is named in the workload's `pod.imagePullSecrets` with an explicit `kind: KubernetesExternalSecret` and `fieldPath: status.outputs.secret_name` -- the pull route that works without a Planton backend, because the cluster reads the credential from your secrets manager

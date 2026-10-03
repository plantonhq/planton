# Secret Store

Creates a NAMESPACED External Secrets Operator store — one backend connection (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao, another Kubernetes cluster, or the fake test backend) that only ExternalSecret resources in the SAME namespace may sync from. The connection surface is identical to Cluster Secret Store — the grain is the only difference: this store's credential Secrets live in its namespace, and its blast radius ends at the namespace boundary. Requires the External Secrets Operator on the cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SecretStore** -- the namespaced ESO custom resource, named after `metadata.name` (ExternalSecrets in the namespace reference it by that name; kind SecretStore is the upstream default)
- **Credential Secret** (only for declared static credentials) -- materialized in the store's own namespace with a deterministic name; the credential never appears inside the store resource itself. Keyless postures materialize nothing.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster.
- **External Secrets Operator** -- deployed on the cluster; this store is a CR that operator reconciles.

### Backend Side

- The identity that reads the backend must exist and be authorized: a Workload-Identity/IRSA-bound ServiceAccount in the store's namespace for the keyless postures, or an exported credential for the static ones.
- For Vault: the KV engine mount path and version must match how the engine was mounted; AppRole needs the role configured on the Vault side.

## Deploy

### Console

Open the deployment store, find **Secret Store**, and click **Deploy**. The creation wizard walks you through the namespace, the backend connection, authentication, and tuning. Start from the **Team-Scoped GCP Secret Manager (Keyless)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSecretStore
metadata:
  name: team-a-gcp
  org: acme-corp
  env: prod
spec:
  namespace:
    value: team-a
  config:
    gcpSecretManager:
      projectId:
        value: acme-prod
      serviceAccountName:
        value: team-a-eso-reader
```

```shell
planton apply -f secret-store.yaml
```

This creates a store in `team-a` that reads GCP Secret Manager keylessly through the team's own Workload-Identity-bound ServiceAccount. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the store to its namespace and reader identity:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: team-a-namespace
      fieldPath: spec.name
  config:
    gcpSecretManager:
      projectId:
        value: acme-prod
      serviceAccountName:
        valueFrom:
          kind: KubernetesServiceAccount
          name: team-a-eso-reader
          fieldPath: status.outputs.service_account_name
```

The InfraPipeline deploys the namespace and the ServiceAccount first, then creates the store against them.

## Key Configuration

These are the most important decisions when configuring the store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The namespace IS the fence** -- only ExternalSecrets in this namespace can use the store; nothing to configure, nothing to get wrong. If a backend should serve several namespaces, use Cluster Secret Store (with its conditions fence) instead.

**The team's own identity** -- keyless via a ServiceAccount in THIS namespace is where the namespaced grain earns its keep: the identity carries the team's cloud grant and nothing outside the namespace can borrow it. A referenced ServiceAccount without an explicit namespace defaults to the store's own — almost always what you mean.

**Static credentials never enter the CR** -- they bind as `$secret/<slug>` references and the IaC modules materialize them as a Kubernetes Secret in this namespace; the store references that Secret. Prefer keyless wherever the backend supports federation — exported keys never expire and never rotate themselves.

**Tuning is the store's clock** -- `config.refreshInterval` re-validates the CONNECTION on a cycle; how often each synced secret refreshes is per-ExternalSecret. `controller_class` only matters with several operator installations.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **GcpProject** | `config.gcpSecretManager.projectId` | `status.outputs.project_id` |
| **KubernetesServiceAccount** | `config.*.serviceAccountName` | `status.outputs.service_account_name` |
| **AwsIamRole** | `config.awsSecretsManager.role` | `status.outputs.role_arn` |
| **AzureKeyVault** | `config.azureKeyVault.vaultUrl` | `status.outputs.vault_uri` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `store_name` | Name of the created SecretStore | Kubernetes External Secret's `store_ref.name` (kind SecretStore, the default) |
| `namespace` | Namespace the store (and its credential Secrets) live in | Placing the ExternalSecrets that use it |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team GCP Secret Manager** -- the team's own Workload-Identity-bound ServiceAccount reads the team's secrets. Start from the **Team-Scoped GCP Secret Manager (Keyless)** preset.

**Vault with AppRole** -- machine-identity auth (role-id + secret-id) for Vault/OpenBao KV v2. Start from the **Team-Scoped Vault KV with AppRole** preset.

**Fake sandbox** -- ESO's built-in fake backend serves declared literals: no external account, no network, fully deterministic — for pipelines and tests. Start from the **Fake Backend Sandbox (Test-Only)** preset.

## Works With

- [**External Secrets Operator**](/infra-catalog/kubernetes-external-secrets-operator) -- must be on the cluster first.
- [**External Secret**](/infra-catalog/kubernetes-external-secret) -- declares each secret to sync through this store, in the same namespace.
- [**Cluster Secret Store**](/infra-catalog/kubernetes-cluster-secret-store) -- the cluster-wide twin, for platform backends every team shares.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the store's home, credential home, and access fence.

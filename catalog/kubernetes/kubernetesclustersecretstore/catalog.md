# Cluster Secret Store

Creates a CLUSTER-scoped External Secrets Operator store — one backend connection (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao, another Kubernetes cluster, or the fake test backend) that ExternalSecret resources in ANY allowed namespace may sync from. The store holds no secret data: it is a named connection plus the identity ESO uses to read the backend. Use the cluster grain for platform-wide backends every team shares (fenced with conditions); use the namespaced Secret Store for a connection that belongs to ONE namespace/team. Requires the External Secrets Operator on the cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ClusterSecretStore** -- the cluster-scoped ESO custom resource, named after `metadata.name` (ExternalSecrets reference it by that name with kind ClusterSecretStore)
- **Credential Secret** (only for declared static credentials) -- materialized in the secrets namespace with a deterministic name; the credential never appears inside the store resource itself. Keyless postures materialize nothing.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster.
- **External Secrets Operator** -- deployed on the cluster; this store is a CR that operator reconciles.

### Kubernetes Cluster

- The identity that reads the backend must exist and be authorized: an IRSA/Workload-Identity-bound ServiceAccount for the keyless postures, or an exported credential for the static ones.
- For Vault: the KV engine mount path and version must match how the engine was mounted; Kubernetes auth needs the role configured on the Vault side.

## Deploy

### Console

Open the deployment store, find **Cluster Secret Store**, and click **Deploy**. The creation wizard walks you through the secrets home, the backend connection, authentication, tuning, and the namespace access fence. Start from the **AWS Secrets Manager with IRSA (Keyless)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClusterSecretStore
metadata:
  name: aws-prod-secrets
  org: acme-corp
  env: prod
spec:
  secretsNamespace:
    value: external-secrets
  config:
    aws:
      region: us-east-1
      serviceAccountName:
        value: eso-aws-reader
      serviceAccountNamespace: external-secrets
```

```shell
planton apply -f cluster-secret-store.yaml
```

This creates a cluster-wide connection to AWS Secrets Manager authenticating keylessly through the referenced ServiceAccount's IRSA binding. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the secrets home and reader identity by reference:

```yaml
spec:
  secretsNamespace:
    valueFrom:
      kind: KubernetesExternalSecretsOperator
      name: eso
      fieldPath: status.outputs.namespace
  config:
    aws:
      region: us-east-1
      serviceAccountName:
        valueFrom:
          kind: KubernetesServiceAccount
          name: eso-aws-reader
          fieldPath: status.outputs.service_account_name
      serviceAccountNamespace: external-secrets
```

The InfraPipeline deploys the operator and the ServiceAccount first, then creates the store against them.

## Key Configuration

These are the most important decisions when configuring the store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The secrets namespace** -- a cluster-scoped store reads its referenced credential Secrets from one explicit namespace (`secretsNamespace`); the operator's install namespace is the convention. Reference your External Secrets Operator resource to inherit its namespace output — the reference also draws the dependency edge.

**Three authentication postures** -- keyless via a referenced ServiceAccount (per-store identities — the multi-team production posture), keyless via the operator's ambient identity (leave the auth block empty), or declared static credentials (materialized as a Kubernetes Secret; the value binds as a `$secret/<slug>` reference, never plaintext). For a cluster-scoped store, a referenced ServiceAccount needs its namespace stated explicitly.

**The access fence** -- with no `conditions`, EVERY namespace may sync from this store. Conditions allow namespaces by exact name, label selector, or name regex; everything unions, and there is no deny rule. A store holding production credentials should not be readable from every dev namespace.

**One store per backend per posture** -- an `aws-prod` store and an `aws-dev` store with different identities beat one store every environment shares; the store's identity bounds what any ExternalSecret using it can sync.

**Tuning is the store's clock** -- `config.refreshInterval` re-validates the CONNECTION on a cycle; how often each synced secret refreshes is per-ExternalSecret. `controllerClass` only matters with several operator installations.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesExternalSecretsOperator** | `secretsNamespace` | `status.outputs.namespace` |
| **KubernetesServiceAccount** | `config.aws.serviceAccountName` (each backend has its own) | `status.outputs.service_account_name` |
| **AwsIamRole** | `config.aws.role` | `status.outputs.role_arn` |
| **GcpProject** | `config.gcpSecretManager.projectId` | `status.outputs.project_id` |
| **AzureKeyVault** | `config.azureKeyVault.vaultUrl` | `status.outputs.vault_uri` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `store_name` | Name of the created ClusterSecretStore | An External Secret's `secretStoreRef.name` (with kind ClusterSecretStore) |
| `secrets_namespace` | Namespace credential Secrets were materialized in | Auditing where the store's credentials live |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**AWS Secrets Manager via IRSA** -- keyless reads through a referenced ServiceAccount's IAM role binding. Start from the **AWS Secrets Manager with IRSA (Keyless)** preset.

**GCP Secret Manager via Workload Identity** -- keyless reads through a GKE Workload Identity binding. Start from the **GCP Secret Manager with Workload Identity (Keyless)** preset.

**Azure Key Vault via Workload Identity** -- keyless reads through an AKS federated credential. Start from the **Azure Key Vault with Workload Identity (Keyless)** preset.

**Vault with Kubernetes auth** -- the ServiceAccount token is exchanged for a Vault token; no Vault credential stored on the cluster. Start from the **Vault KV with Kubernetes Auth (Keyless)** preset.

## Works With

- [**External Secrets Operator**](/infra-catalog/kubernetes-external-secrets-operator) -- must be on the cluster first; cluster-scoped stores reference its namespace output as their secrets home.
- [**External Secret**](/infra-catalog/kubernetes-external-secret) -- declares each secret to sync through this store (kind ClusterSecretStore).
- [**Secret Store**](/infra-catalog/kubernetes-secret-store) -- the namespaced twin, for connections that belong to one team.
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) -- carries the IRSA / Workload Identity binding the keyless postures reference as the reader identity.

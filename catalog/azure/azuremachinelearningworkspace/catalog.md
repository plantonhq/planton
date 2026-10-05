# Azure Machine Learning Workspace

Deploys an Azure Machine Learning workspace -- the central home a data-science team keeps its experiments, models, endpoints, datastores, and compute in. The workspace plugs into a storage account, a key vault, and an application-insights component (all required, all referenced by typed refs), plus an optional container registry.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Machine Learning workspace** -- the ARM workspace object with its identity, companion-service attachments, optional CMK encryption, managed virtual network, and serverless-compute settings
- **FQDN outbound rules** (optional) -- one ARM child per entry allowing outbound traffic by domain name under approved-outbound isolation
- **Private-endpoint outbound rules** (optional) -- one ARM child per entry; the managed VNet creates a private endpoint to the referenced Azure resource
- **Service-tag outbound rules** (optional) -- one ARM child per entry allowing outbound traffic to an Azure service tag on given protocol/ports

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **The three companion services** -- an AzureStorageAccount (general-purpose, hierarchical namespace OFF), an AzureKeyVault, and an AzureApplicationInsights to reference. An AzureContainerRegistry is optional.

### Azure Subscription

- **The default storage account must not have hierarchical namespace enabled** -- ARM rejects Data Lake Gen2 accounts as workspace default storage.
- **Deletion is a soft delete** -- a deleted workspace holds its name until purged; recreating under the same name fails until then.

## Deploy

### Console

Open the deployment store, find **Azure Machine Learning Workspace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team Workspace** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningWorkspace
metadata:
  name: ml-prod
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: ml-platform-rg
  name: acme-ml-prod
  applicationInsightsId:
    valueFrom:
      name: ml-insights
  keyVaultId:
    valueFrom:
      name: ml-vault
  storageAccountId:
    valueFrom:
      name: mlartifacts
  identity:
    type: SYSTEM_ASSIGNED
```

```shell
planton apply -f azure-machine-learning-workspace.yaml
```

This creates a system-identity workspace attached to its three companion services. An Infra Job tracks the provisioning in real time.

### InfraChart

In an ML-platform chart the order is: resource group → storage account + key vault + log analytics → application insights → **workspace** → datastores and compute. Wire the companion services by reference:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: ml-platform-rg
      fieldPath: status.outputs.resource_group_name
  applicationInsightsId:
    valueFrom:
      kind: AzureApplicationInsights
      name: ml-insights
      fieldPath: status.outputs.application_insights_id
  keyVaultId:
    valueFrom:
      kind: AzureKeyVault
      name: ml-vault
      fieldPath: status.outputs.key_vault_id
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: mlartifacts
      fieldPath: status.outputs.storage_account_id
```

The InfraPipeline resolves the dependency graph and deploys the companion services before the workspace.

## Key Configuration

These are the most important decisions when configuring the workspace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The companion services are forever** -- `applicationInsightsId`, `keyVaultId`, `storageAccountId`, and `containerRegistryId` are all fixed at creation; re-pointing any of them replaces the workspace. Choose accounts sized for the workspace's lifetime.

**Identity** -- required. `SYSTEM_ASSIGNED` is the simple path; `USER_ASSIGNED` (with `primaryUserAssignedIdentity`) lets you grant storage/vault access before the workspace exists.

**Managed network isolation** -- `DISABLED` (default), `ALLOW_INTERNET_OUTBOUND`, or `ALLOW_ONLY_APPROVED_OUTBOUND`. Under approved-outbound, the three outbound-rule lists define what is reachable; rule names share one namespace across all three types.

**Public network access** -- leave enabled (the default) unless you have private endpoints in place; a private workspace with no-public-IP serverless compute also needs a serverless subnet.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureApplicationInsights** | `applicationInsightsId` | `status.outputs.application_insights_id` |
| **AzureKeyVault** | `keyVaultId` | `status.outputs.key_vault_id` |
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |
| **AzureContainerRegistry** | `containerRegistryId` (optional) | `status.outputs.container_registry_id` |
| **AzureUserAssignedIdentity** | `identity.identityIds`, `primaryUserAssignedIdentity` (optional) | `status.outputs.identity_id` |
| **AzureSubnet** | `serverlessCompute.subnetId` (optional) | `status.outputs.subnet_id` |
| **AzureKeyVaultKey** | `encryption.keyId` (optional) | `status.outputs.versionless_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `machine_learning_workspace_id` | ARM ID of the workspace | Datastores, compute, endpoints, and deployments reference it as their workspace |
| `workspace_guid` | The immutable workspace GUID | Data-plane SDKs, diagnostic settings |
| `discovery_url` | The regional discovery URL | SDK endpoint resolution |
| `system_assigned_identity_principal_id` | The system identity's principal ID | Storage / Key Vault role assignments |

The outbound-rule ID maps (`fqdn_outbound_rule_ids`, `private_endpoint_outbound_rule_ids`, `service_tag_outbound_rule_ids`) echo the managed-VNet rules back for inspection; no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team workspace** -- system identity on the standard companion trio. Start from the **Team Workspace** preset.

**Feature store** -- a `FEATURE_STORE`-flavor workspace (with the required `featureStore` block) backing online/offline feature serving. Deployed alongside, not instead of, the regular training workspace. Start from the **Feature Store** preset.

**Private hardened workspace** -- public access off, approved-outbound isolation, explicit outbound rules. Start from the **Private Hardened Workspace** preset.

## Works With

- [**Azure Machine Learning Datastore**](/infra-catalog/azure-machine-learning-datastore) -- saved data connections on the workspace
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the default artifact storage
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the workspace's secret store
- [**Azure Application Insights**](/infra-catalog/azure-application-insights) -- workspace telemetry
- [**Azure Container Registry**](/infra-catalog/azure-container-registry) -- environment images

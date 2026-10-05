# Azure AI Foundry Hub

Creates an Azure AI Foundry hub -- the shared foundation (security, storage, network posture) a company's AI teams create their Foundry projects in. The hub requires a key vault and a storage account at creation (both attachments are fixed for the hub's life), optionally wires Application Insights and a container registry, and owns the identity, encryption, and managed-network posture every project inside it inherits. Deletion is a soft delete: the ghost keeps holding the hub name until purged.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **AI Foundry Hub** -- ARM-wise an ML workspace of kind "Hub" (`Microsoft.MachineLearningServices/workspaces`), attached to your key vault and storage account, optionally wired to Application Insights and a container registry

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureKeyVault and an AzureStorageAccount** -- the hub's required companion services (both attachments are fixed at creation).

### Azure Subscription

- **Storage account shape** -- use a general-purpose account WITHOUT hierarchical namespace; the hub is an ML workspace at ARM, which rejects Data Lake Gen2 accounts as default storage.
- **Soft delete** -- a deleted hub becomes a purgeable ghost that keeps holding the hub name; recreating under the same name fails until the ghost is purged. Planton destroys purge automatically; for a hub deleted outside Planton, the Azure portal's "Recently deleted" view (per region) is the only place ghosts are listed and purged -- no CLI or REST API lists them.

## Deploy

### Console

Open the deployment store, find **Azure AI Foundry Hub**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team AI Hub** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAiFoundry
metadata:
  name: team-hub
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: ai-platform-rg
  name: acme-ai-hub
  keyVaultId:
    valueFrom:
      name: ai-platform-kv
  storageAccountId:
    valueFrom:
      name: ai-platform-storage
  identity:
    type: SYSTEM_ASSIGNED
  friendlyName: Team AI Hub
```

```shell
planton apply -f azure-ai-foundry.yaml
```

This creates a hub with a system-assigned identity, attached to the referenced key vault and storage account, in the referenced resource group. An Infra Job tracks the provisioning in real time.

### InfraChart

In an AI-platform chart the order is: resource group → key vault + storage account → **hub** → projects. Wire each layer with ValueFromRef so the InfraPipeline deploys them in dependency order:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: ai-platform-rg
      fieldPath: status.outputs.resource_group_name
  keyVaultId:
    valueFrom:
      kind: AzureKeyVault
      name: ai-platform-kv
      fieldPath: status.outputs.key_vault_id
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: ai-platform-storage
      fieldPath: status.outputs.storage_account_id
```

The InfraPipeline resolves the dependency graph, provisions the resource group, key vault, and storage account first, then creates the hub with the resolved ARM IDs.

## Key Configuration

These are the most important decisions when configuring the hub. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Identity** -- `SYSTEM_ASSIGNED` is the simple default (grant access after creation). Use `USER_ASSIGNED` with `primaryUserAssignedIdentity` when grants must exist BEFORE the hub -- required for CMK encryption.

**Encryption** -- the CMK block is fixed at creation, and its `keyId` must be a VERSIONED key URL (the hub's own contract -- rotation does not auto-propagate; re-point the field to rotate). This deliberately differs from the classic ML workspace's versionless guidance.

**Managed network** -- `ALLOW_ONLY_APPROVED_OUTBOUND` is the isolation posture; hubs manage their outbound rules from the Foundry portal/API (azurerm models no outbound-rule children for hubs).

**Attachments** -- key vault and storage are fixed at creation; Application Insights and the container registry attach or re-point in place (unlike the classic ML workspace, where the registry is ForceNew).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureKeyVault** | `keyVaultId` | `status.outputs.key_vault_id` |
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |
| **AzureApplicationInsights** | `applicationInsightsId` | `status.outputs.application_insights_id` |
| **AzureContainerRegistry** | `containerRegistryId` | `status.outputs.container_registry_id` |
| **AzureKeyVaultKey** | `encryption.keyId` | `status.outputs.key_id` (the VERSIONED URL) |
| **AzureUserAssignedIdentity** | `identity.identityIds[]` | `status.outputs.identity_id` |
| **AzureUserAssignedIdentity** | `primaryUserAssignedIdentity` | `status.outputs.identity_id` |
| **AzureUserAssignedIdentity** | `encryption.userAssignedIdentityId` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ai_foundry_id` | ARM ID of the hub | What AzureAiFoundryProject references as its `aiServicesHubId` |
| `ai_foundry_name` | The hub's name | Operational tooling |
| `workspace_guid` | The hub's immutable GUID | Data-plane SDKs, diagnostics |
| `discovery_url` | The hub's regional discovery URL | SDK endpoint resolution |
| `system_assigned_identity_principal_id` | The system identity's principal ID | Key vault / storage role assignments |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Shared team hub** -- the simple foundation: system-assigned identity, public endpoint, no CMK. The right start for most teams; the trade is that access grants happen after creation. Start from the **Team AI Hub** preset.

**Regulated estate** -- CMK encryption, disabled public access, approved-outbound isolation, user-assigned identity so key grants exist before the hub does. Every hardening choice here is fixed at creation or one-way, so start hardened rather than retrofitting. Start from the **CMK-Hardened AI Hub** preset.

## Works With

- [**Azure AI Foundry Project**](/infra-catalog/azure-ai-foundry-project) -- the per-team workspace created inside this hub
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- required secrets companion
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- required artifacts companion
- [**Azure AI Search Service**](/infra-catalog/azure-search-service) -- retrieval for the projects' RAG applications
- [**Azure Cognitive Account**](/infra-catalog/azure-cognitive-account) -- the Azure OpenAI models the projects call

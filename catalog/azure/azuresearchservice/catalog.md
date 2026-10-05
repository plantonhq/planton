# Azure AI Search Service

Creates an Azure AI Search service -- the managed search-and-retrieval engine AI applications index and query their own data with (keyword, vector, and semantic search; the standard retrieval companion to Azure OpenAI). Capacity is SKU × partitions × replicas -- partitions scale storage and indexing, replicas scale query throughput and availability -- and the SKU upgrades in place only along basic → standard → standard2 → standard3; every other SKU change replaces the service.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Search Service** -- the service itself (`Microsoft.Search/searchServices`) with its SKU, capacity, auth posture, network controls, and identity
- **Shared Private Links** (optional) -- one ARM child per `sharedPrivateLinkServices` entry (`.../sharedPrivateLinkResources/{name}`), giving indexers private reach to data sources behind private endpoints

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureResourceGroup** -- the service lives in one.

### Azure Subscription

- **The name is GLOBALLY unique** -- it forms the endpoint `{name}.search.windows.net`; a name-taken error means a genuine global collision.
- **Higher tiers need quota approval** -- `standard2`, `standard3`, and both storage-optimized SKUs require a quota increase request to Microsoft before ARM accepts them.
- **One free service per subscription** -- the `free` SKU is a shared-cluster evaluation tier.

## Deploy

### Console

Open the deployment store, find **Azure AI Search Service**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production Search Service** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureSearchService
metadata:
  name: acme-search
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: ai-platform-rg
  name: acme-search-prod
  sku: standard
  replicaCount: 3
  identity:
    type: SYSTEM_ASSIGNED
```

```shell
planton apply -f azure-search-service.yaml
```

This creates a standard-tier search service named `acme-search-prod` with three replicas (the 99.9% read-write SLA threshold) and a system-assigned identity for keyless indexer access to data sources. An Infra Job tracks the provisioning in real time.

### InfraChart

In an AI-platform chart the order is: resource group → **search service** alongside the Azure OpenAI account; applications consume `endpoint` plus a key (or Entra RBAC), and indexers reach data sources via the service identity or shared private links:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: ai-platform-rg
      fieldPath: status.outputs.resource_group_name
  sharedPrivateLinkServices:
    - name: to-blob
      subresourceName: blob
      targetResourceId:
        valueFrom:
          kind: AzureStorageAccount
          name: corpus-store
          fieldPath: status.outputs.storage_account_id
      requestMessage: "Search indexers need private reach to the corpus"
```

The InfraPipeline resolves the dependency graph, deploying the resource group and storage account before the service and its private link.

## Key Configuration

These are the most important decisions when configuring the service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Capacity** -- SKU × partitions × replicas. Partitions scale storage/indexing, replicas scale queries and availability (3+ replicas = 99.9% read-write SLA). Counts resize in place; the SKU upgrades in place ONLY along basic → standard → standard2 → standard3 (anything else replaces the service).

**Auth posture** -- default is API keys. Set `authenticationFailureMode` to add Entra RBAC alongside keys; set `localAuthenticationEnabled: false` for RBAC-only (keys stop working, and the failure mode must stay unset -- validated at manifest time).

**Semantic ranking** -- `semanticSearchSku: free` for evaluation (1000 requests/month), `standard` for production RAG; not available on the free service SKU.

**Shared private links** -- each entry sits "Pending" until the target resource's owner approves the connection; creation itself never needs the target side's consent.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `identity.identityIds[]` | `status.outputs.identity_id` |
| *(any private-linkable kind)* | `sharedPrivateLinkServices[].targetResourceId` | named explicitly in `valueFrom` (storage, SQL, Key Vault, ...) |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `search_service_id` | ARM ID of the service | Scope for Azure Role Assignments granting data-plane RBAC; diagnostic-setting targets |
| `endpoint` | `https://{name}.search.windows.net` | What applications and SDKs call |
| `primary_key` / `secondary_key` | Admin API keys (sensitive) | Application config via managed secrets; empty when local auth is disabled |
| `default_query_key` | The built-in read-only query key (sensitive) | Client-side query access |
| `system_assigned_identity_principal_id` | The system identity's principal ID | Azure Role Assignment `principalId` -- data-source grants for indexers |

The `search_service_name`, `customer_managed_key_encryption_compliance_status`, and name-keyed `shared_private_link_service_ids` outputs are readbacks for operational and governance tooling rather than values other Infra Components typically wire in.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production search** -- standard SKU, SLA replicas, RBAC alongside keys. Start from the **Production Search Service** preset.

**Development** -- the cheap basic tier. Start from the **Development Basic Search** preset.

**RAG retrieval** -- semantic ranking for Azure OpenAI grounding. Start from the **Semantic RAG Search** preset.

## Works With

- [**Azure Cognitive Account**](/infra-catalog/azure-cognitive-account) -- the Azure OpenAI models RAG applications pair with this service
- [**Azure AI Foundry Hub**](/infra-catalog/azure-ai-foundry) -- Foundry projects consume search for agent retrieval
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the classic indexer data source (and shared-private-link target)
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where the service lives

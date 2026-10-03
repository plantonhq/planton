# Azure Cognitive Account

Deploys an Azure AI services account -- the container Azure AI capabilities are provisioned and billed through: Azure OpenAI model deployments, the multi-service AI Services account behind AI Foundry, and the single-service accounts (Speech, Vision, Language, Content Safety, ...). The account owns the endpoint, the access keys, the network perimeter, and the responsible-AI policy; model deployments and AI Foundry projects are separate Infra Components created onto it. Deletion is a soft delete: the ghost keeps holding the account name until purged.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **AI services account** -- the endpoint, keys, kind (OpenAI / AIServices / single-service), SKU, identity, network perimeter, and encryption settings
- **Responsible-AI blocklists** (optional) -- one ARM child per `raiBlocklists` entry, keyed by name (the `rai_blocklist_ids` output republishes each ARM ID)
- **Responsible-AI policies** (optional) -- one ARM child per `raiPolicies` entry, keyed by name (the `rai_policy_ids` output republishes each ARM ID); model deployments select a policy via `raiPolicyName`

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Model and SKU availability differs per region** -- Azure OpenAI models especially. Pick the region where the models you plan to deploy exist.
- **A deleted account is a soft-deleted ghost** that keeps holding the account name until purged; recreating under the same name fails until then (the module purges on destroy by default).

## Deploy

### Console

Open the deployment store, find **Azure Cognitive Account**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Azure OpenAI Account** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCognitiveAccount
metadata:
  name: openai-prod
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: ai-rg
      fieldPath: status.outputs.resource_group_name
  name: acme-openai-prod
  kind: OpenAI
  skuName: S0
  customSubdomainName: acme-openai-prod
```

```shell
planton apply -f azure-cognitive-account.yaml
```

This creates an S0 Azure OpenAI account with a custom subdomain, ready for model deployments -- the account object itself carries no idle cost for OpenAI (billing follows deployment usage). An Infra Job tracks the provisioning in real time.

### InfraChart

In an AI-platform chart the order is: resource group → **account** → model deployments and projects. Wire the resource group with ValueFromRef so the InfraPipeline deploys it first:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: ai-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, provisions the resource group first, then the account -- and the deployments and projects that reference this account's `cognitive_account_id` deploy after it.

## Key Configuration

These are the most important decisions when configuring the account. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Kind** -- the central choice. `OpenAI` for Azure OpenAI model deployments; `AIServices` for the multi-service account that also backs AI Foundry projects and agents; the rest are single-service accounts. Changing kind replaces the account except between `OpenAI` and `AIServices` (in-place upgrade).

**Custom subdomain** -- required before network ACLs work and before Entra ID (token) authentication. Set it at creation: it can be added later to an account without one, but changing an existing value replaces the account.

**Auth posture** -- `localAuthEnabled: false` disables the access keys and forces Entra ID tokens (the hardened posture; the key outputs are then empty).

**Responsible AI** -- define `raiPolicies` on the account and point each model deployment's `raiPolicyName` at one; custom word/pattern lists ride `raiBlocklists`. Severity-less binary filters (Jailbreak, Indirect Attack, Protected Material Text/Code) deploy through the Terraform provisioner only -- the classic Pulumi SDK requires a severity on every filter while Azure rejects one on the binary names.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (optional) | `networkAcls.virtualNetworkRules[].subnetId`, `networkInjection.subnetId` | `status.outputs.subnet_id` |
| **AzureKeyVaultKey** (optional) | `customerManagedKey.keyVaultKeyId` | `status.outputs.versionless_id` |
| **AzureStorageAccount** (optional) | `storage[].storageAccountId` | `status.outputs.storage_account_id` |
| **AzureUserAssignedIdentity** (optional) | `identity.identityIds` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cognitive_account_id` | ARM ID of the account | A model deployment's or project's `cognitiveAccountId` |
| `cognitive_account_name` | Name of the account | Operational tooling |
| `endpoint` | The endpoint URL applications call | Application configuration |
| `primary_access_key` | Access key (secret; empty when local auth is off) | Application secrets |
| `secondary_access_key` | Rotation key (secret) | Zero-downtime key rotation |
| `system_assigned_identity_principal_id` | The system identity's principal ID | Key Vault / storage role grants |
| `rai_blocklist_ids` | ARM ID of each blocklist, keyed by name | Policy automation (`status.outputs.rai_blocklist_ids.<name>`) |
| `rai_policy_ids` | ARM ID of each policy, keyed by name | Policy automation (`status.outputs.rai_policy_ids.<name>`) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Azure OpenAI account** -- an `OpenAI`-kind S0 account with a custom subdomain, ready for model deployments. Start from the **Azure OpenAI Account** preset.

**AI Foundry account** -- an `AIServices`-kind account with project management and a system identity, ready for projects and agents. Start from the **AI Foundry Account** preset.

**Locked-down account** -- deny-by-default network ACLs, Entra-ID-only auth, restricted outbound. Start from the **Private Hardened Account** preset.

## Works With

- [**Azure Cognitive Deployment**](/infra-catalog/azure-cognitive-deployment) -- the model deployments applications call
- [**Azure Cognitive Account Project**](/infra-catalog/azure-cognitive-account-project) -- AI Foundry team workspaces on the account
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- customer-managed encryption
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- network rules and agent injection

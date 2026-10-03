# Azure Cognitive Deployment

Deploys a model onto an Azure AI services account -- which actual model applications call (gpt-4o, text-embedding-3-large, ...), at which throughput class and capacity. The deployment's NAME is the model parameter applications pass to the account's endpoint. The spec carries the model (format, name, optional pinned version), the SKU (throughput class and capacity), the version-upgrade policy, and the responsible-AI policy selection.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Model deployment** -- an ARM child of the account (`.../accounts/{account}/deployments/{name}`): the model (format/name/version), the SKU (throughput class + capacity), the version-upgrade policy, and the responsible-AI policy selection

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureCognitiveAccount** of kind `OpenAI` or `AIServices` -- the deployment is created onto it.

### Azure Subscription

- **Model availability is regional** -- the ACCOUNT's region decides which models can deploy (GlobalStandard reaches the widest set).
- **Capacity draws from per-subscription, per-region quota** -- a rejected create with a quota error is a quota request away, not a module defect.

## Deploy

### Console

Open the deployment store, find **Azure Cognitive Deployment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **GPT-4o Chat** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCognitiveDeployment
metadata:
  name: chat
  org: acme-corp
  env: prod
spec:
  cognitiveAccountId:
    valueFrom:
      kind: AzureCognitiveAccount
      name: openai-prod
      fieldPath: status.outputs.cognitive_account_id
  name: chat
  model:
    format: OpenAI
    name: gpt-5.4-mini
  sku:
    name: GlobalStandard
    capacity: 50
```

```shell
planton apply -f azure-cognitive-deployment.yaml
```

This creates a pay-per-token GlobalStandard deployment of a mini-class chat model, rate-limited at 50K tokens per minute, callable as `chat` on the referenced account's endpoint. An Infra Job tracks the provisioning in real time.

### InfraChart

In an AI-platform chart the order is: account → **deployments** (one per model); applications consume the account's `endpoint` output plus each deployment's name. Wire the account with ValueFromRef so the InfraPipeline deploys it first:

```yaml
spec:
  cognitiveAccountId:
    valueFrom:
      kind: AzureCognitiveAccount
      name: openai-prod
      fieldPath: status.outputs.cognitive_account_id
```

The InfraPipeline resolves the dependency graph, provisions the account first, then creates each deployment with the resolved account ARM ID.

## Key Configuration

These are the most important decisions when configuring the deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU name** -- the billing model. `Standard`/`GlobalStandard`/`DataZoneStandard` bill per token (capacity is a rate limit, no idle cost); the `...Batch` variants serve asynchronous batch jobs; the `...ProvisionedManaged` variants bill reserved PTU capacity CONTINUOUSLY while the deployment exists -- the expensive class.

**Capacity** -- thousands of tokens-per-minute (pay-per-token) or PTUs (provisioned). Updates in place; this is the scale knob.

**Version policy** -- leave `model.version` unset to track Azure's default, or pin it with `versionUpgradeOption: NO_AUTO_UPGRADE` when compliance demands a frozen model.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureCognitiveAccount** | `cognitiveAccountId` | `status.outputs.cognitive_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `deployment_id` | ARM ID of the deployment | Operational tooling |
| `deployment_name` | The deployment's name | The model parameter applications pass to the endpoint |
| `model_version` | The deployed version ARM resolved | Version pinning and audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Chat (mini model)** -- a current mini-class chat model on GlobalStandard with modest capacity. Start from the **Chat (Mini Model)** preset. Model names age: Azure rejects catalog-"Deprecating" models for new deployments, so check `az cognitiveservices model list` for what is currently GenerallyAvailable.

**Embeddings** -- text-embedding-3-large behind a rate limit. Start from the **Text Embeddings** preset.

## Works With

- [**Azure Cognitive Account**](/infra-catalog/azure-cognitive-account) -- the parent account (endpoint, keys, perimeter)
- [**Azure Cognitive Account Project**](/infra-catalog/azure-cognitive-account-project) -- AI Foundry workspaces on the same account

# AzureCognitiveDeployment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureCognitiveDeploymentSpec** defines a model deployment on an
Azure AI services account (ARM:
Microsoft.CognitiveServices/accounts/{account}/deployments/{name})
-- which actual model an application calls: "gpt-4o" version X on
deployment "chat", "text-embedding-3-large" on deployment
"embeddings". The account (AzureCognitiveAccount, kind "OpenAI" or
"AIServices") owns the endpoint and keys; the deployment decides
the model, the throughput class (SKU), and the capacity.

**Billing follows the SKU**: "Standard" and the Global/DataZone
variants bill per token (capacity is a rate limit in thousands of
tokens-per-minute, not idle cost); the ProvisionedManaged variants
bill per PTU capacity CONTINUOUSLY while the deployment exists --
the expensive class.

**The deployment is an ARM child of its account** -- it has no
region, resource group, or tags of its own (ARM derives all three
through the account; the provider's schema carries none).

**ForceNew fields**: `name`, `cognitive_account_id`, the model's
`format` and `name`, and the sku's `name`/`tier`/`size`/`family`.
Model `version`, sku `capacity`, rai_policy_name,
version_upgrade_option and dynamic throttling update in place.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCognitiveDeployment
metadata:
  name: test-cognitive-deployment
spec:
  cognitiveAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.CognitiveServices/accounts/acme-openai-test
  name: chat
  model:
    format: OpenAI
    # A GenerallyAvailable model from the live catalog -- ARM rejects
    # "Deprecating"-lifecycle models for new deployments
    # (ServiceModelDeprecating), so examples must age with the catalog.
    name: gpt-5.4-mini
  sku:
    name: GlobalStandard
    capacity: 10
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cognitiveAccountId` | `string \| valueFrom` | yes |  | AzureCognitiveAccount (`status.outputs.cognitive_account_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.model` | `AzureCognitiveDeploymentModel` | yes |  |  |
| `spec.model.format` | `string` | yes |  |  |
| `spec.model.name` | `string` | yes |  |  |
| `spec.model.version` | `string` |  |  |  |
| `spec.sku` | `AzureCognitiveDeploymentSku` | yes |  |  |
| `spec.sku.name` | `string` | yes |  |  |
| `spec.sku.tier` | `enum` |  |  |  |
| `spec.sku.size` | `string` |  |  |  |
| `spec.sku.family` | `string` |  |  |  |
| `spec.sku.capacity` | `int32` |  | `1` |  |
| `spec.raiPolicyName` | `string` |  |  |  |
| `spec.versionUpgradeOption` | `enum` |  |  |  |
| `spec.dynamicThrottlingEnabled` | `bool` |  |  |  |

## Field Details

### spec.cognitiveAccountId

`string | valueFrom` · required

The Azure AI services account the model deploys onto, by ARM ID.
The account's kind must support model deployments ("OpenAI" or
"AIServices"). Fixed at creation.

- references: AzureCognitiveAccount (`status.outputs.cognitive_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureCognitiveAccount, name: <that resource's name>, fieldPath: status.outputs.cognitive_account_id}} -- a bare string does not parse

### spec.name

`string` · required

The deployment's name, unique on the account -- what applications
pass as the model/deployment parameter when calling the endpoint
("chat", "embeddings"). Changing the name replaces the
deployment.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.model

`AzureCognitiveDeploymentModel` · required

The model being deployed.

- rule: {"required":true}

### spec.model.format

`string` · required

The model's format -- the publisher namespace ARM catalogs the
model under. "OpenAI" for every Azure OpenAI model; partner
models carry their publisher's format string.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.model.name

`string` · required

The model's catalog name, e.g. "gpt-5.4-mini",
"text-embedding-3-large". Regional availability differs -- the
account's region decides which models can deploy. Models AGE:
ARM rejects a model whose catalog lifecycle is "Deprecating" for
NEW deployments (error ServiceModelDeprecating) well before its
final retirement date -- existing deployments keep running. Pick
from the currently GenerallyAvailable catalog:
`az cognitiveservices model list -l <region>
 --query "[?model.lifecycleStatus=='GenerallyAvailable'].model.name"`.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.model.version

`string`

The model version, e.g. "2024-08-06". Leave unset to deploy the
model's current default version (the value is read back from
ARM). Updating the version upgrades the deployment in place.

### spec.sku

`AzureCognitiveDeploymentSku` · required

The deployment's throughput class and capacity.

- rule: {"required":true}

### spec.sku.name

`string` · required

The throughput class (the wire values). "Standard" (regional
pay-per-token), "GlobalStandard" (Azure-wide routing, the widest
model availability), "DataZoneStandard" (EU/US data-zone
routing), the "...Batch" variants (asynchronous batch
inference), and the "...ProvisionedManaged" variants (reserved
PTU capacity -- bills continuously while the deployment exists).
Fixed at creation.

- rule: {"required":true,"string":{"in":["Standard","DataZoneBatch","DataZoneProvisionedManaged","DataZoneStandard","GlobalBatch","GlobalProvisionedManaged","GlobalStandard","ProvisionedManaged"]}}

### spec.sku.tier

`enum`

The SKU tier. Rarely needed -- ARM derives it from the SKU name.
Fixed at creation.

Allowed values (use exactly as shown):

- `azure_cognitive_deployment_sku_tier_unspecified` -- Not specified: the property is omitted and ARM derives the tier from the SKU name.
- `FREE` -- Wire value "Free".
- `BASIC` -- Wire value "Basic".
- `STANDARD` -- Wire value "Standard".
- `PREMIUM` -- Wire value "Premium".
- `ENTERPRISE` -- Wire value "Enterprise".

### spec.sku.size

`string`

The SKU size code. Rarely needed. Fixed at creation.

### spec.sku.family

`string`

The SKU family code. Rarely needed. Fixed at creation.

### spec.sku.capacity

`int32` · optional (explicit presence)

The capacity: thousands of tokens-per-minute for the
pay-per-token SKUs (a rate limit, not idle cost), or PTUs for
the ProvisionedManaged SKUs (billed while the deployment
exists). Unspecified applies the provider default of 1. Updates
in place -- this is the scale knob.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.raiPolicyName

`string`

The responsible-AI (content-filter) policy applied to this
deployment, by policy NAME on the same account (a rai_policies
entry of the AzureCognitiveAccount spec, or a policy created in
the portal). Leave unset for Azure's default policy -- the
service assigns one, so the value is read back from ARM rather
than defaulted here.

### spec.versionUpgradeOption

`enum`

What happens when Azure retires or supersedes the deployed model
version. Unspecified applies the provider default,
ONCE_NEW_DEFAULT_VERSION_AVAILABLE.

Allowed values (use exactly as shown):

- `azure_cognitive_deployment_version_upgrade_option_unspecified` -- Not specified: the provider applies "OnceNewDefaultVersionAvailable".
- `ONCE_CURRENT_VERSION_EXPIRED` -- Upgrade when the deployed version expires (wire value "OnceCurrentVersionExpired").
- `ONCE_NEW_DEFAULT_VERSION_AVAILABLE` -- Upgrade whenever a new default version ships (wire value "OnceNewDefaultVersionAvailable") -- the provider's default.
- `NO_AUTO_UPGRADE` -- Never auto-upgrade; version changes are yours to make (wire value "NoAutoUpgrade"). The pinned-model compliance posture.

### spec.dynamicThrottlingEnabled

`bool`

Let the service raise the deployment's rate limits
opportunistically (dynamic throttling).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureCognitiveDeployment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.deployment_id` | `string` | The Azure Resource Manager ID of the deployment. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.CognitiveServices/accounts/{account}/deployments/{name} |
| `status.outputs.deployment_name` | `string` | The deployment's name -- what applications pass as the model/deployment parameter when calling the account's endpoint. |
| `status.outputs.model_version` | `string` | The deployed model's version as ARM reports it -- the resolved value when the spec left version unset. Engine nuance (live-proven): with version unset, the Pulumi engine populates this at create while Terraform leaves it EMPTY until the next refresh or import (the v5 provider does not read the resolved version back at create). Pin model.version when automation needs this output deterministically on both engines. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cognitiveAccountId` | AzureCognitiveAccount | `status.outputs.cognitive_account_id` |

## See Also

- [Overview](../README.md)

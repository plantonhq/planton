# AzureCognitiveAccount

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureCognitiveAccountSpec** defines an Azure AI services account
(ARM: Microsoft.CognitiveServices/accounts) -- the container every
Azure AI capability is provisioned and billed through: Azure OpenAI
model deployments, the multi-service AI Services account that backs
AI Foundry projects, and the single-service accounts (Speech,
Vision, Language, Content Safety, ...). The account owns the
endpoint, the access keys, the network perimeter, and the
responsible-AI (content-filter) policy; model deployments
(AzureCognitiveDeployment) and projects
(AzureCognitiveAccountProject) are created onto it.

**The kind decides everything else.** "OpenAI" hosts Azure OpenAI
model deployments; "AIServices" is the multi-service account (and
the ONLY kind that can enable project management or agent network
injection -- it is also where users of the retired
azurerm_ai_services resource land); the remaining values are the
single-service accounts. Several fields are only meaningful -- or
only legal -- on specific kinds; those contracts are enforced here
so the error lands in seconds, not at deploy time.

**ForceNew fields**: `name`, `region`, `resource_group`, and the
four metrics_advisor_* fields. Changing `kind` also replaces the
account UNLESS the change is between "OpenAI" and "AIServices"
(Azure upgrades those in place). `custom_subdomain_name` can be SET
once on an account that never had one, but changing an existing
value replaces the account.

**Deletion is a soft delete.** A deleted account becomes a
purgeable ghost that keeps holding the account name (the Key Vault
recycle-bin pattern); recreating under the same name fails until
the ghost is purged.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCognitiveAccount
metadata:
  name: test-cognitive-account
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: acme-openai-test
  kind: OpenAI
  skuName: S0
  customSubdomainName: acme-openai-test
  raiBlocklists:
    - name: blocked-terms
      description: Custom blocked words for chat deployments.
  raiPolicies:
    - name: strict-chat
      basePolicyName: Microsoft.Default
      contentFilters:
        - name: Hate
          filterEnabled: true
          blockEnabled: true
          source: Prompt
          severityThreshold: LOW
        # Graded filters only: severity-less filters (the binary
        # Jailbreak/Indirect Attack/Protected Material filters) deploy
        # through Terraform ONLY -- PARITY-EXCEPTION on the spec's
        # severity_threshold field. The canonical example stays
        # engine-portable.
        - name: Violence
          filterEnabled: true
          blockEnabled: true
          source: Completion
          severityThreshold: MEDIUM
      mode: BLOCKING
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.kind` | `string` | yes |  |  |
| `spec.skuName` | `string` | yes |  |  |
| `spec.projectManagementEnabled` | `bool` |  |  |  |
| `spec.customSubdomainName` | `string` |  |  |  |
| `spec.customerManagedKey` | `AzureCognitiveAccountCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.identityClientId` | `string` |  |  |  |
| `spec.dynamicThrottlingEnabled` | `bool` |  |  |  |
| `spec.fqdns` | `[]string` |  |  |  |
| `spec.identity` | `AzureCognitiveAccountIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.localAuthEnabled` | `bool` |  | `true` |  |
| `spec.metricsAdvisorAadClientId` | `string` |  |  |  |
| `spec.metricsAdvisorAadTenantId` | `string` |  |  |  |
| `spec.metricsAdvisorSuperUserName` | `string` |  |  |  |
| `spec.metricsAdvisorWebsiteName` | `string` |  |  |  |
| `spec.networkAcls` | `AzureCognitiveAccountNetworkAcls` |  |  |  |
| `spec.networkAcls.defaultAction` | `string` | yes |  |  |
| `spec.networkAcls.ipRules` | `[]string` |  |  |  |
| `spec.networkAcls.virtualNetworkRules` | `[]AzureCognitiveAccountVirtualNetworkRule` |  |  |  |
| `spec.networkAcls.virtualNetworkRules[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.networkAcls.virtualNetworkRules[].ignoreMissingVnetServiceEndpoint` | `bool` |  |  |  |
| `spec.networkAcls.bypass` | `enum` |  |  |  |
| `spec.networkInjection` | `AzureCognitiveAccountNetworkInjection` |  |  |  |
| `spec.networkInjection.scenario` | `string` | yes |  |  |
| `spec.networkInjection.subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.outboundNetworkAccessRestricted` | `bool` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.qnaRuntimeEndpoint` | `string` |  |  |  |
| `spec.customQuestionAnsweringSearchServiceId` | `string` |  |  |  |
| `spec.customQuestionAnsweringSearchServiceKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.storage` | `[]AzureCognitiveAccountStorage` |  |  |  |
| `spec.storage[].storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.storage[].identityClientId` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.raiBlocklists` | `[]AzureCognitiveAccountRaiBlocklist` |  |  |  |
| `spec.raiBlocklists[].name` | `string` | yes |  |  |
| `spec.raiBlocklists[].description` | `string` |  |  |  |
| `spec.raiBlocklists[].tags` | `map<string, string>` |  |  |  |
| `spec.raiPolicies` | `[]AzureCognitiveAccountRaiPolicy` |  |  |  |
| `spec.raiPolicies[].name` | `string` | yes |  |  |
| `spec.raiPolicies[].basePolicyName` | `string` | yes |  |  |
| `spec.raiPolicies[].contentFilters` | `[]AzureCognitiveAccountRaiPolicyContentFilter` | yes |  |  |
| `spec.raiPolicies[].contentFilters[].name` | `string` | yes |  |  |
| `spec.raiPolicies[].contentFilters[].filterEnabled` | `bool` |  |  |  |
| `spec.raiPolicies[].contentFilters[].blockEnabled` | `bool` |  |  |  |
| `spec.raiPolicies[].contentFilters[].source` | `string` | yes |  |  |
| `spec.raiPolicies[].contentFilters[].severityThreshold` | `enum` |  |  |  |
| `spec.raiPolicies[].mode` | `enum` |  |  |  |
| `spec.raiPolicies[].tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the account lives in, e.g. "eastus". Model and
SKU availability differ per region (Azure OpenAI models
especially). Changing the region replaces the account.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the account is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The account's name, unique within the resource group: starts with
an alphanumeric character, then alphanumerics, periods, dashes or
underscores (the provider's own rule). Changing the name replaces
the account -- and the OLD name stays held by the soft-deleted
ghost until purged.

- rule: {"required":true,"string":{"minLen":"2","pattern":"^[a-zA-Z0-9][a-zA-Z0-9_.-]+$"}}

### spec.kind

`string` · required

Which AI service the account provides (the ARM account kind --
the wire values). The two flagship choices: "OpenAI" (Azure
OpenAI model deployments) and "AIServices" (the multi-service
account that also backs AI Foundry projects and agents). The rest
are single-service accounts. Changing kind replaces the account
unless the change is between "OpenAI" and "AIServices" (in-place
upgrade -- how an OpenAI account grows into the full AI Services
surface).

- rule: {"required":true,"string":{"in":["AIServices","Academic","AnomalyDetector","Bing.Autosuggest","Bing.Autosuggest.v7","Bing.CustomSearch","Bing.Search","Bing.Search.v7","Bing.Speech","Bing.SpellCheck","Bing.SpellCheck.v7","CognitiveServices","ComputerVision","ContentModerator","ConversationalLanguageUnderstanding","ContentSafety","CustomSpeech","CustomVision.Prediction","CustomVision.Training","Emotion","Face","FormRecognizer","ImmersiveReader","LUIS","LUIS.Authoring","MetricsAdvisor","OpenAI","Personalizer","QnAMaker","Recommendations","SpeakerRecognition","Speech","SpeechServices","SpeechTranslation","TextAnalytics","TextTranslation","WebLM"]}}

### spec.skuName

`string` · required

The account's pricing tier (the wire values). "S0" is the
standard paid tier for OpenAI and AIServices accounts; "F0" is
the free tier where the service offers one. Which SKUs a kind
accepts (and in which regions) is decided by ARM at deploy time.

- rule: {"required":true,"string":{"in":["C2","C3","C4","D3","DC0","E0","F0","F1","P0","P1","P2","S","S0","S1","S2","S3","S4","S5","S6"]}}

### spec.projectManagementEnabled

`bool`

Allow AzureCognitiveAccountProject resources (AI Foundry
projects) to be created on this account. Only legal on the
"AIServices" kind, and the account must carry a managed identity.
Disabling it later on an account that is not "OpenAI"-kind
replaces the account.

### spec.customSubdomainName

`string`

A custom subdomain for the account's endpoint
("{value}.cognitiveservices.azure.com"). Required before network
ACLs can be configured, and required for Entra ID (token-based)
authentication. Can be SET once on an account created without
one; CHANGING an existing value replaces the account.

### spec.customerManagedKey

`AzureCognitiveAccountCustomerManagedKey`

Encrypt the account's data with your own Key Vault key instead of
Microsoft-managed keys. The account needs a managed identity with
wrap/unwrap access on the key BEFORE this is configured.

### spec.customerManagedKey.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key (data-plane URL, e.g.
"https://{vault}.vault.azure.net/keys/{name}"). Reference an
AzureKeyVaultKey's versionless_id so rotation propagates without
intervention; pin a versioned URL only under a compliance regime
that demands a frozen version.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.customerManagedKey.identityClientId

`string`

The client ID of the USER-ASSIGNED identity that unwraps the key.
Leave unset to use the account's system-assigned identity.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

### spec.dynamicThrottlingEnabled

`bool`

Let the service raise the account's default rate limits
opportunistically (dynamic throttling). Not supported on the
"OpenAI" and "AIServices" kinds (their capacity is managed per
deployment instead).

### spec.fqdns

`[]string`

Outbound destinations the account may call when
outbound_network_access_restricted is true -- an FQDN allowlist
(e.g. a search service the account grounds on).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.identity

`AzureCognitiveAccountIdentity`

The account's managed identity. Required for project management
(AI Foundry), customer-managed keys, and user-owned storage
access via identity.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the account; USER_ASSIGNED brings identities you manage
(grantable Key Vault / storage access BEFORE the account exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_cognitive_account_identity_type_unspecified` -- Not specified: the account has no managed identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the account.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the account, by ARM ID. Reference
AzureUserAssignedIdentity resources so Key Vault / storage grants
can be composed before the account is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.localAuthEnabled

`bool` · optional (explicit presence)

Whether the account's access KEYS work. Unspecified applies true
(ARM's default). Set false to force Entra ID (token) auth only --
the recommended hardened posture; note the account's key outputs
are then empty.

- default: `true`

### spec.metricsAdvisorAadClientId

`string`

Metrics Advisor only: the Entra ID application (client) ID bound
to the account. Fixed at creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

### spec.metricsAdvisorAadTenantId

`string`

Metrics Advisor only: the Entra ID tenant ID. Fixed at creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

### spec.metricsAdvisorSuperUserName

`string`

Metrics Advisor only: the super-user name. Fixed at creation.

### spec.metricsAdvisorWebsiteName

`string`

Metrics Advisor only: the website name. Fixed at creation.

### spec.networkAcls

`AzureCognitiveAccountNetworkAcls`

The account's network perimeter: default action, allowed IPs /
CIDRs, allowed VNet subnets, and the trusted-services bypass.
Requires custom_subdomain_name (the provider's own contract --
network rules only work against the custom endpoint).

### spec.networkAcls.defaultAction

`string` · required

What happens to traffic matching no rule: "Allow" or "Deny" (the
wire values). "Deny" plus ip_rules / virtual_network_rules is the
locked-down posture.

- rule: {"required":true,"string":{"in":["Allow","Deny"]}}

### spec.networkAcls.ipRules

`[]string`

Public IPv4 addresses or CIDR ranges allowed through the
perimeter.

- rule: {"repeated":{"items":{"string":{"pattern":"^(\\d{1,3}\\.){3}\\d{1,3}(/\\d{1,2})?$"}}}}

### spec.networkAcls.virtualNetworkRules

`[]AzureCognitiveAccountVirtualNetworkRule`

VNet subnets allowed through the perimeter (service endpoints).

### spec.networkAcls.virtualNetworkRules[].subnetId

`string | valueFrom` · required

The subnet's ARM ID. The subnet needs the
"Microsoft.CognitiveServices" service endpoint enabled (or set
ignore_missing_vnet_service_endpoint).
On a diagram this is a firewall rule -- the account admits traffic
from the subnet and lives in its resource group -- so the
reference is access, not placement (every PaaS kind's
network-rule subnets carry the same mark).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.networkAcls.virtualNetworkRules[].ignoreMissingVnetServiceEndpoint

`bool`

Accept the rule even if the subnet has no CognitiveServices
service endpoint yet (it will not take effect until one exists).

### spec.networkAcls.bypass

`enum`

Let trusted Azure services through a "Deny" default action.
Unspecified omits the property (ARM's default). Only legal on the
"OpenAI", "AIServices" and "TextAnalytics" kinds.

Allowed values (use exactly as shown):

- `azure_cognitive_account_network_acls_bypass_unspecified` -- Not specified: the property is omitted and ARM applies its default.
- `AZURE_SERVICES` -- Trusted Azure services may bypass the perimeter (wire value "AzureServices").
- `NONE` -- Nothing bypasses the perimeter (wire value "None").

### spec.networkInjection

`AzureCognitiveAccountNetworkInjection`

Inject the account's agent workloads into your own VNet subnet
(the AI Foundry Agent Service network-injection scenario). Only
legal on the "AIServices" kind.

### spec.networkInjection.scenario

`string` · required

The injection scenario. Azure currently defines exactly one:
"agent" (the wire value) -- AI Foundry Agent Service workloads.

- rule: {"required":true,"string":{"in":["agent"]}}

### spec.networkInjection.subnetId

`string | valueFrom` · required

The delegated subnet agents are injected into. NOTE (deletion
behavior): Azure removes the subnet's service association link
asynchronously after the account deletes -- the subnet cannot be
deleted until that completes (the module waits).
On a diagram the account lives in its resource group and its
agents' compute attaches to this subnet, so the reference is
access, not placement -- the account is never drawn inside the
subnet it injects into (the Azure-SSIS runtime's rule).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.outboundNetworkAccessRestricted

`bool`

Restrict the account's OUTBOUND calls to the fqdns allowlist
(data-loss prevention). ARM's default is unrestricted.

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the account's endpoint answers the public internet.
Unspecified applies true (ARM's default). Set false to make the
account reachable only through private endpoints.

- default: `true`

### spec.qnaRuntimeEndpoint

`string`

QnAMaker only: the QnA runtime endpoint URL (required for that
kind).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.customQuestionAnsweringSearchServiceId

`string`

TextAnalytics (custom question answering) only: the ARM ID of the
Azure AI Search service backing custom question answering.
Plain string this wave; becomes a typed AzureSearchService
reference when that kind registers (recorded in-place upgrade).

### spec.customQuestionAnsweringSearchServiceKey

`string | valueFrom` · sensitive

TextAnalytics (custom question answering) only: the Search
service's API key. Reference a secret rather than embedding the
literal in manifests.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.storage

`[]AzureCognitiveAccountStorage`

Bring-your-own-storage: the storage accounts the service stores
customer data in (Speech, Custom Vision and friends).

### spec.storage[].storageAccountId

`string | valueFrom` · required

The storage account's ARM ID.

The account reads training data and writes results here and lives in
its own resource group, so on a diagram the reference is access, not
placement.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.storage[].identityClientId

`string`

The client ID of the USER-ASSIGNED identity that accesses the
storage account. Leave unset to use the system-assigned identity.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uuid":true}}

### spec.tags

`map<string, string>`

Free-form tags applied to the account, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

### spec.raiBlocklists

`[]AzureCognitiveAccountRaiBlocklist`

Responsible-AI blocklists on the account: named containers for
custom blocked words/patterns that rai_policies can reference.
Each deploys as its own ARM child; ids surface name-keyed in the
rai_blocklist_ids output. (Blocklist ITEMS are managed through
the service's data plane, not ARM.)

### spec.raiBlocklists[].name

`string` · required

The blocklist's name, unique on the account: 2-64 alphanumerics,
hyphens or underscores (the provider's own rule). The blocklist's
ARM id surfaces in the rai_blocklist_ids output under this name.
Changing the name replaces the blocklist.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_-]{2,64}$"}}

### spec.raiBlocklists[].description

`string`

What the blocklist is for.

### spec.raiBlocklists[].tags

`map<string, string>`

Free-form tags on the blocklist object.

### spec.raiPolicies

`[]AzureCognitiveAccountRaiPolicy`

Responsible-AI (content-filter) policies on the account:
named filter configurations that model deployments select via
their rai_policy_name. Each deploys as its own ARM child; ids
surface name-keyed in the rai_policy_ids output.

- rule: severity_threshold is not applicable to the 'Jailbreak', 'Indirect Attack', 'Protected Material Text' and 'Protected Material Code' filters -- they are binary

### spec.raiPolicies[].name

`string` · required

The policy's name, unique on the account. The policy's ARM id
surfaces in the rai_policy_ids output under this name. Changing
the name replaces the policy.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.raiPolicies[].basePolicyName

`string` · required

The built-in policy this one extends -- "Microsoft.Default" (the
standard filters) or "Microsoft.DefaultV2". Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.raiPolicies[].contentFilters

`[]AzureCognitiveAccountRaiPolicyContentFilter` · required

The content filters the policy applies -- one entry per filter
category and direction (e.g. "Hate" on source "Prompt", "Violence"
on source "Completion", "Jailbreak" on "Prompt").

- rule: {"repeated":{"minItems":"1"}}

### spec.raiPolicies[].contentFilters[].name

`string` · required

The filter's name -- a category Azure defines: the severity-based
filters ("Hate", "Sexual", "Violence", "Selfharm") and the binary
filters ("Jailbreak", "Indirect Attack", "Protected Material
Text", "Protected Material Code", "Profanity", or a custom
blocklist's name).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.raiPolicies[].contentFilters[].filterEnabled

`bool`

Whether the filter evaluates content.

### spec.raiPolicies[].contentFilters[].blockEnabled

`bool`

Whether matching content is BLOCKED (versus only annotated).

### spec.raiPolicies[].contentFilters[].source

`string` · required

Which content the filter applies to (the wire values): "Prompt"
(what users send) or "Completion" (what the model returns), plus
the agent-scoped sources "PreRun"/"PostRun" and
"PreToolCall"/"PostToolCall".

- rule: {"required":true,"string":{"in":["Completion","PostRun","PostToolCall","PreRun","PreToolCall","Prompt"]}}

### spec.raiPolicies[].contentFilters[].severityThreshold

`enum`

For the severity-based filters: the lowest severity that
triggers the filter. Not applicable to the binary filters
(Jailbreak, Indirect Attack, Protected Material Text/Code).

PARITY-EXCEPTION: severity-less filters deploy through Terraform
ONLY. Terraform (azurerm v5) made this property optional and
REJECTS it on the binary filters; the classic Pulumi SDK bridges
the pre-v5 provider, which requires a severity on EVERY content
filter -- so a filter without one (any binary filter, or a graded
filter relying on ARM's default) is inexpressible on the Pulumi
engine and its module fails loudly. Unblock: a v5-bridged
pulumi-azure major.

Allowed values (use exactly as shown):

- `azure_cognitive_account_rai_policy_content_level_unspecified` -- Not specified: the property is omitted.
- `LOW` -- Wire value "Low".
- `MEDIUM` -- Wire value "Medium".
- `HIGH` -- Wire value "High".

### spec.raiPolicies[].mode

`enum`

How filtering is applied. Unspecified omits the property (ARM
applies its default).

Allowed values (use exactly as shown):

- `azure_cognitive_account_rai_policy_mode_unspecified` -- Not specified: the property is omitted and ARM applies its default.
- `DEFAULT` -- Wire value "Default".
- `BLOCKING` -- Filters run before the response streams (wire value "Blocking").
- `ASYNCHRONOUS_FILTER` -- Filters run asynchronously alongside streaming (wire value "AsynchronousFilter") -- lower latency, content may briefly stream before annotation.
- `DEFERRED` -- Wire value "Deferred".

### spec.raiPolicies[].tags

`map<string, string>`

Free-form tags on the policy object.

## Validation Rules

- `project_management_requires_aiservices`: project_management_enabled is only supported when kind is 'AIServices'
- `project_management_requires_identity`: project_management_enabled requires a managed identity -- configure the identity block
- `dynamic_throttling_unsupported_on_openai_kinds`: dynamic_throttling_enabled is not supported when kind is 'OpenAI' or 'AIServices'
- `network_acls_require_custom_subdomain`: network_acls requires custom_subdomain_name -- network rules only apply to the account's custom endpoint
- `network_acls_bypass_kind_gate`: network_acls.bypass can only be set when kind is 'OpenAI', 'AIServices' or 'TextAnalytics'
- `network_injection_requires_aiservices`: network_injection is only supported when kind is 'AIServices'
- `qna_maker_requires_runtime_endpoint`: kind 'QnAMaker' requires qna_runtime_endpoint
- `qna_runtime_endpoint_only_for_qna_maker`: qna_runtime_endpoint is only used when kind is 'QnAMaker' -- it would be silently ignored on this kind
- `qna_search_service_only_for_text_analytics`: custom_question_answering_search_service_id/key can only be set when kind is 'TextAnalytics'
- `metrics_advisor_fields_only_for_metrics_advisor`: the metrics_advisor_* fields can only be set when kind is 'MetricsAdvisor'
- `rai_blocklist_names_unique`: rai_blocklists names must be unique on the account -- each is the key the rai_blocklist_ids output uses
- `rai_policy_names_unique`: rai_policies names must be unique on the account -- each is the key the rai_policy_ids output uses

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureCognitiveAccount, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cognitive_account_id` | `string` | The Azure Resource Manager ID of the account -- what model deployments and projects reference as their cognitive_account_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.CognitiveServices/accounts/{name} |
| `status.outputs.cognitive_account_name` | `string` | The name of the account. ARM addresses deployments, projects and responsible-AI children as children of this name. |
| `status.outputs.endpoint` | `string` | The account's endpoint URL -- what applications call (with a key or an Entra ID token). With a custom_subdomain_name this is "https://{subdomain}.cognitiveservices.azure.com/"; without one, the regional shared endpoint. |
| `status.outputs.primary_access_key` | `string` | The account's primary access key. Marked sensitive in both engines. Empty when local_auth_enabled is false (token auth only). |
| `status.outputs.secondary_access_key` | `string` | The account's secondary access key (for zero-downtime rotation). Marked sensitive in both engines. Empty when local_auth_enabled is false. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the account's system-assigned identity, when one is enabled -- what Key Vault and storage grants bind to. |
| `status.outputs.rai_blocklist_ids` | `map<string, string>` | The ARM ID of each responsible-AI blocklist on the account, keyed by the blocklist's name from the spec. Example valueFrom fieldPath: status.outputs.rai_blocklist_ids.competitor-names |
| `status.outputs.rai_policy_ids` | `map<string, string>` | The ARM ID of each responsible-AI policy on the account, keyed by the policy's name from the spec. Example valueFrom fieldPath: status.outputs.rai_policy_ids.strict-chat |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.networkAcls.virtualNetworkRules[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.networkInjection.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.storage[].storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureCognitiveAccountProject | `spec.cognitiveAccountId` | `status.outputs.cognitive_account_id` |
| AzureCognitiveDeployment | `spec.cognitiveAccountId` | `status.outputs.cognitive_account_id` |

## See Also

- [Overview](../README.md)

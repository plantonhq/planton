# AzureLogAnalyticsWorkspace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureLogAnalyticsWorkspaceSpec** defines the configuration for creating an
Azure Log Analytics Workspace.

Log Analytics Workspaces are the central data platform for Azure Monitor. They
collect, store, and analyze log and performance data from Azure resources,
on-premises servers, and third-party services. Workspaces are the foundation
for monitoring in Azure -- they power Container Insights (AKS), Application
Insights, Microsoft Sentinel (SIEM), diagnostic settings, and log-query alert
rules. Most Azure estates run a small number of workspaces (often one per
environment or per region) that many resources feed into, so the workspace is
a long-lived governance boundary: retention, ingestion quotas, network access
posture, and authentication mode are all set here.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the
# CapacityReservation sku with its commitment tier, the security and
# network posture flags at non-default values, a user-assigned identity,
# a default DCR, and user tags.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: test-law
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  workspaceName: test-workspace
  sku: CAPACITY_RESERVATION
  reservationCapacityInGbPerDay: 100
  retentionInDays: 90
  dailyQuotaGb: 50
  identity:
    type: USER_ASSIGNED
    userAssignedIdentityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  localAuthenticationEnabled: false
  internetIngestionEnabled: false
  internetQueryEnabled: false
  allowResourceOnlyPermissions: false
  cmkForQueryForced: true
  immediateDataPurgeOn30DaysEnabled: true
  dataCollectionRuleId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Insights/dataCollectionRules/test-dcr
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.workspaceName` | `string` | yes |  |  |
| `spec.sku` | `enum` |  |  |  |
| `spec.reservationCapacityInGbPerDay` | `int32` |  |  |  |
| `spec.retentionInDays` | `int32` |  | `30` |  |
| `spec.dailyQuotaGb` | `double` |  | `-1` |  |
| `spec.identity` | `AzureLogAnalyticsWorkspaceIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.localAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.internetIngestionEnabled` | `bool` |  | `true` |  |
| `spec.internetQueryEnabled` | `bool` |  | `true` |  |
| `spec.allowResourceOnlyPermissions` | `bool` |  | `true` |  |
| `spec.cmkForQueryForced` | `bool` |  |  |  |
| `spec.immediateDataPurgeOn30DaysEnabled` | `bool` |  |  |  |
| `spec.dataCollectionRuleId` | `string \| valueFrom` |  |  | AzureMonitorDataCollectionRule (`status.outputs.data_collection_rule_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Log Analytics Workspace will be deployed.
Examples: "eastus", "westus2", "westeurope", "southeastasia".
Choose a region close to the resources that will send logs to this workspace
to minimize data egress costs and latency. Cross-region ingestion works but
bills inter-region bandwidth and adds latency.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the workspace will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.workspaceName

`string` · required

The name of the Log Analytics Workspace. Unique within the resource group.
4-63 characters: letters, digits, and hyphens; must start and end with a
letter or digit.

**ForceNew**: Changing this destroys and recreates the workspace -- and with
it all ingested data. Treat the name as permanent.

- rule: workspace name must be 4-63 characters of letters, digits, or hyphens, and must start and end with a letter or digit
- rule: {"required":true,"string":{"minLen":"4","maxLen":"63"}}

### spec.sku

`enum`

The pricing tier (SKU) of the workspace. Unspecified deploys PER_GB_2018 --
Azure's pay-as-you-go model and the right choice for nearly every workspace.
Choose CAPACITY_RESERVATION only at sustained high ingestion (100+ GB/day),
where commitment tiers discount ingestion; it requires
reservation_capacity_in_gb_per_day and carries Azure's 31-day commitment
period on the chosen tier. PER_NODE and STANDALONE are legacy pre-2018
pricing models kept for existing estates -- do not choose them for new
workspaces.

Switching between PER_GB_2018 and CAPACITY_RESERVATION updates in place;
any other SKU change forces the workspace to be destroyed and recreated
(Azure's own transition rule).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_log_analytics_workspace_sku_unspecified` -- Not specified -- deploys Azure's recommended PER_GB_2018 pay-as-you-go tier.
- `PER_GB_2018` -- Pay-as-you-go per GB ingested -- Azure's recommended tier and the right choice for nearly every workspace.
- `CAPACITY_RESERVATION` -- Commitment tier with discounted ingestion. Requires reservation_capacity_in_gb_per_day; worthwhile at sustained 100+ GB/day.
- `PER_NODE` -- Legacy pre-2018 per-node (OMS) pricing -- kept for existing estates only.
- `STANDALONE` -- Legacy pre-2018 standalone per-GB pricing -- kept for existing estates only.

### spec.reservationCapacityInGbPerDay

`int32` · optional (explicit presence)

The commitment-tier capacity in GB of ingestion per day. Required with --
and only meaningful with -- the CAPACITY_RESERVATION sku. Azure only sells
fixed tiers: 100, 200, 300, 400, 500, 1000, 2000, 5000, 10000, 25000, or
50000 GB/day. Ingestion beyond the reserved capacity bills at the
pay-as-you-go rate, so pick the tier just below the observed daily volume.
Raising the tier restarts Azure's 31-day commitment period.

- rule: {"int32":{"in":[100,200,300,400,500,1000,2000,5000,10000,25000,50000]}}

### spec.retentionInDays

`int32` · optional (explicit presence)

The number of days to retain data in the workspace at the workspace level.
Range: 30 to 730 days. PER_GB_2018 includes 31 days free (90 when Microsoft
Sentinel is enabled); beyond that, retention bills per GB per month. For
compliance workloads 90-365 days is typical. Per-table retention overrides
are managed in Azure directly and are not part of this workspace-level dial.
Default: 30

- default: `30`
- rule: {"int32":{"lte":730,"gte":30}}

### spec.dailyQuotaGb

`double` · optional (explicit presence)

The daily ingestion quota in GB.
Set to -1 for unlimited ingestion (no cap).
Set to a positive value to cap daily ingestion and prevent cost overruns --
when the cap is reached, ingestion stops until the next UTC day, so a cap on
a production workspace is also a data-loss dial. Use it deliberately.
Default: -1 (unlimited)

- default: `-1`
- rule: {"double":{"gte":-1}}

### spec.identity

`AzureLogAnalyticsWorkspaceIdentity`

The workspace's managed identity, used when the workspace itself needs to
access other Azure resources -- most notably reading a customer-managed key
when the workspace is attached to a dedicated Log Analytics cluster, or
running queries against linked storage.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a service
principal bound to the workspace's lifecycle) or USER_ASSIGNED (bring
identities from user_assigned_identity_ids, shareable across resources
and grantable before the workspace exists).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_log_analytics_workspace_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the workspace's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity entries -- shareable across resources and grantable before the workspace exists.

### spec.identity.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities to attach -- required when (and only
meaningful when) type is USER_ASSIGNED. Each entry references an
AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.localAuthenticationEnabled

`bool` · optional (explicit presence)

Whether local authentication methods (workspace shared keys) are allowed in
addition to Microsoft Entra ID. Azure's default is true. Set false for a
keyless posture: agents and ingestion APIs must then authenticate with Entra
identities, and the shared-key outputs stop being usable credentials.
Default: true (Azure's own default; flip deliberately -- legacy agents and
some ingestion paths still require shared keys)

- default: `true`

### spec.internetIngestionEnabled

`bool` · optional (explicit presence)

Whether the workspace accepts log ingestion over the public internet.
Set false to force ingestion through Azure Monitor Private Link Scope
(AMPLS) private endpoints only.
Default: true (Azure's default)

- default: `true`

### spec.internetQueryEnabled

`bool` · optional (explicit presence)

Whether the workspace serves log queries over the public internet.
Set false to force queries through Azure Monitor Private Link Scope (AMPLS)
private endpoints only -- the portal's Logs blade then also requires private
connectivity.
Default: true (Azure's default)

- default: `true`

### spec.allowResourceOnlyPermissions

`bool` · optional (explicit presence)

Whether users can query only the data of resources they have Azure RBAC
permission to view (resource-context access), without needing permission on
the workspace itself. This is Azure's modern access model and its default.
Set false to require explicit workspace-level permissions for every query
(workspace-context access) -- the stricter, centralized model some security
teams mandate.
Default: true (Azure's default)

- default: `true`

### spec.cmkForQueryForced

`bool`

Whether customer-managed storage is mandatory for query artifacts (saved
queries and query results). Only relevant for regulated estates that pair
the workspace with linked storage accounts so query artifacts never rest on
Microsoft-managed storage. Leave false otherwise.

### spec.immediateDataPurgeOn30DaysEnabled

`bool`

Whether data is purged immediately after 30 days instead of being retained
in Azure's grace store. Relevant for strict data-residency/right-to-erasure
compliance: normally Azure keeps data recoverable for a short period beyond
the retention window; this switch removes that grace period.

### spec.dataCollectionRuleId

`string | valueFrom`

The Azure Monitor Data Collection Rule to set as the workspace's
default DCR (applied to data flowing in without an explicit rule), by
ARM resource ID. Can be a literal ARM ID or a reference to an
AzureMonitorDataCollectionRule output. Leave unset for Azure's default
handling.

- references: AzureMonitorDataCollectionRule (`status.outputs.data_collection_rule_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMonitorDataCollectionRule, name: <that resource's name>, fieldPath: status.outputs.data_collection_rule_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Tags to apply to the workspace, merged over the Planton-derived
metadata tags (user values win on key conflicts). ARM tags are Azure's
first-class governance surface -- Azure Policy enforces them and
Microsoft Cost Management groups by them.

## Validation Rules

- `log_analytics_workspace_capacity_requires_reservation_sku`: reservation_capacity_in_gb_per_day can only be set with the CAPACITY_RESERVATION sku -- on pay-as-you-go and legacy skus Azure rejects a commitment-tier capacity
- `log_analytics_workspace_reservation_sku_requires_capacity`: the CAPACITY_RESERVATION sku requires reservation_capacity_in_gb_per_day -- pick the commitment tier just below the workspace's observed daily ingestion

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.workspace_id` | `string` | The Azure Resource Manager ID of the Log Analytics Workspace. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.OperationalInsights/workspaces/{name} This is what downstream resources reference: Application Insights workspace binding, AKS Container Insights, Container App Environment log destinations, diagnostic settings, and log-query alert scopes. |
| `status.outputs.workspace_name` | `string` | The name of the Log Analytics Workspace. |
| `status.outputs.workspace_customer_id` | `string` | The workspace customer ID -- the GUID agents and direct-ingestion APIs identify the workspace by (the portal calls it "Workspace ID" on the agents page; distinct from the ARM resource ID above). |
| `status.outputs.resource_group_name` | `string` | The name of the resource group containing the workspace. |
| `status.outputs.primary_shared_key` | `string` (sensitive) | The primary shared key for agent authentication. Secret-bearing: used by Azure Monitor agents and direct ingestion APIs to authenticate when sending data. Unusable as a credential when local_authentication_enabled is false (keyless posture). |
| `status.outputs.secondary_shared_key` | `string` (sensitive) | The secondary shared key for agent authentication. Secret-bearing: a backup key that allows rotation without downtime -- point agents at the secondary, regenerate the primary, then swap back. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the workspace's system-assigned managed identity. Empty unless the identity block enables SYSTEM_ASSIGNED. Grant this principal access (for example to a Key Vault key) when the workspace itself must read other resources. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.dataCollectionRuleId` | AzureMonitorDataCollectionRule | `status.outputs.data_collection_rule_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAksCluster | `spec.omsAgent.logAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| AzureAksCluster | `spec.microsoftDefender.logAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| AzureApplicationInsights | `spec.workspaceId` | `status.outputs.workspace_id` |
| AzureContainerAppEnvironment | `spec.logAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| AzureContainerInstance | `spec.diagnosticsLogAnalytics.workspaceId` | `status.outputs.workspace_customer_id` |
| AzureContainerInstance | `spec.diagnosticsLogAnalytics.workspaceKey` | `status.outputs.primary_shared_key` |
| AzureFirewallPolicy | `spec.insights.defaultLogAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| AzureFirewallPolicy | `spec.insights.logAnalyticsWorkspaces[].workspaceId` | `status.outputs.workspace_id` |
| AzureMonitorDataCollectionRule | `spec.destinations.logAnalytics[].workspaceResourceId` | `status.outputs.workspace_id` |
| AzureMonitorDiagnosticSetting | `spec.logAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| AzureMonitorScheduledQueryAlert | `spec.scope` | `status.outputs.workspace_id` |
| AzureNetworkWatcherFlowLog | `spec.trafficAnalytics.workspaceId` | `status.outputs.workspace_customer_id` |
| AzureNetworkWatcherFlowLog | `spec.trafficAnalytics.workspaceResourceId` | `status.outputs.workspace_id` |

## See Also

- [Overview](../README.md)

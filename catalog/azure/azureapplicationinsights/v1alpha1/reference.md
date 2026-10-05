# AzureApplicationInsights

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureApplicationInsightsSpec** defines the configuration for creating an
Azure Application Insights resource.

Application Insights is Azure's Application Performance Management (APM)
service. It provides deep observability into web applications and services --
tracking request rates, response times, failure rates, dependency calls,
exceptions, page views, and custom telemetry. It is the standard APM layer in
Azure, consumed by Function Apps, Web Apps, Container Apps, and any
application instrumented with the Application Insights SDK or OpenTelemetry.

This kind models workspace-based Application Insights only: telemetry is
stored in a Log Analytics Workspace the component references. Classic
(non-workspace) Application Insights was retired by Azure in February 2024,
so the workspace binding is required here even though the underlying API
still tolerates its absence for migrated legacy resources. Once set, the
workspace binding can be moved to another workspace but never removed.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the NODE_JS
# application type (an irregular-casing wire value, "Node.JS"), the
# privacy/auth/network posture flags at non-default values, the daily
# cap dials, and user tags.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureApplicationInsights
metadata:
  name: test-appinsights
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  applicationInsightsName: test-appinsights
  applicationType: NODE_JS
  workspaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.OperationalInsights/workspaces/test-law
  retentionInDays: 180
  dailyDataCapInGb: 10
  dailyDataCapNotificationsEnabled: false
  samplingPercentage: 50
  ipMaskingEnabled: false
  localAuthenticationEnabled: false
  internetIngestionEnabled: false
  internetQueryEnabled: false
  forceCustomerStorageForProfiler: true
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.applicationInsightsName` | `string` | yes |  |  |
| `spec.applicationType` | `enum` |  |  |  |
| `spec.workspaceId` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`) |
| `spec.retentionInDays` | `int32` |  | `90` |  |
| `spec.dailyDataCapInGb` | `double` |  | `100` |  |
| `spec.dailyDataCapNotificationsEnabled` | `bool` |  | `true` |  |
| `spec.samplingPercentage` | `double` |  | `100` |  |
| `spec.ipMaskingEnabled` | `bool` |  | `true` |  |
| `spec.localAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.internetIngestionEnabled` | `bool` |  | `true` |  |
| `spec.internetQueryEnabled` | `bool` |  | `true` |  |
| `spec.forceCustomerStorageForProfiler` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where Application Insights will be deployed.
Examples: "eastus", "westus2", "westeurope", "southeastasia".
Should match the region of the application(s) being monitored to minimize
telemetry ingestion latency and data residency concerns.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where Application Insights will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.applicationInsightsName

`string` · required

The name of the Application Insights resource. Unique within the resource
group. Length: 1 to 260 characters.

**ForceNew**: Changing this destroys and recreates the resource -- and the
instrumentation key and connection string with it. Treat the name as
permanent once applications are wired to the resource.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"260"}}

### spec.applicationType

`enum`

The type of application being monitored. This sets the resource's kind and
shapes which experiences the portal surfaces (server-side APM views,
mobile analytics, etc.). Unspecified deploys WEB -- the right choice for
web applications and services of any language, and the value the vast
majority of resources use. OpenTelemetry-instrumented workloads should
also use WEB.

**ForceNew**: Changing this destroys and recreates the resource.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_application_insights_application_type_unspecified` -- Not specified -- deploys WEB, the right choice for web applications and services of any language.
- `WEB` -- Web applications and services (ASP.NET, Spring Boot, Django, Express, and anything OpenTelemetry-instrumented).
- `JAVA` -- Java applications (standalone, not web).
- `NODE_JS` -- Node.js applications.
- `OTHER` -- General/other application types.
- `IOS` -- iOS applications.
- `PHONE` -- Windows Phone applications (legacy).
- `STORE` -- Windows Store applications (legacy).
- `MOBILE_CENTER` -- App Center-managed applications.

### spec.workspaceId

`string | valueFrom` · required

The Log Analytics Workspace that stores this resource's telemetry.
Required -- workspace-based Application Insights is the only mode Azure
still supports for new resources. Can be a literal workspace ARM ID or a
reference to an AzureLogAnalyticsWorkspace output.

Once set, the binding can be repointed to a different workspace (telemetry
flows to the new workspace from that moment; history stays in the old one)
but can never be removed.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.retentionInDays

`int32` · optional (explicit presence)

The number of days to retain telemetry data.
Azure only allows specific values: 30, 60, 90, 120, 180, 270, 365, 550, 730.
For workspace-based resources the effective retention is governed by the
workspace; this value applies to the classic-experience tables and is kept
aligned with the workspace's retention in most estates.
Default: 90

- default: `90`
- rule: {"int32":{"in":[30,60,90,120,180,270,365,550,730]}}

### spec.dailyDataCapInGb

`double` · optional (explicit presence)

The daily telemetry data cap in GB.
When the cap is reached, telemetry ingestion stops until the next UTC day --
a cost guard that is also a data-loss dial; size it above normal daily
volume in production.
Default: 100 (Azure's default)

- default: `100`
- rule: {"double":{"gte":0}}

### spec.dailyDataCapNotificationsEnabled

`bool` · optional (explicit presence)

Whether a notification email is sent when the daily data cap is reached.
Default: true (be told when telemetry starts being dropped)

- default: `true`

### spec.samplingPercentage

`double` · optional (explicit presence)

The percentage of telemetry data to sample (0-100).
Reducing sampling percentage lowers data volume and cost while still
providing statistically representative telemetry. Common production values
are 25-50%. Set to 100 for full fidelity (all telemetry collected).
Default: 100

- default: `100`
- rule: {"double":{"lte":100,"gte":0}}

### spec.ipMaskingEnabled

`bool` · optional (explicit presence)

Whether client IP addresses are masked to 0.0.0.0 in stored telemetry.
Azure's default is to mask (GDPR-friendly); set false to store real client
IPs -- only do this deliberately, with privacy review, when geo/diagnostic
fidelity genuinely requires it.
Default: true (mask client IPs)

- default: `true`

### spec.localAuthenticationEnabled

`bool` · optional (explicit presence)

Whether local authentication (instrumentation-key-only ingestion) is
allowed in addition to Microsoft Entra ID. Set false for a keyless
posture: SDKs must then authenticate with Entra identities and a bare
instrumentation key no longer authorizes ingestion.
Default: true (Azure's default; most SDK setups still ingest by key)

- default: `true`

### spec.internetIngestionEnabled

`bool` · optional (explicit presence)

Whether the resource accepts telemetry ingestion over the public internet.
Set false to force ingestion through Azure Monitor Private Link Scope
(AMPLS) private endpoints only.
Default: true (Azure's default)

- default: `true`

### spec.internetQueryEnabled

`bool` · optional (explicit presence)

Whether the resource serves telemetry queries over the public internet.
Set false to force queries through Azure Monitor Private Link Scope
(AMPLS) private endpoints only.
Default: true (Azure's default)

- default: `true`

### spec.forceCustomerStorageForProfiler

`bool`

Whether the .NET Profiler and Snapshot Debugger must write to a
customer-owned storage account (BYO storage) instead of Microsoft-managed
storage. Only relevant for regulated estates with data-residency
requirements on profiler artifacts; the storage account is linked to the
resource outside this spec.

### spec.tags

`map<string, string>`

Tags to apply to the resource, merged over the Planton-derived
metadata tags (user values win on key conflicts). ARM tags are Azure's
first-class governance surface -- Azure Policy enforces them and
Microsoft Cost Management groups by them.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureApplicationInsights, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.application_insights_id` | `string` | The Azure Resource Manager ID of the Application Insights resource. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Insights/components/{name} Referenced by metric alerts (web-test availability criteria) and diagnostic settings targeting this resource. |
| `status.outputs.application_insights_name` | `string` | The name of the Application Insights resource. |
| `status.outputs.instrumentation_key` | `string` | The instrumentation key for classic SDK configuration. Secret-bearing: it authorizes telemetry ingestion while local authentication is enabled. Microsoft recommends the connection_string for new applications; the key remains for SDKs that have not migrated. |
| `status.outputs.connection_string` | `string` | The connection string for SDK configuration. Secret-bearing: contains the instrumentation key plus the ingestion endpoint in a single string -- the recommended way to configure Application Insights and OpenTelemetry SDKs. Referenced by: AzureFunctionApp, AzureLinuxWebApp, and AzureContainerAppEnvironment (Dapr telemetry). |
| `status.outputs.app_id` | `string` | The Application ID for API access -- used when programmatically querying telemetry via the Application Insights REST API. Not a secret (queries additionally require an API key or Entra token). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.workspaceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAiFoundry | `spec.applicationInsightsId` | `status.outputs.application_insights_id` |
| AzureApplicationInsightsStandardWebTest | `spec.applicationInsightsId` | `status.outputs.application_insights_id` |
| AzureFunctionApp | `spec.applicationInsightsConnectionString` | `status.outputs.connection_string` |
| AzureFunctionAppFlexConsumption | `spec.applicationInsightsConnectionString` | `status.outputs.connection_string` |
| AzureLinuxWebApp | `spec.applicationInsightsConnectionString` | `status.outputs.connection_string` |
| AzureMachineLearningWorkspace | `spec.applicationInsightsId` | `status.outputs.application_insights_id` |
| AzureMonitorMetricAlert | `spec.webTestAvailabilityCriteria.componentId` | `status.outputs.application_insights_id` |
| AzureMonitorScheduledQueryAlert | `spec.scope` | `status.outputs.application_insights_id` |

## See Also

- [Overview](../README.md)

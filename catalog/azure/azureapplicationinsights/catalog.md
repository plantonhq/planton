# Azure Application Insights

Deploys an Azure Application Insights resource backed by a Log Analytics Workspace, with configurable application type, telemetry sampling, data retention, and daily ingestion caps. Classic non-workspace mode is not supported, and the privacy and access dials -- IP masking, Entra-only ingestion, private-link-only ingestion and query -- are all part of the spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Application Insights Resource** -- a workspace-based Application Insights instance in the specified Azure region and resource group, configured with the chosen application type, retention period, daily data cap, and sampling percentage
- **Log Analytics Integration** -- telemetry data is stored in the referenced Log Analytics Workspace (classic non-workspace mode is not supported)
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where Application Insights will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A Log Analytics Workspace** to store telemetry data. Workspace-based mode is required; classic mode is deprecated. Provide the workspace resource ID directly or reference an AzureLogAnalyticsWorkspace Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Application Insights**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Web Application Insights** preset in the [Presets](#presets) tab to pre-populate a full-fidelity monitoring configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureApplicationInsights
metadata:
  name: platform-apm
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  applicationInsightsName: acme-app-insights
  workspaceId:
    valueFrom:
      name: acme-logs
```

```shell
planton apply -f app-insights.yaml
```

This creates an Application Insights resource on Azure's defaults -- the WEB application type, 100% sampling, 90-day retention, and a 100 GB daily cap -- storing telemetry in the referenced workspace. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire Application Insights to a resource group and Log Analytics Workspace deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  workspaceId:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-logs
      fieldPath: status.outputs.workspace_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and Log Analytics Workspace first, then provisions Application Insights with the resolved values.

## Key Configuration

These are the most important decisions when configuring Application Insights. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Application type** -- The `applicationType` field categorizes the monitored workload. Unspecified deploys `WEB` -- right for web applications and services of any language, OpenTelemetry included. The closed vocabulary also offers `JAVA`, `NODE_JS`, `OTHER`, `IOS`, `PHONE`, `STORE`, and `MOBILE_CENTER` (values map to Azure's exact case-sensitive API strings in the IaC modules). This is a ForceNew field -- changing it after creation replaces the resource and its credentials.

**Sampling percentage** -- The `samplingPercentage` field controls what fraction of telemetry is collected (0-100). Full sampling (100%) gives complete fidelity but costs more at scale. Production workloads with high traffic commonly use 25-50% to reduce volume while maintaining statistical accuracy.

**Daily data cap** -- The `dailyDataCapInGb` field sets a hard limit on daily ingestion. When reached, telemetry collection stops until the next UTC day. Use this in development or staging to prevent cost surprises from logging storms. Defaults to 100 GB.

**Retention period** -- The `retentionInDays` field controls how long telemetry is queryable. Azure allows specific values (30, 60, 90, 120, 180, 270, 365, 550, 730). The free tier includes 90 days; longer retention incurs per-GB monthly charges. For workspace-based resources the workspace's own retention governs the workspace tables.

**Privacy and access posture** -- Five optional dials, all defaulting to Azure's open posture: `ipMaskingEnabled` (client IPs masked to 0.0.0.0 -- the GDPR-friendly default), `localAuthenticationEnabled` (false = Entra-only ingestion; bare instrumentation keys stop authorizing), `internetIngestionEnabled` and `internetQueryEnabled` (false = the corresponding path requires Azure Monitor Private Link Scope private endpoints), and `forceCustomerStorageForProfiler` (BYO storage for .NET Profiler and Snapshot Debugger artifacts).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureLogAnalyticsWorkspace** | `workspaceId` | `status.outputs.workspace_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `application_insights_id` | Azure Resource Manager ID of the Application Insights resource | Availability web tests, diagnostic settings, Azure Policy assignments |
| `application_insights_name` | Name of the Application Insights resource | Azure CLI references, portal navigation |
| `instrumentation_key` | Classic SDK instrumentation key (secret; inert when local authentication is disabled) | Legacy SDK configuration (connection string preferred) |
| `connection_string` | SDK connection string with ingestion endpoint (secret) | Application environment variables for Function Apps, Web Apps, Container Apps |
| `app_id` | Application ID for REST API access | Programmatic telemetry queries via the Application Insights API |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard full-fidelity monitoring** -- 100% sampling, 90-day retention, and 100 GB daily cap. Suitable for development, staging, and moderate-traffic production workloads where complete telemetry visibility is needed. Start from the **Standard Web Application Insights** preset.

**Production with cost-controlled sampling** -- 25% sampling with a 10 GB daily ingestion cap. Reduces telemetry volume by 75% while maintaining statistically representative performance data. Designed for high-traffic production APIs where monitoring budget is constrained. Start from the **Production Application Insights (Sampled, Cost-Controlled)** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where Application Insights is created
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- provides the workspace that stores telemetry data
- [**Azure Application Insights Standard Web Test**](/infra-catalog/azure-application-insights-standard-web-test) -- availability probes that store their results in this resource via `application_insights_id`
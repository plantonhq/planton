# AzureApplicationInsights: Research & Design Documentation

## 1. What Is Azure Application Insights?

Azure Application Insights is Azure's Application Performance Management (APM) service, part of Azure Monitor. It provides deep observability into live applications by collecting telemetry data including:

- **Request telemetry**: HTTP requests, response codes, durations
- **Dependency tracking**: Calls to databases, HTTP APIs, Azure services
- **Exception telemetry**: Unhandled and handled exceptions with stack traces
- **Performance counters**: CPU, memory, GC metrics
- **Custom events and metrics**: Application-specific telemetry
- **Page views**: Browser-side telemetry (for web apps)
- **Availability tests**: Synthetic monitoring from Azure data centers

Application Insights supports auto-instrumentation (zero-code) for .NET, Java, Node.js, and Python, plus manual instrumentation via SDKs or OpenTelemetry for any language.

### Workspace-Based vs Classic

Microsoft introduced workspace-based Application Insights in 2020. The key difference:

- **Classic**: Data stored in Application Insights' own storage. Limited query capabilities, no cross-resource correlation, deprecated.
- **Workspace-based**: Data stored in a Log Analytics Workspace. Full KQL query support, cross-resource queries, longer retention, unified billing.

**This component only supports workspace-based mode.** Classic mode is deprecated by Microsoft (announced 2024, retirement planned for 2025-2026) and should not be used for new deployments.

## 2. Deployment Landscape

### Level 0: Azure Portal

The portal provides a guided experience for creating Application Insights. It auto-suggests a workspace and offers application type selection.

### Level 1: Azure CLI

```bash
az monitor app-insights component create \
  --app my-app-insights \
  --location eastus \
  --resource-group my-rg \
  --application-type web \
  --workspace /subscriptions/.../workspaces/my-law \
  --retention-time 90
```

### Level 2: ARM Templates / Bicep

```bicep
resource appInsights 'Microsoft.Insights/components@2020-02-02' = {
  name: 'my-app-insights'
  location: 'eastus'
  kind: 'web'
  properties: {
    Application_Type: 'web'
    WorkspaceResourceId: law.id
    RetentionInDays: 90
    SamplingPercentage: 50
    IngestionMode: 'LogAnalytics'
  }
}
```

### Level 3: Terraform

```hcl
resource "azurerm_application_insights" "main" {
  name                = "my-app-insights"
  location            = "eastus"
  resource_group_name = "my-rg"
  application_type    = "web"
  workspace_id        = azurerm_log_analytics_workspace.main.id
  retention_in_days   = 90
  sampling_percentage = 50
  daily_data_cap_in_gb = 100
}
```

### Level 4: Pulumi

```go
insights, _ := appinsights.NewInsights(ctx, "my-ai",
  &appinsights.InsightsArgs{
    Name:               pulumi.String("my-app-insights"),
    Location:           pulumi.String("eastus"),
    ResourceGroupName:  pulumi.String("my-rg"),
    ApplicationType:    pulumi.String("web"),
    WorkspaceId:        pulumi.String(law.ID()),
    RetentionInDays:    pulumi.Int(90),
    SamplingPercentage: pulumi.Float64(50),
  })
```

### Level 5: Planton (This Component)

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureApplicationInsights
metadata:
  name: platform-ai
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: my-app-insights
  workspace_id:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-law
      fieldPath: status.outputs.workspace_id
  retention_in_days: 90
  sampling_percentage: 50
```

Planton adds composability (StringValueOrRef for resource_group and workspace_id) and consistency (same API pattern across all cloud providers) on top of Terraform/Pulumi.

## 3. Application Types

### What application_type Controls

The `application_type` field affects:
1. **Default dashboard tiles** in the Azure Portal
2. **Which Application Map features** are enabled
3. **Suggested diagnostic tools** in the UI

It does NOT affect the underlying telemetry collection or data format. The same data is collected regardless of application type.

### Available Types

| Value | Use Case |
|-------|----------|
| `web` | Web applications (ASP.NET, Spring Boot, Django, Express, Flask, Rails, etc.) |
| `java` | Standalone Java applications (not web) |
| `Node.JS` | Node.js applications |
| `other` | All other application types (Go, Python CLI, .NET console, etc.) |
| `MobileCenter` | Mobile applications (App Center integration) |
| `phone`, `store`, `ios` | Mobile-specific types (niche) |

### Why We Include 4 Types (80/20)

The four included types (`web`, `java`, `Node.JS`, `other`) cover 95%+ of server-side and web workloads. Mobile-centric types (`MobileCenter`, `phone`, `store`, `ios`) are niche and typically managed through Azure App Center rather than infrastructure-as-code.

### ForceNew Behavior

The `application_type` is marked `ForceNew` in Azure. Changing it after creation destroys and recreates the Application Insights resource. This is important because it means losing the existing instrumentation key and connection string. Applications must be updated with the new connection string.

## 4. Sampling Strategy

### How Sampling Works

Application Insights sampling controls the percentage of telemetry data that is collected and stored. It operates at the ingestion level:

- **100%**: All telemetry collected (full fidelity)
- **50%**: Half of telemetry collected (statistically representative)
- **25%**: Quarter of telemetry collected (good for high-volume apps)
- **0%**: No telemetry collected (effectively disabled)

### Why Sampling Matters

For high-traffic applications, 100% sampling can generate enormous data volumes:
- A web app handling 1000 req/s generates ~3 GB/day at 100% sampling
- At 25% sampling, that drops to ~750 MB/day
- Cost reduction is roughly proportional to sampling reduction

### Sampling Is Statistically Valid

Azure's adaptive sampling algorithm preserves statistical accuracy:
- Request counts are adjusted (100 sampled requests at 50% = 200 reported)
- Percentile calculations remain accurate
- Anomaly detection still works
- Correlation between requests and dependencies is preserved

### Recommended Values

| Environment | Sampling | Rationale |
|-------------|----------|-----------|
| Development | 100% | Full telemetry for debugging |
| Staging | 50% | Balance between cost and debugging |
| Production (low traffic) | 100% | Cost is manageable |
| Production (high traffic) | 25-50% | Cost control with representative data |

### Why We Default to 100%

Full sampling is the safest default because:
1. It ensures complete data during initial setup and debugging
2. Users can tune it down once they understand their data volume
3. Missing telemetry during troubleshooting is worse than higher cost

## 5. Data Retention

### Azure's Retention Model for Application Insights

Unlike Log Analytics Workspace (which accepts a range of 30-730), Application Insights only allows specific retention values:

| Days | Use Case |
|------|----------|
| 30 | Development/testing -- minimum cost |
| 60 | Short-term production monitoring |
| 90 | Standard production (default) -- covers most incident investigations |
| 120 | Extended production monitoring |
| 180 | Half-year retention for trend analysis |
| 270 | Three-quarter year for seasonal analysis |
| 365 | Annual retention for compliance |
| 550 | Extended compliance |
| 730 | Maximum retention for regulated industries |

### Why We Default to 90 Days

90 days is the Azure default and covers most production use cases:
1. Most incident investigations look at the past 1-4 weeks
2. Quarterly trend analysis needs ~90 days of data
3. Beyond 90 days, costs increase significantly per GB stored
4. Regulatory requirements vary -- users who need more will set it explicitly

## 6. Daily Data Cap

### How It Works

- `daily_data_cap_in_gb` sets a hard limit on daily telemetry ingestion
- When the cap is reached, data collection stops until midnight UTC
- Azure sends a notification email when the cap is approaching (configurable)
- The counter resets at midnight UTC daily

### Azure Default: 100 GB

Azure defaults to 100 GB/day, which is generous for most applications. Unlike Log Analytics Workspace (which defaults to -1/unlimited), Application Insights uses a positive default because application telemetry can be verbose (stack traces, request bodies, custom properties).

### When to Use Lower Caps

- **Development**: 0.5-1 GB/day (debug logging can spike unexpectedly)
- **Staging**: 5-10 GB/day (load tests can generate huge telemetry volumes)
- **Production**: 50-100 GB/day (depends on traffic volume and sampling)

### When NOT to Cap

- **Critical production applications**: Losing telemetry during an incident is unacceptable
- **Security monitoring**: Gaps in telemetry data create audit compliance risks

## 7. Planton Design: The 80/20 Applied

### Fields We Include (80% of Users Need These)

| Field | Justification |
|-------|---------------|
| `region` | Every resource needs a location |
| `resource_group` | Azure requirement, with StringValueOrRef for composability |
| `name` | Every resource needs a name |
| `application_type` | Controls portal experience, ForceNew -- must be set at creation |
| `workspace_id` | Required for workspace-based mode, with StringValueOrRef |
| `retention_in_days` | Primary cost and compliance lever |
| `daily_data_cap_in_gb` | Cost protection for all environments |
| `sampling_percentage` | The most impactful cost control lever for high-traffic apps |

### Fields We Exclude (20% or Less Need These)

| Azure Field | Why Excluded |
|-------------|-------------|
| `disable_ip_masking` | Privacy setting, rare to change from default |
| `local_authentication_disabled` | Advanced security, most users keep default |
| `internet_ingestion_enabled` | Rare, only for fully private setups |
| `internet_query_enabled` | Rare, only for fully private setups |
| `force_customer_storage_for_profiler` | Niche profiler configuration |
| `daily_data_cap_notifications_disabled` | Best left enabled (default) |

Any of these can be added in a future iteration if demand emerges.

## 8. Integration Points

### As a Monitoring Layer Resource

Application Insights sits at Layer 2 in Azure infra charts (above Resource Group at Layer 0 and Log Analytics Workspace at Layer 1). It is consumed by:

- **AzureFunctionApp**: References `connection_string` for APM integration
- **AzureLinuxWebApp**: References `connection_string` for APM integration
- **AzureContainerApp**: Uses `connection_string` as an environment variable

### Output Design

We export 4 outputs:

1. **app_insights_id**: The ARM resource ID. Used when other Azure resources need to reference this Application Insights instance.
2. **instrumentation_key**: The classic authentication key. Still supported but Microsoft recommends `connection_string` for new applications.
3. **connection_string**: The recommended SDK configuration value. Contains the instrumentation key, ingestion endpoint, and other configuration in a single string. This is the primary output consumed by downstream application resources.
4. **app_id**: The application ID for programmatic access via the Application Insights REST API.

### Sensitive Outputs

The `instrumentation_key` and `connection_string` are marked as sensitive in both the Pulumi module (automatic) and Terraform module (`sensitive = true`). They are authentication credentials that should not be exposed in logs or state files.

## 9. Scope Boundaries

### What This Component Does

- Creates an Azure Application Insights resource (workspace-based)
- Configures application type, retention, daily cap, and sampling
- Tags the resource with Planton metadata
- Exports resource ID, instrumentation key, connection string, and app ID

### What This Component Does NOT Do

- **Smart Detection rules**: Azure auto-creates "Failure Anomalies" detection. Custom smart detection rules are separate configuration.
- **Availability tests**: Synthetic monitoring (web tests) are separate resources.
- **Custom dashboards**: Azure Portal dashboards are presentation-layer constructs.
- **Alerts**: Metric and log alerts are separate resources applied to Application Insights data.
- **Application Map**: Automatically generated from telemetry, no configuration needed.
- **Profiler/Snapshot Debugger**: Advanced debugging features configured via the portal or ARM.
- **Continuous Export**: Data export to storage accounts is a separate configuration.
- **Work Item Integration**: DevOps integration is portal-level configuration.

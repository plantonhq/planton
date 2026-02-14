# Standard Application Insights

This preset creates an Azure Application Insights resource with full telemetry collection (100% sampling), 90-day retention, and 100 GB daily cap. This is the standard configuration for development and small-to-medium production workloads where full observability is needed and data volume is manageable.

## When to Use

- Web applications, APIs, and microservices that need APM (Application Performance Management)
- Development and staging environments where full telemetry fidelity is important for debugging
- Production workloads with moderate traffic where 100 GB/day is sufficient

## Key Configuration Choices

- **Application type** (`applicationType: web`) -- Optimized for web applications. Change to `java`, `Node.JS`, or `other` for non-web workloads
- **Full sampling** (`samplingPercentage: 100`) -- Collects all telemetry data. No data loss but higher cost at scale
- **90-day retention** (`retentionInDays: 90`) -- Free tier includes 90 days. Sufficient for most debugging and performance analysis
- **100 GB daily cap** (`dailyDataCapInGb: 100`) -- Azure default. Prevents runaway costs from telemetry spikes
- **Workspace-based** -- Requires a Log Analytics Workspace (classic mode is deprecated)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (match the application's region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-app-insights-name>` | Name for the Application Insights resource | Your naming convention |
| `<log-analytics-workspace-id>` | Full ARM resource ID of the Log Analytics Workspace | Azure portal or `AzureLogAnalyticsWorkspace` status outputs |

## Related Presets

- **02-production-sampled** -- Use instead for high-traffic production workloads with cost-controlled sampling

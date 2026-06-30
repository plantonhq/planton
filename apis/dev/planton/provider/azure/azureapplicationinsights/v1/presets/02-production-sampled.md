# Production Application Insights with Sampling

This preset creates an Azure Application Insights resource with 25% adaptive sampling and a 10 GB daily ingestion cap. This is the cost-optimized configuration for high-traffic production workloads where full telemetry fidelity is not needed and controlling monitoring costs is a priority.

## When to Use

- High-traffic production APIs and web applications generating large telemetry volumes
- Cost-conscious environments where monitoring budget is constrained
- Workloads where statistically representative sampling provides sufficient insight

## Key Configuration Choices

- **25% sampling** (`samplingPercentage: 25`) -- Collects 1 in 4 telemetry items. Reduces data volume by 75% while maintaining statistically representative performance data. Increase to 50% for more granularity
- **10 GB daily cap** (`dailyDataCapInGb: 10`) -- Hard limit that stops ingestion when reached. Prevents cost surprises from traffic spikes or logging storms
- **90-day retention** (`retentionInDays: 90`) -- Same as standard; retention cost is per-GB regardless of sampling

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (match the application's region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-app-insights-name>` | Name for the Application Insights resource | Your naming convention |
| `<log-analytics-workspace-id>` | Full ARM resource ID of the Log Analytics Workspace | Azure portal or `AzureLogAnalyticsWorkspace` status outputs |

## Related Presets

- **01-standard** -- Use instead for development or moderate-traffic workloads where full telemetry is needed

# AzureApplicationInsights Examples

## Minimal Configuration

The simplest possible Application Insights with all defaults -- web type, 90-day
retention, 100 GB daily cap, 100% sampling.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: my-ai
spec:
  region: eastus
  resource_group: my-resource-group
  name: my-app-insights
  workspace_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.OperationalInsights/workspaces/my-law
```

## Development Environment

A development instance with reduced sampling and low daily cap to minimize costs.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: dev-ai
  org: mycompany
  env: development
spec:
  region: eastus
  resource_group: dev-monitoring-rg
  name: dev-platform-ai
  application_type: web
  workspace_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.OperationalInsights/workspaces/dev-law
  retention_in_days: 30
  daily_data_cap_in_gb: 1.0
  sampling_percentage: 25
```

## Production Environment

A production instance with extended retention and moderate sampling for cost balance.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: prod-ai
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-monitoring-rg
  name: prod-platform-ai
  application_type: web
  workspace_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.OperationalInsights/workspaces/prod-law
  retention_in_days: 90
  daily_data_cap_in_gb: 100
  sampling_percentage: 50
```

## Java Application Monitoring

An Application Insights instance configured for Java applications.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: java-ai
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-apps-rg
  name: prod-java-api-ai
  application_type: java
  workspace_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.OperationalInsights/workspaces/prod-law
  retention_in_days: 90
```

## Node.js Application Monitoring

An Application Insights instance configured for Node.js applications.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: node-ai
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-apps-rg
  name: prod-node-api-ai
  application_type: Node.JS
  workspace_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.OperationalInsights/workspaces/prod-law
```

## Infra Chart Wiring -- Full Monitoring Stack

This example demonstrates the primary use case: wiring Application Insights into an
infra chart with Resource Group and Log Analytics Workspace.

### Resource Group (Layer 0)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: platform-rg
  org: mycompany
  env: production
spec:
  name: prod-platform-rg
  region: eastus
```

### Log Analytics Workspace (Layer 1)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: platform-law
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-platform-law
  retention_in_days: 90
```

### Application Insights (Layer 2)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: platform-ai
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-platform-ai
  workspace_id:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-law
      fieldPath: status.outputs.workspace_id
  retention_in_days: 90
  sampling_percentage: 50
```

## Enterprise Compliance Configuration

A high-retention instance for regulated environments.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: compliance-ai
  org: enterprise-corp
  env: production
spec:
  region: eastus
  resource_group: compliance-monitoring-rg
  name: compliance-app-insights
  workspace_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/compliance-rg/providers/Microsoft.OperationalInsights/workspaces/compliance-law
  retention_in_days: 730
  daily_data_cap_in_gb: 100
  sampling_percentage: 100
```

## Best Practices

1. **Use workspace-based mode** -- This component enforces workspace-based Application
   Insights by requiring `workspace_id`. Classic mode is deprecated by Microsoft.

2. **Co-locate with your workspace** -- Deploy Application Insights in the same region
   as the Log Analytics Workspace it connects to, which should also be the same region
   as the monitored applications.

3. **Tune sampling for production** -- Full sampling (100%) in production can be
   expensive. Start with 50% and increase only if you need complete telemetry fidelity.
   Sampling is statistically representative, so 50% sampling still gives accurate metrics.

4. **Set daily caps for non-production** -- Use `daily_data_cap_in_gb` in dev/staging
   to prevent cost surprises from debug logging or load testing.

5. **Use connection_string, not instrumentation_key** -- The `connection_string` output
   is the recommended way to configure SDKs. It contains the instrumentation key plus
   endpoint configuration, and supports future features like regional endpoints.

6. **One App Insights per application** -- While you can share a single instance across
   multiple apps, separate instances provide cleaner telemetry isolation and independent
   lifecycle management.

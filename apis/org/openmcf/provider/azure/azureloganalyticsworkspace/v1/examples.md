# AzureLogAnalyticsWorkspace Examples

## Minimal Configuration

The simplest possible workspace with all defaults -- PerGB2018 SKU, 30-day retention,
unlimited ingestion.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: my-law
spec:
  region: eastus
  resource_group: my-resource-group
  name: my-workspace
```

## Development Environment

A development workspace with short retention to minimize costs.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: dev-law
  org: mycompany
  env: development
spec:
  region: eastus
  resource_group: dev-monitoring-rg
  name: dev-platform-law
  retention_in_days: 30
  daily_quota_gb: 1.0
```

## Production Environment

A production workspace with extended retention and no ingestion cap.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: prod-law
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-monitoring-rg
  name: prod-platform-law
  sku: PerGB2018
  retention_in_days: 90
  daily_quota_gb: -1
```

## Enterprise Compliance Configuration

A workspace configured for regulatory compliance with maximum retention.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: compliance-law
  org: enterprise-corp
  env: production
spec:
  region: eastus
  resource_group: compliance-monitoring-rg
  name: compliance-audit-law
  sku: PerGB2018
  retention_in_days: 730
  daily_quota_gb: -1
```

## Infra Chart Wiring -- with AzureResourceGroup

This example demonstrates the primary use case: wiring the workspace to a resource
group via `valueFrom` in an infra chart.

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
  sku: PerGB2018
  retention_in_days: 90
```

### Application Insights (Layer 2) -- references workspace

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: platform-appinsights
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
```

## Cost Control Configuration

A workspace with a daily ingestion cap to prevent cost surprises. When the cap is
reached, ingestion stops until the next UTC day.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: cost-controlled-law
  org: mycompany
  env: staging
spec:
  region: eastus
  resource_group: staging-monitoring-rg
  name: staging-law
  retention_in_days: 30
  daily_quota_gb: 5.0
```

## Best Practices

1. **Start with PerGB2018 SKU** -- pay-as-you-go pricing is the most cost-effective
   for most workloads. Only use CapacityReservation when you consistently ingest
   more than 100 GB/day.

2. **Set retention based on compliance** -- 30 days for dev, 90 days for production,
   365-730 days for regulated industries.

3. **Use daily_quota_gb for non-production** -- prevent runaway costs in dev/staging
   environments where log volume can spike unexpectedly.

4. **One workspace per environment** -- share a workspace across resources in the same
   environment, but separate production from non-production.

5. **Co-locate with resources** -- deploy the workspace in the same region as the
   resources sending logs to minimize egress costs and latency.

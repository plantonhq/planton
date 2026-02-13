# AzureResourceGroup Examples

## Minimal Configuration

The simplest possible resource group -- just a name and region.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: my-rg
spec:
  name: my-resource-group
  region: eastus
```

## Development Environment

A resource group for development workloads with organizational metadata.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: dev-platform-rg
  org: mycompany
  env: development
spec:
  name: dev-platform-rg
  region: eastus
```

## Production Environment

A production resource group in a European region.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: prod-platform-rg
  org: mycompany
  env: production
spec:
  name: prod-platform-rg
  region: westeurope
```

## Enterprise Multi-Tier Architecture

In enterprise architectures, you typically create multiple resource groups for
different tiers. Each resource group provides an RBAC boundary and cost tracking unit.

### Networking Resource Group

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: prod-network-rg
  org: mycompany
  env: production
spec:
  name: prod-network-rg
  region: eastus
```

### Database Resource Group

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: prod-database-rg
  org: mycompany
  env: production
spec:
  name: prod-database-rg
  region: eastus
```

### Application Resource Group

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: prod-app-rg
  org: mycompany
  env: production
spec:
  name: prod-app-rg
  region: eastus
```

## Infra Chart Wiring Example

This example shows how downstream resources reference the resource group via `valueFrom`.
This is the primary use case in infra charts.

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

### Log Analytics Workspace (Layer 1) -- references resource group

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: platform-law
  org: mycompany
  env: production
spec:
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  region: eastus
  name: prod-platform-law
  sku: PerGB2018
  retention_in_days: 90
```

## Best Practices

1. **Naming convention** -- use a consistent naming pattern like `{env}-{purpose}-rg`
   (e.g., `prod-platform-rg`, `dev-database-rg`)

2. **Region selection** -- choose a region close to your users and one that supports
   all the Azure services you plan to use

3. **One resource group per tier** -- in enterprise architectures, separate networking,
   databases, and applications into different resource groups for RBAC isolation and
   cost tracking

4. **Don't over-segment** -- too many resource groups create management overhead.
   A good rule of thumb is one resource group per deployment boundary
   (resources that are deployed and lifecycle-managed together)

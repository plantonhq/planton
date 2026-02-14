# AzureContainerAppEnvironment Examples

## Minimal Environment (Consumption Plan)

The simplest configuration. No VNet injection, no logging, Consumption-only compute.
Apps are publicly accessible via Azure-assigned domain and static IP.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerAppEnvironment
metadata:
  name: dev-env
spec:
  region: eastus
  resource_group: dev-rg
  name: dev-apps-env
```

## Environment with Log Analytics

Centralized logging via Log Analytics Workspace. Enables KQL queries over
container app logs, Azure Monitor alerts, and dashboard integration.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerAppEnvironment
metadata:
  name: staging-env
spec:
  region: westus2
  resource_group: staging-rg
  name: staging-apps-env
  log_analytics_workspace_id: /subscriptions/sub-id/resourceGroups/shared-rg/providers/Microsoft.OperationalInsights/workspaces/central-law
```

## VNet-Injected Production Environment

VNet injection for private connectivity to databases, storage, and other VNet
resources. Zone-redundant for high availability across 3 availability zones.

The subnet must be /21 or larger (2048+ IPs).

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerAppEnvironment
metadata:
  name: prod-env
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: production-rg
  name: prod-apps-env
  infrastructure_subnet_id: /subscriptions/sub-id/resourceGroups/network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/container-apps
  log_analytics_workspace_id: /subscriptions/sub-id/resourceGroups/shared-rg/providers/Microsoft.OperationalInsights/workspaces/prod-law
  zone_redundancy_enabled: true
```

## Internal Environment (VNet-Only Access)

Internal load balancer mode restricts all apps to VNet-only access.
No public internet ingress. Use for backend microservices, internal APIs,
and workloads that should never be exposed publicly.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerAppEnvironment
metadata:
  name: internal-env
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: production-rg
  name: internal-apps-env
  infrastructure_subnet_id: /subscriptions/sub-id/resourceGroups/network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/internal-apps
  log_analytics_workspace_id: /subscriptions/sub-id/resourceGroups/shared-rg/providers/Microsoft.OperationalInsights/workspaces/prod-law
  internal_load_balancer_enabled: true
  zone_redundancy_enabled: true
```

## Environment with Dedicated Workload Profiles

Workload profiles provide dedicated compute alongside the default Consumption plan.
Use dedicated profiles for workloads needing guaranteed resources, predictable
performance, or GPU access.

The "Consumption" profile is always available -- do not add it here.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerAppEnvironment
metadata:
  name: multi-tier-env
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: production-rg
  name: multi-tier-apps-env
  infrastructure_subnet_id: /subscriptions/sub-id/resourceGroups/network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/container-apps
  log_analytics_workspace_id: /subscriptions/sub-id/resourceGroups/shared-rg/providers/Microsoft.OperationalInsights/workspaces/prod-law
  zone_redundancy_enabled: true
  workload_profiles:
    - name: general
      workload_profile_type: D4
      minimum_count: 2
      maximum_count: 8
    - name: high-memory
      workload_profile_type: E8
      minimum_count: 0
      maximum_count: 4
```

## Infra Chart: valueFrom Pattern

In the `container-apps-environment` infra chart, the environment references
upstream resources (ResourceGroup, Subnet, LAW) via `valueFrom` and is itself
referenced by downstream ContainerApp resources.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerAppEnvironment
metadata:
  name: platform-env
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: platform-env
  infrastructure_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: apps-subnet
      fieldPath: status.outputs.subnet_id
  log_analytics_workspace_id:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: central-law
      fieldPath: status.outputs.workspace_id
  zone_redundancy_enabled: true
```

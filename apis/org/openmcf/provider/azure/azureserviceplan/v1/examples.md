# AzureServicePlan Examples

## Minimal Linux Plan (Premium v3)

The simplest production-ready configuration. Defaults to Linux OS.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServicePlan
metadata:
  name: web-plan
spec:
  region: eastus
  resource_group: my-rg
  name: web-plan
  sku_name: P1v3
```

## Windows Plan (Standard)

Explicit Windows OS type for .NET Framework or Windows-specific workloads.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServicePlan
metadata:
  name: dotnet-plan
spec:
  region: westus2
  resource_group: my-rg
  name: dotnet-plan
  os_type: Windows
  sku_name: S1
```

## Premium Zone-Redundant Plan

High-availability configuration with instances distributed across availability zones.
Worker count of 3 ensures even distribution across 3 zones.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServicePlan
metadata:
  name: ha-plan
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: production-rg
  name: ha-web-plan
  sku_name: P2v3
  worker_count: 6
  zone_balancing_enabled: true
  per_site_scaling_enabled: true
```

## Elastic Premium Plan for Azure Functions

Serverless compute with pre-warmed instances and cost-controlled maximum scale.
`maximum_elastic_worker_count` caps the serverless scale-out at 50 workers.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServicePlan
metadata:
  name: functions-plan
spec:
  region: westeurope
  resource_group: serverless-rg
  name: event-processing-plan
  sku_name: EP1
  maximum_elastic_worker_count: 50
```

## Consumption Plan for Azure Functions (Pay-per-execution)

Minimal-cost serverless plan. Azure manages all scaling automatically.
No `worker_count` needed -- instances scale to 0 when idle.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServicePlan
metadata:
  name: consumption-plan
spec:
  region: eastus
  resource_group: dev-rg
  name: lightweight-functions
  sku_name: Y1
```

## Infra Chart: valueFrom Pattern

In an infra chart, the Service Plan references an `AzureResourceGroup` via `valueFrom`
and is itself referenced by downstream `AzureFunctionApp` or `AzureLinuxWebApp` resources.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServicePlan
metadata:
  name: app-plan
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: foundation-rg
      fieldPath: status.outputs.resource_group_name
  name: app-plan
  sku_name: P1v3
  worker_count: 3
  zone_balancing_enabled: true
```

## Basic Development Plan

Low-cost plan for development and testing. No zone redundancy or scaling features.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServicePlan
metadata:
  name: dev-plan
spec:
  region: eastus
  resource_group: dev-rg
  name: dev-plan
  sku_name: B1
```

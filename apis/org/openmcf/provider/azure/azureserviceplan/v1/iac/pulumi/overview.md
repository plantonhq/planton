# AzureServicePlan Pulumi Module: Architecture Overview

## Resource Graph

The AzureServicePlan module creates a single resource:

```
AzureServicePlan
└── appservice.ServicePlan (azurerm_service_plan)
```

This is one of the simplest Azure resources -- a single plan resource with no bundled
sub-resources. The complexity lives in the SKU-dependent feature validation, which
Azure's API handles at deployment time.

## Data Flow

```
AzureServicePlanStackInput
├── target.metadata    → Azure tags (resource, resource_name, resource_kind, org, env)
├── target.spec.region → ServicePlanArgs.Location
├── target.spec.resource_group → locals.ResourceGroupName (via .GetValue())
├── target.spec.name   → ServicePlanArgs.Name
├── target.spec.os_type → ServicePlanArgs.OsType (default: "Linux")
├── target.spec.sku_name → ServicePlanArgs.SkuName
├── target.spec.worker_count → ServicePlanArgs.WorkerCount (optional)
├── target.spec.zone_balancing_enabled → ServicePlanArgs.ZoneBalancingEnabled (optional)
├── target.spec.per_site_scaling_enabled → ServicePlanArgs.PerSiteScalingEnabled (optional)
└── target.spec.maximum_elastic_worker_count → ServicePlanArgs.MaximumElasticWorkerCount (optional)
```

## Output Wiring

```
ServicePlan.ID()   → plan_id    (referenced by AzureFunctionApp, AzureLinuxWebApp)
ServicePlan.Name   → plan_name  (informational)
spec.os_type       → os_type    (informational)
spec.sku_name      → sku_name   (informational)
```

## Design Notes

- **No bundled sub-resources**: Unlike databases (server + DBs + firewall rules) or
  load balancers (LB + pools + probes + rules), a Service Plan is a single resource.
- **Optional field handling**: Optional fields (`worker_count`, `zone_balancing_enabled`,
  `per_site_scaling_enabled`, `maximum_elastic_worker_count`) are only set on the
  `ServicePlanArgs` when the spec field is non-nil, allowing Azure defaults to apply.
- **os_type defaulting**: The `os_type` defaults to `"Linux"` via proto default annotation.
  The module resolves this with `spec.GetOsType()` and falls back to `"Linux"` if empty.

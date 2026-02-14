# AzureContainerAppEnvironment Pulumi Module: Architecture Overview

## Resource Graph

The AzureContainerAppEnvironment module creates a single resource:

```
AzureContainerAppEnvironment
└── containerapp.Environment (azurerm_container_app_environment)
```

This is a single-resource component. The complexity lives in the conditional
configuration: VNet injection, logging destination auto-derivation, and
workload profile expansion.

## Data Flow

```
AzureContainerAppEnvironmentStackInput
├── target.metadata    → Azure tags (resource, resource_name, resource_kind, org, env)
├── target.spec.region → EnvironmentArgs.Location
├── target.spec.resource_group → locals.ResourceGroupName (via .GetValue())
├── target.spec.name   → EnvironmentArgs.Name
├── target.spec.infrastructure_subnet_id → EnvironmentArgs.InfrastructureSubnetId (optional)
├── target.spec.log_analytics_workspace_id → EnvironmentArgs.LogAnalyticsWorkspaceId + LogsDestination (optional)
├── target.spec.internal_load_balancer_enabled → EnvironmentArgs.InternalLoadBalancerEnabled (optional)
├── target.spec.zone_redundancy_enabled → EnvironmentArgs.ZoneRedundancyEnabled (optional)
└── target.spec.workload_profiles → EnvironmentArgs.WorkloadProfiles (optional, expanded)
```

## Output Wiring

```
Environment.ID()                      → environment_id     (referenced by AzureContainerApp)
Environment.DefaultDomain             → default_domain     (DNS configuration)
Environment.StaticIpAddress           → static_ip_address  (DNS records, firewall rules)
Environment.PlatformReservedCidr      → platform_reserved_cidr (network planning)
Environment.PlatformReservedDnsIpAddress → platform_reserved_dns_ip_address (DNS debugging)
```

## Design Notes

- **logs_destination auto-derived**: When `log_analytics_workspace_id` is provided,
  `LogsDestination` is automatically set to `"log-analytics"`. When omitted, logs are
  streaming-only. This eliminates a redundant field from the user-facing spec.
- **Consumption profile auto-added**: Azure always includes the Consumption workload
  profile. The `workload_profiles` field only carries user-defined dedicated profiles.
  The module passes them directly without adding Consumption.
- **Optional field handling**: Optional fields (`infrastructure_subnet_id`,
  `log_analytics_workspace_id`, `internal_load_balancer_enabled`,
  `zone_redundancy_enabled`) are only set on `EnvironmentArgs` when non-nil,
  allowing Azure defaults to apply.
- **ForceNew awareness**: Most fields on this resource are ForceNew (name, region,
  subnet, internal LB, zone redundancy). Changes to these fields cause environment
  replacement. This is documented in the spec.proto comments.

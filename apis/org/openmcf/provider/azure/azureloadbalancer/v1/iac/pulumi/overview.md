# AzureLoadBalancer Pulumi Module -- Architecture Overview

## Resource Flow

```
Stack Input (AzureLoadBalancerStackInput)
  │
  ├── target: AzureLoadBalancer (api + spec + metadata)
  └── provider_config: AzureProviderConfig (credentials)
        │
        ▼
  initializeLocals()
  ├── Extracts resource group name via .GetValue()
  ├── Derives frontend config name: "{name}-frontend"
  ├── Builds Azure tags from metadata
  └── Returns Locals struct
        │
        ▼
  Resources()
  ├── Creates Azure provider (auth via service principal)
  ├── Builds frontend IP configuration:
  │   ├── If public_ip_id set → PublicIpAddressId (public LB)
  │   └── If subnet_id set → SubnetId + optional PrivateIpAddress (internal LB)
  ├── Creates lb.LoadBalancer (Standard SKU, hardcoded)
  │   ├── Name, Location, ResourceGroupName, Tags
  │   └── Single FrontendIpConfiguration
  ├── For each spec.BackendPools[i]:
  │   └── Creates lb.BackendAddressPool
  │       ├── Name, LoadbalancerId
  │       └── DependsOn: LoadBalancer
  ├── For each spec.HealthProbes[i]:
  │   └── Creates lb.Probe
  │       ├── Name, Protocol, Port, RequestPath (if Http/Https)
  │       ├── IntervalInSeconds (default 15), NumberOfProbes (default 2)
  │       └── DependsOn: LoadBalancer
  ├── For each spec.Rules[i]:
  │   └── Creates lb.Rule
  │       ├── Name, Protocol, FrontendPort, BackendPort
  │       ├── FrontendIpConfigurationName (auto-derived)
  │       ├── BackendAddressPoolIds (looked up by name)
  │       ├── ProbeId (looked up by name)
  │       ├── IdleTimeoutInMinutes (default 4)
  │       ├── EnableFloatingIp (default false)
  │       ├── DisableOutboundSnat (default false)
  │       └── DependsOn: LoadBalancer, BackendPool, Probe
  └── Exports outputs:
      ├── lb_id (ARM resource ID)
      ├── lb_name
      ├── frontend_ip_address (from first frontend config)
      ├── frontend_ip_configuration_id
      └── backend_pool_id (first pool's ID)
```

## Design Decisions

### Standard SKU Only

Basic SKU was retired by Azure in September 2025 and lacks zone redundancy,
outbound rules, and SLA guarantees. Gateway SKU is for NVA chaining (extremely
niche). Standard is the only production-viable SKU.

### Public vs Internal (No `is_internal` Boolean)

The LB mode is determined by which frontend field is set:
- `public_ip_id` → public LB
- `subnet_id` → internal LB

This eliminates the contradiction risk of a boolean that could disagree with
the actual frontend configuration.

### Frontend Config Name Auto-Derived

The frontend IP configuration name is automatically derived as `"{name}-frontend"`.
Load balancing rules reference this name internally. Users never need to specify it.

### Backend Pool Membership External

Backend pools only define names. Actual instance membership (VMs, VMSS, NICs) is
managed externally through AKS node pools, VMSS configurations, or NIC-to-pool
associations. This keeps the LB lifecycle independent of backend instance lifecycle.

### Default Handling

Fields with OpenMCF defaults (interval_in_seconds=15, number_of_probes=2,
idle_timeout_in_minutes=4, enable_floating_ip=false, disable_outbound_snat=false)
are resolved by OpenMCF middleware before the Pulumi module runs. The module
uses `rule.GetXxx()` to access the resolved values.

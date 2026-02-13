# AzurePublicIp Pulumi Module -- Architecture Overview

## Resource Flow

```
Stack Input (AzurePublicIpStackInput)
  │
  ├── target: AzurePublicIp (api + spec + metadata)
  └── provider_config: AzureProviderConfig (credentials)
        │
        ▼
  initializeLocals()
  ├── Extracts resource group name via .GetValue()
  ├── Builds Azure tags from metadata
  └── Returns Locals struct
        │
        ▼
  Resources()
  ├── Creates Azure provider (auth via service principal)
  ├── Builds PublicIpArgs
  │   ├── SKU: "Standard" (hardcoded)
  │   ├── AllocationMethod: "Static" (hardcoded)
  │   ├── DomainNameLabel (conditional)
  │   ├── Zones (conditional)
  │   └── IdleTimeoutInMinutes (from spec)
  ├── Creates network.PublicIp resource
  └── Exports outputs:
      ├── public_ip_id (ARM resource ID)
      ├── ip_address (allocated static IPv4)
      ├── fqdn (if domain_name_label set)
      └── public_ip_name
```

## Design Decisions

### Standard SKU Only

Azure retired the Basic SKU for Public IP Addresses on September 30, 2025.
Standard SKU is the only viable option for new deployments. This is hardcoded
in the module rather than exposed as a spec field.

### Static Allocation Only

Standard SKU requires static allocation (Azure API rejects Dynamic + Standard).
Since SKU is always Standard, allocation method is always Static. Both are
hardcoded to keep the spec clean.

### IdleTimeoutInMinutes Default Handling

The proto field uses `optional int32` with a default of 4 via `org.openmcf.shared.options.default`.
The OpenMCF middleware resolves defaults before the Pulumi module runs, so
`spec.GetIdleTimeoutInMinutes()` always returns the correct value (user-specified or 4).

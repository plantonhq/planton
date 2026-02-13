# AzureSubnet Pulumi Module -- Architecture Overview

## Resource Flow

```
Stack Input (AzureSubnetStackInput)
  |
  +-- target: AzureSubnet (api + spec + metadata)
  +-- provider_config: AzureProviderConfig (credentials)
        |
        v
  initializeLocals()
  +-- Extracts resource group name via .GetValue()
  +-- Extracts VNet name from ARM resource ID (string split)
  +-- Builds Azure tags from metadata
  +-- Returns Locals struct
        |
        v
  Resources()
  +-- Creates Azure provider (auth via service principal)
  +-- Builds SubnetArgs
  |   +-- Name, ResourceGroupName, VirtualNetworkName
  |   +-- AddressPrefixes: [address_prefix] (wrapped in list)
  |   +-- ServiceEndpoints (conditional)
  |   +-- Delegations (conditional)
  |   +-- PrivateEndpointNetworkPolicies (from spec, default "Disabled")
  |   +-- PrivateLinkServiceNetworkPoliciesEnabled (from spec, default true)
  +-- Creates network.Subnet resource
  +-- Exports outputs:
      +-- subnet_id (ARM resource ID)
      +-- subnet_name
      +-- address_prefix
```

## Design Decisions

### VNet Name Extraction

The spec references the VNet via its ARM resource ID (from `AzureVpc.status.outputs.vnet_id`).
The Pulumi Azure Classic provider requires the VNet name (not the full ID) as the
`VirtualNetworkName` argument. The locals initialization extracts the name by splitting
the ARM ID by "/" and taking the last segment.

### No Tags on Subnet Resource

Azure subnets do not support tags directly (unlike most other Azure resources).
Tags are built in locals for consistency with other components but are not applied
to the subnet resource itself.

### Single Address Prefix

The spec uses singular `address_prefix` (string) rather than plural `address_prefixes`
(list). The Pulumi resource expects a list, so the module wraps the value:
`AddressPrefixes: pulumi.StringArray{pulumi.String(spec.AddressPrefix)}`.

### Conditional Delegation

Delegation is fully optional. When `spec.Delegation` is nil, no delegation block
is added to the subnet. When present, the module creates a single delegation with
an optional actions list (Azure defaults actions if not provided).

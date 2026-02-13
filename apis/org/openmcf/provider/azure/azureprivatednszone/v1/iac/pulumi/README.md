# AzurePrivateDnsZone Pulumi Module

## Overview

This Pulumi module creates an Azure Private DNS Zone with a Virtual Network link. It provisions two Azure resources:

1. **`privatedns.Zone`** -- The private DNS zone (global, no region)
2. **`privatedns.ZoneVirtualNetworkLink`** -- Links the zone to a VNet for DNS resolution

## Architecture

```
Stack Input
    ├── Target (AzurePrivateDnsZone)
    │   ├── Metadata (name, org, env)
    │   └── Spec
    │       ├── resource_group (StringValueOrRef)
    │       ├── name (zone name)
    │       ├── vnet_id (StringValueOrRef)
    │       └── registration_enabled (bool)
    └── ProviderConfig (Azure credentials)

Resources Created
    ├── privatedns.Zone
    │   ├── Name: spec.name
    │   ├── ResourceGroupName: spec.resource_group
    │   └── Tags: standard Azure tags
    └── privatedns.ZoneVirtualNetworkLink
        ├── Name: "{metadata.name}-vnet-link"
        ├── PrivateDnsZoneName: zone.Name
        ├── VirtualNetworkId: spec.vnet_id
        └── RegistrationEnabled: spec.registration_enabled

Stack Outputs
    ├── zone_id: Azure resource ID of the zone
    └── zone_name: Name of the private DNS zone
```

## Usage

```bash
# Build
make build

# Run with Pulumi
pulumi up --stack dev

# Debug
./debug.sh
```

## Provider

Uses `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns` for the private DNS resources.

# AzurePrivateEndpoint Pulumi Module

## Overview

This Pulumi module creates an Azure Private Endpoint with an optional Private DNS Zone Group. It provisions one or two Azure resources:

1. **`privatelink.Endpoint`** -- The private endpoint that connects privately to an Azure service
2. **`privatelink.EndpointPrivateDnsZoneConfig`** (optional) -- DNS zone group that automatically registers the private endpoint's IP address in a Private DNS Zone

## Architecture

```
Stack Input
    ├── Target (AzurePrivateEndpoint)
    │   ├── Metadata (name, org, env)
    │   └── Spec
    │       ├── region
    │       ├── resource_group (StringValueOrRef)
    │       ├── name (endpoint name)
    │       ├── subnet_id (StringValueOrRef)
    │       ├── private_connection_resource_id (StringValueOrRef)
    │       ├── subresource_names ([]string)
    │       └── private_dns_zone_id (StringValueOrRef, optional)
    └── ProviderConfig (Azure credentials)

Resources Created
    ├── privatelink.Endpoint
    │   ├── Name: spec.name
    │   ├── Location: spec.region
    │   ├── ResourceGroupName: spec.resource_group
    │   ├── SubnetId: spec.subnet_id
    │   ├── PrivateServiceConnection
    │   │   ├── Name: "{metadata.name}-connection"
    │   │   ├── IsManualConnection: false
    │   │   ├── PrivateConnectionResourceId: spec.private_connection_resource_id
    │   │   └── SubresourceNames: spec.subresource_names
    │   └── Tags: standard Azure tags
    └── privatelink.EndpointPrivateDnsZoneConfig (if private_dns_zone_id provided)
        ├── Name: "{metadata.name}-dns-zone-group"
        └── PrivateDnsZoneId: spec.private_dns_zone_id

Stack Outputs
    ├── private_endpoint_id: Azure resource ID of the private endpoint
    ├── private_ip_address: Private IP address allocated to the endpoint
    └── network_interface_id: Azure resource ID of the network interface
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

Uses `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatelink` for the private endpoint resources.

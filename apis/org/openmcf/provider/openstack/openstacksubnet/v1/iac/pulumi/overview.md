# OpenStackSubnet Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackSubnetStackInput
  ├── target: OpenStackSubnet (api.proto)
  │   ├── metadata.name → subnet name
  │   └── spec: OpenStackSubnetSpec
  │       ├── network_id (StringValueOrRef FK → OpenStackNetwork)
  │       ├── cidr
  │       ├── ip_version (default: 4)
  │       ├── gateway_ip / no_gateway
  │       ├── enable_dhcp (default: true)
  │       ├── dns_nameservers[]
  │       ├── allocation_pools[]
  │       ├── description
  │       ├── tags[]
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

  initializeLocals()
  ├── Resolve network_id from StringValueOrRef → locals.NetworkId
  └── Store references for subnet()

         │
         ▼

  subnet()
  ├── Map spec fields → networking.SubnetArgs
  ├── Handle conditional fields (gateway, DHCP, pools)
  ├── networking.NewSubnet()
  └── Export outputs: subnet_id, name, cidr, gateway_ip, network_id, region
```

## Resource Mapping

| Spec Field | Pulumi SubnetArgs Field | Behavior |
|---|---|---|
| `network_id` | `NetworkId` | Required. Resolved from StringValueOrRef |
| `cidr` | `Cidr` | Required. Passed directly |
| `ip_version` | `IpVersion` | Optional. Set when present |
| `gateway_ip` | `GatewayIp` | Set when non-empty. Mutually exclusive with NoGateway |
| `no_gateway` | `NoGateway` | Set when true. Mutually exclusive with GatewayIp |
| `enable_dhcp` | `EnableDhcp` | Set when present (default: true via middleware) |
| `dns_nameservers` | `DnsNameservers` | Set when non-empty |
| `allocation_pools` | `AllocationPools` | Mapped to SubnetAllocationPoolArgs array |
| `description` | `Description` | Set when non-empty |
| `tags` | `Tags` | Set when non-empty |
| `region` | `Region` | Set when non-empty |

## Outputs

All outputs match the `OpenStackSubnetStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `subnet_id` | `createdSubnet.ID()` |
| `name` | `createdSubnet.Name` |
| `cidr` | `createdSubnet.Cidr` |
| `gateway_ip` | `createdSubnet.GatewayIp` |
| `network_id` | `createdSubnet.NetworkId` |
| `region` | `createdSubnet.Region` |

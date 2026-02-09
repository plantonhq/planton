# OpenStackLoadBalancer Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackLoadBalancerStackInput
  ├── target: OpenStackLoadBalancer (api.proto)
  │   ├── metadata.name → load balancer name
  │   └── spec: OpenStackLoadBalancerSpec
  │       ├── vip_subnet_id (StringValueOrRef FK → OpenStackSubnet)
  │       ├── vip_address
  │       ├── description
  │       ├── admin_state_up (default: true)
  │       ├── flavor_id
  │       ├── tags[]
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

  initializeLocals()
  ├── Resolve vip_subnet_id from StringValueOrRef → locals.VipSubnetId
  └── Store references for loadBalancer()

         │
         ▼

  loadBalancer()
  ├── Map spec fields → loadbalancer.LoadBalancerArgs
  ├── Handle conditional fields (vip_address, admin_state_up, flavor_id, tags)
  ├── loadbalancer.NewLoadBalancer()
  └── Export outputs: loadbalancer_id, name, vip_address, vip_port_id, vip_subnet_id, region
```

## Resource Mapping

| Spec Field | Pulumi LoadBalancerArgs Field | Behavior |
|---|---|---|
| `vip_subnet_id` | `VipSubnetId` | Required. Resolved from StringValueOrRef |
| `vip_address` | `VipAddress` | Set when non-empty. ForceNew |
| `description` | `Description` | Set when non-empty |
| `admin_state_up` | `AdminStateUp` | Set when present (default: true via middleware) |
| `flavor_id` | `FlavorId` | Set when non-empty. ForceNew |
| `tags` | `Tags` | Set when non-empty |
| `region` | `Region` | Set when non-empty |

## Outputs

All outputs match the `OpenStackLoadBalancerStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `loadbalancer_id` | `createdLb.ID()` |
| `name` | `createdLb.Name` |
| `vip_address` | `createdLb.VipAddress` |
| `vip_port_id` | `createdLb.VipPortId` |
| `vip_subnet_id` | `createdLb.VipSubnetId` |
| `region` | `createdLb.Region` |

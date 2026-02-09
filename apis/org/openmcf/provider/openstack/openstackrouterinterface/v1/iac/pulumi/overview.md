# OpenStackRouterInterface Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackRouterInterfaceStackInput
  ├── target: OpenStackRouterInterface (api.proto)
  │   ├── metadata.name → resource name (Pulumi identifier)
  │   └── spec: OpenStackRouterInterfaceSpec
  │       ├── router_id (required StringValueOrRef FK → OpenStackRouter)
  │       ├── subnet_id (required StringValueOrRef FK → OpenStackSubnet)
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

  initializeLocals()
  ├── Resolve router_id from StringValueOrRef → locals.RouterId
  ├── Resolve subnet_id from StringValueOrRef → locals.SubnetId
  └── Store references for routerInterface()

         │
         ▼

  routerInterface()
  ├── Map spec fields → networking.RouterInterfaceArgs
  ├── networking.NewRouterInterface()
  └── Export outputs: port_id, router_id, subnet_id, region
```

## Resource Mapping

| Spec Field | Pulumi RouterInterfaceArgs Field | Behavior |
|---|---|---|
| `router_id` | `RouterId` | Required. Resolved from StringValueOrRef. Passed as `pulumi.String()` |
| `subnet_id` | `SubnetId` | Required. Resolved from StringValueOrRef. Passed as `pulumi.StringPtr()` |
| `region` | `Region` | Optional. Set when non-empty |

## Outputs

All outputs match the `OpenStackRouterInterfaceStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `port_id` | `createdRouterInterface.PortId` |
| `router_id` | `createdRouterInterface.RouterId` |
| `subnet_id` | `createdRouterInterface.SubnetId` |
| `region` | `createdRouterInterface.Region` |

The `port_id` is the UUID of the port that OpenStack auto-creates on the subnet when the router interface is attached. This is also the Terraform resource ID.

## Note on Resource Naming

OpenStack router interfaces do not have a user-visible "name" attribute. The `metadata.name` from the KRM manifest is used as the Pulumi resource name for state tracking, but no name is passed to the OpenStack API.

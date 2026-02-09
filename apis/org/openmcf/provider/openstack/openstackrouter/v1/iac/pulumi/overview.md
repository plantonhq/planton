# OpenStackRouter Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackRouterStackInput
  ├── target: OpenStackRouter (api.proto)
  │   ├── metadata.name → router name
  │   └── spec: OpenStackRouterSpec
  │       ├── external_network_id (optional StringValueOrRef FK → OpenStackNetwork)
  │       ├── admin_state_up (default: true)
  │       ├── enable_snat (requires external_network_id)
  │       ├── distributed (DVR mode, create-time only)
  │       ├── external_fixed_ips[] (requires external_network_id)
  │       ├── description
  │       ├── tags[]
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

  initializeLocals()
  ├── Resolve external_network_id from StringValueOrRef → locals.ExternalNetworkId (if present)
  └── Store references for router()

         │
         ▼

  router()
  ├── Map spec fields → networking.RouterArgs
  ├── Handle optional external gateway (external_network_id, enable_snat, external_fixed_ips)
  ├── Handle optional DVR (distributed)
  ├── networking.NewRouter()
  └── Export outputs: router_id, name, external_network_id, external_gateway_ip, region
```

## Resource Mapping

| Spec Field | Pulumi RouterArgs Field | Behavior |
|---|---|---|
| `external_network_id` | `ExternalNetworkId` | Optional. Resolved from StringValueOrRef. Set only when present |
| `admin_state_up` | `AdminStateUp` | Optional. Set when present (default: true via middleware) |
| `enable_snat` | `EnableSnat` | Optional. Set when present. CEL-guarded: requires external_network_id |
| `distributed` | `Distributed` | Optional. Set when present. Create-time only |
| `external_fixed_ips` | `ExternalFixedIps` | Mapped to RouterExternalFixedIpArray. CEL-guarded: requires external_network_id |
| `description` | `Description` | Set when non-empty |
| `tags` | `Tags` | Set when non-empty |
| `region` | `Region` | Set when non-empty |

## Outputs

All outputs match the `OpenStackRouterStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `router_id` | `createdRouter.ID()` |
| `name` | `createdRouter.Name` |
| `external_network_id` | `createdRouter.ExternalNetworkId` |
| `external_gateway_ip` | `createdRouter.ExternalFixedIps[0].IpAddress` (via ApplyT) |
| `region` | `createdRouter.Region` |

The `external_gateway_ip` output uses Pulumi's `ApplyT` to safely extract the first IP from the computed `ExternalFixedIps` array, returning an empty string when no external gateway is configured.

# OpenStackLoadBalancerPool Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackLoadBalancerPoolStackInput
  ├── target: OpenStackLoadBalancerPool (api.proto)
  │   ├── metadata.name → pool name
  │   └── spec: OpenStackLoadBalancerPoolSpec
  │       ├── listener_id (StringValueOrRef FK → OpenStackLoadBalancerListener)
  │       ├── protocol (HTTP, HTTPS, TCP, UDP, PROXY)
  │       ├── lb_method (ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP, SOURCE_IP_PORT)
  │       ├── persistence (optional: type + cookie_name)
  │       ├── description
  │       ├── admin_state_up (default: true)
  │       ├── tags[]
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

  initializeLocals()
  ├── Resolve listener_id from StringValueOrRef → locals.ListenerId
  └── Store references for pool()

         │
         ▼

  pool()
  ├── Map spec fields → loadbalancer.PoolArgs
  ├── Handle optional persistence block
  ├── loadbalancer.NewPool()
  └── Export outputs: pool_id, name, protocol, lb_method, region
```

## Resource Mapping

| Spec Field | Pulumi PoolArgs Field | Behavior |
|---|---|---|
| `listener_id` | `ListenerId` | Required. Resolved from StringValueOrRef |
| `protocol` | `Protocol` | Required. Passed directly |
| `lb_method` | `LbMethod` | Required. Passed directly |
| `persistence` | `Persistence` | Optional. Mapped to PoolPersistenceArgs |
| `description` | `Description` | Set when non-empty |
| `admin_state_up` | `AdminStateUp` | Set when present (default: true via middleware) |
| `tags` | `Tags` | Set when non-empty |
| `region` | `Region` | Set when non-empty |

## Outputs

All outputs match the `OpenStackLoadBalancerPoolStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `pool_id` | `createdPool.ID()` |
| `name` | `createdPool.Name` |
| `protocol` | `createdPool.Protocol` |
| `lb_method` | `createdPool.LbMethod` |
| `region` | `createdPool.Region` |

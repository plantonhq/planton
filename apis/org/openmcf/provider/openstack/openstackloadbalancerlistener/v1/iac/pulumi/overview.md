# OpenStackLoadBalancerListener Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackLoadBalancerListenerStackInput
  ├── target: OpenStackLoadBalancerListener (api.proto)
  │   ├── metadata.name → listener name
  │   └── spec: OpenStackLoadBalancerListenerSpec
  │       ├── loadbalancer_id (StringValueOrRef FK → OpenStackLoadBalancer)
  │       ├── protocol (HTTP | HTTPS | TCP | UDP | TERMINATED_HTTPS)
  │       ├── protocol_port (1-65535)
  │       ├── description
  │       ├── connection_limit (optional, -1 = unlimited)
  │       ├── default_tls_container_ref (required for TERMINATED_HTTPS)
  │       ├── insert_headers (map<string,string>)
  │       ├── allowed_cidrs[]
  │       ├── admin_state_up (default: true)
  │       ├── tags[]
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

  initializeLocals()
  ├── Resolve loadbalancer_id from StringValueOrRef → locals.LoadBalancerId
  └── Store references for listener()

         │
         ▼

  listener()
  ├── Map spec fields → loadbalancer.ListenerArgs
  ├── Handle conditional fields (description, connection_limit, tls ref, headers, cidrs, tags)
  ├── loadbalancer.NewListener()
  └── Export outputs: listener_id, name, protocol, protocol_port, region
```

## Resource Mapping

| Spec Field | Pulumi ListenerArgs Field | Behavior |
|---|---|---|
| `loadbalancer_id` | `LoadbalancerId` | Required. Resolved from StringValueOrRef |
| `protocol` | `Protocol` | Required. One of HTTP, HTTPS, TCP, UDP, TERMINATED_HTTPS |
| `protocol_port` | `ProtocolPort` | Required. 1-65535 |
| `description` | `Description` | Set when non-empty |
| `connection_limit` | `ConnectionLimit` | Set when present (-1 = unlimited) |
| `default_tls_container_ref` | `DefaultTlsContainerRef` | Set when non-empty. Required for TERMINATED_HTTPS |
| `insert_headers` | `InsertHeaders` | Set when non-empty. Mapped to pulumi.StringMap |
| `allowed_cidrs` | `AllowedCidrs` | Set when non-empty |
| `admin_state_up` | `AdminStateUp` | Set when present (default: true via middleware) |
| `tags` | `Tags` | Set when non-empty |
| `region` | `Region` | Set when non-empty |

## Outputs

All outputs match the `OpenStackLoadBalancerListenerStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `listener_id` | `createdListener.ID()` |
| `name` | `createdListener.Name` |
| `protocol` | `createdListener.Protocol` |
| `protocol_port` | `createdListener.ProtocolPort` |
| `region` | `createdListener.Region` |

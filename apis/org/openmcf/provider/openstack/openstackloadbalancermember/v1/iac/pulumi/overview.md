# OpenStackLoadBalancerMember Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackLoadBalancerMemberStackInput
  |-- target: OpenStackLoadBalancerMember (api.proto)
  |   |-- metadata.name -> member name
  |   +-- spec: OpenStackLoadBalancerMemberSpec
  |       |-- pool_id (StringValueOrRef FK -> OpenStackLoadBalancerPool)
  |       |-- address (backend server IP)
  |       |-- protocol_port (backend server port)
  |       |-- subnet_id (optional StringValueOrRef FK -> OpenStackSubnet)
  |       |-- weight (optional, 0-256)
  |       |-- admin_state_up (default: true)
  |       |-- tags[]
  |       +-- region
  +-- provider_config: OpenStackProviderConfig

         |
         v

  initializeLocals()
  |-- Resolve pool_id from StringValueOrRef -> locals.PoolId
  |-- Resolve optional subnet_id -> locals.SubnetId
  +-- Store references for member()

         |
         v

  member()
  |-- Map spec fields -> loadbalancer.MemberArgs
  |-- Handle optional weight, subnet_id, admin_state_up, tags
  |-- loadbalancer.NewMember()
  +-- Export outputs: member_id, name, address, protocol_port, weight, region
```

## Resource Mapping

| Spec Field | Pulumi MemberArgs Field | Behavior |
|---|---|---|
| `pool_id` | `PoolId` | Required. Resolved from StringValueOrRef |
| `address` | `Address` | Required. Passed directly |
| `protocol_port` | `ProtocolPort` | Required. Passed directly |
| `subnet_id` | `SubnetId` | Set when present. Resolved from StringValueOrRef |
| `weight` | `Weight` | Set when present (optional int32) |
| `admin_state_up` | `AdminStateUp` | Set when present (default: true via middleware) |
| `tags` | `Tags` | Set when non-empty |
| `region` | `Region` | Set when non-empty |

## Outputs

All outputs match the `OpenStackLoadBalancerMemberStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `member_id` | `createdMember.ID()` |
| `name` | `createdMember.Name` |
| `address` | `createdMember.Address` |
| `protocol_port` | `createdMember.ProtocolPort` |
| `weight` | `createdMember.Weight` |
| `region` | `createdMember.Region` |

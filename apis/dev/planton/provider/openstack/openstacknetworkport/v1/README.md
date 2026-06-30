# OpenStackNetworkPort

An OpenStack Neutron port provides stable network identity (MAC address, fixed IPs, security groups) on a virtual network. Explicit port creation is preferred over instance-inline networking when you need stable IP addresses, pre-provisioned network identities for InfraChart orchestration, or fine-grained security group assignments.

## When to Use

- **Stable IPs that survive instance rebuilds** -- Create the port first, assign IPs, then attach to an instance. Rebuilding the instance preserves the port and its addresses.
- **InfraChart orchestration** -- Wire ports to subnets and security groups created in the same chart using `value_from` FK references.
- **Floating IP targets** -- Create a port and associate a floating IP with it via `OpenStackFloatingIp` or `OpenStackFloatingIpAssociate`.
- **Multiple security groups** -- Assign multiple security groups to a single network interface.
- **Specific MAC addresses** -- Required for DPDK, network bonding, or license-tied MAC scenarios.

## Foreign Key Relationships

| Field | FK Target | Required |
|-------|-----------|----------|
| `network_id` | `OpenStackNetwork.status.outputs.network_id` | Yes |
| `fixed_ips[].subnet_id` | `OpenStackSubnet.status.outputs.subnet_id` | No |
| `security_group_ids[]` | `OpenStackSecurityGroup.status.outputs.security_group_id` | No |

## Key Design Decisions

- **`repeated StringValueOrRef` for `security_group_ids`** -- First component in Planton to use repeated FK fields. Each element independently resolves as a literal UUID or `value_from` reference, enabling InfraChart DAG wiring to multiple security groups.
- **`StringValueOrRef` inside nested `FixedIp` message** -- First component with FK annotations inside a nested message. Enables `value_from` references to subnets created in the same InfraChart.
- **`no_security_groups` field** -- Explicitly removes all security groups (including the default SG that OpenStack auto-applies). Without this field, an empty `security_group_ids` list means "apply the default SG."
- **`port_security_enabled`** -- Inherits from the network if omitted. Explicitly set to `false` for ports that need unrestricted traffic (network appliances, certain LB configurations).

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `network_id` | StringValueOrRef | Yes | Network to create the port on (ForceNew) |
| `fixed_ips` | repeated FixedIp | No | IP allocations from subnets |
| `security_group_ids` | repeated StringValueOrRef | No | Security groups to apply |
| `no_security_groups` | bool | No | Remove all SGs including default |
| `admin_state_up` | optional bool | No | Admin state (default: true) |
| `mac_address` | string | No | Specific MAC address (ForceNew) |
| `port_security_enabled` | optional bool | No | Port security enforcement |
| `description` | string | No | Human-readable description |
| `tags` | repeated string | No | Tags for filtering/organization |
| `region` | string | No | Region override |

## Outputs

| Output | Description |
|--------|-------------|
| `port_id` | Port UUID (primary FK target) |
| `mac_address` | Assigned MAC address |
| `all_fixed_ips` | All assigned IP addresses |
| `all_security_group_ids` | All applied security group UUIDs |
| `region` | OpenStack region |

## Downstream Consumers

- `OpenStackFloatingIp.spec.port_id` -- Built-in floating IP association
- `OpenStackFloatingIpAssociate.spec.port_id` -- DAG-visible floating IP association
- `OpenStackInstance.spec.networks[].port` -- Instance network attachment (future)

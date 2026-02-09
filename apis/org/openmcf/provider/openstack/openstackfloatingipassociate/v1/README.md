# OpenStackFloatingIpAssociate

Associates an existing OpenStack Neutron floating IP with a port, providing external connectivity to the resource attached to that port. This is the DAG-visible "join" resource for floating IP management in InfraCharts.

## When to Use

- **InfraChart DAG visibility** -- When the floating IP allocation and port creation are separate chart components and you need the association to appear as an explicit node in the dependency graph.
- **Decoupled lifecycle** -- When floating IPs are long-lived (e.g., DNS-registered) and ports are recreated during instance replacements.
- **Explicit dependency edges** -- When you need InfraChart orchestration to clearly show "this floating IP is bound to this port."

## When to Use OpenStackFloatingIp Instead

- When allocation and association happen together in a single manifest.
- When DAG visibility of the association is not needed.
- When simplicity is preferred over explicit decomposition.

## Foreign Key Relationships

| Field | FK Target | Required |
|-------|-----------|----------|
| `floating_ip` | `OpenStackFloatingIp.status.outputs.address` | Yes |
| `port_id` | `OpenStackNetworkPort.status.outputs.port_id` | Yes |

**Note**: The `floating_ip` FK targets the **address** output (e.g., "203.0.113.42"), not the UUID. This matches the Terraform provider's `floating_ip` attribute which accepts either an IP address or a floating IP UUID.

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `floating_ip` | StringValueOrRef | Yes | Floating IP address or UUID to associate |
| `port_id` | StringValueOrRef | Yes | Port to bind the floating IP to |
| `fixed_ip` | string | No | Specific fixed IP on multi-IP ports |
| `region` | string | No | Region override (ForceNew) |

## Outputs

| Output | Description |
|--------|-------------|
| `id` | Terraform resource ID |
| `floating_ip` | The associated floating IP address |
| `port_id` | The associated port UUID |
| `fixed_ip` | The mapped fixed IP (computed if not specified) |
| `region` | OpenStack region |

## Relationship to Other Components

```
OpenStackFloatingIp (allocates public IP)
        |
        | status.outputs.address
        v
OpenStackFloatingIpAssociate (this component)
        ^
        | status.outputs.port_id
        |
OpenStackNetworkPort (provides network endpoint)
```

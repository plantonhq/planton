# Floating IP Allocation Only

This preset allocates a floating IP from an external network without associating it with a port. The floating IP is reserved but not yet bound to any instance. Use `OpenStackFloatingIpAssociate` as a separate resource to bind it later -- this gives better DAG visibility in InfraCharts.

## When to Use

- Allocating floating IPs ahead of time for DNS pre-configuration or firewall whitelisting
- InfraCharts where allocation and association should be separate DAG nodes
- Reserving public IPs before the target instance or port exists

## Key Configuration Choices

- **Allocation only** -- no `portId` set, so the IP is reserved but not bound
- **Auto-assigned address** -- OpenStack picks any available IP from the external network

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<external-network-id>` | ID of the external (provider) network to allocate from | OpenStack admin or `OpenStackNetwork` (external) status outputs |

## Related Presets

- **02-with-port-association** -- Use instead when allocation and association should happen in a single resource

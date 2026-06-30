# Floating IP with Port Association

This preset allocates a floating IP and immediately associates it with a port, providing external connectivity to whatever is attached to that port (typically an instance). This is a single-resource approach -- simpler than using separate allocation and association resources.

## When to Use

- Simple deployments where a single instance needs a public IP
- Cases where the target port already exists at the time the floating IP is created
- Standalone manifests (not InfraCharts) where DAG node separation is unnecessary

## Key Configuration Choices

- **Immediate association** (`portId`) -- the floating IP is bound to the port upon creation
- **Auto-assigned address** -- OpenStack picks any available IP from the external network

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<external-network-id>` | ID of the external (provider) network to allocate from | OpenStack admin or `OpenStackNetwork` (external) status outputs |
| `<port-id>` | ID of the port to associate the floating IP with | OpenStack console or `OpenStackNetworkPort` status outputs |

## Related Presets

- **01-allocation-only** -- Use instead when allocation and association should be separate DAG nodes in InfraCharts

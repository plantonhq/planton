# Standard Load Balancer

This preset creates an Octavia load balancer with a VIP on the specified subnet. The load balancer itself is just the VIP endpoint -- attach listeners, pools, members, and monitors to complete the traffic path. This is the entry point for all Octavia load balancing configurations.

## When to Use

- Any workload that needs traffic distribution across multiple backend servers
- The first step in building an Octavia load balancer stack (LB -> Listener -> Pool -> Members -> Monitor)

## Key Configuration Choices

- **Auto-assigned VIP** -- Octavia picks an available IP from the subnet (add `vipAddress` to request a specific IP)
- **Admin state up** -- default (true), load balancer is active immediately
- **No flavor** -- uses Octavia's default resource limits (add `flavorId` for custom bandwidth/connection limits)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<subnet-id>` | ID of the subnet for the VIP address | OpenStack console or `OpenStackSubnet` status outputs |

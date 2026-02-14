# Port with No Security Groups

This preset creates a port with all security groups removed, including the default security group that OpenStack normally applies. Traffic flows unrestricted through this port. Use this for load balancer VIPs, network appliance ports, or any port where security group filtering would interfere with traffic.

## When to Use

- Load balancer VIP ports that must accept traffic on all protocols and ports
- Network appliance ports (firewalls, routers) that manage their own traffic filtering
- Trunk ports or DPDK ports where security groups are incompatible

## Key Configuration Choices

- **No security groups** (`noSecurityGroups: true`) -- explicitly removes all SGs, bypassing OpenStack's default SG behavior
- **Explicit subnet** (`fixedIps[].subnetId`) -- ensures the IP comes from the specified subnet
- **Admin state up** -- default (true), port forwards traffic immediately

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<network-id>` | ID of the network to create the port on | OpenStack console or `OpenStackNetwork` status outputs |
| `<subnet-id>` | ID of the subnet to allocate an IP from | OpenStack console or `OpenStackSubnet` status outputs |

## Related Presets

- **01-standard-fixed-ip** -- Use instead when the port should have the project's default security group applied

# Isolated Subnet (No Gateway)

This preset creates an isolated subnet with no gateway and no DHCP. It is designed for backend networks where instances use statically assigned IPs and no routing to external networks is needed -- storage replication, database clusters, or heartbeat networks.

## When to Use

- Storage networks (Ceph, NFS, iSCSI) where traffic stays within the subnet
- Database replication links that should not be routable externally
- Heartbeat or cluster-interconnect networks
- Any backend network where static IP assignment is preferred

## Key Configuration Choices

- **No gateway** (`noGateway: true`) -- subnet has no default route; traffic cannot leave the L2 domain
- **DHCP disabled** (`enableDhcp: false`) -- IPs are assigned statically via port fixed_ips or instance configuration
- **/24 CIDR on 10.0.0.0** -- private RFC 1918 range, 254 usable addresses

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<network-id>` | ID of the network this subnet belongs to | OpenStack console or `OpenStackNetwork` status outputs |

## Related Presets

- **01-standard-dhcp** -- Use instead when instances need automatic IP assignment and external routing

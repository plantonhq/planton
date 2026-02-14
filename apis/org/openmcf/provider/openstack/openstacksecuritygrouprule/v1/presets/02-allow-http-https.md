# Allow HTTPS Ingress Rule

This preset creates a standalone security group rule that allows inbound HTTPS (TCP port 443) from any source. This is the most common standalone rule for web-facing services. For HTTP (port 80), duplicate this preset and change the port range to 80.

## When to Use

- Adding HTTPS access to an existing security group for web services
- InfraCharts where the security group is created in a separate manifest
- Load balancer or reverse proxy security groups that need public HTTPS ingress

## Key Configuration Choices

- **Ingress only** (`direction: ingress`) -- allows inbound HTTPS connections
- **Open to all** (`remoteIpPrefix: 0.0.0.0/0`) -- accepts HTTPS from any IPv4 source
- **HTTPS port only** (`portRangeMin: 443`, `portRangeMax: 443`) -- single port, not a range
- **ForceNew** -- all fields are immutable; changing any field recreates the rule

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<security-group-id>` | ID of the security group to add this rule to | OpenStack console or `OpenStackSecurityGroup` status outputs |

## Related Presets

- **01-allow-ssh** -- Use alongside this preset for SSH access from a trusted network

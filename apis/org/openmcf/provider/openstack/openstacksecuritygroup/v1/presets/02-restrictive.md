# Restrictive Security Group (Zero-Trust Baseline)

This preset creates a security group with OpenStack's default egress rules deleted, providing a zero-trust starting point. Only explicitly defined rules are active: SSH from a trusted CIDR and all outbound IPv4. Add more rules as needed for your specific workload.

## When to Use

- Security-sensitive environments that require explicit rule documentation for every allowed flow
- Compliance workloads (PCI-DSS, HIPAA) where default-allow egress is unacceptable
- Backend services that should only be reachable from specific sources

## Key Configuration Choices

- **Default rules deleted** (`deleteDefaultRules: true`) -- starts with zero rules, not OpenStack's default allow-all-egress
- **SSH restricted** -- only from a trusted CIDR
- **Explicit egress** -- IPv4 egress is re-added manually (IPv6 egress omitted; add if needed)
- **Stateful** -- default mode; return traffic is automatically allowed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<trusted-cidr>` | CIDR of the network allowed to SSH (e.g., `10.0.0.0/8` or `203.0.113.50/32`) | Your network admin or VPN configuration |

## Related Presets

- **01-web-server** -- Use instead for a standard web server with HTTP/HTTPS open to the internet

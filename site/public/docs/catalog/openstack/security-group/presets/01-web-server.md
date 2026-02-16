---
title: "Web Server Security Group"
description: "This preset creates a security group for a typical web server: SSH from a trusted network, HTTP and HTTPS from anywhere, and unrestricted egress (OpenStack's default). The default egress rules are..."
type: "preset"
rank: "01"
presetSlug: "01-web-server"
componentSlug: "security-group"
componentTitle: "Security Group"
provider: "openstack"
icon: "package"
order: 1
---

# Web Server Security Group

This preset creates a security group for a typical web server: SSH from a trusted network, HTTP and HTTPS from anywhere, and unrestricted egress (OpenStack's default). The default egress rules are kept, providing full outbound connectivity for package updates, API calls, and response traffic.

## When to Use

- Web servers or reverse proxies exposed to the internet
- Application servers behind a load balancer that also need direct SSH access
- Standard web-tier instances serving HTTP/HTTPS traffic

## Key Configuration Choices

- **SSH restricted** (`remoteIpPrefix: <trusted-cidr>`) -- SSH is not open to the world; limited to a trusted network CIDR
- **HTTP/HTTPS open** (`remoteIpPrefix: 0.0.0.0/0`) -- web traffic allowed from any source
- **Default egress kept** -- OpenStack's automatic allow-all-egress rules remain (no `deleteDefaultRules`)
- **Stateful** -- default mode; return traffic is automatically allowed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<trusted-cidr>` | CIDR of the network allowed to SSH (e.g., `10.0.0.0/8` or your office IP `203.0.113.50/32`) | Your network admin or VPN configuration |

## Related Presets

- **02-restrictive** -- Use instead for a zero-trust baseline where all rules (including egress) are explicitly defined

# Production Web Server

This preset creates a production-grade compute instance on a medium-sized node with Ubuntu 22.04 LTS, VPC networking, firewall protection, and a cloud-init script that applies security updates on first boot. Suitable for web servers, API backends, and reverse proxies.

## When to Use

- Production web applications and API servers
- Reverse proxies (Nginx, Caddy, Traefik) fronting backend services
- Any internet-facing workload that needs VPC isolation and firewall protection

## Key Configuration Choices

- **Medium instance** (`size: g3.medium`) -- balanced CPU/RAM for most web workloads; scale to `g3.large` or `g3.xlarge` for heavier traffic
- **Ubuntu 22.04 LTS** (`image: ubuntu-jammy`) -- long-term support, widely documented, excellent package ecosystem
- **VPC networking** (`network`) -- private network for secure east-west traffic between instances
- **Web firewall** (`firewallIds`) -- attach the web-tier firewall allowing HTTP/HTTPS/SSH
- **Cloud-init** (`userData`) -- automatic security updates on first boot; extend with your application setup
- **Tags** (`tags: [web, production]`) -- enables tag-based firewall auto-assignment and organization

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | `CivoVpc` status outputs |
| `<web-firewall-id>` | Firewall ID for web-tier rules | `CivoFirewall` status outputs |

## Related Presets

- **02-development** -- Use instead for dev/test instances without firewall or cloud-init complexity

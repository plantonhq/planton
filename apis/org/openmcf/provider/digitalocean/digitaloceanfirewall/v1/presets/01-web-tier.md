# Web Tier Firewall

This preset creates a DigitalOcean Cloud Firewall for web-facing Droplets. It allows inbound HTTP/HTTPS from anywhere, restricts SSH to a management CIDR, and permits all outbound traffic. The firewall is applied via tag-based targeting to any Droplet tagged `web`.

## When to Use

- Web servers, reverse proxies, or API gateways exposed to the internet
- Any Droplet that serves HTTP/HTTPS traffic to end users
- Production web tier requiring SSH restricted to a management network

## Key Configuration Choices

- **HTTPS + HTTP inbound** (`ports 443, 80`) -- open to all IPv4 and IPv6 addresses. This is the standard web-facing configuration.
- **Restricted SSH** (`port 22`) -- limited to a specific management CIDR. Never expose SSH to `0.0.0.0/0` in production.
- **Permissive outbound** -- all TCP, UDP, and ICMP outbound allowed. Droplets can reach external APIs, package repositories, and services.
- **Tag-based targeting** (`tags: [web]`) -- automatically applied to any Droplet tagged `web`. Preferred over explicit Droplet IDs for scalability.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-management-cidr>` | CIDR block for SSH access (e.g., `203.0.113.0/24` or your VPN IP) | Your network admin or VPN provider |

## Related Presets

- **02-database-tier** -- Use for backend Droplets that should only accept traffic from the web tier, not from the public internet

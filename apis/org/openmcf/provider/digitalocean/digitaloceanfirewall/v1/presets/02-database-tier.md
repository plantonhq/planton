# Database Tier Firewall

This preset creates a DigitalOcean Cloud Firewall for database Droplets. It restricts inbound access to PostgreSQL (port 5432) from web-tier Droplets only, and limits SSH to a management CIDR. Applied via tag-based targeting to any Droplet tagged `database`.

## When to Use

- Self-managed PostgreSQL, MySQL (change port to 3306), or other database servers
- Backend Droplets that must only accept connections from the application tier
- Any database that should never be directly reachable from the public internet

## Key Configuration Choices

- **Tag-based source filtering** (`sourceTags: [web]`) -- only Droplets tagged `web` can reach port 5432. This is the recommended DigitalOcean pattern for multi-tier architectures.
- **PostgreSQL port** (`portRange: "5432"`) -- change to `3306` for MySQL, `6379` for Redis, `27017` for MongoDB, etc.
- **Restricted SSH** -- SSH access limited to management CIDR, consistent with the web-tier firewall.
- **Outbound**: all TCP allowed (for package updates, replication); DNS (UDP 53) allowed for name resolution.
- **Tag-based targeting** (`tags: [database]`) -- automatically applied to any Droplet tagged `database`.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-management-cidr>` | CIDR block for SSH access (e.g., `203.0.113.0/24` or your VPN IP) | Your network admin or VPN provider |

## Related Presets

- **01-web-tier** -- Companion firewall for the web-facing Droplets that connect to these database Droplets

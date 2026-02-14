# Web Tier Firewall

This preset creates a firewall allowing inbound HTTP (80), HTTPS (443), and restricted SSH (22) access. It is the most common configuration for internet-facing web servers, API gateways, and reverse proxies running inside a Civo VPC.

## When to Use

- Web servers and API backends exposed to the internet
- Reverse proxies and load balancers that accept public HTTP/HTTPS traffic
- Any compute instance serving web traffic that also needs SSH for administration

## Key Configuration Choices

- **HTTP + HTTPS open to all** (`0.0.0.0/0`) -- standard for public-facing web services
- **SSH restricted** (`<your-admin-cidr>/32`) -- locked to a single admin IP; never open SSH to `0.0.0.0/0` in production
- **No egress rules** -- Civo allows all outbound traffic by default when no egress rules are specified
- **Tag-based targeting** (`tags: [web]`) -- any instance in the same network tagged `web` inherits this firewall automatically

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | Civo dashboard or `CivoVpc` status outputs |
| `<your-admin-cidr>` | Your public IP address for SSH access | `curl ifconfig.me` from your admin machine |

## Related Presets

- **02-database-tier** -- Use for backend database instances that should only accept traffic from the application tier

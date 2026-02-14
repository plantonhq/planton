# Web Server Security Group

This preset creates a security group for web-facing instances using an allowlist model. Inbound traffic is dropped by default, with explicit rules accepting SSH (from a restricted CIDR), HTTP, and HTTPS from anywhere. All outbound traffic is allowed.

## When to Use

- Web servers, API servers, or reverse proxies that serve HTTP/HTTPS traffic
- Instances that need SSH access restricted to specific admin IP ranges
- Any public-facing server where only known ports should be open

## Key Configuration Choices

- **Allowlist model** (`inboundDefaultPolicy: drop`) -- all inbound traffic is blocked unless an explicit rule accepts it; the most secure approach for production
- **SSH restricted** (`portRange: "22"`, `ipRange: <your-admin-cidr>`) -- SSH access limited to your office or VPN CIDR; prevents brute-force attacks from the internet
- **HTTP/HTTPS open** (ports 80 and 443 from `0.0.0.0/0`) -- allows public web traffic from any source
- **All outbound allowed** (`outboundDefaultPolicy: accept`) -- instances can reach any external service for package updates, API calls, and DNS
- **Stateful** (`stateful: true`) -- return traffic for accepted connections is automatically permitted

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-admin-cidr>` | CIDR range for SSH access (e.g., `203.0.113.0/24` or `198.51.100.10/32`) | Your network administrator or VPN provider |

## Related Presets

- **02-deny-all-allowlist** -- Use instead for backend instances (databases, workers) that should only accept traffic from specific private IPs

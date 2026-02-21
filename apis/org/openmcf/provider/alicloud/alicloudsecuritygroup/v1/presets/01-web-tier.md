# Web Tier Security Group

This preset creates a security group suitable for public-facing web servers or load balancers. It allows HTTP (port 80) and HTTPS (port 443) inbound from any source, with unrestricted outbound access.

## When to Use

- ECS instances running web applications behind ALB or NLB
- Public-facing API servers
- Reverse proxy or ingress controller nodes
- Any resource that needs to accept HTTP/HTTPS traffic from the internet

## Key Configuration Choices

- **HTTP + HTTPS ingress from 0.0.0.0/0** -- Allows traffic from any IPv4 address. Restrict to specific CIDRs if your traffic comes from known IP ranges (e.g., CDN edge nodes).
- **All egress allowed** -- Web servers typically need outbound access for package updates, API calls, and database connections. Restrict if your security policy requires it.
- **Default inner_access_policy (Accept)** -- Instances in the same security group can communicate freely, which is standard for web tier instances behind a load balancer.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region |
| `<your-vpc-id>` | VPC ID that this security group belongs to | Alibaba Cloud VPC console or `AlicloudVpc` stack outputs |
| `<your-sg-name>` | Security group name (2-128 chars) | Choose a descriptive name |

## Related Presets

- **02-database-tier** -- Use for database instances that should only accept internal VPC traffic
- **03-bastion-host** -- Use for bastion/jump hosts that need SSH access

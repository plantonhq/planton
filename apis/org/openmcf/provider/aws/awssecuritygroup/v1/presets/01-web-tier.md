# Web Tier Security Group

This preset creates a security group for internet-facing web servers or load balancers. It allows inbound HTTP (80) and HTTPS (443) traffic from any source and permits all outbound traffic. This is the most common security group pattern for ALBs, web servers, and API gateways.

## When to Use

- Application Load Balancers that serve public internet traffic
- Internet-facing web servers or reverse proxies
- Any resource that needs to accept HTTP/HTTPS connections from the public internet

## Key Configuration Choices

- **HTTP and HTTPS ingress from anywhere** (`0.0.0.0/0`) -- Allows all IPv4 traffic on ports 80 and 443; restrict to specific CIDRs if your application is not fully public
- **All outbound traffic** (protocol `-1`) -- Unrestricted egress for health checks, API calls, dependency downloads, and responses
- **Rule descriptions** -- Each rule has a human-readable description for AWS console clarity

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<vpc-id>` | VPC ID where this security group will be created | AWS VPC console or `AwsVpc` status outputs |

## Related Presets

- **02-database-tier** -- Use for database resources that should only accept connections from the application tier
- **03-bastion** -- Use for bastion hosts that accept SSH from specific trusted IPs

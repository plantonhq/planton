# Database Tier Security Group

This preset creates a locked-down security group for database instances (RDS, PolarDB, Redis, MongoDB). It only allows connections on standard database ports from the VPC CIDR range, with no public access and no unrestricted outbound rules.

## When to Use

- RDS instances running MySQL, PostgreSQL, or SQL Server
- PolarDB clusters
- Redis (KVStore) instances
- MongoDB instances
- Any database resource that should only be accessible from within the VPC

## Key Configuration Choices

- **inner_access_policy: Drop** -- Prevents databases from communicating with each other within the same security group unless an explicit rule allows it. This is a security hardening measure.
- **VPC CIDR as source** -- Only allows connections from within the VPC. Replace `<your-vpc-cidr>` with your actual VPC CIDR block (e.g., `10.0.0.0/8`).
- **No egress rules** -- Databases rarely need outbound internet access. Stateful return traffic for accepted ingress connections is automatically allowed.
- **Priority ordering** -- Rules are numbered 1-3 for clarity; adjust if you add more rules.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID that this security group belongs to | Alibaba Cloud VPC console or `AlicloudVpc` stack outputs |
| `<your-sg-name>` | Security group name (2-128 chars) | Choose a descriptive name |
| `<your-vpc-cidr>` | VPC CIDR block (e.g., `10.0.0.0/8`, `172.16.0.0/12`) | Your AlicloudVpc spec.cidrBlock |

## Related Presets

- **01-web-tier** -- Use for public-facing web servers
- **03-bastion-host** -- Use for bastion hosts that need SSH + database access

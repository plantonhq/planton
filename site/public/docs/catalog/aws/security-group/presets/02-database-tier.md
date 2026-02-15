---
title: "Database Tier Security Group"
description: "This preset creates a security group for database instances that only accepts connections from the application tier. Ingress is restricted to PostgreSQL port 5432 from a specific CIDR block..."
type: "preset"
rank: "02"
presetSlug: "02-database-tier"
componentSlug: "security-group"
componentTitle: "Security Group"
provider: "aws"
icon: "package"
order: 2
---

# Database Tier Security Group

This preset creates a security group for database instances that only accepts connections from the application tier. Ingress is restricted to PostgreSQL port 5432 from a specific CIDR block (typically the application subnet range). Adjust the port for MySQL (3306), MongoDB (27017), or other database engines.

## When to Use

- RDS instances, Aurora clusters, or DocumentDB clusters that should only be accessible from application servers
- Any database resource in a private subnet that needs restricted network access
- Defense-in-depth networking where databases are isolated from direct internet or bastion access

## Key Configuration Choices

- **PostgreSQL port** (`fromPort/toPort: 5432`) -- Default for PostgreSQL; change to 3306 for MySQL, 27017 for MongoDB, or 6379 for Redis
- **Application tier CIDR only** (`<app-tier-cidr>`) -- Restricts ingress to the application subnet range; never use `0.0.0.0/0` for databases
- **All outbound traffic** -- Permits egress for replication, backups, and AWS service communication

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<vpc-id>` | VPC ID where this security group will be created | AWS VPC console or `AwsVpc` status outputs |
| `<app-tier-cidr>` | CIDR block of the application subnet (e.g., `10.0.1.0/24`) | Your VPC subnet configuration or `AwsVpc` status outputs |

## Related Presets

- **01-web-tier** -- Use for internet-facing resources (ALBs, web servers)
- **03-bastion** -- Use for bastion hosts that accept SSH from trusted IPs

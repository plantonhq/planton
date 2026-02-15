---
title: "Database Tier Firewall"
description: "This preset creates a firewall that restricts inbound access to standard database ports (PostgreSQL 5432, MySQL 3306) from the application tier CIDR only. No public internet access is permitted,..."
type: "preset"
rank: "02"
presetSlug: "02-database-tier"
componentSlug: "firewall"
componentTitle: "Firewall"
provider: "civo"
icon: "package"
order: 2
---

# Database Tier Firewall

This preset creates a firewall that restricts inbound access to standard database ports (PostgreSQL 5432, MySQL 3306) from the application tier CIDR only. No public internet access is permitted, ensuring databases are only reachable from trusted internal sources.

## When to Use

- Database instances (PostgreSQL, MySQL) that should only accept connections from application servers
- Backend services that must not be directly accessible from the internet
- Any data-tier workload following a multi-tier security architecture

## Key Configuration Choices

- **PostgreSQL + MySQL ports** (`5432`, `3306`) -- covers the two most common database engines; remove whichever port you don't use
- **Restricted source CIDR** (`<app-tier-cidr>`) -- only the application subnet can reach the database; never use `0.0.0.0/0`
- **No SSH rule** -- database instances should be administered via a bastion host, not direct SSH
- **No egress rules** -- all outbound traffic allowed by default (needed for package updates, backups)
- **Tag-based targeting** (`tags: [database]`) -- any instance tagged `database` inherits this firewall

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | Civo dashboard or `CivoVpc` status outputs |
| `<app-tier-cidr>` | CIDR block of your application tier (e.g., `10.0.0.0/24`) | Your VPC network plan |

## Related Presets

- **01-web-tier** -- Use for internet-facing web servers that need HTTP/HTTPS and restricted SSH access

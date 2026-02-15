---
title: "Development MySQL Database"
description: "This preset creates a minimal MySQL 8.0 database for development and testing. Single node (no replicas), smallest instance size, no firewall. Keeps cost low while providing a fully managed MySQL..."
type: "preset"
rank: "02"
presetSlug: "02-mysql-development"
componentSlug: "database"
componentTitle: "Database"
provider: "civo"
icon: "package"
order: 2
---

# Development MySQL Database

This preset creates a minimal MySQL 8.0 database for development and testing. Single node (no replicas), smallest instance size, no firewall. Keeps cost low while providing a fully managed MySQL instance for application development.

## When to Use

- Development and testing environments using MySQL
- Applications being migrated from MySQL-based stacks
- Quick prototyping where a managed database is needed without production overhead

## Key Configuration Choices

- **MySQL 8.0** (`engine: mysql`, `engineVersion: "8.0"`) -- latest LTS major version
- **No replicas** (`replicas` omitted) -- single primary node to minimize cost
- **Small instance** (`sizeSlug: g3.db.small`) -- lowest cost for dev workloads
- **No firewall** -- simplifies connectivity during development; add firewall rules for staging
- **VPC networking** (`networkId`) -- private network access even in dev

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | `CivoVpc` status outputs |

## Related Presets

- **01-postgresql-production** -- Use instead for production PostgreSQL deployments with replicas and firewall protection

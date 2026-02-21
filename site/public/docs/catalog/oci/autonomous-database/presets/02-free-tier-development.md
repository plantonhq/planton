---
title: "Free Tier Development"
description: "This preset creates an Always Free Autonomous Database for development and experimentation at zero cost. The database is limited to 2 ECPUs and 20 GB of usable storage but provides the full..."
type: "preset"
rank: "02"
presetSlug: "02-free-tier-development"
componentSlug: "autonomous-database"
componentTitle: "Autonomous Database"
provider: "oci"
icon: "package"
order: 2
---

# Free Tier Development

This preset creates an Always Free Autonomous Database for development and experimentation at zero cost. The database is limited to 2 ECPUs and 20 GB of usable storage but provides the full Autonomous Database feature set for prototyping, learning, and testing application integrations.

## When to Use

- Prototyping a new application against a real Oracle Autonomous Database
- Learning Oracle SQL, PL/SQL, or APEX development without incurring costs
- CI/CD environments that need a disposable database for integration testing
- Personal development environments or hackathon projects

## Key Configuration Choices

- **Always Free** (`isFreeTier: true`) -- no charges accrue for this database. It is automatically paused after extended inactivity and reclaimed if inactive for an extended period. Not suitable for production.
- **2 ECPUs** (`computeCount: 2`) -- the Always Free minimum. Auto-scaling is not available on free tier.
- **Public endpoint** (no `subnetId`) -- the database is accessible over the internet via its public secure access URL. Acceptable for development; use private endpoints for production.
- **OLTP workload** (`dbWorkload: oltp`) -- can be changed to `dw` or `ajd` depending on the development use case. The Always Free tier supports all workload types.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the database | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<admin-password>` | Administrator password (12-30 chars, must include uppercase, lowercase, and numeric) | Generate a password; complexity rules apply even on free tier |

## Related Presets

- **01-serverless-oltp** -- Use instead for production OLTP workloads with private endpoint and auto-scaling
- **03-serverless-data-warehouse** -- Use instead for analytics and reporting workloads

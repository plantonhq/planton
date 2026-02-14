# Development PostgreSQL

This preset creates a single-node PostgreSQL database for development and testing. No VPC is required, and the smallest node size keeps costs minimal. Ideal for local development, CI/CD test databases, or staging environments.

## When to Use

- Development and testing environments
- CI/CD pipelines needing ephemeral databases
- Staging workloads where HA is not required
- Cost-sensitive scenarios

## Key Configuration Choices

- **Single node** (`nodeCount: 1`) -- no HA; suitable for non-production. Single point of failure is acceptable.
- **Smallest size** (`sizeSlug: db-s-1vcpu-1gb`) -- minimal cost; upgrade for heavier dev workloads.
- **No VPC** -- VPC omitted; cluster uses default networking. Add `vpc` for staging that mirrors production.
- **PostgreSQL 16** (`engine: pg`, `engineVersion: "16"`) -- match production version for consistency.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc3` | Target DigitalOcean region slug | [DigitalOcean Regions API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Regions) |

## Related Presets

- **01-postgresql-ha** -- Use for production workloads requiring HA and VPC isolation
- **03-redis** -- Use for caching instead of relational storage

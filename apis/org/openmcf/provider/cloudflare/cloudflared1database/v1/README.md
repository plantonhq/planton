# Cloudflare D1 Database

Provision and manage Cloudflare D1 databases using OpenMCF's unified API.

## Overview

Cloudflare D1 is a serverless SQLite database built for edge compute. Unlike traditional connection-based databases, D1 integrates seamlessly with Cloudflare Workers through bindings, offering a genuinely serverless experience with pay-per-query pricing and no idle costs.

This component provides a clean, protobuf-defined API for provisioning D1 databases, following the **80/20 principle**: exposing only the essential configuration fields that 80% of users need while keeping the API simple.

## Key Features

- **Serverless SQLite**: Built on SQLite, designed for edge-first applications
- **Multiple Regions**: Deploy databases close to your users (Western Europe, Eastern Europe, Asia Pacific, Oceania, Western/Eastern North America)
- **Read Replication (Beta)**: Enable automatic read replication for global applications
- **Automatic Backups**: 30-day Point-in-Time Recovery (PITR) via D1 Time Travel
- **Simple Configuration**: Just `account_id` and `database_name` to get started

## Prerequisites

1. **Cloudflare Account**: Active Cloudflare account with D1 access
2. **API Token**: Cloudflare API token with D1 permissions
3. **OpenMCF CLI**: Install from [openmcf.org](https://openmcf.org)

## Quick Start

### Minimal Configuration

Create a D1 database with the bare minimum configuration:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: my-dev-db
spec:
  account_id: "your-cloudflare-account-id"
  database_name: "my-app-dev-db"
```

Deploy:

```bash
planton apply -f database.yaml
```

### With Region Specification

Specify a region for optimal latency:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: my-prod-db
spec:
  account_id: "your-cloudflare-account-id"
  database_name: "my-app-production-db"
  region: "enam"  # Eastern North America
```

### With Read Replication (Beta)

Enable read replication for global applications:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: my-global-db
spec:
  account_id: "your-cloudflare-account-id"
  database_name: "my-app-global-db"
  region: "weur"
  read_replication:
    mode: "auto"
```

**⚠️ Important**: Enabling read replication requires updating your Worker code to use the [D1 Sessions API](https://developers.cloudflare.com/d1/worker-api/) to maintain data consistency.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | string | Your Cloudflare account ID (required) |
| `database_name` | string | Unique name for the database (max 64 characters, required) |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Geographic region for the primary database. Valid values: `weur`, `eeur`, `apac`, `oc`, `wnam`, `enam`. If omitted, Cloudflare selects a default. |
| `read_replication` | object | Read replication configuration (Beta). Contains `mode` field: `"auto"` or `"disabled"`. |

### Supported Regions

- `weur` - Western Europe
- `eeur` - Eastern Europe
- `apac` - Asia Pacific
- `oc` - Oceania
- `wnam` - Western North America
- `enam` - Eastern North America

## Outputs

After deployment, the following outputs are available:

- `database_id`: The unique identifier of the created D1 database
- `database_name`: The name of the database (same as input)

A Worker reaches D1 through its `d1` binding (referencing `database_id`); there is
no connection string.

Access outputs:

```bash
planton output database_id
planton output database_name
```

## Schema Management

**Important**: This component provisions the database **resource** (the container). It does **not** create tables or manage schema.

Schema management is handled exclusively via Cloudflare's Wrangler CLI:

```bash
# Create a migration
npx wrangler d1 migrations create my-app-db create_users_table

# Apply migrations
npx wrangler d1 migrations apply my-app-db --remote
```

This architectural separation is by design. See [docs/README.md](docs/README.md) for detailed explanation of the "Orchestration Gap."

## Multi-Environment Pattern

### Separate Databases Per Environment

Preview environments are **not** a database property. They are a binding-level pattern. Create separate databases for each environment:

**Development Database**:
```yaml
metadata:
  name: my-app-dev
spec:
  database_name: "my-app-dev-db"
  region: "weur"
```

**Production Database**:
```yaml
metadata:
  name: my-app-prod
spec:
  database_name: "my-app-production-db"
  region: "enam"
  read_replication:
    mode: "auto"
```

Configure Worker bindings (`wrangler.toml`) to point to the appropriate database based on deployment context.

## Common Use Cases

### 1. Development Database

Simple database for local development:

```yaml
spec:
  account_id: "abc123..."
  database_name: "my-app-dev-db"
  region: "weur"
```

### 2. Production Database with Global Replication

Database optimized for global users:

```yaml
spec:
  account_id: "abc123..."
  database_name: "my-app-production-db"
  region: "enam"
  read_replication:
    mode: "auto"
```

### 3. Preview/Staging Database

Separate database for preview deployments:

```yaml
spec:
  account_id: "abc123..."
  database_name: "my-app-preview-db"
  region: "wnam"
```

## Best Practices

1. **Use Descriptive Names**: Name databases clearly: `my-app-prod-db`, `my-app-preview-db`
2. **Choose Regions Wisely**: Deploy databases close to your primary user base
3. **Plan for Replication**: Only enable read replication when you're ready to refactor Worker code to use Sessions API
4. **Separate Environments**: Create distinct databases for dev/preview/prod
5. **Version Control Configs**: Store database manifests in git alongside application code
6. **Use Wrangler for Schema**: Manage schema via Wrangler migrations, not at the database resource level

## What This Component Does NOT Include

Following the 80/20 principle, these fields are **intentionally excluded**:

- ❌ `preview_branch`: Architecturally incorrect. Preview environments are a Worker binding pattern, not a database property.
- ❌ `primary_key`: Schema-level construct. Managed via Wrangler migrations, not at the resource level.

## Backup and Recovery

D1 automatically provides **Time Travel** (Point-in-Time Recovery):

- **Retention**: 30 days for paid plans, 7 days for free plans
- **Automatic**: No configuration needed
- **Restoration**: Via Wrangler CLI

```bash
# Restore to specific timestamp
npx wrangler d1 time-travel restore my-app-db --timestamp=2025-11-01T12:00:00Z
```

**Note**: Restoration is destructive and in-place. It rewinds the database to the specified point in time.

## Monitoring

Monitor your D1 database via:

1. **Cloudflare Dashboard**: View query volume, latency (p50, p90, p95), and storage size
2. **GraphQL Analytics API**: Programmatic access to metrics for integration with Grafana, Datadog, etc.
3. **Worker Logging**: Log query execution times in your Worker code for detailed debugging

## Troubleshooting

### "Database already exists" Error

The database name must be unique within your Cloudflare account. Choose a different name or delete the existing database.

### Schema Not Created After Deployment

This is expected. The component provisions the database resource only. Create tables via Wrangler migrations:

```bash
npx wrangler d1 migrations apply my-app-db --remote
```

### Read Replication Causing Data Inconsistency

Ensure your Worker code uses the [D1 Sessions API](https://developers.cloudflare.com/d1/worker-api/). Without it, replicas may serve stale data.

## Examples

For detailed usage examples, see [examples.md](examples.md).

## Architecture Details

For in-depth architectural guidance, deployment methods comparison, and production best practices, see [docs/README.md](docs/README.md).

## Terraform and Pulumi

This component supports both Pulumi (default) and Terraform:

- **Pulumi**: `iac/pulumi/` - Go-based implementation
- **Terraform**: `iac/tf/` - HCL-based implementation

Both produce identical infrastructure. Choose based on your team's preference.

## Support

- **Documentation**: [docs/README.md](docs/README.md)
- **Cloudflare D1 Docs**: [developers.cloudflare.com/d1](https://developers.cloudflare.com/d1)
- **Wrangler CLI Guide**: [developers.cloudflare.com/workers/wrangler](https://developers.cloudflare.com/workers/wrangler)
- **OpenMCF**: [openmcf.org](https://openmcf.org)

## License

This component is part of OpenMCF and follows the same license.


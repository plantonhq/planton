# Zalando Postgres Operator with Cloudflare R2 Backup

This preset deploys the Zalando Postgres Operator with WAL-G continuous archiving to Cloudflare R2 storage. All PostgreSQL clusters managed by this operator instance will use the configured R2 bucket for base backups and WAL archiving, enabling point-in-time recovery.

## When to Use

- Production PostgreSQL clusters where data durability is critical
- You use Cloudflare R2 for S3-compatible object storage
- You need automated daily base backups with continuous WAL archiving
- You want point-in-time recovery and clone-from-backup capabilities

## Key Configuration Choices

- **Backup schedule** (`0 2 * * *`) -- daily base backup at 2:00 AM UTC; adjust to your maintenance window
- **WAL-G backup enabled** (`true`) -- enables continuous WAL archiving to R2 between base backups
- **WAL-G restore enabled** (`true`) -- allows restoring PostgreSQL clusters from R2 backups
- **Clone WAL-G restore enabled** (`true`) -- allows cloning a PostgreSQL cluster from an R2 backup (useful for staging environments)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-cloudflare-account-id>` | Cloudflare account ID (used to construct the R2 endpoint URL) | Cloudflare dashboard > Account ID |
| `<your-r2-backup-bucket>` | Name of the R2 bucket for storing PostgreSQL backups | Cloudflare dashboard > R2 > Buckets |
| `<your-r2-access-key-id>` | R2 API access key ID | Cloudflare dashboard > R2 > Manage API Tokens |
| `<your-r2-secret-access-key>` | R2 API secret access key | Cloudflare dashboard > R2 > Manage API Tokens |

## Related Presets

- **01-standard** -- Deploys the operator without backup configuration

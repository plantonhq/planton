# Bucket and Backups

Stigmer with its files in an S3-compatible bucket and its bundled database archived there from its first minute, with a daily full backup kept for 30 days and a standby instance. A SeaweedFS in the same environment provides the bucket by reference; AWS S3, Cloudflare R2 and MinIO work the same way.

## When to Use

- Any install whose records you want back after losing the cluster
- Before turning on per-session sandboxes, which need the bucket

## Key Configuration Choices

- **`object_store.s3`** — the server and the database's backups use the key pair; no runner ever holds it (runners get short-lived signed URLs)
- **`database.bundled.backup`** — the schedule is a six-field cron with seconds first, in UTC
- **cert-manager on the cluster** — the backup plugin needs it, and nothing in this preset installs it; deploy a Cert Manager first
- **Save the keys** — run the `keys_export_command` output once and keep the file outside the cluster: a restore into a new cluster needs it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `bucket` | The bucket's name | Create it in your store first |
| `storage` | Your KubernetesSeaweedFs | The SeaweedFS resource in this environment |
| `access_key_id`, `$secret/stigmer-bucket-secret-key` | The bucket's key pair | Your store's credentials |

## Related Presets

- **05-per-session-sandboxes** — sandboxes that read and write through this bucket

# Backups to Cloudflare R2

The zero-config platform whose own database archives continuously to a
Cloudflare R2 bucket you own — every write as WAL, a base backup nightly,
30 days kept — declared entirely by reference to the bucket and token
resources that the catalog created, so no key is ever typed. The same
declaration, with `recover_from` added, brings the platform back from that
archive as itself: every record and every user, not a fresh install.

## When to Use

- Any platform whose records you would mind losing — organizations,
  environments, connections, projects, pipeline history, members, the
  identity realm — which is every platform a team runs
- An archive deliberately OUTSIDE the cluster's cloud provider, so no
  provider mishap can take the platform and its recovery path down
  together (R2 has no egress fees and speaks S3)

## Prerequisites

- A `CloudflareR2Bucket` for the archive (its `jurisdiction` is fixed at
  creation and decides which host serves it) and a
  `CloudflareAccountApiToken` scoped to that bucket with `Workers R2
  Storage Bucket Item Write` — both declared in the catalog, in the same
  InfraChart or another the platform can reference
- cert-manager on the cluster: the operator installs CloudNativePG's
  backup engine (the Barman Cloud plugin) beside CloudNativePG and needs
  it for the TLS between the two
- Operator chart 0.15.0 or newer (the first whose definition knows
  `backup`)

## Key Configuration Choices

- **Everything by reference** — `account_id` and `jurisdiction` follow the
  bucket's outputs; `access_key_id` and `secret_access_key` follow the
  token's. The module materializes the credential as a Secret
  (`<platform>-postgres-backup-creds`) before the platform, so the database
  is born archiving, and a rotated token is a new key pair the Secret
  follows on the next apply
- **One path per platform, shared safely** — the archive is filed beneath
  `destination_path` under the platform's own server name
  (`status.backup.serverName`), so several platforms can share a bucket
  and a path
- **The first base backup is immediate** — `schedule` is a SIX-field cron
  (seconds first) for the nightly ones; WAL alone restores nothing, so the
  platform is protected from the first minute, not the first night
- **Retention is the store's** — `retention_policy: 30d` prunes base
  backups and their WAL after each backup; `8w` and `6m` are the other
  units
- **Watch the column** — `kubectl get plantonplatform` reads `BACKUP`
  `Healthy` when archiving works and `Failing` in the plugin's own words
  when it does not; a failing backup never takes a working platform out of
  Ready

## The Boundary

Only the platform's PostgreSQL is covered. The secrets manager (OpenBAO)
keeps its data on its own volume outside this archive, so a restored
platform comes back with every record and an empty secrets manager: the
credentials behind its connections are re-entered after a restore.

## Placeholders to Replace

- `acme-platform-backups` — your `CloudflareR2Bucket` resource's name (and,
  in `destination_path`, its `bucket_name`)
- `acme-platform-backups-writer` — your `CloudflareAccountApiToken`
  resource's name

## Related Presets

- **01-zero-config** — the same platform with no backup; the database
  lives on one volume in the cluster
- **02-ingress-tls** — expose the platform at a real hostname; combine
  with this preset for a team-ready install that survives its cluster
- **03-eks** — the EKS-shaped variant; swap the `r2` arm for `s3` with
  `keyless: true` and an IRSA role in `service_account_annotations` to
  archive to S3 without a key

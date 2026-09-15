# Backups to Cloudflare R2

The zero-config platform whose own database archives continuously to a
Cloudflare R2 bucket you own — every write as WAL, a base backup nightly,
30 days kept — declared entirely by reference to the bucket and token
resources that the catalog created, so no key is ever typed. The archive
carries the bundled secrets manager with the records, because the vault
stores its data in that same database; the one thing it cannot carry is
the keys that open the vault, so the preset names a Secret you own for
them. The same declaration, with `recover_from` added and that Secret in
place, brings the platform back from the archive as itself: every record,
every user, and every secret, not a fresh install.

## When to Use

- Any platform whose records you would mind losing — organizations,
  environments, connections, projects, pipeline history, members, the
  identity realm, and the credentials behind every connection — which is
  every platform a team runs
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
- The operator chart whose definition knows `backup` and the vault's
  `init_secret_name` (the GUIDE names the floor)

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
- **The vault's keys outlive the platform** — `vault.init_secret_name`
  names a Secret you own; the operator writes the vault's unseal keys and
  root token into it at first boot and never deletes it, so a platform
  destroy leaves it standing. A backup with neither this nor a cloud seal
  is refused: the archive would carry every secret and no way to open it

## What Comes Back

Everything the platform held: every record the control plane keeps, the
identity realm with its users, and the secrets manager's contents — every
connection credential, every managed secret, the license and OIDC signing
keys — because the vault stores in the same database the archive carries.
On the bad day, recreate `planton-vault-keys` in the new cluster from the
copy you kept, declare the platform again with `recover_from`, and the
operator unseals the restored vault with it. `status.backup.vault` says
what the archive covers and names the Secret to keep. What no archive
brings back is that Secret itself when it was lost with the cluster and
never copied — so copy it. A cloud key through `vault.auto_unseal` is the
other posture: the restored vault opens itself with no Secret to carry.

## Placeholders to Replace

- `acme-platform-backups` — your `CloudflareR2Bucket` resource's name (and,
  in `destination_path`, its `bucket_name`)
- `acme-platform-backups-writer` — your `CloudflareAccountApiToken`
  resource's name
- `planton-vault-keys` — the name you want the vault's keys Secret to have;
  the operator creates it, you keep a copy of it

## Related Presets

- **01-zero-config** — the same platform with no backup; the database
  lives on one volume in the cluster
- **02-ingress-tls** — expose the platform at a real hostname; combine
  with this preset for a team-ready install that survives its cluster
- **03-eks** — the EKS-shaped variant; swap the `r2` arm for `s3` with
  `keyless: true` and an IRSA role in `service_account_annotations` to
  archive to S3 without a key

---
title: "Backup and Recovery"
description: "Back up your self-hosted Planton's database to an object store you own — S3, GCS, Azure Blob, or Cloudflare R2 — and bring the whole platform back from it on any cluster"
icon: security
order: 25
tags:
  - Self-Hosting
  - Backup
  - Recovery
  - Kubernetes
---

# Backup and Recovery

A self-hosted Planton is, at bottom, one PostgreSQL database. Organizations, environments, cloud connections, projects, deployment history, members, and the sign-in realm with its users are all rows in it. Declare a backup and every change streams to an object store you own, with a full copy taken on a schedule, so the platform can be restored to any moment inside the retention window — on the same cluster or a new one. Declare nothing and it runs exactly as before, with no archive.

## What is in the backup, and what is not

**In it:** everything the platform's database holds. After a restore, the organizations are there, the environments and their cloud resources are there, the deployment history is there, and people sign in with the passwords they had, because the identity realm is in the same database.

**Not in it:** the secrets manager. OpenBAO keeps its data on its own volume, outside this backup. Every record that points at a secret comes back; the value behind it does not. After a restore, the credentials behind your cloud connections and any config secrets are re-entered. Also outside: the cache (rebuilt on start), the state files of the infrastructure you deployed (they live in your state backend, not in Planton's database), and running workloads on other clusters (untouched; they reconnect).

## Declare the backup

The backup is one block on the platform resource, `database.postgresql.backup`, and it names an object store in the backend's own vocabulary. On Cloudflare R2 nothing is typed: the bucket and the credential are catalog resources too, and the platform references their outputs.

Four resources, applied in order. First the bucket:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareR2Bucket
metadata:
  name: acme-platform-backups
spec:
  bucketName: acme-platform-backups
  accountId: <your 32-character Cloudflare account id>
  jurisdiction: default
```

Then a token scoped to that one bucket with the `Workers R2 Storage Bucket Item Write` permission group. Permission groups are Cloudflare's own identifiers; fetch the list with `GET /accounts/{account_id}/tokens/permission_groups` and pin the UUID for that group:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAccountApiToken
metadata:
  name: acme-platform-backups-writer
spec:
  account_id: "<your account id>"
  name: acme-platform-backups-writer
  policies:
    - effect: allow
      permission_group_ids:
        - "<UUID of Workers R2 Storage Bucket Item Write>"
      resources:
        "com.cloudflare.edge.r2.bucket.<your account id>_default_acme-platform-backups": "*"
  status: active
```

The token resource exports itself as the S3 key pair R2 authenticates, so the platform never sees the token's value. Then the platform, with the backup block referencing both:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  create_namespace: true
  version: v0.0.62
  database:
    postgresql:
      backup:
        object_store:
          destination_path: s3://acme-platform-backups/platform
          r2:
            account_id:
              valueFrom:
                kind: CloudflareR2Bucket
                name: acme-platform-backups
                fieldPath: status.outputs.account_id
            jurisdiction:
              valueFrom:
                kind: CloudflareR2Bucket
                name: acme-platform-backups
                fieldPath: status.outputs.jurisdiction
            credentials:
              access_key_id:
                valueFrom:
                  kind: CloudflareAccountApiToken
                  name: acme-platform-backups-writer
                  fieldPath: status.outputs.r2_access_key_id
              secret_access_key:
                valueFrom:
                  kind: CloudflareAccountApiToken
                  name: acme-platform-backups-writer
                  fieldPath: status.outputs.r2_secret_access_key
        retention_policy: 30d
        schedule: "0 0 2 * * *"
```

A few things to know about this block:

- `destination_path` is the bucket plus a path. Several platforms can share one path: each files its archive beneath it under its own server name (below), so they never touch each other's.
- `retention_policy` is a number of days, weeks, or months (`30d`, `8w`, `6m`). Any moment inside the window is restorable; older backups and their WAL are pruned. Empty means 30 days.
- `schedule` is a **six-field** cron with seconds first — `0 0 2 * * *` is daily at 02:00 UTC. The five-field Kubernetes form is refused. The first full backup runs the moment the backup is declared, whatever the schedule says; the write stream alone restores nothing.
- The module writes the credential into a Kubernetes Secret before the platform resource exists, in the same apply, so the database is born archiving. Rotating the token is a new key pair the Secret follows on the next apply.

The other backends take the same block with a different arm. **S3**: `s3` with `region`, and either `keyless: true` (the database pods' IAM role through IRSA or an instance profile, bound with `service_account_annotations` on the backup block) or `access_keys` as references to your organization secrets; an `endpoint_url` makes it any S3-compatible store, which then needs access keys. **GCS**: `gcs` with `keyless: true` (Workload Identity; the identity needs `roles/storage.objectAdmin` and `roles/storage.legacyBucketReader`) or a `service_account_key_json` reference, and a `gs://` destination. **Azure Blob**: `azure_blob` with `storage_account` and either `keyless: true` (Workload Identity) or a `connection_string` reference, and an `https://` container URL as the destination.

Declaring a backup on a running platform is an ordinary edit: archiving starts at the next reconcile, no reinstall. On a fresh install with a backup declared, the operator waits for the referenced credential to exist before it starts the database, rather than starting it unarchived; the status says what it is waiting for.

The full declaration is the catalog kind's [backups-to-R2 preset](https://github.com/plantonhq/planton/blob/main/catalog/kubernetes/kubernetesplantonplatform/presets/06-backups-to-r2.yaml), and the kind's [guide](https://github.com/plantonhq/planton/blob/main/catalog/kubernetes/kubernetesplantonplatform/GUIDE.md) has the operator's reasoning. The console's platform wizard has a Backups step that writes this block for you; screenshots follow the next platform release.

> **Note:** The backup needs the operator chart that installs CloudNativePG's backup engine (the Barman Cloud plugin) beside CloudNativePG, and that engine needs cert-manager on the cluster for its own certificate. On a cluster without cert-manager the status names it.

## Read the archive's state

`kubectl get plantonplatform -n planton` has a `BACKUP` column. `NotConfigured` means no backup is declared. `Deploying`, `Healthy`, and `Failing` are the engine's own words; a failing backup never takes a working platform out of Ready, so watch the column.

The detail is on the status:

```bash
kubectl get plantonplatform planton -n planton -o jsonpath='{.status.backup}'
```

It carries the **server name** the archive is filed under (the database's name plus the first eight characters of the platform's install id, so every install's archive is its own), the first recoverability point, the last successful full backup, and, when something is wrong, what the operator is waiting for. Copy the server name somewhere outside the cluster: a restore declares it.

## Restore

A restore is a declaration too. Declare the platform again — on the same cluster after a loss, or on a new one — with `database.postgresql.recover_from` naming the same store and the source's server name:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  create_namespace: true
  version: v0.0.62
  database:
    postgresql:
      recover_from:
        server_name: planton-postgres-3f9a1c2e
        object_store:
          destination_path: s3://acme-platform-backups/platform
          r2:
            # the same references as the backup block
        # target_time: "2026-09-13T20:30:00Z"   # a point in time; empty = the latest
      backup:
        # keep the backup block too, so the restored platform archives itself
```

What happens: the database is created from the archive rather than empty, replayed to the latest moment or to `target_time` (RFC 3339). The platform comes up with its organizations, users, environments, connections, and history. The sign-in admin works again: the operator re-establishes the administrator's credential after the restore, so the admin password in the new platform's Secret (`planton-identity-bootstrap-admin`) is the one that signs in, and every other user signs in with the password they had. The restored platform then archives under a **new** server name, so it never writes over the archive it restored from; keep `backup` declared beside `recover_from` and the two share one bucket and one path.

Two rules the operator holds:

- **A restore is honored at first creation only.** Once the platform's database exists, editing `recover_from` changes nothing; the status names the procedure. To restore, declare a new platform (or delete the lost one first). Nothing here ever destroys data to honor a declaration.
- **When the source is gone,** the server name is the folder under the destination path in the bucket's own listing.

After the restore, re-enter what the backup does not carry: the credentials behind your cloud connections and any config secrets, in the console where each connection lives.

## Requirements

- Operator chart `0.15.0` or newer for the backup; `0.16.0` or newer for a restore that re-establishes the sign-in admin on its own.
- cert-manager on the cluster (the backup engine's certificate).
- An object store the database pods can reach, and a credential for it — by reference, never typed into the manifest.

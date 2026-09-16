---
title: "Backup and Recovery"
description: "Back up your self-hosted Planton -- its records and its secrets -- to an object store you own, and bring the whole platform back from it on any cluster with nobody carrying keys"
icon: security
order: 25
tags:
  - Self-Hosting
  - Backup
  - Recovery
  - Kubernetes
---

# Backup and Recovery

A self-hosted Planton is, at bottom, one PostgreSQL database -- and its bundled secrets manager stores inside that same database. Organizations, environments, cloud connections, projects, deployment history, members, the sign-in realm with its users, and every secret the platform holds are all rows in it. Declare a backup and every change streams to an object store you own, with a full copy taken on a schedule, so the platform can be restored to any moment inside the retention window -- on the same cluster or a new one, records and secrets together, to the same instant. Declare nothing and it runs exactly as before, with no archive.

## What is in the backup, and what is not

**In it:** everything the platform's database holds. After a restore, the organizations are there, the environments and their cloud resources are there, the deployment history is there, and people sign in with the passwords they had, because the identity realm is in the same database.

**Also in it:** the secrets manager. The bundled vault stores its data in that same database, so the credentials behind your cloud connections, your config secrets, and the platform's signing keys come back with the records. Keyless cloud connections keep working after a restore because the platform's signing key is the same key.

**Not in it, by design:** the key that opens the vault. The vault's data is encrypted, and what decrypts it is either a key in your cloud or a Secret you own -- never anything in the archive. That is what makes the archive safe to keep anywhere, and it is the one thing you decide when you declare the backup. The section below is that decision.

**Also not in it:** the cache (rebuilt on start), the state files of the infrastructure you deployed (they live in your state backend, not in Planton's database), and running workloads on other clusters (untouched; they reconnect).

## Protect the vault's keys

A backup is refused unless the vault's keys outlive the platform. The platform resource refuses the declaration otherwise, in one sentence, because an archive that carries every secret with no way to open them is not a backup. Choose one of two ways.

**A key in your cloud.** `vault.auto_unseal` seals the vault with a key you already trust -- Google Cloud KMS, AWS KMS, Azure Key Vault, or a central OpenBao's transit engine -- and a restored vault opens itself, on a cluster that has never seen it, with no person involved. On Google Cloud the key is declared by reference to catalog resources you create beside the platform:

```yaml
  vault:
    auto_unseal:
      gcp_kms:
        project:
          value: my-gcp-project
        region: asia-south1
        key_ring:
          valueFrom:
            kind: GcpKmsKeyRing
            name: planton-vault-unseal
            fieldPath: status.outputs.key_ring_name
        crypto_key:
          valueFrom:
            kind: GcpKmsKey
            name: planton-vault-unseal
            fieldPath: status.outputs.key_name
        workload_identity_service_account:
          valueFrom:
            kind: GcpServiceAccount
            name: planton-vault-unseal
            fieldPath: status.outputs.email
    init_secret_name: planton-vault-keys
```

Three facts about a cloud key. The key and its grants must exist before the platform, because the vault checks its seal the moment its server starts; a key it cannot reach is a crash-loop the platform's status explains. On Google Cloud the seal identity needs two roles on the key, `roles/cloudkms.cryptoKeyEncrypterDecrypter` and `roles/cloudkms.viewer` -- the second is the one people forget, and without it the vault never starts. And the seal is decided once: a platform restored from this archive must declare the same key.

**A Secret you own.** Without a cloud key, the vault uses its built-in key shares, and `vault.init_secret_name` names a Kubernetes Secret in the platform's namespace that you own:

```yaml
  vault:
    init_secret_name: planton-vault-keys
```

The operator writes the vault's unseal keys and root token into that Secret at first boot, never deletes it, and never writes over a Secret that already holds a vault's keys. Deleting the platform leaves the Secret standing. But a namespace the platform resource created is deleted with the platform and takes every Secret in it -- so the day the platform is born, copy the Secret out of the cluster and keep it where your other break-glass material lives:

```bash
kubectl get secret planton-vault-keys -n planton -o yaml > planton-vault-keys.yaml
```

That file is what a restore into a new cluster needs. Under a cloud key the same Secret exists and holds the recovery keys and the root token instead; nothing running needs it, but it is your break-glass, so keep a copy of it too.

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

The token resource exports itself as the S3 key pair R2 authenticates, so the platform never sees the token's value. Then the platform, with the backup block referencing both and the vault's keys in a Secret you own:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  create_namespace: true
  version: v0.0.65
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
  vault:
    init_secret_name: planton-vault-keys
```

A few things to know about this block:

- `destination_path` is the bucket plus a path. Several platforms can share one path: each files its archive beneath it under its own server name (below), so they never touch each other's.
- `retention_policy` is a number of days, weeks, or months (`30d`, `8w`, `6m`). Any moment inside the window is restorable; older backups and their WAL are pruned. Empty means 30 days.
- `schedule` is a **six-field** cron with seconds first -- `0 0 2 * * *` is daily at 02:00 UTC. The five-field Kubernetes form is refused. The first full backup runs the moment the backup is declared, whatever the schedule says; the write stream alone restores nothing.
- The module writes the credential into a Kubernetes Secret before the platform resource exists, in the same apply, so the database is born archiving. Rotating the token is a new key pair the Secret follows on the next apply.
- `vault.init_secret_name` (or `vault.auto_unseal`) is required beside a backup, for the reason above.

The other backends take the same block with a different arm. **S3**: `s3` with `region`, and either `keyless: true` (the database pods' IAM role through IRSA or an instance profile, bound with `service_account_annotations` on the backup block) or `access_keys` as references to your organization secrets; an `endpoint_url` makes it any S3-compatible store, which then needs access keys. **GCS**: `gcs` with `keyless: true` (Workload Identity; the identity needs `roles/storage.objectAdmin` and `roles/storage.legacyBucketReader`, and its binding names the database's ServiceAccount, `<platform>-postgres`) or a `service_account_key_json` reference, and a `gs://` destination. **Azure Blob**: `azure_blob` with `storage_account` and either `keyless: true` (Workload Identity) or a `connection_string` reference, and an `https://` container URL as the destination.

Declaring a backup on a running platform is an ordinary edit: archiving starts at the next reconcile, no reinstall. On a fresh install with a backup declared, the operator waits for the referenced credential to exist before it starts the database, rather than starting it unarchived; the status says what it is waiting for.

The full declaration is the catalog kind's [backups-to-R2 preset](https://github.com/plantonhq/planton/blob/main/catalog/kubernetes/kubernetesplantonplatform/presets/06-backups-to-r2.yaml), and the kind's [guide](https://github.com/plantonhq/planton/blob/main/catalog/kubernetes/kubernetesplantonplatform/GUIDE.md) has the operator's reasoning, the complete Google Cloud resource set for a keyless backup with a Cloud KMS seal, and the bad-day runbook. The console's platform wizard writes this block for you: the Backups step declares the archive and the Secrets step decides what opens the vault; screenshots follow the next platform release.

> **Note:** The backup needs the operator chart that installs CloudNativePG's backup engine (the Barman Cloud plugin) beside CloudNativePG, and that engine needs cert-manager on the cluster for its own certificate. On a cluster without cert-manager the status names it.

## Read the archive's state

`kubectl get plantonplatform -n planton` has a `BACKUP` column. `NotConfigured` means no backup is declared. `Deploying`, `Healthy`, and `Failing` are the engine's own words; a failing backup never takes a working platform out of Ready, so watch the column.

The detail is on the status:

```bash
kubectl get plantonplatform planton -n planton -o jsonpath='{.status.backup}'
```

It carries the **server name** the archive is filed under (the database's name plus the first eight characters of the platform's install id, so every install's archive is its own), the first recoverability point, the last successful full backup, and, when something is wrong, what the operator is waiting for. Copy the server name somewhere outside the cluster: a restore declares it.

It also carries `vault`, the archive's answer for the secrets manager: `covered` (the archive carries the vault's data), `seal` (`shamir` for the built-in key shares, or the cloud seal's name), `initSecretName` (the Secret to keep a copy of), and a sentence for your posture. Read it once after declaring the backup; it tells you the one thing to keep.

## Restore

A restore is a declaration too. Declare the platform again -- on the same cluster after a loss, or on a new one -- with `database.postgresql.recover_from` naming the same store and the source's server name, the same `vault` block the source had, and `backup` beside it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  create_namespace: true
  version: v0.0.65
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
  vault:
    init_secret_name: planton-vault-keys   # the same Secret name the source had
```

What happens: the database is created from the archive rather than empty, replayed to the latest moment or to `target_time` (RFC 3339). The platform comes up with its organizations, users, environments, connections, and history. The sign-in admin works again: the operator re-establishes the administrator's credential after the restore, so the admin password in the new platform's Secret (`planton-identity-bootstrap-admin`) is the one that signs in, and every other user signs in with the password they had. The vault comes back with every secret, and the operator gives the platform a fresh vault credential of its own, so nothing is carried across by hand. The restored platform then archives under a **new** server name, so it never writes over the archive it restored from; keep `backup` declared beside `recover_from` and the two share one bucket and one path.

**The vault's step, by seal:**

- **A key in your cloud.** Nothing to carry. The key, its two grants, and the vault's Workload Identity binding must exist for the new cluster (the ring and the key never died; the binding is per cluster), and the restored vault opens itself from your key. Then recreate the keys Secret from your copy: until you do, the platform works and the vault's status says so plainly -- the vault came back from the archive and your key opened it, but its break-glass (the root token and recovery keys) is gone until the Secret is back.
- **A Secret you own.** Recreate the Secret in the new cluster's namespace from the copy you kept, under the same name, before you declare the platform (or the moment the namespace exists -- the operator restores the database first and needs the Secret only after). The operator unseals the restored vault with the keys in it. If you forget, nothing is lost: the vault's status refuses in one sentence that names the archive, the Secret, its keys, and this step, and the next pass unseals once the Secret is back.

Two rules the operator holds:

- **A restore is honored at first creation only.** Once the platform's database exists, editing `recover_from` changes nothing; the status names the procedure. To restore, declare a new platform (or delete the lost one first). Nothing here ever destroys data to honor a declaration.
- **The seal is decided once.** A restore must declare the seal the source had -- the same cloud key, or the built-in seal with the same Secret. A different seal is refused before anything runs, because the vault's data can only be opened by the key that sealed it.
- **When the source is gone,** the server name is the folder under the destination path in the bucket's own listing.

After the restore there is nothing to re-enter. Your cloud connections, config secrets, and signing keys came back with the records.

## Requirements

- An operator chart that carries the vault in the archive: the chart's release notes name the first version that does, and this page will name it here when it ships. Earlier charts back up the database alone and leave the secrets manager outside the archive.
- Operator chart `0.16.3` or newer for the database backup itself. Earlier charts declared the same fields, but their backup engine could not finish installing (a permission the operator lacked), a restore could come back as an empty database when that happened, and a restored platform's sign-in server could stall; `0.16.3` is the first chart on which the database's backup and restore have been proven end to end on a live platform.
- cert-manager on the cluster (the backup engine's certificate).
- An object store the database pods can reach, and a credential for it -- by reference, never typed into the manifest.
- For a cloud seal: the key, its grants, and the vault's identity, created before the platform. For the built-in seal: the keys Secret, copied out of the cluster the day the platform is born.

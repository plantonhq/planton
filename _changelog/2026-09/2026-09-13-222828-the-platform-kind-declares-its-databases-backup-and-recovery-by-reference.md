# The platform kind declares its database's backup and recovery, by reference

**Date**: September 13, 2026
**Type**: Feature
**Components**: KubernetesPlantonPlatform (catalog kind, Terraform and Pulumi modules), Presets, Reference Docs

## Summary

`KubernetesPlantonPlatform` gains `database.postgresql.backup`, `database.postgresql.recover_from`, and `prerequisites.postgres_backup_plugin` — the declaration the operator has accepted since chart 0.15.0, in the catalog's own vocabulary. A self-hosted Planton now says, in the manifest that declares it, where its own database is backed up (an S3, GCS, Azure Blob, or Cloudflare R2 bucket the adopter owns, a schedule, a retention) and, when the day comes, which archive to restore itself from. On R2 the whole declaration is composed from other catalog resources: the account and jurisdiction follow a `CloudflareR2Bucket`'s outputs, the credential follows a bucket-scoped `CloudflareAccountApiToken`'s S3 key pair, and no key is typed. Both engines materialize the credential as a Secret before the platform resource, in the same apply, so the database is born archiving.

## Problem Statement / Motivation

The operator could back up the platform's database and restore it (`spec.database.postgresql.backup` with an `r2` arm, `recoverFrom`, the `BACKUP` column) but the catalog kind could not declare either, and `planton chart build` is schema-strict — so a platform declared from the catalog, which is every platform the desktop and the console install, had a database that lived on one volume in one namespace with `BACKUP NotConfigured` and no way to say otherwise. The sibling `KubernetesPostgres` kind already spoke a complete object-store vocabulary with R2 by reference; the platform kind did not.

### Pain Points

- The `controls.yaml` entry for backup-recovery read "not modeled on this spec yet"; the honest reinstall guidance was "credentials and database volumes die together".
- An adopter who wanted the backup the operator offers had to bypass the catalog and hand-write the CR, then hand-create the credentials Secret, in the right order.
- The catalog had two ways to say "R2" only if the platform kind copied the CRD's Secret-name shape instead of the PostgreSQL kind's declared-credential shape.

## What Changed

### The schema (`v1alpha1/spec.proto`)

- `KubernetesPlantonPlatformPostgresql` gains `backup = 4` and `recover_from = 5`.
- `...PostgresqlBackup`: `object_store` (required), `retention_policy` (default `30d`, `^[1-9][0-9]*[dwm]$`), `schedule` (default `0 0 2 * * *`; a six-field cron, seconds first — the operator's CRD carries no pattern here, so the catalog is the only place a five-field cron is caught with a sentence), `service_account_annotations` (the kind's existing map shape, where a keyless posture binds the database pods' cloud identity).
- `...PostgresqlRecoverFrom`: `object_store` (required), `server_name` (required), `target_time` (RFC 3339, checked in the catalog since the operator does not).
- `...ObjectStore`: `destination_path` plus exactly one of `s3`, `gcs`, `azure_blob`, `r2`, with the path scheme pinned per arm — the `KubernetesPostgres` kind's rules and wording, so a platform's backup and a database's backup read alike.
- The arms mirror the operator's postures: S3 keyless or access keys (plus an optional private-CA PEM for S3-compatible endpoints), GCS keyless or a service-account key, Azure Blob keyless or a connection string (the operator honors no account-plus-key posture, so the catalog offers none), and R2 with `account_id` and `jurisdiction` as references onto `CloudflareR2Bucket` and `credentials` as references onto `CloudflareAccountApiToken`'s `r2_access_key_id` / `r2_secret_access_key`.
- `KubernetesPlantonPlatformPrerequisites` gains `postgres_backup_plugin` (`auto` | `skip`).
- Doc comments carry what a manifest author needs at the field: what comes back after a restore and what does not (the secrets manager's data is outside the archive), the server-name rule, that R2 has no keyless posture, the GCS two-roles fact.

### Both engines

- Terraform (`iac/tf`): the mirrored `optional(object({...}))` shapes in `variables.tf` (cross-checked against the generator's output for the kind — identical); one rendering context per declared store in `locals.tf`; the credentials Secret's data per arm under the operator's key names (`ACCESS_KEY_ID`/`SECRET_ACCESS_KEY`, `APPLICATION_CREDENTIALS`, `AZURE_STORAGE_CONNECTION_STRING`), null for a keyless posture; the endpoint-CA Secret for an S3 arm with `endpoint_ca_pem`; the CR's `objectStore` per arm with `credentialsSecretName` only when a Secret exists; `kubernetes_secret_v1` resources for `<platform>-postgres-backup-creds`, `-recovery-creds`, `-backup-endpoint-ca`, `-recovery-endpoint-ca`, created after the namespace and before the CR (`depends_on`).
- Pulumi (`iac/pulumi/module`): `object_store_secrets.go` computes and creates the same Secrets (a pure listing function plus the resource creation); `platform_cr.go` renders `backup`, `recoverFrom`, and `prerequisites.postgresBackupPlugin` in the same shape; `main.go` adds the Secrets to the CR's `DependsOn`.
- The R2 endpoint is the operator's to compose from account and jurisdiction, so neither engine carries an R2 host table; account and jurisdiction pass through.

### Reader surfaces

- `reference.md` regenerated (and the catalog-wide `reference-graph.yaml`, which gains the four new foreign-key edges; the two Cloudflare kinds' pages gain their "referenced by" rows).
- `controls.yaml`: `backup-recovery` and `backup-retention-policy` move from `not_applicable` to `configurable`, with the PostgreSQL-only boundary named.
- `GUIDE.md`: "Back up the platform's own database, and bring it back" — the by-reference R2 story, the `BACKUP` column's words, the recovery declaration and the server-name rule, and the boundary stated once; the destroy section re-trued.
- `catalog.md`, `README.md`: the new fields, the two new consumed dependencies, a new common pattern, three new "works with" entries.
- Preset `06-backups-to-r2` (yaml + md): nightly, 30 days, everything by reference; pins the operator chart floor in a comment.
- `e2e/manifest.yaml` (the full-surface offline-proof input and the reference page's example): the R2 arm with self-evident placeholder literals — an offline plan cannot resolve a reference — and `postgres_backup_plugin: auto`.
- `cost.yaml`: an exclusion naming the object store's own bill.
- `_rules/component/presets/validate-planton-presets.mdc`: the checklist line that said presets never use `valueFrom:` (contradicted by 739 presets and the preset gate) now says the opposite for anything the catalog can mint.

## How to check

- `go test ./catalog/kubernetes/kubernetesplantonplatform/v1alpha1/` — 68 specs (28 new): a positive per posture, a negative per CEL rule with its message asserted.
- `go test ./catalog/kubernetes/kubernetesplantonplatform/iac/pulumi/module/` — the rendering contract: nothing renders when nothing is declared; the R2 backup names `<platform>-postgres-backup-creds`; recovery renders beside backup with its own Secret; keyless postures name no Secret; S3 with keys and a private CA materializes both Secrets.
- Offline renders from inside `iac/tf` and `iac/pulumi` with `--manifest ../../e2e/manifest.yaml`: both engines produce the same CR spec (`backup.objectStore.{destinationPath, r2.{accountId, credentialsSecretName, jurisdiction}}`, `retentionPolicy`, `schedule`; `prerequisites.postgresBackupPlugin: auto`) and one Secret `planton-postgres-backup-creds` in the platform's namespace; with `--manifest ../../presets/01-zero-config.yaml` neither renders a Secret or any backup key.
- `make protos` (buf lint, stubs, the protovalidate-java conformance gate over the new CEL, gazelle), `terraform validate`, `go test ./pkg/explain/refgen/`, `go test ./pkg/presetvalidity/`, `go test ./pkg/anatomy/`, `go test ./pkg/finops/estimategen/` — all green.

## Not in this change

- No live proof: the operator's backup and recovery path is exercised live by the first platform that declares it; the catalog's own e2e lane runs on a Kind cluster with no bucket.
- The console's platform detail page and the public self-hosting docs describe the declaration once a release carries the field.
- Two candidates recorded, not done: migrating this kind's `variables.tf` onto the generator's drift allowlist; one shared Opaque-Secret helper for the four Kubernetes modules that each carry their own.

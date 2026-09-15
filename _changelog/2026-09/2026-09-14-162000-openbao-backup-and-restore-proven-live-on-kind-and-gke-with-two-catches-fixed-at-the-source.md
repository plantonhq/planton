# OpenBao backup and restore proven live on kind and GKE, with two catches fixed at the source

**Date**: September 14, 2026
**Type**: Fix
**Components**: KubernetesOpenBao (Pulumi and Terraform modules, spec comments, guide, presets, E2E fixtures and seed script), Kubernetes E2E harness (gcp-gke batch), component update rule

## Summary

The vault's backup and restore story now has live evidence on both engines and both cluster classes: on kind, a Shamir vault backs up to the catalog's own SeaweedFS store and a fresh vault on a transit seal restores from it; on GKE, a vault sealed by a Cloud KMS key backs up keylessly to Google Cloud Storage through Workload Identity and, separately, to a Cloudflare R2 bucket by reference, and a fresh vault on the same key restores from each — the newest snapshot on GCS, a named one on R2 — with a blind Terraform import round-trip proposing no real change and the batch torn down to zero residue. The kind's E2E profile is green. Two things the live runs caught are fixed where they live rather than around them.

## What Changed

### The seal's identity needs two roles on the key, not one

The GCP KMS seal wrapper reads the key's metadata when the server configures its seal at start — a key-existence check — and `roles/cloudkms.cryptoKeyEncrypterDecrypter` carries only the encrypt and decrypt permissions. A vault whose identity held that one role, exactly as the field comment, the guide's resource-set table, two preset explainers, and the real-cluster batch documented, crash-looped on `Error configuring seal "gcpckms": ... Permission 'cloudkms.cryptoKeys.get' denied` before it could initialize. Every one of those places now names both roles — `cryptoKeyEncrypterDecrypter` to use the key and `roles/cloudkms.viewer` to read it — with the failure the missing one produces, and the batch grants both through two `GcpKmsKeyIamMember` resources.

### The backup job waits for its server before it acts

The snapshot container talks to the active-leader Service, which has no endpoints until a server is initialized and unsealed, and a fresh unseal, a rolling restart, or a leader election empties it for a few seconds. A run that fired inside that window died on `connection refused` in under a second; with `restartPolicy: OnFailure` and `backoffLimit: 3` the Job spent its retries in fifteen seconds and deleted its pod, taking every log line with it. The snapshot script now polls `bao status` (exit 0 only for an unsealed, initialized server) for up to five minutes before the first request, in both engines byte-identically; every failure explanation after that is unchanged.

### The proof explains itself

The seed script that stages the bad day prints the server's own log when init never opens, and captures each attempt's logs while a backup run lives, then prints the Job's conditions, the namespace's events, and the captured lines when a run fails — the record above came from exactly that. The source fixtures take a once-a-year schedule so the only snapshots in a lane are the ones the seed triggers by hand (a live source on the hourly default snapshotted mid-lane and `restore.latest` picked it; the CronJob controller refuses an impossible date, so New Year's midnight UTC is the rarest real moment).

### `latest` beside a live source

The `restore.latest` field, the guide's bad-day runbook, and the console's restore step now say what the lane learned: "newest under the prefix" is a moving target while the source is alive and snapshotting on its own schedule, so a clone or a migration rehearsal names a `snapshot_key`; the bad day, where the source is gone, is where `latest` is right.

### The update rule

Two teachings join the component update rule: an IAM grant named anywhere is verified against the client's first call in its SDK source, never inferred from the role's name; and a module-owned satellite that reaches its own server waits, bounded, for the server to answer before it acts.

## Verification

Kind, both engines: `with-backup` (the login recipe verbatim, a run from the CronJob, the store listed with the job's identity from inside the cluster, retention pruning a seeded stale object) and `behavioral-backup-restore` (the transit seal, the restore through the source's snapshot, marker A back and marker B absent, the init token gone, a replaced pod Ready alone). GKE, both engines: `gke-gcs-backup-restore` and `gke-r2-backup-restore` with the KMS seal; the Terraform GCS lane's blind import round-trip re-imported five resources with no real change proposed; `teardown.sh` then `audit.sh` reported zero batch residue beyond the KMS ring and key shells GCP keeps by design. Offline: spec tests, `e2e-build`, `e2e-vet`, the cross-engine script parity guard, both Pulumi packages built and vetted, `tofu validate`, the reference and proto-docs freshness gates, `validate-manifest` on the touched fixtures.

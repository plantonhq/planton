# OpenBao backs up its Raft state to an object store and restores a fresh cluster by declaration

**Date**: September 14, 2026
**Type**: Feature
**Components**: KubernetesOpenBao (catalog kind, Pulumi and Terraform modules, presets, guide, E2E), Kubernetes E2E harness (verifier, gcp-gke batch), repository guards

## Summary

`KubernetesOpenBao` gains a `backup` block and a `restore` block. A team declares where its vault's snapshots go — an S3 or S3-compatible bucket, Google Cloud Storage, Azure Blob, or Cloudflare R2, each in that store's own vocabulary and by reference to the catalog's bucket, identity, and token kinds — and both engines render a module-owned CronJob that logs in to OpenBao with its own ServiceAccount, streams a Raft snapshot with the server's own `bao` CLI, ships it with rclone, and prunes by age. On the bad day a fresh vault with the same seal key declares `restore` (a named snapshot, or the newest) and a one-shot Job installs it once the operator hands over the fresh cluster's initial root token. Every failure a person can meet prints what happened and the exact next step with the real names, and the never-proven chart snapshot agent the kind used to render is gone. The proof surface is authored end to end — four scenarios, their fixtures and seed script, the verifier's backup and restore proofs, the GKE batch's seal node set — and awaits its first live run.

## Problem Statement / Motivation

A vault holds the one copy of every secret its consumers depend on, and the kind's backup story was a thin chart block that named one S3 host, asked the operator to hand-create a Secret whose documented key names were wrong (the chart read `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`; the comment said `access_key` / `secret_key`, so a pod built from the docs failed on every run), had never run live, and could not restore. Postgres and MongoDB already kept the promise a stateful kind on this platform makes — backed up and brought back by declaration, to four stores, keyless where the cloud allows — and OpenBao did not.

### Pain Points

- One store (S3 through `s3cmd`), static keys only, no GCS, no Azure, no R2, no keyless posture anywhere.
- A Secret-key contract documented from the agent's README rather than the chart template that consumed it; the resulting CronJob could never have worked.
- No restore: the snapshots, had any landed, were a file with no declared way back into a cluster.
- Nothing in the E2E surface exercised any of it.

## What Changed

### The schema (`v1alpha1/spec.proto`)

- `snapshot_agent` (field 11) is reserved out; `backup = 14` and `restore = 15` arrive.
- `KubernetesOpenBaoBackup`: `schedule` (5-field cron, default hourly), `retention_days` (object age under the prefix, default 14, `0` keeps forever), `object_store` (a `prefix` plus exactly one of `s3` / `gcs` / `azure_blob` / `r2`), the shared `KubernetesWorkloadIdentity` for the job's keyless cloud identity, `auth` (the Kubernetes auth mount and role), `images` (mirror overrides for the `openbao` and `rclone` containers), `resources`.
- The arms carry the catalog's postures: `s3` keyless (IRSA) or access keys, with `endpoint_url` by reference to a `KubernetesSeaweedFs` for the S3-compatible case; `gcs` keyless (Workload Identity) or a service-account key by reference; `azure_blob` keyless, key, or connection string; `r2` with `bucket`, `account_id`, `jurisdiction` by reference to `CloudflareR2Bucket` and `credentials` by reference to `CloudflareAccountApiToken`'s S3 pair — byte-faithful to the MongoDB arm.
- `KubernetesOpenBaoRestore`: exactly one of `snapshot_key` or `latest`, plus `root_token` (the shared `KubernetesSecretKey`) naming the Secret the operator fills after `bao operator init`.
- Rules that refuse a broken manifest with a sentence: `backup` requires `server.ha` (snapshots exist only for Raft) and `service_account.auth_delegator_enabled`; a keyless arm requires `backup.workload_identity`; `restore` requires `backup` and an `auto_unseal` arm (the same key on source and target is what makes a restore one call); the R2 account id and jurisdiction formats.
- Field comments carry what a manifest author needs at the field: the four-command login recipe with the rendered names, the one-prefix-per-live-vault rule, that restore mode suspends the schedule, the two GCS bucket roles, that the seal is checked at server START and a missing or unreachable key crash-loops the pod, that a transit key is created on first use but the engine must exist.

### Both engines, byte-identical

- A ServiceAccount `<name>-backup` carrying the workload-identity annotation (and the AKS pod label on the jobs), a scripts ConfigMap `<name>-backup-scripts`, a credentials Secret `<name>-backup-credentials` for declared keys (absent on keyless arms), the CronJob `<name>-backup` (a `snapshot` init container on the server's image, an `upload` container on rclone's; `concurrencyPolicy: Forbid`; both containers as OpenBao's uid 100 so the upload can read the 0600 snapshot), and — when `restore` is declared — the Job `<name>-restore-<8hex>` (a `fetch` init container, a `restore` container that waits for the target to be initialized and unsealed, then `bao operator raft snapshot restore`), created without awaiting completion in both engines and with no finished-Job TTL, while the CronJob renders suspended.
- rclone is configured entirely through `RCLONE_CONFIG_STORE_*` environment variables (provider `Cloudflare` / `AWS` / `Other`, `no_check_bucket`, `bucket_policy_only` on GCS); no config file with credentials is ever rendered. The R2 host follows `pkg/cloudflare/r2` in Go and the same table in HCL.
- Four scripts, a Go constant and an HCL heredoc kept byte-identical: the snapshot script prints the login recipe with real names on every way the login can fail, the upload script names the exact permission each store needs on AccessDenied, the fetch script says what to list when a key is missing, the restore script quotes OpenBao's own same-key refusal and prints the two closing steps (remove `restore`, delete the Secret).
- Five new outputs (`backup_service_account_name`, `backup_policy_name`, `backup_auth_role`, `backup_cron_job_name`, `restore_job_name`); import-map rows for every satellite; the `batch/cronjobs` + `batch/jobs` permission row; the kind's name budget drops to 45 characters when backups are declared (CronJob names cap at 52).

### The proof surface (authored; the lanes have not yet run)

- Scenarios: `with-backup` (kind: Shamir single-node Raft backing up to the catalog's own SeaweedFS by reference), `behavioral-backup-restore` (kind: a fresh vault on a transit seal served by a dev-mode OpenBao key holder, restoring by snapshot key from a source vault the seed script initialized and backed up), `gke-gcs-backup-restore` (keyless GCS, GCP KMS seal, `restore.latest`), `gke-r2-backup-restore` (R2 by value from the batch, restore by key). Fixtures for the namespace, the store, the key holder, and the three source vaults; one seed script for every restore lane (init with recovery shares, the recipe verbatim, marker before, one snapshot, marker after, the key published).
- The verifier: an auto-unseal init path (`recovery_shares`; the server refuses `secret_shares` under any auto-unseal seal), presence-only verification for `fixture-*` manifests, a key-holder arm that enables `transit/` and proves an encrypt, THE BACKUP PROOF (recipe verbatim through the server's own `bao`, a run from the CronJob, the store listed from inside the cluster with the job's own identity through a probe Job built from the CronJob's pod template, retention proven against a seeded stale object), and THE RESTORE PROOF (the CronJob asserted suspended, the target initialized and its token handed over, the Job completed, marker A present and marker B absent through the source's token, the target's init token dead, a replaced pod back Ready with no unseal).
- The gcp-gke batch: a keyless backup identity with bindings on both vaults' backup ServiceAccounts and bucket roles, and a seal node set (server identity, its bindings, a permanent Cloud KMS key ring, a per-bootstrap key, the encrypter-decrypter grant); the audit asserts the residue GCP leaves by design.
- `e2e/manifest.yaml` exercises the s3 keyless (IRSA) arm — the one no lane can prove — so the offline plan and preview cover it.

### Reader surfaces

- `GUIDE.md`: how backups and restore work, the login recipe, the GKE and Cloudflare R2 resource-set tables, the bad-day runbook for a declared restore, and the by-hand runbook for a Shamir vault.
- Presets `04-gke-ha-gcs-backups` (KMS seal and keyless GCS by reference) and `05-production-ha-r2-backups` (R2 by reference, any cluster), with explainers; `catalog.md`, `README.md`, `cost.yaml`, `controls.yaml`, the identity-and-access-platform chart's README, and the regenerated reference pages.
- The update rule learns three timeless lessons: verify a spec comment's Secret keys against the template that reads them; never await or expire a Job that waits for a human; keep cross-engine script text as a marked pair.
- A new repository guard, `hack/guards/ensure_cross_engine_script_parity.sh` (with its lint workflow), diffs every `# parity:`-marked Terraform heredoc against its Go constant; OpenBao's four scripts and Locust's login backend are its first pairs.

## How to check

- `go test ./catalog/kubernetes/kubernetesopenbao/v1alpha1/` — every arm and posture valid, every refusal firing.
- `planton tofu plan` and `planton pulumi preview` against `e2e/manifest.yaml` and one manifest per arm and posture: the same objects from both engines, no credential value outside a Secret.
- `make e2e-build && make e2e-vet`; `go test ./catalog/kubernetes/aa_e2e/verify/`; `planton validate-manifest` on every scenario, fixture, and preset.
- `bash hack/guards/ensure_cross_engine_script_parity.sh` — five pairs byte-identical.
- The lanes: `go test -tags=e2e -run 'TestKubernetesOpenBao_(Pulumi|Terraform)' ./e2e/` on a kind cluster, and on the gcp-gke batch after `bootstrap.sh`.

## Not in this change

- The live runs of the four lanes (the profile stays `pending_proof` until they pass).
- The Azure Blob arm and the S3 keyless arm live: proven offline, deferred until an AKS or EKS batch exists.
- A declarative in-vault bootstrap (OpenBao's self-initialization) that would remove the login recipe — a different init posture for the whole kind, recorded as a candidate.
- The console's wizard and detail pages for the new blocks.

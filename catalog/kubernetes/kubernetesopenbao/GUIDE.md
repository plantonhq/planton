# KubernetesOpenBao Guide

The judgment this guide carries: when a user asks for Vault, this is the
catalog's answer — and unlike most components, deploying it is not the
end of the story: a fresh server sits NotReady BY DESIGN until a human
(or runbook) initializes it, and a proposal that does not say so hands
the user what looks like a failed deploy.

Substitutes for: Vault (Linux Foundation-governed secrets manager,
MPL-2.0 fork; speaks the Vault API — clients and integrations connect
unchanged, and the External Secrets store's Vault arm points at it
directly).

## Answering "give me Vault"

Propose this kind and say what you did: OpenBao is the catalog's
Vault-compatible secrets manager. Never fall back to a Helm-release
Vault chart — the substitution workflow is the
[catalog guide](../../_docs/GUIDE.md)'s first law.

## Say the init step out loud

Initialization and unsealing are runtime API operations no deployment
tool can perform declaratively — until they happen, pods report NotReady
on purpose (the full lifecycle, including why the Services stay
addressable for the init calls, is on [reference.md](v1alpha1/reference.md)). Two
composition consequences:

- The proposal must include the one-time `bao operator init` handoff —
  otherwise the user reads the NotReady pods as a broken deploy.
- Choose auto-unseal (cloud KMS arms, or the transit engine of another
  OpenBao/Vault) in the manifest so RESTARTS need no human; only the
  one-time initialization remains manual. Dev mode skips the ceremony
  entirely but is never for real secrets — the reference page is blunt
  about why.

## The self-hosted secrets chain

OpenBao is the backend that completes an in-cluster External Secrets
story with no cloud dependency: this component + a
[KubernetesClusterSecretStore](../kubernetesclustersecretstore/GUIDE.md)
whose Vault arm points at its endpoint + KubernetesExternalSecret
declarations in each consuming namespace. Every hop of that chain is a
typed, referenceable node.

## Backups and restore

A vault holds the one copy of every secret its consumers depend on, so a
proposal that deploys OpenBao without saying where its backups go is
incomplete. The kind's `backup` block declares scheduled Raft snapshots to
an object store — S3 (or any S3-compatible store), Google Cloud Storage,
Azure Blob, or Cloudflare R2 — each in that store's own vocabulary, by
reference to the catalog's bucket, identity, and token kinds; the `restore`
block declares a fresh cluster's recovery from that store. What the module
renders and what it asks of the operator:

- **A CronJob, `<name>-backup`, on the server's own image plus rclone.**
  Every run logs in to OpenBao with its own ServiceAccount (`<name>-backup`)
  through the Kubernetes auth method, streams a snapshot with the `bao` CLI,
  ships it to `<prefix>/<name>-<UTC timestamp>.snap`, and prunes objects
  under the prefix older than `retention_days`. Snapshots exist only for
  integrated Raft storage — `backup` requires `server.ha` (single-node Raft
  is `ha.replicas: 1`).
- **One prefix per live vault.** Retention prunes under the prefix, so two
  live vaults must never share one; a restore target deliberately declares
  its source's prefix, and that is the only sharing there is.
- **The login recipe is the one step the module cannot take.** OpenBao is
  sealed at deploy; the policy, auth mount, and role the job logs in with
  are API writes INSIDE the vault, after initialization. The recipe is four
  commands (next section); until it runs, every backup run fails and its log
  prints the recipe with the real names. The kind exports the names it
  rendered (`backup_service_account_name`, `backup_policy_name`,
  `backup_auth_role`, `backup_cron_job_name`) so the recipe is copy-paste
  from the outputs.
- **A declared restore requires auto-unseal — the SAME key on source and
  target.** A snapshot is protected by the seal key that took it; with the
  same KMS key or the same transit key on both sides a restore is a single
  call. Shamir clusters back up like any other; their restore is the manual
  runbook below.
- **Restore mode suspends backups.** While `restore` is declared the
  target's CronJob renders `suspend: true`: a fresh vault sharing the
  source's prefix would otherwise snapshot an empty vault into it, prune the
  source's snapshots, and let `restore.latest` pick its own empty snapshot.
  After the restore completes, remove `restore` and apply again to resume.
- **The seal is checked at server START.** Every seal backend reaches for
  its key while the server configures itself; a key that is missing or
  unreachable makes the pod crash-loop with "Error configuring seal". Create
  the KMS key (or enable the transit engine) and grant the identity BEFORE
  the vault, in the same dependency-ordered set.

Keyless where the cloud allows: `gcs.keyless` on GKE (Workload Identity),
`s3.keyless` on EKS (IRSA), `azure_blob.keyless` on AKS — each paired with
`backup.workload_identity`, the shared identity seam that annotates the job's
ServiceAccount. R2 has no keyless posture from any cluster; its credential
is a `CloudflareAccountApiToken`, referenced.

## The login recipe

Run once, after `bao operator init`, with a token that can manage auth
methods and policies (the initial root token works), against the vault's
API. The names are the kind's outputs; with the defaults they are all
`<name>-backup` and the auth mount is `kubernetes`:

```bash
bao policy write <name>-backup - <<'POLICY'
path "sys/storage/raft/snapshot" { capabilities = ["read"] }
POLICY
bao auth enable -path=kubernetes kubernetes
bao write auth/kubernetes/config kubernetes_host="https://kubernetes.default.svc:443"
bao write auth/kubernetes/role/<name>-backup \
  bound_service_account_names=<name>-backup \
  bound_service_account_namespaces=<namespace> \
  token_policies=<name>-backup token_ttl=1h
```

Taking a snapshot is a plain `read` on `sys/storage/raft/snapshot` — not a
sudo operation — so that one path is the whole policy. The Kubernetes auth
method validates the job's token through TokenReview, which is why `backup`
requires `service_account.auth_delegator_enabled` (the default). To take a
snapshot now rather than at the next schedule:
`kubectl create job --from=cronjob/<name>-backup -n <namespace> <name>-backup-now`.

After a restore, the restored state carries the SOURCE's role, bound to the
source's ServiceAccount name and namespace: a target with the same name and
namespace resumes backups untouched; a renamed one runs the recipe again.

## Disaster recovery on GKE: the resource set

"Sealed by a key no human holds, backed up keylessly, restorable by
declaration" is eight catalog resources on the GCP side and the Kubernetes
side together — every one a kind in this catalog, wired by reference. The
`gcp-gke` lane deploys exactly this set.

| # | Resource | What it is for | Wiring |
|---|---|---|---|
| 1 | `GcpServiceAccount` (e.g. `bao-unseal`) | The SERVER's identity: wraps and unwraps the master key — KEYLESS | — |
| 2 | `GcpServiceAccount` (e.g. `bao-backup`) | The BACKUP JOB's identity: writes and prunes snapshots — KEYLESS. Two identities on purpose: the seal key and the snapshot bucket are different blast radii | — |
| 3 | `GcpKmsKeyRing` | Holds the unseal key. Permanent by GCP design — it can never be deleted and its name is occupied forever | `location` = the region the vault runs in |
| 4 | `GcpKmsKey` | The unseal key. A declared restore needs the SAME key on source and target | `key_ring_id` by reference to #3; `deletion_policy: PREVENT` in production — destroying the key destroys every vault sealed by it |
| 5 | `GcpKmsKeyIamMember` | Lets the server use the key | `crypto_key_id` by reference to #4's `key_id`, `role: roles/cloudkms.cryptoKeyEncrypterDecrypter`, `member` by reference to #1's `member` |
| 6 | `GcpGcsBucket` | The snapshot store | `iam_members`: **two** roles for #2 — `roles/storage.objectAdmin` AND `roles/storage.legacyBucketReader` (rclone reads the bucket's attributes before writing; objectAdmin alone does not carry `storage.buckets.get`) — `member` by reference to #2's `member` |
| 7 | `GcpGkeWorkloadIdentityBinding` (two per vault) | Lets the KSAs act as the identities | for the server: `ksa_name` = the vault's `metadata.name` (the chart names the ServiceAccount after the release) bound to #1; for the job: `ksa_name` = `<name>-backup` bound to #2; `ksa_namespace` = the vault's namespace. A restore target is another vault and needs its own pair |
| 8 | `KubernetesOpenBao` (the production vault) | HA + auto-unseal + backups | `server.ha`, `auto_unseal.gcp_kms` with `key_ring` and `crypto_key` by reference to #3/#4 (bare names) and `workload_identity_service_account` by reference to #1, `backup.object_store.gcs.bucket` by reference to #6 with `keyless: true`, `backup.workload_identity.gke.service_account_email` by reference to #2, a `prefix` of its own |
| 9 | `KubernetesOpenBao` (the restore target, on the bad day) | Restore | the same `auto_unseal` (the same key), the same `backup` block INCLUDING the source's `prefix`, `restore.latest: true` (or a `snapshot_key`), `restore.root_token` naming the Secret you will create after init; its own #7 pair |

The validated manifests for this set are the `gcp-gke` lane's own:
`e2e/fixture-gke-gcs-source.yaml` (#8), `e2e/scenarios/gke-gcs-backup-restore.yaml`
(#9), and the GCP side under `../aa_e2e/realcluster/gcp-gke/manifests/`
(#1–#7). The `04-gke-ha-gcs-backups` preset is #8 as a starting point.

## Disaster recovery with Cloudflare R2: the resource set

The same story with the snapshots outside the cloud that runs the vault —
in a Cloudflare R2 bucket declared from the catalog. R2 speaks S3, so rclone
reaches it through its S3 code path, but the vault declares the store in
R2's own terms and the module does the translation (the jurisdiction's
endpoint host, region `auto`, path-style addressing, the token as an S3 key
pair). The seal is whatever the cluster offers (the KMS trio above on GKE);
the store is R2 from anywhere. The `gcp-gke` lane deploys this set beside
the GCS one.

| # | Resource | What it is for | Wiring |
|---|---|---|---|
| 1 | `CloudflareR2Bucket` | The snapshot store | `jurisdiction` fixed at creation (`default`, `eu`, `fedramp`, `us`) — it decides which host serves the bucket; exports `bucket_name`, `account_id`, `jurisdiction` |
| 2 | `CloudflareAccountApiToken` (e.g. `bao-snapshots-writer`) | The credential — R2 has NO keyless posture from any cluster | one policy: permission group `Workers R2 Storage Bucket Item Write` on resource `com.cloudflare.edge.r2.bucket.<account>_<jurisdiction>_<bucket>` (least privilege: objects in this bucket only); exports the token as the S3 key pair, `r2_access_key_id` + `r2_secret_access_key` |
| 3 | `KubernetesOpenBao` (the production vault) | HA + auto-unseal + backups | `server.ha`, an `auto_unseal` arm, `backup.object_store.r2` with `bucket`, `account_id`, `jurisdiction` by reference to #1 and `credentials` by reference to #2, a `prefix` of its own. No `backup.workload_identity` — nothing on the cluster side identifies the job to R2 |
| 4 | `KubernetesOpenBao` (the restore target) | Restore | the same `auto_unseal` (the same key), the same `r2` store and `prefix`, `restore.snapshot_key` (or `latest`), `restore.root_token` |

Three R2 facts join the rules above:

- **The token is the key.** Rotating the token (or deleting and recreating
  it) mints a new key pair; the vault follows the references on its next
  apply. A token without an R2 permission group authenticates and then
  fails every upload with AccessDenied — the permission group is the grant,
  there is no bucket-side policy to attach.
- **The module owns the S3 dialect.** rclone's Cloudflare provider profile
  (path-style addressing, no multipart ETags) is selected for the `r2` arm;
  nothing S3-shaped is typed in the manifest.
- **Emptying the bucket is yours.** Deleting a `CloudflareR2Bucket` that
  still holds objects is refused by Cloudflare (the provider has no
  force-destroy); retire a snapshot store by emptying the bucket over the S3
  API (`aws s3 rm --recursive`, against the account's R2 endpoint with the
  token's pair) before destroying it.

The validated manifests for this set are the `gcp-gke` lane's own:
`e2e/fixture-gke-r2-source.yaml` (#3), `e2e/scenarios/gke-r2-backup-restore.yaml`
(#4), and the Cloudflare side under `../aa_e2e/realcluster/gcp-gke/manifests/`
(#1–#2). The `05-production-ha-r2-backups` preset is #3 as a starting point.

## Restore on the bad day

The original is gone; the snapshots are in the store; the seal key still
exists. Declare a fresh `KubernetesOpenBao` with the same `auto_unseal`
key, the source's `backup` block (same store, same `prefix`), and a
`restore` block — `latest: true`, or the exact `snapshot_key` from the
store's listing — naming the Secret the root token will live in:

```yaml
  restore:
    latest: true
    rootToken:
      name: bao-init-root
      key: token
```

Then the one manual step every OpenBao has: initialize the fresh cluster.
With auto-unseal, init returns recovery keys and a root token, and the
server unseals itself.

```bash
kubectl exec -n <namespace> <name>-0 -- bao operator init -recovery-shares=1 -recovery-threshold=1 -format=json
kubectl create secret generic bao-init-root -n <namespace> --from-literal=token=<root_token>
kubectl logs -n <namespace> job/<restore_job_name> -f   # the restore_job_name output
```

The restore Job was already waiting for that Secret; it fetches the
snapshot, installs it, and prints the two closing steps: remove `restore`
from the spec and apply again (backups resume), and delete the Secret (the
token it held no longer exists — the restored state carries the SOURCE's
tokens, policies, auth methods, and login role). Read the vault with the
source's root token from here on. The `behavioral-backup-restore` lane runs
exactly this on a kind cluster through the transit seal.

## Restore without auto-unseal: by hand

A Shamir vault has no key the module can point a restore at — the humans
holding the shares are the key. Restore it by hand, with OpenBao's own
tools; the module's snapshots are plain Raft snapshots and need nothing
else:

1. Deploy a fresh `KubernetesOpenBao` on Shamir with the source's `backup`
   block and NO `restore` block; initialize and unseal it as usual.
2. Fetch the snapshot from the store (`rclone copyto` with the same
   remote the job uses, or any S3/GCS/Azure client) and copy it into pod 0.
3. `bao operator raft snapshot restore -force <file>`. Without `-force`
   OpenBao refuses a snapshot taken under other unseal keys ("could not
   verify hash file, possibly the snapshot is using a different set of
   unseal keys"); `-force` installs it anyway.
4. The server seals itself after a forced restore, because the installed
   state is protected by the SOURCE's unseal keys. Unseal every pod with
   the source's shares, then log in with the source's root token.

The declared restore exists to make steps 2–4 disappear; the seal key is
what makes that possible.

## Namespace ownership — the infra exception

A dedicated namespace with `createNamespace: true` is the normal
single-tenant shape — the
[namespace-ownership pattern](../../_patterns/namespace-ownership.md)'s
sole-tenant case.

## On the diagram

OpenBao renders in the shared-cluster layer; the cluster store's backend
points at it and ExternalSecrets draw "reads from" edges into the store
— the whole path from application credential back to the vault is
visible, hop by hop.

## Pairs well with

- KubernetesExternalSecretsOperator + KubernetesClusterSecretStore +
  KubernetesExternalSecret — the self-hosted chain above.
- KubernetesIngress / route kinds — only when the API must be reachable
  from outside the cluster; composed, never embedded.
- GcpGcsBucket / CloudflareR2Bucket + CloudflareAccountApiToken /
  KubernetesSeaweedFs — the snapshot store, by reference from `backup`;
  GcpServiceAccount + GcpGkeWorkloadIdentityBinding for the keyless job
  identity; the GcpKmsKeyRing / GcpKmsKey / GcpKmsKeyIamMember trio for the
  seal a declared restore depends on (the resource sets above).

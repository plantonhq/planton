# GKE Production HA with Cloud KMS Auto-Unseal and GCS Backups

The production vault on GKE with disaster recovery built in: three Raft
servers that unseal themselves through a Cloud KMS key, and hourly Raft
snapshots landing keylessly in a Google Cloud Storage bucket — the seal
key, the bucket, and both identities declared by reference to the catalog's
GCP kinds, so nothing here is typed twice and nothing is a credential. This
is the production half of the GKE resource set in the component guide; the
restore half is a second `KubernetesOpenBao` on the SAME key with a
`restore` block, on the day the original is gone.

Two identities on purpose: the server's (wraps the master key with the KMS
key) and the backup job's (writes and prunes snapshots). The seal key and
the snapshot bucket are different blast radii, and each identity gets
exactly the grant its job needs — `cloudkms.cryptoKeyEncrypterDecrypter` on
the key, `storage.objectAdmin` plus `storage.legacyBucketReader` on the
bucket. Each needs a `GcpGkeWorkloadIdentityBinding`: the server's on the
ServiceAccount named after the vault, the job's on `<name>-backup`.

Initialization is still yours, once (`bao operator init` returns RECOVERY
keys and a root token; the seal unseals the server). Then the one thing the
module cannot do inside the vault: run the four-command login recipe the
guide prints so the backup job can authenticate — until it runs, every
backup run fails and its log prints the recipe with the real names.

The seal is checked when the server STARTS: declare the KMS key ring, key,
and grant before this resource in the same dependency-ordered set, or the
pods crash-loop with "Error configuring seal" until they exist. And treat
the key as the vault itself — `deletion_policy: PREVENT` on the `GcpKmsKey`;
destroying it destroys every vault and every snapshot sealed by it.

Change first: the GCP project and region, the referenced resource names
(the KMS trio, the two service accounts, the bucket), the `prefix` (one per
live vault, forever), then the schedule and retention to your recovery
objectives.

See [04-gke-ha-gcs-backups.yaml](./04-gke-ha-gcs-backups.yaml) for the
manifest.

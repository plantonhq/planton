package module

// The four scripts the backup CronJob and the restore Job run, delivered
// through the module-owned `<name>-backup-scripts` ConfigMap. Two run on
// the server's own image with the `bao` CLI (snapshot, restore); two run
// on the rclone image (upload, fetch). They are POSIX sh — both images
// are Alpine — and they are the surface a person reads on the bad day,
// so every failure names what happened and the exact next step, one
// root cause per message, with the names this module rendered.
//
// PARITY: the Terraform module carries this exact content in its locals
// (`backup_scripts` in scripts.tf), with `${` written as `$${` for HCL —
// keep the rendered text byte-identical across engines; each heredoc there
// carries a `# parity:` marker naming its constant here, and the
// repository's cross-engine script parity guard diffs the pair on every
// change.
//
// Environment contract (set by the module on each container):
//   snapshot: BAO_ADDR, BAO_CACERT (TLS only), BAO_CLIENT_TIMEOUT,
//             BAO_AUTH_PATH, BAO_ROLE, BACKUP_POLICY, BACKUP_SERVICE_ACCOUNT,
//             BACKUP_NAMESPACE, SNAPSHOT_DIR, SNAPSHOT_PREFIX
//   upload:   RCLONE_CONFIG_STORE_* (the remote), STORE_PATH, STORE_KIND,
//             RETENTION_DAYS, SNAPSHOT_DIR
//   fetch:    RCLONE_CONFIG_STORE_* (the remote), STORE_PATH, STORE_ROOT,
//             STORE_KIND, RESTORE_SNAPSHOT_KEY or RESTORE_LATEST, SNAPSHOT_DIR
//   restore:  BAO_ADDR, BAO_CACERT (TLS only), BAO_CLIENT_TIMEOUT, BAO_TOKEN
//             (from the operator's Secret), ROOT_TOKEN_SECRET, ROOT_TOKEN_KEY,
//             RELEASE_NAME, RELEASE_NAMESPACE, SNAPSHOT_DIR

// ConfigMap keys — the file names under the scripts mount.
const (
	scriptSnapshot = "snapshot.sh"
	scriptUpload   = "upload.sh"
	scriptFetch    = "fetch.sh"
	scriptRestore  = "restore.sh"
)

// backupScripts is the ConfigMap's data, keyed by file name.
func backupScripts() map[string]string {
	return map[string]string{
		scriptSnapshot: snapshotScript,
		scriptUpload:   uploadScript,
		scriptFetch:    fetchScript,
		scriptRestore:  restoreScript,
	}
}

// snapshotScript logs in through the Kubernetes auth method with the
// pod's projected ServiceAccount token and streams a Raft snapshot to
// the scratch volume. Its failure messages carry the login recipe with
// the rendered names: the recipe is the one step this module cannot take
// (OpenBao is sealed at deploy time), so the job teaches it every time
// it is missing.
const snapshotScript = `#!/bin/sh
# Takes a Raft snapshot of the OpenBao cluster into the shared scratch
# volume. Authenticates through the Kubernetes auth method with this
# pod's ServiceAccount token; a snapshot is a plain read on
# sys/storage/raft/snapshot (not a sudo operation), so the policy the
# recipe below writes is the whole authorization.
set -eu

recipe() {
  cat <<EOF

The backup job cannot log in to OpenBao yet. Run this recipe once, with a
token that can manage auth methods and policies (the initial root token
from 'bao operator init' works), against $BAO_ADDR:

  bao policy write $BACKUP_POLICY - <<'POLICY'
  path "sys/storage/raft/snapshot" { capabilities = ["read"] }
  POLICY
  bao auth enable -path=$BAO_AUTH_PATH kubernetes
  bao write auth/$BAO_AUTH_PATH/config kubernetes_host="https://kubernetes.default.svc:443"
  bao write auth/$BAO_AUTH_PATH/role/$BAO_ROLE \
    bound_service_account_names=$BACKUP_SERVICE_ACCOUNT \
    bound_service_account_namespaces=$BACKUP_NAMESPACE \
    token_policies=$BACKUP_POLICY token_ttl=1h

The next scheduled run picks it up; to run one now:
  kubectl create job --from=cronjob/$BACKUP_SERVICE_ACCOUNT -n $BACKUP_NAMESPACE $BACKUP_SERVICE_ACCOUNT-now
EOF
}

jwt=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
# BAO_ADDR is the active-leader Service, which has no endpoints
# until a server is initialized and unsealed -- and a fresh unseal, a
# rolling restart, or a leader election empties it for a few seconds. A run
# that starts inside that window must wait, not spend the Job's backoff
# limit in seconds: poll for a reachable, unsealed server (bao status exits
# 0 only then) for up to five minutes before the login below explains a
# failure.
tries=0
until status_out=$(bao status 2>&1); do
  tries=$((tries + 1))
  if [ "$tries" -ge 60 ]; then
    echo "OpenBao at $BAO_ADDR did not answer 'bao status' as an unsealed server within 5 minutes: $status_out"
    break
  fi
  sleep 5
done
login_out=$(bao write -field=token "auth/$BAO_AUTH_PATH/login" role="$BAO_ROLE" jwt="$jwt" 2>&1) || {
  echo "Login to OpenBao at $BAO_ADDR failed: $login_out"
  case "$login_out" in
    *"no handler for route"*)
      echo "The Kubernetes auth method is not enabled at auth/$BAO_AUTH_PATH."
      recipe ;;
    *"invalid role name"*)
      echo "The role '$BAO_ROLE' does not exist under auth/$BAO_AUTH_PATH."
      recipe ;;
    *"permission denied"*|*"service account name not authorized"*|*"namespace not authorized"*)
      echo "The role '$BAO_ROLE' exists but is not bound to ServiceAccount $BACKUP_SERVICE_ACCOUNT in namespace $BACKUP_NAMESPACE. Re-run the role command from the recipe with those two names."
      recipe ;;
    *"certificate"*|*"x509"*|*"tls"*)
      echo "TLS verification against $BAO_ADDR failed. The server certificate must include the active Service name (the host in BAO_ADDR) in its dnsNames; add it to the KubernetesCertificate and let it re-issue." ;;
    *"connection refused"*|*"no such host"*|*"i/o timeout"*)
      echo "OpenBao is not reachable at $BAO_ADDR. This is the active-leader Service: it has no endpoints until a server is initialized and unsealed ('bao operator init' then unseal, or auto-unseal after init)." ;;
    *"Vault is sealed"*|*"Vault is not initialized"*)
      echo "OpenBao is sealed or not initialized. Initialize and unseal it first; backups resume on the next scheduled run." ;;
  esac
  exit 1
}
export BAO_TOKEN="$login_out"

stamp=$(date -u +%Y%m%dT%H%M%SZ)
file="$SNAPSHOT_DIR/$SNAPSHOT_PREFIX-$stamp.snap"
save_out=$(bao operator raft snapshot save "$file" 2>&1) || {
  echo "Taking the snapshot failed: $save_out"
  case "$save_out" in
    *"raft storage is not in use"*)
      echo "This server does not run integrated Raft storage; snapshots exist only for Raft storage (server.raft). The spec rule should have refused this — re-check the deployed manifest." ;;
    *"permission denied"*)
      echo "The token from role '$BAO_ROLE' cannot read sys/storage/raft/snapshot. The policy '$BACKUP_POLICY' must grant capabilities = [\"read\"] on that path; re-run the policy command from the recipe."
      recipe ;;
  esac
  exit 1
}
echo "$(basename "$file")" > "$SNAPSHOT_DIR/.latest"
echo "Snapshot written: $file ($(wc -c < "$file") bytes)"
`

// uploadScript moves the snapshot the previous container wrote into the
// declared store and prunes objects older than the retention window. The
// remote is configured entirely through RCLONE_CONFIG_STORE_* variables
// the module sets — no config file exists.
const uploadScript = `#!/bin/sh
# Ships the snapshot the previous container wrote to the declared object
# store and prunes objects under the prefix older than RETENTION_DAYS.
set -eu

denied() {
  echo "The store refused the request. The identity or key this job runs with needs write and list on the store:"
  case "$STORE_KIND" in
    r2)    echo "  Cloudflare R2: the API token needs the 'Workers R2 Storage Bucket Item Write' permission group scoped to this bucket." ;;
    gcs)   echo "  Google Cloud Storage: grant roles/storage.objectAdmin AND roles/storage.legacyBucketReader on the bucket to the job's identity (a GcpGcsBucket's iam_members, one entry each)." ;;
    azure) echo "  Azure Blob: the identity needs 'Storage Blob Data Contributor' on the container, or the account key / connection string must be current." ;;
    *)     echo "  S3: the identity or key needs s3:PutObject, s3:ListBucket, and s3:DeleteObject on the bucket and prefix." ;;
  esac
}

file=$(cat "$SNAPSHOT_DIR/.latest")
if ! out=$(rclone copyto "$SNAPSHOT_DIR/$file" "$STORE_PATH/$file" 2>&1); then
  echo "Upload of $file to $STORE_PATH failed: $out"
  case "$out" in
    *AccessDenied*|*"403"*|*"Forbidden"*|*"AuthorizationFailure"*|*"AuthenticationFailed"*) denied ;;
    *NoSuchBucket*|*"404"*|*"ContainerNotFound"*) echo "The bucket or container in $STORE_PATH does not exist. The store is declared by reference to the catalog's bucket kind; check that resource is deployed and that its name matches." ;;
    *"no such host"*|*"connection refused"*|*"i/o timeout"*) echo "The store endpoint is not reachable from this cluster. For an S3-compatible endpoint, check endpoint_url and the cluster's egress." ;;
    *"certificate"*|*"x509"*) echo "TLS verification against the store endpoint failed. For a self-signed S3-compatible endpoint set ca_pem on the s3 arm." ;;
  esac
  exit 1
fi
echo "Uploaded $STORE_PATH/$file"

if [ "${RETENTION_DAYS:-0}" -gt 0 ]; then
  if ! out=$(rclone delete "$STORE_PATH" --min-age "${RETENTION_DAYS}d" -v 2>&1); then
    echo "Pruning objects older than $RETENTION_DAYS days under $STORE_PATH failed: $out"
    case "$out" in
      *AccessDenied*|*"403"*|*"Forbidden"*) denied ;;
    esac
    exit 1
  fi
  pruned=$(printf '%s\n' "$out" | grep -c 'Deleted' || true)
  echo "Pruned $pruned snapshot(s) older than $RETENTION_DAYS days under $STORE_PATH"
else
  echo "Retention is 0: keeping every snapshot (the bucket's own lifecycle rules decide)."
fi
`

// fetchScript downloads the snapshot the restore declaration names — or
// the newest under the prefix — into the scratch volume for the restore
// container.
const fetchScript = `#!/bin/sh
# Fetches the declared snapshot (or the newest under the prefix) from the
# store into the shared scratch volume.
set -eu

if [ "${RESTORE_LATEST:-}" = "true" ]; then
  listing=$(rclone lsf "$STORE_PATH" --files-only --format tp 2>&1) || {
    echo "Listing $STORE_PATH failed: $listing"
    echo "Check the store's credentials or identity, and that the bucket exists."
    exit 1
  }
  key=$(printf '%s\n' "$listing" | grep '\.snap$' | sort | tail -n 1 | cut -d';' -f2-)
  if [ -z "$key" ]; then
    echo "No snapshot found under $STORE_PATH."
    echo "The source has not written one yet, or its backup prefix differs from this cluster's. List the store with: rclone lsf $STORE_PATH"
    exit 1
  fi
  src="$STORE_PATH/$key"
  echo "Newest snapshot under $STORE_PATH: $key"
else
  src="$STORE_ROOT/$RESTORE_SNAPSHOT_KEY"
fi

if ! out=$(rclone copyto "$src" "$SNAPSHOT_DIR/restore.snap" 2>&1); then
  echo "Fetching $src failed: $out"
  case "$out" in
    *"not found"*|*NoSuchKey*|*"404"*|*"object not found"*) echo "No object at $src. The key is relative to the bucket (e.g. <prefix>/<name>-<UTC>.snap); list the store with: rclone lsf $STORE_PATH" ;;
    *AccessDenied*|*"403"*|*"Forbidden"*) echo "The store refused the read. The identity or key this job runs with needs read and list on the bucket." ;;
  esac
  exit 1
fi
echo "Fetched $src ($(wc -c < "$SNAPSHOT_DIR/restore.snap") bytes)"
`

// restoreScript installs the fetched snapshot into the target cluster.
// It waits for the target to be initialized and unsealed first (the
// operator creates the token Secret seconds after `bao operator init`,
// and an auto-unseal target may still be coming up), and its failure
// messages name the one thing that most often goes wrong on the bad day:
// a target sealed with a different key than the snapshot.
const restoreScript = `#!/bin/sh
# Installs the fetched snapshot into this cluster with the initial root
# token the operator placed in a Secret after 'bao operator init'.
set -eu

echo "Waiting for OpenBao at $BAO_ADDR to be initialized and unsealed..."
tries=0
while :; do
  if status=$(bao status -format=json 2>&1); then
    break
  fi
  case "$status" in
    *'"initialized": false'*|*'"initialized":false'*)
      echo "OpenBao is not initialized. Run 'bao operator init' against it; with auto-unseal it unseals itself, then put the returned root token in Secret $ROOT_TOKEN_SECRET under key $ROOT_TOKEN_KEY (this Job read that Secret to start, so it may already hold a token from a previous init — recreate it if so)." ;;
    *'"sealed": true'*|*'"sealed":true'*)
      echo "OpenBao is initialized but sealed. With auto-unseal it unseals on its own shortly after init; if it stays sealed, the seal's KMS key or transit token is not reachable — check the server logs." ;;
    *)
      echo "OpenBao is not answering yet: $status" ;;
  esac
  tries=$((tries + 1))
  if [ "$tries" -ge 60 ]; then
    echo "Gave up after 10 minutes. Fix the condition above; this Job retries on its own."
    exit 1
  fi
  sleep 10
done

if ! out=$(bao operator raft snapshot restore "$SNAPSHOT_DIR/restore.snap" 2>&1); then
  echo "Installing the snapshot failed: $out"
  case "$out" in
    *"could not verify hash file"*)
      echo "The snapshot was taken under a different seal key than this cluster's. A declared restore requires the SAME auto_unseal key on source and target (the same KMS key, or the same transit key on the same key holder). Point auto_unseal at the source's key and redeploy; a snapshot cannot be moved across keys declaratively." ;;
    *"permission denied"*)
      echo "The token in Secret $ROOT_TOKEN_SECRET/$ROOT_TOKEN_KEY cannot install snapshots. It must be the initial root token 'bao operator init' printed for THIS cluster." ;;
    *"raft storage is not in use"*)
      echo "This server does not run integrated Raft storage; a restore needs Raft storage (server.raft)." ;;
  esac
  echo "OpenBao seals itself when an install fails part-way. Delete the server pods in namespace $RELEASE_NAMESPACE so they restart and auto-unseal (kubectl delete pod -n $RELEASE_NAMESPACE -l app.kubernetes.io/instance=$RELEASE_NAME), fix the cause above, then change or re-declare 'restore' to run again."
  exit 1
fi

cat <<EOF
Restore complete. This cluster now carries the snapshot's state: every secret, policy, auth method, and token — including the source's backup login role and its root token; the token this Job used no longer exists.

Two things to do now:
  1. Remove the 'restore' block from the KubernetesOpenBao spec and apply again. Backups are suspended while 'restore' is declared; removing it deletes this Job and resumes the schedule.
  2. Delete the Secret that carried the token: kubectl delete secret -n $RELEASE_NAMESPACE $ROOT_TOKEN_SECRET
EOF
`

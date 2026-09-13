#!/usr/bin/env bash
# SETUP for the *-backup-restore scenarios: brings the SOURCE vault (the
# lane's fixture-*-source.yaml, deployed and Running but — like every fresh
# OpenBao — uninitialized) to the state the restore proof reads back, and
# publishes which snapshot it took. The script seeds the VAULT, not the
# store: the source declares whichever store its lane uses (an in-cluster S3
# store, GCS, R2), and the snapshot lands wherever the module wired it, so
# one script serves every restore lane.
#
#   1. init         -> `bao operator init` with recovery shares (the source is
#                      auto-unsealed: transit on kind, Cloud KMS on GKE) —
#                      the root token is kept in a Secret in the lane
#                      namespace so the verifier can read the restored state
#                      with the token that state carries
#   2. the recipe   -> the four commands the backup job prints when it cannot
#                      log in, run VERBATIM with the names the module rendered
#   3. marker A     -> a KV v2 secret written BEFORE the snapshot
#   4. a snapshot   -> one run created from the backup CronJob, waited to
#                      completion; its object key is derived from the job's
#                      own environment and log
#   5. marker B     -> a KV v2 secret written AFTER the snapshot
#
# A restored target carrying A and not B proves exactly one snapshot's worth
# of state came back. The snapshot's key is only known once the job names
# it, so the script publishes it for the scenario manifest
# (${E2E_SETUP:OPENBAO_SNAPSHOT_KEY}); the `latest` lane ignores it.
# Everything lands inside fixture-owned resources, so DEPENDENCIES-DOWN
# removes it. Idempotent per lane: an initialized source is not re-inited,
# the recipe's writes upsert, markers upsert, the run Job is named per lane.
set -euo pipefail

ns="e2e-bao-lab"
src="e2e-bao-src"
pod="${src}-0"
root_secret="${src}-root-token"
kv_mount="e2e-dr"
run_job="${src}-backup-seed-${E2E_RUN_ID}"

# bao runs inside the server pod, where the chart sets BAO_ADDR to the local
# listener; the token rides as an environment variable per call and is never
# echoed.
bao_on_src() {
  local token="$1"; shift
  kubectl exec -n "${ns}" "${pod}" -c openbao -- env "BAO_TOKEN=${token}" bao "$@"
}

# ---- 1. init (once) --------------------------------------------------------
if kubectl get secret "${root_secret}" -n "${ns}" >/dev/null 2>&1; then
  echo "  [seed] source already initialized (Secret ${root_secret} exists)"
  root_token="$(kubectl get secret "${root_secret}" -n "${ns}" -o jsonpath='{.data.token}' | base64 -d)"
else
  echo "  [seed] initializing the source vault (recovery shares — the seal auto-unseals)"
  for i in $(seq 1 60); do
    if init_json="$(kubectl exec -n "${ns}" "${pod}" -c openbao -- bao operator init -recovery-shares=1 -recovery-threshold=1 -format=json 2>/dev/null)"; then
      break
    fi
    init_json=""
    sleep 5
  done
  [ -n "${init_json}" ] || { echo "  [seed] the source never accepted 'bao operator init' (is its seal reachable? the server refuses to start until the transit engine or KMS key exists)" >&2; exit 1; }
  root_token="$(printf '%s' "${init_json}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["root_token"])')"
  kubectl create secret generic "${root_secret}" -n "${ns}" --from-literal=token="${root_token}" >/dev/null
  echo "  [seed] source initialized; root token kept in Secret ${ns}/${root_secret}"
fi

# Auto-unseal: the server unseals itself right after init; wait for it.
for i in $(seq 1 60); do
  if bao_on_src "${root_token}" status >/dev/null 2>&1; then break; fi
  sleep 5
done
bao_on_src "${root_token}" status >/dev/null || { echo "  [seed] the source stayed sealed after init — check the seal (server logs name the cause)" >&2; exit 1; }
echo "  [seed] source is initialized and unsealed"

# ---- 2. the login recipe, verbatim -----------------------------------------
# The names are the module's: policy, role, and ServiceAccount are all
# `<name>-backup`; the auth mount is the spec default. These are the exact
# four commands the backup job's log prints when the recipe has not run.
policy="${src}-backup"
role="${src}-backup"
sa="${src}-backup"
mount="kubernetes"
echo "  [seed] running the login recipe (policy ${policy}, auth/${mount}, role ${role} -> ${ns}/${sa})"
kubectl exec -n "${ns}" "${pod}" -c openbao -- env "BAO_TOKEN=${root_token}" sh -c "bao policy write ${policy} - <<'POLICY'
path \"sys/storage/raft/snapshot\" { capabilities = [\"read\"] }
POLICY"
bao_on_src "${root_token}" auth enable -path="${mount}" kubernetes 2>/dev/null \
  || echo "  [seed] auth/${mount} already enabled"
bao_on_src "${root_token}" write "auth/${mount}/config" kubernetes_host="https://kubernetes.default.svc:443" >/dev/null
bao_on_src "${root_token}" write "auth/${mount}/role/${role}" \
  bound_service_account_names="${sa}" \
  bound_service_account_namespaces="${ns}" \
  token_policies="${policy}" token_ttl=1h >/dev/null

# ---- 3. marker A -----------------------------------------------------------
bao_on_src "${root_token}" secrets enable -path="${kv_mount}" -version=2 kv 2>/dev/null \
  || echo "  [seed] ${kv_mount}/ already mounted"
bao_on_src "${root_token}" kv put -mount="${kv_mount}" marker-a value="marker-a-${E2E_RUN_ID}" >/dev/null
echo "  [seed] marker A written to ${kv_mount}/marker-a"

# ---- 4. one snapshot, from the CronJob -------------------------------------
cronjob="${src}-backup"
kubectl delete job "${run_job}" -n "${ns}" --ignore-not-found >/dev/null
kubectl create job --from="cronjob/${cronjob}" "${run_job}" -n "${ns}" >/dev/null
echo "  [seed] backup run ${run_job} created from cronjob/${cronjob}"
if ! kubectl wait --for=condition=complete "job/${run_job}" -n "${ns}" --timeout=10m >/dev/null 2>&1; then
  echo "  [seed] backup run ${run_job} did not complete; its logs:" >&2
  kubectl logs "job/${run_job}" -n "${ns}" --all-containers=true >&2 || true
  exit 1
fi
# The object key relative to the bucket: the snapshot container logs the file
# it wrote; the job's own environment carries the store root (`store:<bucket>`)
# and the store path (`store:<bucket>/<prefix>`), so the key is the path
# beyond the root plus the file name — exactly what `restore.snapshotKey`
# and the fetch script expect.
file="$(kubectl logs "job/${run_job}" -n "${ns}" -c snapshot | sed -n 's|^Snapshot written: .*/\([^/ ]*\.snap\).*|\1|p' | tail -n 1)"
[ -n "${file}" ] || { echo "  [seed] the snapshot container did not log the file it wrote" >&2; exit 1; }
store_root="$(kubectl get cronjob "${cronjob}" -n "${ns}" -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[?(@.name=="upload")].env[?(@.name=="STORE_ROOT")].value}')"
store_path="$(kubectl get cronjob "${cronjob}" -n "${ns}" -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[?(@.name=="upload")].env[?(@.name=="STORE_PATH")].value}')"
prefix="${store_path#"${store_root}"}"
prefix="${prefix#/}"
snapshot_key="${prefix:+${prefix}/}${file}"
echo "  [seed] snapshot landed at ${store_path}/${file} (key ${snapshot_key})"

# ---- 5. marker B -----------------------------------------------------------
bao_on_src "${root_token}" kv put -mount="${kv_mount}" marker-b value="marker-b-${E2E_RUN_ID}" >/dev/null
echo "  [seed] marker B written to ${kv_mount}/marker-b (after the snapshot — must NOT come back)"

echo "OPENBAO_SNAPSHOT_KEY=${snapshot_key}" >> "${E2E_SETUP_OUTPUT}"
echo "OPENBAO_SEED_BACKUP_JOB=${run_job}" >> "${E2E_SETUP_OUTPUT}"

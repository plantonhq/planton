#!/usr/bin/env bash
# gcp-gke batch audit: enumerates every GCP and Cloudflare resource class the
# batch can leave behind and fails on any survivor — except the residue GCP
# leaves by design (the KMS key ring, the destroyed key's shell), which it
# asserts is in its expected state instead. Read-only; safe at any time.
# Source env.sh first so the KMS probes know the region and this batch's id.
set -uo pipefail

# Either the bootstrap-time variables or the env.sh the bootstrap wrote.
GCP_PROJECT_ID="${GCP_PROJECT_ID:-${PLANTON_E2E_GCP_PROJECT:-}}"
: "${GCP_PROJECT_ID:?set GCP_PROJECT_ID or source env.sh}"
CLOUDFLARE_ACCOUNT_ID="${CLOUDFLARE_ACCOUNT_ID:-${PLANTON_E2E_GKE_R2_ACCOUNT_ID:-}}"
: "${CLOUDFLARE_ACCOUNT_ID:?set CLOUDFLARE_ACCOUNT_ID or source env.sh}"
: "${CLOUDFLARE_API_TOKEN:?set CLOUDFLARE_API_TOKEN (the Cloudflare side is read through it)}"
rc=0

echo "==> service accounts"
for sa in planton-e2e-gke-pg-backup planton-e2e-gke-mongo-backup planton-e2e-gke-openbao-backup planton-e2e-gke-openbao-unseal planton-e2e-gke-langfuse; do
  if gcloud iam service-accounts describe "${sa}@${GCP_PROJECT_ID}.iam.gserviceaccount.com" --project "${GCP_PROJECT_ID}" >/dev/null 2>&1; then
    echo "  SURVIVOR: service account ${sa}"; rc=1
  fi
done

echo "==> buckets"
for bucket in "planton-e2e-gke-backups-${GCP_PROJECT_ID}" "planton-e2e-gke-langfuse-${GCP_PROJECT_ID}"; do
  if gcloud storage buckets describe "gs://${bucket}" >/dev/null 2>&1; then
    echo "  SURVIVOR: bucket ${bucket}"; rc=1
  fi
done

# Workload Identity bindings live ON the service accounts (IAM policy of the
# GSA); a surviving binding implies a surviving account, caught above.

# The OpenBao seal's KMS objects are the residue GCP leaves BY DESIGN, and the
# audit asserts exactly that shape rather than calling it a survivor: the key
# ring cannot be deleted (its destroy is an abandon), and a key cannot be
# deleted either — destroy schedules every version for destruction and the
# key shell stays. So: the ring is expected; the key of THIS batch (named by
# the bootstrap's id, read from env.sh) must exist with no ENABLED version
# left — an enabled version means the destroy never ran.
echo "==> KMS (expected residue: the ring and the destroyed key's shell)"
GCP_REGION="${GCP_REGION:-${PLANTON_E2E_GCP_REGION:-}}"
kms_ring="planton-e2e-gke-openbao-unseal"
if [ -n "${GCP_REGION}" ] && gcloud kms keyrings describe "${kms_ring}" --location "${GCP_REGION}" --project "${GCP_PROJECT_ID}" >/dev/null 2>&1; then
  echo "  ring ${kms_ring} present in ${GCP_REGION} (expected: rings are permanent)"
  if [ -n "${PLANTON_E2E_GKE_BATCH_ID:-}" ]; then
    kms_key="${kms_ring}-${PLANTON_E2E_GKE_BATCH_ID}"
    enabled="$(gcloud kms keys versions list --key "${kms_key}" --keyring "${kms_ring}" --location "${GCP_REGION}" --project "${GCP_PROJECT_ID}" --filter='state=ENABLED' --format='value(name)' 2>/dev/null || true)"
    if [ -n "${enabled}" ]; then
      echo "  SURVIVOR: key ${kms_key} still has an ENABLED version — the key node was not destroyed (destroy_node gcpkmskey in teardown.sh)"; rc=1
    else
      echo "  key ${kms_key}: no enabled version (expected: destroy schedules every version for destruction; the shell stays)"
    fi
  else
    echo "  (PLANTON_E2E_GKE_BATCH_ID unset: source env.sh to audit this batch's key versions)"
  fi
elif [ -z "${GCP_REGION}" ]; then
  echo "  (GCP_REGION unset: source env.sh to audit the KMS ring)"
else
  echo "  ring ${kms_ring} absent (never bootstrapped in ${GCP_REGION})"
fi

cf_api="https://api.cloudflare.com/client/v4/accounts/${CLOUDFLARE_ACCOUNT_ID}"
echo "==> R2 bucket"
if curl -sf -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" "${cf_api}/r2/buckets/planton-e2e-gke-backups" >/dev/null 2>&1; then
  echo "  SURVIVOR: R2 bucket planton-e2e-gke-backups (if it still holds objects: empty it over the S3 API with the writer token, then destroy the node)"; rc=1
fi

echo "==> R2 writer token"
if curl -s -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" "${cf_api}/tokens?per_page=50" | python3 -c 'import json,sys; sys.exit(0 if any(t.get("name")=="planton-e2e-gke-backups-writer" for t in json.load(sys.stdin).get("result") or []) else 1)' 2>/dev/null; then
  echo "  SURVIVOR: account API token planton-e2e-gke-backups-writer"; rc=1
fi

# The ComputeClass lanes make GKE create nodes (and auto-created node pools)
# for the classes they declare; the behavioral verifier waits for the node to
# be reclaimed, but a lane killed mid-proof leaves billed capacity behind. A
# node still labelled with an E2E class name is a survivor; GKE deletes the
# emptied auto-created pool once its last node is gone.
echo "==> compute-class nodes"
kubeconfig="${PLANTON_E2E_KUBECONFIG:-}"
if [ -n "${kubeconfig}" ]; then
  class_nodes="$(kubectl --kubeconfig "${kubeconfig}" get nodes -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.metadata.labels.cloud\.google\.com/compute-class}{"\n"}{end}' 2>/dev/null | awk '$2 ~ /^e2e-gke-computeclass/')"
  if [ -n "${class_nodes}" ]; then
    echo "  SURVIVOR: nodes of an E2E compute class (delete the class's pods, or the class, and GKE reclaims them):"
    echo "${class_nodes}" | sed 's/^/    /'
    rc=1
  fi
else
  echo "  (PLANTON_E2E_KUBECONFIG unset: source env.sh to audit compute-class nodes)"
fi

if [ "${rc}" -eq 0 ]; then echo "==> audit clean: zero batch residue in ${GCP_PROJECT_ID} and Cloudflare account ${CLOUDFLARE_ACCOUNT_ID}"; else echo "==> audit FAILED"; fi
exit "${rc}"

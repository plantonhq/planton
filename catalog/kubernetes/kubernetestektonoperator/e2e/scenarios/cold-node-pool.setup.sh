#!/usr/bin/env bash
# Seeds the cold-node-pool scenario (cold-node-pool.yaml): publishes a node
# label no node carries yet as COLD_POOL_LABEL, and puts it on a node 60s
# from now -- a node joining an empty pool. The label is per lane (derived
# from E2E_RUN_ID), so two lanes on one cluster never hand each other a
# node early, and it is taken back 30 minutes later so a persistent cluster
# keeps no residue. It touches nothing else: the operator's pods simply wait
# Pending until the label arrives.
set -euo pipefail

: "${E2E_RUN_ID:?the runner sets E2E_RUN_ID}"
: "${E2E_SETUP_OUTPUT:?the runner sets E2E_SETUP_OUTPUT}"

lane="$(printf '%s' "${E2E_RUN_ID}" | cksum | cut -d' ' -f1)"
label="cold-pool.e2e.planton.dev/lane-${lane}"
node="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"

# A label left by an earlier run of this lane would hand the pods a node
# at once and prove nothing.
kubectl label node "${node}" "${label}-" >/dev/null 2>&1 || true

echo "COLD_POOL_LABEL=${label}" >>"${E2E_SETUP_OUTPUT}"

nohup bash -c "sleep 60 \
  && kubectl label node '${node}' '${label}=true' --overwrite >/dev/null \
  && sleep 1800 \
  && kubectl label node '${node}' '${label}-' >/dev/null 2>&1" >/dev/null 2>&1 &

echo "  [setup] ${node} joins the pool as ${label} in 60s"

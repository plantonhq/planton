#!/usr/bin/env bash
# SETUP-phase step for the KubernetesStigmer lanes (planton.dev/e2e-setup-script
# on each scenario; the SETUP section of e2e/README.md).
#
# The lanes share a kind cluster with every other tier-4 kind. Some of those
# keep their definitions when destroyed, by design: the catalog's CloudNativePG
# operator and backup plugin always do, and the agent-sandbox controller does
# by default. A definition left behind with no controller running it is
# exactly what KubernetesStigmer refuses (the orphaned-prerequisite refusal)
# and what Stigmer's operator would otherwise trust and wait on forever. So
# before a lane deploys, this removes such leftovers, with the delete the
# refusal's own sentence teaches.
#
# A definition is a leftover only when nothing runs it:
#   - one Helm installed (meta.helm.sh/release-name and release-namespace
#     annotations) whose release has no Deployment left
#     (app.kubernetes.io/instance=<release> in <release-namespace>);
#   - an agent-sandbox definition the catalog's controller kind wrote
#     (planton.ai/crd-source) with no agent-sandbox-controller Deployment on
#     the cluster, the name upstream's release always installs it under.
# A definition Stigmer's own operator installed
# (app.kubernetes.io/managed-by=stigmer-operator) is the operator's to
# manage and is never touched; so is one whose controller runs, such as a
# fixture this lane deployed in front of the Stigmer.
set -euo pipefail

groups='postgresql.cnpg.io|barmancloud.cnpg.io|agents.x-k8s.io|extensions.agents.x-k8s.io'

kubectl get crd -o json | jq -r --arg groups "^(${groups})$" '
  .items[]
  | select(.spec.group | test($groups))
  | select((.metadata.labels["app.kubernetes.io/managed-by"] // "") != "stigmer-operator")
  | [.metadata.name,
     (.metadata.annotations["meta.helm.sh/release-name"] // ""),
     (.metadata.annotations["meta.helm.sh/release-namespace"] // ""),
     (.metadata.labels["planton.ai/crd-source"] // "")]
  | @tsv' |
while IFS=$'\t' read -r name release release_namespace crd_source; do
  if [[ -n "${release}" && -n "${release_namespace}" ]]; then
    running=$(kubectl -n "${release_namespace}" get deployment -l "app.kubernetes.io/instance=${release}" -o name 2>/dev/null || true)
  elif [[ -n "${crd_source}" && "${name}" == *.agents.x-k8s.io ]]; then
    running=$(kubectl get deployment -A --field-selector metadata.name=agent-sandbox-controller -o name 2>/dev/null || true)
  else
    continue
  fi
  if [[ -z "${running}" ]]; then
    echo "clean slate: removing definition ${name}, left by an earlier lane with no controller running it"
    kubectl delete crd "${name}" --wait=true
  fi
done

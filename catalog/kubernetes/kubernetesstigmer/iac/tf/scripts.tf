# The two cluster-side checks, word for word the Pulumi module's
# (module/scripts.go, which carries the full rationale): the restore check,
# run when the keys gate is created (main.tf), and the cleanup wait, run when
# a destroy reaches terraform_data.cleanup_wait. Their `# parity:` markers let
# hack/guards/ensure_cross_engine_script_parity.sh prove each copy is
# byte-identical to its Go twin; the only difference HCL allows is `$${` for
# a shell `${`.

locals {
  # parity: scripts.go keysCheckScript
  keys_check_script = <<EOT
# Stigmer's restore check: a restore needs the keys the archive is sealed with.
if [ -z "$STIGMER_RECOVER_FROM" ]; then
  exit 0
fi
if ! command -v kubectl >/dev/null 2>&1; then
  echo "the restore check needs kubectl on PATH to look for the keys Secret $STIGMER_KEYS_SECRET; install kubectl where this deploy runs" >&2
  exit 1
fi
if [ -z "$KUBECONFIG" ] && [ -n "$KUBE_CONFIG_PATH" ]; then
  KUBECONFIG="$KUBE_CONFIG_PATH"
  export KUBECONFIG
fi
set -- --namespace "$STIGMER_NAMESPACE"
if [ -n "$KUBE_CTX" ]; then
  set -- "$@" --context "$KUBE_CTX"
fi
if kubectl "$@" get stigmers.stigmer.ai "$STIGMER_NAME" >/dev/null 2>&1; then
  exit 0
fi
if kubectl "$@" get secret "$STIGMER_KEYS_SECRET" >/dev/null 2>&1; then
  exit 0
fi
echo "observed: database.bundled.recover_from restores archive $STIGMER_RECOVER_FROM, and the keys Secret $STIGMER_KEYS_SECRET is not in namespace $STIGMER_NAMESPACE" >&2
echo "meaning: the archive is sealed with the source install's encryption keys; without them the operator would mint new keys and the restored records could not be read" >&2
echo "next step: apply the keys Secret you saved from the source install (its keys_export_command output) with kubectl -n $STIGMER_NAMESPACE apply -f $STIGMER_KEYS_SECRET.yaml, then deploy again" >&2
exit 1
EOT

  # parity: scripts.go cleanupWaitScript
  cleanup_wait_script = <<EOT
# Stigmer's cleanup wait: the operator removes what it installed once no Stigmer is left.
if ! command -v kubectl >/dev/null 2>&1; then
  echo "the cleanup wait needs kubectl on PATH to watch Stigmer's operator clean up; without it the destroy goes on, and the operator's leftovers stay: see the KubernetesStigmer README, 'What a destroy leaves'" >&2
  exit 0
fi
if [ -z "$KUBECONFIG" ] && [ -n "$KUBE_CONFIG_PATH" ]; then
  KUBECONFIG="$KUBE_CONFIG_PATH"
  export KUBECONFIG
fi
if [ -n "$KUBE_CTX" ]; then
  set -- --context "$KUBE_CTX"
fi
live=$(kubectl "$@" get stigmers.stigmer.ai --all-namespaces -o go-template='{{range .items}}{{if not .metadata.deletionTimestamp}}{{.metadata.namespace}}/{{.metadata.name}} {{end}}{{end}}' 2>/dev/null)
if [ -n "$live" ]; then
  echo "a Stigmer still runs ($live), so the operator keeps what it installed; nothing to wait for"
  exit 0
fi
leftovers() {
  out=""
  if kubectl "$@" get crd clusters.postgresql.cnpg.io >/dev/null 2>&1; then
    if kubectl "$@" --namespace "$STIGMER_NAMESPACE" get clusters.postgresql.cnpg.io "$STIGMER_NAME-postgres" >/dev/null 2>&1; then
      out="$out
  the database Cluster $STIGMER_NAMESPACE/$STIGMER_NAME-postgres"
    fi
  fi
  if kubectl "$@" get crd objectstores.barmancloud.cnpg.io >/dev/null 2>&1; then
    if kubectl "$@" --namespace "$STIGMER_NAMESPACE" get objectstores.barmancloud.cnpg.io "$STIGMER_NAME-backups" >/dev/null 2>&1; then
      out="$out
  the backup ObjectStore $STIGMER_NAMESPACE/$STIGMER_NAME-backups"
    fi
  fi
  pods=$(kubectl "$@" --namespace "$STIGMER_NAMESPACE" get pods -l "cnpg.io/cluster=$STIGMER_NAME-postgres" -o name 2>/dev/null)
  if [ -n "$pods" ]; then
    out="$out
  the database's pods in $STIGMER_NAMESPACE (they archive their last changes before they stop)"
  fi
  sandboxes=$(kubectl "$@" get namespaces -l app.kubernetes.io/managed-by=stigmer-operator,app.kubernetes.io/component=sandboxes -o name 2>/dev/null)
  if [ -n "$sandboxes" ]; then
    out="$out
  the sandbox namespace(s): $(echo $sandboxes | sed 's#namespace/##g')"
  fi
  for prerequisite in "Agent Substrate=workerpools.ate.dev" "agent-sandbox=sandboxes.agents.x-k8s.io" "the Barman Cloud plugin=objectstores.barmancloud.cnpg.io" "CloudNativePG=clusters.postgresql.cnpg.io"; do
    name=$${prerequisite%%=*}
    crd=$${prerequisite#*=}
    manager=$(kubectl "$@" get crd "$crd" -o go-template='{{index .metadata.labels "app.kubernetes.io/managed-by"}}' 2>/dev/null)
    if [ "$manager" != "stigmer-operator" ]; then
      continue
    fi
    if [ -n "$(kubectl "$@" get "$crd" --all-namespaces -o name 2>/dev/null)" ]; then
      continue
    fi
    if [ "$crd" = "clusters.postgresql.cnpg.io" ] || [ "$crd" = "objectstores.barmancloud.cnpg.io" ]; then
      if [ -n "$(kubectl "$@" get pods --all-namespaces -l cnpg.io/cluster -o name 2>/dev/null)" ]; then
        continue
      fi
    fi
    out="$out
  $name (the operator installed it; definition $crd)"
  done
  printf '%s' "$out"
}
ready=$(kubectl "$@" --namespace "$STIGMER_OPERATOR_NAMESPACE" get deployment "$STIGMER_OPERATOR_RELEASE" -o go-template='{{if .status.readyReplicas}}{{.status.readyReplicas}}{{end}}' 2>/dev/null)
remaining=$(leftovers "$@")
if [ -z "$remaining" ]; then
  echo "Stigmer's operator has nothing left to clean up"
  exit 0
fi
if [ -z "$ready" ]; then
  echo "Stigmer's operator has no ready replica in $STIGMER_OPERATOR_NAMESPACE, so it cannot clean up; the destroy goes on, and these stay on the cluster:$remaining"
  echo "remove them by hand, in this order: kubectl delete namespace -l app.kubernetes.io/managed-by=stigmer-operator,app.kubernetes.io/component=sandboxes; then each listed prerequisite's definitions and namespace (kubectl get crd -l app.kubernetes.io/managed-by=stigmer-operator)"
  exit 0
fi
echo "waiting up to $${STIGMER_WAIT_SECONDS}s for Stigmer's operator to remove what it installed:$remaining"
deadline=$(( $(date +%s) + STIGMER_WAIT_SECONDS ))
next_report=$(( $(date +%s) + 30 ))
while :; do
  sleep 5
  remaining=$(leftovers "$@")
  if [ -z "$remaining" ]; then
    echo "Stigmer's operator finished cleaning up"
    exit 0
  fi
  now=$(date +%s)
  if [ "$now" -ge "$deadline" ]; then
    echo "Stigmer's operator did not finish within $${STIGMER_WAIT_SECONDS}s; the destroy goes on, and these stay on the cluster:$remaining"
    echo "the operator's log says why: kubectl -n $STIGMER_OPERATOR_NAMESPACE logs deployment/$STIGMER_OPERATOR_RELEASE; remove them by hand, in this order: kubectl delete namespace -l app.kubernetes.io/managed-by=stigmer-operator,app.kubernetes.io/component=sandboxes; then each listed prerequisite's definitions and namespace (kubectl get crd -l app.kubernetes.io/managed-by=stigmer-operator)"
    exit 0
  fi
  if [ "$now" -ge "$next_report" ]; then
    echo "still waiting for:$remaining"
    next_report=$(( now + 30 ))
  fi
done
EOT
}

# The readiness gate: this resource reports done only when the operator can
# serve the resources it defines, so a KubernetesTekton applied next is never
# refused for an operator that is still starting.
#
# WHY A GATE: the release manifest ships no health checks, and the operator's
# webhook registers its own admission configuration at RUNTIME (a
# TektonInstallerSet the webhook creates on start; the configuration is not a
# document of the manifest). Until the webhook runs, a TektonConfig reaches the
# API server with none of its defaults filled in and is refused
# ("spec.pipeline.options: Required value"). On a node pool that starts with no
# node, that window lasts as long as the first node takes to join.
#
# WHAT IT WAITS FOR, in order: both operator Deployments rolled out (their pods
# scheduled and running), then a server-side dry run of a TektonConfig coming
# back with the operator's defaults filled in (spec.profile, which only the
# defaulting webhook sets) -- the exact admission a KubernetesTekton's
# TektonConfig takes, asked without writing anything. The Deployments alone
# are not enough: measured on a fresh install, defaulting went live 5-10s
# after both reported rolled out. When a TektonConfig already exists the dry
# run's create collides with it only after passing admission, and that
# counts: the operator has served one before.
#
# THE STAND-IN: a ConfigMap applied after the CRD group. It depends on the CRDs,
# so a destroy deletes it FIRST and the teardown order the CRD group needs is
# untouched; deleting it does nothing. force_new replaces it on any change, so a
# new operator release gates again. The gate runs as its create-time
# provisioner: a failure marks it for replacement, so the next apply gates
# again. The Pulumi module declares the same ConfigMap and runs the same
# script as a BeforeCreate hook (readiness_gate.go).
#
# kubectl must be on PATH where this module runs; the Planton runner image
# carries it.

locals {
  readiness_gate_config_map_name = "tekton-operator-readiness"

  # The gate's own labels (upstream style, like every document of the release
  # manifest): the ConfigMap is the module's, not the operator's, so it does
  # not borrow the manifest's app labels.
  readiness_gate_labels = {
    "app.kubernetes.io/name"    = "tekton-operator-readiness"
    "app.kubernetes.io/part-of" = "tekton-operator"
  }

  readiness_gate_data = {
    purpose         = "Marks the Tekton operator ready to serve: created only after both operator Deployments rolled out and the API server returned a defaulted TektonConfig. Deleting it changes nothing."
    operatorRelease = local.operator_release
  }

  # The inputs the script reads, beside the kubeconfig environment the
  # deploy already carries (KUBECONFIG, KUBE_CONFIG_PATH, KUBE_CTX).
  readiness_gate_environment = {
    TEKTON_OPERATOR_NAMESPACE  = local.namespace
    TEKTON_OPERATOR_DEPLOYMENT = local.operator_deployment_name
    TEKTON_WEBHOOK_DEPLOYMENT  = local.webhook_deployment_name
  }

  # parity: readiness_gate.go readinessGateScript
  readiness_gate_script = <<EOT
# The Tekton operator's readiness gate. It exits 0 once the operator can
# serve a TektonConfig, and otherwise says what it waited for and what to
# check. Inputs: TEKTON_OPERATOR_NAMESPACE, TEKTON_OPERATOR_DEPLOYMENT,
# TEKTON_WEBHOOK_DEPLOYMENT, and the deploy's kubeconfig environment.
if ! command -v kubectl >/dev/null 2>&1; then
  echo "the Tekton operator's readiness gate needs kubectl on PATH to ask the cluster whether the operator is serving; install kubectl where this deploy runs" >&2
  exit 1
fi
if [ -z "$KUBECONFIG" ] && [ -n "$KUBE_CONFIG_PATH" ]; then
  KUBECONFIG="$KUBE_CONFIG_PATH"
  export KUBECONFIG
fi
set -- --namespace "$TEKTON_OPERATOR_NAMESPACE"
if [ -n "$KUBE_CTX" ]; then
  set -- "$@" --context "$KUBE_CTX"
fi
if ! kubectl "$@" rollout status "deployment/$TEKTON_OPERATOR_DEPLOYMENT" "deployment/$TEKTON_WEBHOOK_DEPLOYMENT" --timeout=10m >&2; then
  echo "the Tekton operator's pods did not become ready within 10m; on a node pool that starts empty, check that a node joined (kubectl get nodes) and that the pods scheduled (kubectl -n $TEKTON_OPERATOR_NAMESPACE get pods)" >&2
  exit 1
fi
answer_file=$(mktemp)
trap 'rm -f "$answer_file"' EXIT
deadline=$(( $(date +%s) + 300 ))
while :; do
  defaulted=$(printf 'apiVersion: operator.tekton.dev/v1alpha1\nkind: TektonConfig\nmetadata:\n  name: config\nspec:\n  targetNamespace: tekton-pipelines\n' | kubectl "$@" create --dry-run=server -o jsonpath='{.spec.profile}' -f - 2>"$answer_file")
  if [ -n "$defaulted" ]; then
    exit 0
  fi
  # A TektonConfig already exists: the operator admitted it before, and the
  # dry run's create only collides with it after passing admission.
  if grep -q '(AlreadyExists)' "$answer_file"; then
    exit 0
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    answer=$(cat "$answer_file")
    if [ -z "$answer" ]; then
      answer="a TektonConfig with none of its defaults filled in"
    fi
    echo "the Tekton operator's pods are running, but its webhook did not start filling in TektonConfig defaults within 5m, so a TektonConfig would be refused; the API server's last answer: $answer; read the webhook's log (kubectl -n $TEKTON_OPERATOR_NAMESPACE logs deployment/$TEKTON_WEBHOOK_DEPLOYMENT)" >&2
    exit 1
  fi
  sleep 5
done
EOT
}

resource "kubectl_manifest" "readiness_gate" {
  yaml_body = yamlencode({
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      name      = local.readiness_gate_config_map_name
      namespace = local.namespace
      labels    = local.readiness_gate_labels
    }
    data = local.readiness_gate_data
  })

  server_side_apply = true
  force_conflicts   = true
  force_new         = true

  # The script rides the environment and the command only evals it, so the
  # deploy log shows one line rather than the whole script.
  provisioner "local-exec" {
    interpreter = ["/bin/sh", "-c"]
    command     = "eval \"$TEKTON_OPERATOR_READINESS_GATE\""
    environment = merge(local.readiness_gate_environment, {
      TEKTON_OPERATOR_READINESS_GATE = local.readiness_gate_script
    })
  }

  depends_on = [kubectl_manifest.crds]
}

package module

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// readinessGate declares the stand-in ConfigMap whose creation waits until
// the operator can serve the resources it defines (Terraform twin:
// kubectl_manifest.readiness_gate in iac/tf/readiness_gate.tf, which carries
// the full rationale). The release manifest ships no health checks, and the
// operator's webhook registers its own admission configuration at RUNTIME, so
// until it runs a TektonConfig reaches the API server undefaulted and is
// refused; on a node pool that starts empty that lasts until a node joins.
//
// The gate is a BeforeCreate hook, not an after hook: a failing before hook
// fails the step before the ConfigMap is recorded, so the next `up` creates
// it again and gates again, on every engine version. ReplaceOnChanges("*")
// with DeleteBeforeReplace mirrors the Terraform twin's force_new, so a new
// operator release gates again there too. The ConfigMap depends on the CRD
// group, so a destroy deletes it first and the CRD group's teardown order is
// untouched; the hook never runs on a destroy or a preview.
func readinessGate(ctx *pulumi.Context, kubernetesProvider pulumi.ProviderResource, crdsGroup pulumi.Resource) error {
	hook, err := ctx.RegisterResourceHook(vars.ReadinessGateConfigMapName, func(*pulumi.ResourceHookArgs) error {
		// Said out loud: the wait can last as long as a first node takes to
		// join, and a silent step would read as a hung deploy.
		_ = ctx.Log.Info("waiting for the Tekton operator to serve: both Deployments rolled out, then a TektonConfig admitted with its defaults", nil)
		if err := runReadinessGate(); err != nil {
			return err
		}
		_ = ctx.Log.Info("the Tekton operator is serving", nil)
		return nil
	}, nil)
	if err != nil {
		return errors.Wrap(err, "registering the tekton-operator readiness gate")
	}

	_, err = kubernetescorev1.NewConfigMap(ctx, vars.ReadinessGateConfigMapName,
		&kubernetescorev1.ConfigMapArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String(vars.ReadinessGateConfigMapName),
				Namespace: pulumi.String(vars.Namespace),
				Labels:    pulumi.ToStringMap(readinessGateLabels()),
			},
			Data: pulumi.ToStringMap(readinessGateData()),
		},
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{crdsGroup}),
		pulumi.ReplaceOnChanges([]string{"*"}),
		pulumi.DeleteBeforeReplace(true),
		pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
			BeforeCreate: []*pulumi.ResourceHook{hook},
		}))
	if err != nil {
		return errors.Wrap(err, "failed to apply the tekton-operator readiness gate")
	}
	return nil
}

// readinessGateLabels are the stand-in's own labels, identical to the
// Terraform twin's readiness_gate_labels (upstream style, like every document
// of the release manifest).
func readinessGateLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":    vars.ReadinessGateConfigMapName,
		"app.kubernetes.io/part-of": vars.Namespace,
	}
}

// readinessGateData is the stand-in's data, identical to the Terraform twin's
// readiness_gate_data. The release is in it so a new operator release
// replaces the ConfigMap and gates again.
func readinessGateData() map[string]string {
	return map[string]string{
		"purpose":         "Marks the Tekton operator ready to serve: created only after both operator Deployments rolled out and the API server returned a defaulted TektonConfig. Deleting it changes nothing.",
		"operatorRelease": vars.OperatorRelease,
	}
}

// readinessGateEnvironment is what the script reads beside the kubeconfig
// environment the deploy already carries (KUBECONFIG, KUBE_CONFIG_PATH,
// KUBE_CTX); the Terraform twin's readiness_gate_environment.
func readinessGateEnvironment() []string {
	return []string{
		"TEKTON_OPERATOR_NAMESPACE=" + vars.Namespace,
		"TEKTON_OPERATOR_DEPLOYMENT=" + vars.OperatorDeploymentName,
		"TEKTON_WEBHOOK_DEPLOYMENT=" + vars.WebhookDeploymentName,
	}
}

// runReadinessGate runs the gate script with the deploy's environment, and on
// failure returns the script's own explanation (bounded, so a chatty kubectl
// never floods the engine's diagnostics).
func runReadinessGate() error {
	cmd := exec.Command("/bin/sh", "-c", readinessGateScript)
	cmd.Env = append(os.Environ(), readinessGateEnvironment()...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	explanation := strings.TrimSpace(string(out))
	if len(explanation) > 4096 {
		explanation = "..." + explanation[len(explanation)-4096:]
	}
	return fmt.Errorf("the Tekton operator is not ready to serve (%v):\n%s", err, explanation)
}

// readinessGateScript is the gate itself. PARITY: the Terraform module carries
// this exact text (iac/tf/readiness_gate.tf, local readiness_gate_script); its
// `# parity:` marker lets the repository's cross-engine parity guard
// (hack/guards/ensure_cross_engine_script_parity.sh) prove the two are
// byte-identical, so the engines can never gate on different conditions.
//
// What it asks, in order: both operator Deployments rolled out, then a
// server-side dry run of a TektonConfig coming back with the operator's
// defaults filled in (spec.profile, which only the defaulting webhook sets)
// -- the exact admission a KubernetesTekton's TektonConfig takes, asked
// without writing anything. On a cluster whose TektonConfig already exists
// the dry run's create collides with it only after passing admission, which
// counts: the operator has served one before.
const readinessGateScript = `# The Tekton operator's readiness gate. It exits 0 once the operator can
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
`

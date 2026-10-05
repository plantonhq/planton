package module

import (
	"github.com/pkg/errors"
	kubernetesotelcollectorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesotelcollector/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one operator-managed OpenTelemetry Collector:
//
//  1. the namespace (optional, create_namespace),
//  2. the opentelemetry.io/v1beta1 OpenTelemetryCollector CR itself —
//     the collector workload (Deployment/DaemonSet/StatefulSet per mode,
//     or sidecar registration), the `<name>-collector` Service with
//     receiver-derived ports, the headless and monitoring Services, and
//     the rendered config ConfigMap are all operator-created from it.
//
// PREREQUISITE: a KubernetesOtelOperator on the cluster (it watches
// every namespace).
func Resources(ctx *pulumi.Context, iacInput *kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the operator's naming budget: the
	// operator derives child names by suffixing
	// ("-collector-networkpolicy" is the longest at 25 characters —
	// feature-gated, but an operator-side gate flip must never break
	// existing collector names; verified in the operator's naming
	// source at the pin) and Kubernetes caps names at 63. Twin: the
	// Terraform module's lifecycle precondition.
	if len(locals.ResourceName) > vars.NameBudget {
		return errors.Errorf(
			"metadata.name %q is %d characters — the operator derives \"<name>-collector-networkpolicy\" "+
				"(25-char suffix) and Kubernetes caps names at 63; use a name of at most %d characters",
			locals.ResourceName, len(locals.ResourceName), vars.NameBudget)
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	if err := otelCollectorCR(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create otel collector custom resource")
	}

	return exportOutputs(ctx, locals)
}

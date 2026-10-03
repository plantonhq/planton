package module

import (
	"github.com/pkg/errors"
	kubernetesflinkdeploymentv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflinkdeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one operator-managed Flink cluster:
//
//  1. the namespace (optional, create_namespace),
//  2. the flink.apache.org/v1beta1 FlinkDeployment CR itself — the
//     JobManager, its TaskManagers, the `<name>-rest` Service and (in
//     application mode) the job they run are all operator-created from
//     it. Exposure composes from Gateway API kinds referencing the
//     exported REST service handle, never from this component.
//
// PREREQUISITE: a KubernetesFlinkOperator whose watch scope covers this
// namespace.
func Resources(ctx *pulumi.Context, iacInput *kubernetesflinkdeploymentv1alpha1.KubernetesFlinkDeploymentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the operator's naming budget: the
	// operator derives `<name>-rest` and `<name>-taskmanager-N-M` child
	// names, and Kubernetes object names cap at 63 characters — past 45
	// the derived names silently break the contract the exported
	// outputs are built on. Twin: the Terraform module's lifecycle
	// precondition.
	if len(locals.ResourceName) > vars.NameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the Flink operator derives child names from it "+
				"(`<name>-rest`, `<name>-taskmanager-N-M`) and Kubernetes names cap at 63 characters; "+
				"use a name of at most %d characters",
			locals.ResourceName, len(locals.ResourceName), vars.NameBudget)
	}

	// FAIL LOUDLY on malformed CPU quantities BEFORE building the CR: the
	// CR's resource.cpu is a NUMBER derived from the spec's quantity
	// strings, and the Terraform twin's tonumber() fails its plan on a
	// value like "abc" — this guard keeps the engines' failure semantics
	// identical instead of silently rendering the default sizing.
	if err := validateCpuQuantities(iacInput.Target.Spec); err != nil {
		return err
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

	if err := flinkDeploymentCR(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create flinkdeployment custom resource")
	}

	return exportOutputs(ctx, locals)
}

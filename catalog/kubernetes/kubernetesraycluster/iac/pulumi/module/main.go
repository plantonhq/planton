package module

import (
	"github.com/pkg/errors"
	kubernetesrayclusterv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesraycluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one operator-managed Ray cluster:
//
//  1. the namespace (optional, create_namespace),
//  2. the ray.io/v1 RayCluster CR itself — the head pod, the worker
//     group pods, the `<name>-head-svc` Service, and (in token auth
//     mode without a bring-your-own Secret) the generated bearer-token
//     Secret named exactly after this resource are all operator-created
//     from it. No ingress resources — exposure composes from Gateway
//     API kinds referencing the exported endpoint handles.
//
// PREREQUISITE: a KubernetesKubeRayOperator whose watch scope covers
// this namespace (cluster-wide with the operator's defaults).
func Resources(ctx *pulumi.Context, iacInput *kubernetesrayclusterv1alpha1.KubernetesRayClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the operator's naming budget: the
	// operator derives `<name>-head-svc` (9-character suffix) and
	// per-group worker pod names (`<name>-<group>-worker-<random>`) —
	// Kubernetes names cap at 63 characters, and 40 keeps every derived
	// name inside the budget with short group names. Twin: the
	// Terraform module's lifecycle precondition.
	if len(locals.ResourceName) > vars.NameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the KubeRay operator derives `<name>-head-svc` and "+
				"per-group worker pod names (`<name>-<group>-worker-…`) from it, and Kubernetes names "+
				"cap at 63 characters; use a name of at most %d characters (and keep worker group names short)",
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

	if err := rayClusterCR(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create raycluster custom resource")
	}

	return exportOutputs(ctx, locals)
}

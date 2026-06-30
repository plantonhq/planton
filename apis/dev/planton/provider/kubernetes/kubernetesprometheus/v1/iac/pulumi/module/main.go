package module

import (
	"github.com/pkg/errors"
	kubernetesprometheusv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesprometheus/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single entry-point consumed by the Planton
// runtime.  It wires together noun-style helpers in a Terraform-like
// top-down order so the flow is easy for DevOps engineers to follow.
func Resources(ctx *pulumi.Context,
	stackInput *kubernetesprometheusv1.KubernetesPrometheusStackInput) error {

	// ----------------------------- locals ---------------------------------
	locals := initializeLocals(ctx, stackInput)

	// ------------------------- kubernetes provider ------------------------
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}
	_ = namespaceDeps

	return nil
}

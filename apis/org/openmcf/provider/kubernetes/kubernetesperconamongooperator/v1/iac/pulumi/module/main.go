package module

import (
	"github.com/pkg/errors"
	kubernetesperconamongooperatorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesperconamongooperator/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the Percona Operator for MongoDB Kubernetes add-on.
func Resources(ctx *pulumi.Context, stackInput *kubernetesperconamongooperatorv1.KubernetesPerconaMongoOperatorStackInput) error {

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

	// ------------------------------ helm ----------------------------------
	if err := helmChart(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to deploy Percona MongoDB Operator Helm chart")
	}

	return nil
}

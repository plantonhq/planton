package module

import (
	"github.com/pkg/errors"
	kubernetesneo4jv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesneo4j/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources sets up all the Pulumi resources for the KubernetesNeo4J component.
func Resources(ctx *pulumi.Context, stackInput *kubernetesneo4jv1.KubernetesNeo4JStackInput) error {
	// Initialize local variables from the stack input.
	locals := initializeLocals(ctx, stackInput)

	// Create the kubernetes provider from the credential in the stack input.
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		stackInput.ProviderConfig,
		"kubernetes",
	)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	// When create_namespace is false, createdNamespace is nil and namespaceDeps is empty.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Install the Neo4j Helm chart, applying user-specified config.
	if err := helmChart(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to deploy neo4j helm chart")
	}

	return nil
}

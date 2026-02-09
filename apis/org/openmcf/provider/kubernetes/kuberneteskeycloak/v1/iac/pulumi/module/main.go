package module

import (
	"github.com/pkg/errors"
	kuberneteskeycloakv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteskeycloak/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kuberneteskeycloakv1.KubernetesKeycloakStackInput) error {
	//initialize locals
	locals := initializeLocals(ctx, stackInput)

	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag.
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

	// TODO: Keycloak Helm chart deployment
	// When implementing, resources should use:
	// - namespaceDeps to depend on the namespace
	// - pulumi.Provider(kubernetesProvider) for Kubernetes resources
	// - pulumi.String(locals.Namespace) for namespace references
	_ = namespaceDeps // will be used when Helm chart deployment is implemented

	return nil
}

package module

import (
	"github.com/pkg/errors"
	kubernetesgatewayapicrdsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgatewayapicrds/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the Gateway API CRDs installation.
func Resources(ctx *pulumi.Context, stackInput *kubernetesgatewayapicrdsv1.KubernetesGatewayApiCrdsStackInput) error {
	// Initialize locals with computed values
	locals := initializeLocals(ctx, stackInput)

	// Set up kubernetes provider from the supplied cluster credential
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// --------------------------------------------------------------------
	// Apply Gateway API CRDs
	//
	// The Gateway API CRDs are cluster-scoped resources that enable
	// Gateway, HTTPRoute, GRPCRoute, and other Gateway API resources.
	//
	// Depending on the channel:
	// - Standard: Gateway, GatewayClass, HTTPRoute, ReferenceGrant
	// - Experimental: Standard + TCPRoute, UDPRoute, TLSRoute, GRPCRoute
	// --------------------------------------------------------------------
	crds, err := pulumiyaml.NewConfigFile(ctx, locals.ResourceName,
		&pulumiyaml.ConfigFileArgs{
			File: locals.ManifestURL,
		},
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to apply Gateway API CRDs")
	}

	// Export outputs
	return exportOutputs(ctx, locals, crds)
}

package module

import (
	"github.com/pkg/errors"
	kubernetesgharunnerscalesetcontrollerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgharunnerscalesetcontroller/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry-point for the GHA Runner Scale Set Controller.
//
// The deployment order is:
// 1. Create namespace (if create_namespace is true)
// 2. Deploy Helm chart with controller configuration
//
// After deployment, users can create AutoScalingRunnerSet resources
// in any namespace to deploy actual GitHub Actions runners.
func Resources(ctx *pulumi.Context,
	in *kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerStackInput) error {

	locals := initializeLocals(ctx, in)

	k8sProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, in.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "setup kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, in, locals, k8sProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Deploy GHA Runner Scale Set Controller via Helm
	if err = ghaRunnerScaleSetController(ctx, locals, k8sProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "deploy gha-runner-scale-set-controller")
	}

	return nil
}

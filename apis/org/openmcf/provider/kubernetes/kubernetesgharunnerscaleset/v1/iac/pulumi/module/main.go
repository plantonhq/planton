package module

import (
	"github.com/pkg/errors"
	kubernetesgharunnerscalesetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesgharunnerscaleset/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry-point for the GHA Runner Scale Set.
//
// The deployment order is:
// 1. Create namespace (if create_namespace is true)
// 2. Create PVCs for persistent volumes
// 3. Create GitHub credentials secret (if not using existing)
// 4. Deploy Helm chart with scale set configuration
//
// After deployment, runners will register with GitHub and start
// picking up jobs matching the scale set name.
func Resources(ctx *pulumi.Context,
	in *kubernetesgharunnerscalesetv1.KubernetesGhaRunnerScaleSetStackInput) error {

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

	// Deploy GHA Runner Scale Set via Helm
	if err = ghaRunnerScaleSet(ctx, locals, k8sProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "deploy gha-runner-scale-set")
	}

	return nil
}

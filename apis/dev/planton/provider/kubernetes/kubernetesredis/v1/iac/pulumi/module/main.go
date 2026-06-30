package module

import (
	"github.com/pkg/errors"
	kubernetesredisv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesredis/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single entry-point consumed by the Planton
// runtime. It wires together noun-style helpers in a Terraform-like
// top-down order so the flow is easy for DevOps engineers to follow.
func Resources(ctx *pulumi.Context, stackInput *kubernetesredisv1.KubernetesRedisStackInput) error {
	// ----------------------------- locals ---------------------------------
	locals := initializeLocals(ctx, stackInput)

	// ------------------------- kubernetes provider ------------------------
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build dependency list so child resources wait for the namespace.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ----------------------------- secrets --------------------------------
	if err := adminPassword(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create admin secret")
	}

	// ------------------------------ helm ----------------------------------
	if err := helmChart(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create helm-chart resources")
	}

	// ----------------------------- ingress --------------------------------
	if locals.KubernetesRedis.Spec.Ingress != nil &&
		locals.KubernetesRedis.Spec.Ingress.Enabled &&
		locals.KubernetesRedis.Spec.Ingress.Hostname != "" {
		if err := loadBalancerIngress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create load-balancer ingress resources")
		}
	}

	return nil
}

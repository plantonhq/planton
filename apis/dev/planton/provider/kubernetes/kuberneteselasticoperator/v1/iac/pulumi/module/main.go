package module

import (
	"github.com/pkg/errors"
	kuberneteselasticoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteselasticoperator/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi entry‑point.
func Resources(ctx *pulumi.Context,
	in *kuberneteselasticoperatorv1.KubernetesElasticOperatorStackInput) error {

	locals := initializeLocals(ctx, in)

	k8sProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, in.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "setup kubernetes provider")
	}

	createdNamespace, err := namespace(ctx, in, locals, k8sProvider)
	if err != nil {
		return errors.Wrap(err, "create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	if err = kubernetesElasticOperator(ctx, locals, k8sProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "deploy elastic operator")
	}

	return nil
}

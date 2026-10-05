package module

import (
	"github.com/pkg/errors"
	kubernetesrabbitmqv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesrabbitmq/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one operator-managed RabbitMQ cluster:
//
//  1. the namespace (optional, create_namespace),
//  2. the rabbitmq.com/v1beta1 RabbitmqCluster itself (the typed SDK
//     catches field/structure drift against the pinned CRD at compile
//     time) — the StatefulSet, both Services, the generated credentials
//     Secret and the erlang-cookie Secret are all operator-created from
//     it. No ingress resources — exposure composes from the client
//     Service's type/annotations or first-class kinds referencing the
//     exported handles.
func Resources(ctx *pulumi.Context, iacInput *kubernetesrabbitmqv1alpha1.KubernetesRabbitMqIacInput) error {
	locals := initializeLocals(ctx, iacInput)

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

	if _, err := cluster(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create rabbitmq cluster")
	}

	exportOutputs(ctx, locals)
	return nil
}

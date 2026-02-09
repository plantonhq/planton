package module

import (
	"github.com/pkg/errors"
	kubernetesgharunnerscalesetcontrollerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesgharunnerscalesetcontroller/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func namespace(ctx *pulumi.Context,
	stackInput *kubernetesgharunnerscalesetcontrollerv1.KubernetesGhaRunnerScaleSetControllerStackInput,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
	if !stackInput.Target.Spec.CreateNamespace {
		return nil, nil
	}
	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:   pulumi.String(locals.Namespace),
				Labels: pulumi.ToStringMap(locals.KubeLabels),
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}
	return createdNamespace, nil
}

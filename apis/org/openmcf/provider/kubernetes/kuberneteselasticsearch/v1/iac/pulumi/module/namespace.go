package module

import (
	"github.com/pkg/errors"
	kuberneteselasticsearchv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteselasticsearch/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func namespace(ctx *pulumi.Context,
	stackInput *kuberneteselasticsearchv1.KubernetesElasticsearchStackInput,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
	if !stackInput.Target.Spec.CreateNamespace {
		return nil, nil
	}

	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaPtrInput(
				&metav1.ObjectMetaArgs{
					Name:   pulumi.String(locals.Namespace),
					Labels: pulumi.ToStringMap(locals.Labels),
				}),
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}

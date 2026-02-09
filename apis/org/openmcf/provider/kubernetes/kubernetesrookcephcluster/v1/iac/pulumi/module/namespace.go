package module

import (
	"github.com/pkg/errors"
	kubernetesrookcephclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesrookcephcluster/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace conditionally creates the Kubernetes namespace based on the
// create_namespace flag.
// Returns the created namespace resource (or nil when create_namespace is false).
func namespace(ctx *pulumi.Context,
	stackInput *kubernetesrookcephclusterv1.KubernetesRookCephClusterStackInput,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
	if !stackInput.Target.Spec.GetCreateNamespace() {
		return nil, nil
	}

	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:   pulumi.String(locals.Namespace),
				Labels: pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.StringMap{
					"pulumi.com/deletionPropagationPolicy": pulumi.String("background"),
				},
			},
		},
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}

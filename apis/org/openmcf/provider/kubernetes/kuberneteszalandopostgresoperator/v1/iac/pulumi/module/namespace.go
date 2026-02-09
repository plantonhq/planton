package module

import (
	"github.com/pkg/errors"
	kuberneteszalandopostgresoperatorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteszalandopostgresoperator/v1"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace conditionally creates the Kubernetes namespace based on the
// create_namespace flag.
// Returns the created namespace resource (or nil when create_namespace is false).
// Terraform equivalent: kubernetes_namespace resource with count.
func namespace(ctx *pulumi.Context,
	stackInput *kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorStackInput,
	locals *Locals,
	kubernetesProvider *pulumikubernetes.Provider,
) (*corev1.Namespace, error) {
	if !stackInput.Target.Spec.CreateNamespace {
		return nil, nil
	}

	createdNamespace, err := corev1.NewNamespace(ctx,
		locals.Namespace,
		&corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.Namespace),
				Labels: pulumi.ToStringMap(locals.KubernetesLabels),
			}),
		},
		pulumi.Provider(kubernetesProvider),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}

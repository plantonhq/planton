package module

import (
	"github.com/pkg/errors"
	kubernetesstrimzikafkaoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesstrimzikafkaoperator/v1"
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
	target *kubernetesstrimzikafkaoperatorv1.KubernetesStrimziKafkaOperator,
	l *locals,
	kubernetesProvider *pulumikubernetes.Provider,
) (*corev1.Namespace, error) {
	if !target.Spec.CreateNamespace {
		return nil, nil
	}

	createdNamespace, err := corev1.NewNamespace(
		ctx,
		l.namespace,
		&corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
				Name:   pulumi.String(l.namespace),
				Labels: l.labels,
			}),
		},
		pulumi.Provider(kubernetesProvider),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", l.namespace)
	}

	return createdNamespace, nil
}

package module

import (
	"github.com/pkg/errors"
	kubernetesclickhousev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesclickhouse/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createOrGetNamespace conditionally creates a Kubernetes namespace or returns nil
// based on the create_namespace flag in the spec.
//
// When create_namespace is true:
//   - Creates a dedicated namespace with resource metadata labels for tracking and organization
//   - All ClickHouse resources will be created within this namespace
//
// When create_namespace is false:
//   - Returns nil without creating it
//   - The namespace must exist before deployment
//   - Resources will be deployed into the existing namespace
func createOrGetNamespace(
	ctx *pulumi.Context,
	locals *Locals,
	spec *kubernetesclickhousev1.KubernetesClickHouseSpec,
	kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
	// If create_namespace is false, use the existing namespace
	if !spec.CreateNamespace {
		return nil, nil
	}

	// Create a new namespace
	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: kubernetesmetav1.ObjectMetaPtrInput(
				&kubernetesmetav1.ObjectMetaArgs{
					Name:   pulumi.String(locals.Namespace),
					Labels: pulumi.ToStringMap(locals.KubernetesLabels),
				}),
		}, pulumi.Provider(kubernetesProvider))

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}

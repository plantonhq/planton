package module

import (
	"github.com/pkg/errors"
	kubernetesharborv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesharbor/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace conditionally creates the Kubernetes namespace that will hold every
// resource in the KubernetesHarbor deployment based on the create_namespace flag.
//
// If create_namespace is true:
//   - Creates a dedicated namespace with resource metadata labels for tracking and organization
//   - All Harbor resources will be created within this namespace
//
// If create_namespace is false:
//   - Returns nil without creating the namespace
//   - The namespace must exist before deployment
//   - Resources will use locals.Namespace directly for the namespace name
func namespace(
	ctx *pulumi.Context,
	locals *Locals,
	spec *kubernetesharborv1.KubernetesHarborSpec,
	kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
	// Only create namespace if the flag is set to true
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

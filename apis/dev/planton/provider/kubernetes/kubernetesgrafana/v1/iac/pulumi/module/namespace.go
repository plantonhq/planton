package module

import (
	"github.com/pkg/errors"
	kubernetesgrafanav1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgrafana/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace conditionally creates a Kubernetes namespace based on the create_namespace flag in the spec.
//
// When create_namespace is true:
//   - Creates a dedicated namespace with resource metadata labels for tracking and organization
//   - All Grafana resources will be created within this namespace
//
// When create_namespace is false:
//   - Returns nil without creating it
//   - The namespace must exist before deployment
//   - Resources will be deployed into the existing namespace
func namespace(
	ctx *pulumi.Context,
	stackInput *kubernetesgrafanav1.KubernetesGrafanaStackInput,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
	// If create_namespace is false, return nil (namespace already exists)
	if !stackInput.Target.Spec.CreateNamespace {
		return nil, nil
	}

	// Create a new namespace
	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: kubernetesmetav1.ObjectMetaPtrInput(
				&kubernetesmetav1.ObjectMetaArgs{
					Name:   pulumi.String(locals.Namespace),
					Labels: pulumi.ToStringMap(locals.Labels),
				}),
		}, pulumi.Provider(kubernetesProvider))

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}

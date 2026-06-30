package module

import (
	"github.com/pkg/errors"
	kubernetessolroperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetessolroperator/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace conditionally creates the Kubernetes namespace based on the
// create_namespace flag.
// Returns the created namespace resource (or nil when create_namespace is false).
// Terraform equivalent: kubernetes_namespace resource with count.
func namespace(ctx *pulumi.Context,
	stackInput *kubernetessolroperatorv1.KubernetesSolrOperatorStackInput,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
) (*corev1.Namespace, error) {
	if !stackInput.Target.Spec.CreateNamespace {
		return nil, nil
	}

	createdNamespace, err := corev1.NewNamespace(ctx, locals.Namespace,
		&corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.Namespace),
				Labels: pulumi.ToStringMap(locals.Labels),
				// CRITICAL: Background Deletion Propagation Policy
				//
				// This annotation prevents namespace deletion from timing out during `pulumi destroy`.
				//
				// Problem: By default, Pulumi uses "Foreground" cascading deletion for namespaces.
				// Kubernetes adds a `foregroundDeletion` finalizer and waits for all resources inside
				// the namespace to be deleted before removing the namespace itself. However, if the
				// Helm release or CRDs are being deleted concurrently, there can be race conditions
				// where finalizers on child resources (like operator-managed CRs) prevent timely cleanup.
				//
				// Solution: Using "background" propagation policy causes Kubernetes to delete the
				// namespace object immediately. The namespace controller then asynchronously cleans up
				// all resources within the namespace. This avoids blocking on child resource finalizers.
				//
				// Reference: https://www.pulumi.com/registry/packages/kubernetes/installation-configuration/
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

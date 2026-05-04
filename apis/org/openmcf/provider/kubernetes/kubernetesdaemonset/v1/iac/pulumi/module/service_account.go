package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceAccount creates a ServiceAccount for the DaemonSet if create_service_account is true.
// Returns the ServiceAccount name to use: non-empty if created or explicitly specified,
// empty string if neither (meaning Kubernetes uses the namespace's "default" SA).
func serviceAccount(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (string, error) {
	spec := locals.KubernetesDaemonSet.Spec

	// If not creating, return the explicitly specified name (may be empty,
	// which tells the DaemonSet to use the namespace default SA).
	if !spec.CreateServiceAccount {
		return spec.ServiceAccountName, nil
	}

	// Determine the service account name for creation
	saName := spec.ServiceAccountName
	if saName == "" {
		saName = locals.KubernetesDaemonSet.Metadata.Name
	}

	// Create the ServiceAccount
	saOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := kubernetescorev1.NewServiceAccount(ctx,
		saName,
		&kubernetescorev1.ServiceAccountArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String(saName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
		},
		saOpts...,
	)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create service account %s", saName)
	}

	return saName, nil
}

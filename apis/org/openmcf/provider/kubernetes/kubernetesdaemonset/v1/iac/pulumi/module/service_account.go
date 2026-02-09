package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceAccount creates a ServiceAccount for the DaemonSet if create_service_account is true.
// Returns the ServiceAccount name to use (either created or specified).
func serviceAccount(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (string, error) {
	spec := locals.KubernetesDaemonSet.Spec

	// Determine the service account name
	saName := spec.ServiceAccountName
	if saName == "" {
		saName = locals.KubernetesDaemonSet.Metadata.Name
	}

	// If not creating, just return the name (assume it exists or use default)
	if !spec.CreateServiceAccount {
		return saName, nil
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

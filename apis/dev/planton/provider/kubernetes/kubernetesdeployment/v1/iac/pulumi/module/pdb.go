package module

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	policyv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/policy/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// podDisruptionBudget creates a PodDisruptionBudget resource if configured.
// PodDisruptionBudgets ensure minimum availability during voluntary disruptions like node maintenance.
func podDisruptionBudget(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// Check if PDB is enabled
	pdbConfig := locals.KubernetesDeployment.Spec.Availability.PodDisruptionBudget
	if pdbConfig == nil || !pdbConfig.Enabled {
		return nil
	}

	pdbSpec := &policyv1.PodDisruptionBudgetSpecArgs{
		Selector: &metav1.LabelSelectorArgs{
			MatchLabels: pulumi.ToStringMap(locals.Labels),
		},
	}

	// Set minAvailable or maxUnavailable (they are mutually exclusive)
	if pdbConfig.MinAvailable != "" {
		pdbSpec.MinAvailable = parseIntOrString(pdbConfig.MinAvailable)
	} else if pdbConfig.MaxUnavailable != "" {
		pdbSpec.MaxUnavailable = parseIntOrString(pdbConfig.MaxUnavailable)
	} else {
		// Default to minAvailable: 1 if neither is specified
		pdbSpec.MinAvailable = pulumi.Int(1)
	}

	pdbArgs := &policyv1.PodDisruptionBudgetArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KubernetesDeployment.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Spec: pdbSpec,
	}

	// Use metadata.name prefix for Pulumi resource ID to avoid state conflicts
	pdbResourceName := fmt.Sprintf("%s-pdb", locals.KubernetesDeployment.Metadata.Name)
	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := policyv1.NewPodDisruptionBudget(ctx,
		pdbResourceName,
		pdbArgs,
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create pod disruption budget")
	}

	return nil
}

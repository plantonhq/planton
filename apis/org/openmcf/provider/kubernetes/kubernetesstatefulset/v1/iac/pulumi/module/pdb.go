package module

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	policyv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/policy/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// podDisruptionBudget creates a PodDisruptionBudget resource if configured.
func podDisruptionBudget(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// Check if PDB is enabled
	if locals.KubernetesStatefulSet.Spec.Availability == nil ||
		locals.KubernetesStatefulSet.Spec.Availability.PodDisruptionBudget == nil ||
		!locals.KubernetesStatefulSet.Spec.Availability.PodDisruptionBudget.Enabled {
		return nil
	}

	pdbConfig := locals.KubernetesStatefulSet.Spec.Availability.PodDisruptionBudget

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
			Name:      pulumi.String(locals.KubernetesStatefulSet.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Spec: pdbSpec,
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := policyv1.NewPodDisruptionBudget(ctx,
		"pdb",
		pdbArgs,
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create pod disruption budget")
	}

	return nil
}

// parseIntOrString converts a string value to the appropriate Pulumi input type
// for Kubernetes IntOrString fields.
func parseIntOrString(value string) pulumi.Input {
	if value == "" {
		return nil
	}

	// Check if it's a percentage
	if strings.HasSuffix(value, "%") {
		return pulumi.String(value)
	}

	// Try to parse as integer
	if intValue, err := strconv.Atoi(value); err == nil {
		return pulumi.Int(intValue)
	}

	// Fallback to string
	return pulumi.String(value)
}

package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nackCrds deploys the NACK JetStream CRDs from the official repository.
// CRDs are deployed as an explicit step (not via Helm chart) to avoid race conditions
// and preview/dry-run issues. This follows Helm's best practices for CRD management.
//
// The CRDs must be registered before:
// 1. The NACK controller can watch them
// 2. Any Stream/Consumer custom resources can be created
//
// Reference: https://helm.sh/docs/chart_best_practices/custom_resource_definitions/
func nackCrds(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	natsHelmChart pulumi.Resource, namespaceDeps []pulumi.ResourceOption) (pulumi.Resource, error) {

	// Skip if NACK controller is not enabled
	if locals.KubernetesNats.Spec.NackController == nil ||
		!locals.KubernetesNats.Spec.NackController.Enabled {
		return nil, nil
	}

	// Deploy CRDs from the versioned URL
	// Using ConfigGroup to apply the multi-document YAML containing all CRDs
	crdsOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		// CRDs should depend on NATS being deployed first
		// (controller needs NATS to connect to)
		pulumi.DependsOn([]pulumi.Resource{natsHelmChart}),
	}, namespaceDeps...)
	crds, err := yaml.NewConfigGroup(ctx, "nack-crds",
		&yaml.ConfigGroupArgs{
			// Fetch CRDs from URL matching the NACK controller version
			// This ensures CRD schema matches the NACK controller version
			Files: []string{locals.NackCrdsUrl},
		},
		crdsOpts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy NACK CRDs")
	}

	return crds, nil
}

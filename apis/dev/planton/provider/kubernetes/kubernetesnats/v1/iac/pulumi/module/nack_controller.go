package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesnatsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesnats/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// nackController deploys the NACK (NATS Controllers for Kubernetes) Helm chart.
// NACK provides a controller that watches JetStream CRDs and reconciles them
// to actual JetStream resources in the NATS cluster.
//
// This must be deployed after:
// 1. NATS Helm chart (provides the NATS server to connect to)
// 2. NACK CRDs (so the controller can watch them)
//
// Reference: https://github.com/nats-io/nack
func nackController(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	nackCrds pulumi.Resource, authPassword pulumi.StringOutput, namespaceDeps []pulumi.ResourceOption) (pulumi.Resource, error) {

	// Skip if NACK controller is not enabled
	if locals.KubernetesNats.Spec.NackController == nil ||
		!locals.KubernetesNats.Spec.NackController.Enabled {
		return nil, nil
	}

	// Build NATS URL for NACK controller
	// If auth is enabled, include credentials in the URL
	var natsUrl pulumi.StringInput
	auth := locals.KubernetesNats.Spec.Auth
	if auth != nil && auth.Enabled && auth.Scheme == kubernetesnatsv1.KubernetesNatsAuthScheme_basic_auth {
		// For basic auth, embed username:password in the URL
		// Format: nats://user:pass@host:port
		serviceName := locals.KubernetesNats.Metadata.Name
		natsUrl = authPassword.ApplyT(func(pass string) string {
			return fmt.Sprintf("nats://%s:%s@%s.%s.svc.cluster.local:%d",
				vars.AdminUsername, pass, serviceName, locals.Namespace, vars.NatsClientPort)
		}).(pulumi.StringOutput)
	} else {
		// No auth or different auth scheme - use plain URL
		natsUrl = pulumi.String(locals.ClientURLInternal)
	}

	// Build Helm values for NACK controller
	values := pulumi.Map{
		"jetstream": pulumi.Map{
			// Enable the JetStream controller
			"enabled": pulumi.Bool(true),
			// NATS URL - points to the NATS service deployed by the NATS Helm chart
			"nats": pulumi.Map{
				"url": natsUrl,
			},
		},
	}

	// Enable control-loop mode if requested
	// Control-loop mode is required for KeyValue, ObjectStore, and Account support
	// and provides more reliable state enforcement
	if locals.KubernetesNats.Spec.NackController.EnableControlLoop {
		values["jetstream"].(pulumi.Map)["additionalArgs"] = pulumi.ToStringArray([]string{"--control-loop"})
	}

	// Deploy NACK Helm chart
	releaseName := fmt.Sprintf("%s-nack", locals.KubernetesNats.Metadata.Name)
	nackOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{nackCrds}),
	}, namespaceDeps...)
	nackRelease, err := helmv3.NewRelease(ctx, releaseName,
		&helmv3.ReleaseArgs{
			Chart:     pulumi.String(vars.NackHelmChartName),
			Version:   pulumi.String(locals.NackHelmChartVersion),
			Namespace: pulumi.String(locals.Namespace),
			RepositoryOpts: helmv3.RepositoryOptsArgs{
				Repo: pulumi.String(vars.NackHelmChartRepoUrl),
			},
			Values: values,
			// Skip CRDs - we install them separately for better control
			SkipCrds: pulumi.Bool(true),
			// Timeout for the release to be ready (5 minutes)
			Timeout: pulumi.Int(300),
		},
		nackOpts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy NACK controller")
	}

	// Export NACK controller status
	ctx.Export(OpNackControllerEnabled, pulumi.Bool(true))
	ctx.Export(OpNackControllerVersion, pulumi.String(locals.NackHelmChartVersion))

	return nackRelease, nil
}

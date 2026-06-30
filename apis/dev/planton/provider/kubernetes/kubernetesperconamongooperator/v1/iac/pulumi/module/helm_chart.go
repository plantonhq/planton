package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// helmChart installs the upstream Percona MongoDB Operator Helm chart and tailors it to the spec.
func helmChart(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	spec := locals.KubernetesPerconaMongoOperator.Spec

	// prepare helm values with resource limits from spec
	helmValues := pulumi.Map{
		// Enable cluster-wide mode to watch all namespaces
		"watchAllNamespaces": pulumi.Bool(true),
		"resources": pulumi.Map{
			"limits": pulumi.Map{
				"cpu":    pulumi.String(spec.Container.Resources.Limits.Cpu),
				"memory": pulumi.String(spec.Container.Resources.Limits.Memory),
			},
			"requests": pulumi.Map{
				"cpu":    pulumi.String(spec.Container.Resources.Requests.Cpu),
				"memory": pulumi.String(spec.Container.Resources.Requests.Memory),
			},
		},
	}

	// deploy the operator via Helm
	// Use computed release name from metadata.name to avoid conflicts when multiple instances share a namespace
	releaseOpts := []pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
	}
	releaseOpts = append(releaseOpts, namespaceDeps...)
	_, err := helm.NewRelease(ctx, locals.HelmReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.HelmReleaseName),
			Namespace:       pulumi.String(locals.Namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(vars.HelmChartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(300),
			Values:          helmValues,
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		releaseOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install percona-operator helm release")
	}

	return nil
}

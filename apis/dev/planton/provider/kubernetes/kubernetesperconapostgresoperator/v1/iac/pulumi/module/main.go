package module

import (
	"github.com/pkg/errors"
	kubernetesperconapostgresoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesperconapostgresoperator/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the Percona Operator for PostgreSQL Kubernetes add-on.
func Resources(ctx *pulumi.Context, stackInput *kubernetesperconapostgresoperatorv1.KubernetesPerconaPostgresOperatorStackInput) error {
	// ----------------------------- locals ---------------------------------
	locals := initializeLocals(ctx, stackInput)

	// ------------------------- kubernetes provider ------------------------
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ helm ----------------------------------
	if err := helmChart(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to deploy Percona Postgres Operator Helm chart")
	}

	return nil
}

// helmChart installs the Percona Postgres Operator Helm chart.
func helmChart(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	// prepare helm values with resource limits from spec
	helmValues := pulumi.Map{
		"resources": pulumi.Map{
			"limits": pulumi.Map{
				"cpu":    pulumi.String(locals.KubernetesPerconaPostgresOperator.Spec.Container.Resources.Limits.Cpu),
				"memory": pulumi.String(locals.KubernetesPerconaPostgresOperator.Spec.Container.Resources.Limits.Memory),
			},
			"requests": pulumi.Map{
				"cpu":    pulumi.String(locals.KubernetesPerconaPostgresOperator.Spec.Container.Resources.Requests.Cpu),
				"memory": pulumi.String(locals.KubernetesPerconaPostgresOperator.Spec.Container.Resources.Requests.Memory),
			},
		},
	}

	// deploy the operator via Helm
	// Use locals.HelmReleaseName to avoid conflicts when multiple instances share a namespace
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
		return errors.Wrap(err, "failed to install kubernetes-percona-postgres-operator helm release")
	}

	return nil
}

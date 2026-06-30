package module

import (
	"github.com/pkg/errors"
	kubernetesperconamysqloperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesperconamysqloperator/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single entry-point consumed by the Planton
// runtime.  It wires together noun-style helpers in a Terraform-like
// top-down order so the flow is easy for DevOps engineers to follow.
func Resources(ctx *pulumi.Context, stackInput *kubernetesperconamysqloperatorv1.KubernetesPerconaMysqlOperatorStackInput) error {
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
		return errors.Wrap(err, "failed to deploy Percona MySQL Operator Helm chart")
	}

	return nil
}

// helmChart installs the Percona MySQL Operator Helm chart.
func helmChart(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	target := locals.KubernetesPerconaMysqlOperator

	// prepare helm values with resource limits from spec
	helmValues := pulumi.Map{
		"resources": pulumi.Map{
			"limits": pulumi.Map{
				"cpu":    pulumi.String(target.Spec.Container.Resources.Limits.Cpu),
				"memory": pulumi.String(target.Spec.Container.Resources.Limits.Memory),
			},
			"requests": pulumi.Map{
				"cpu":    pulumi.String(target.Spec.Container.Resources.Requests.Cpu),
				"memory": pulumi.String(target.Spec.Container.Resources.Requests.Memory),
			},
		},
	}

	// deploy the operator via Helm
	// Use locals.HelmReleaseName for the Kubernetes release name to avoid conflicts
	// when multiple instances are deployed to the same namespace
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
		return errors.Wrap(err, "failed to install percona-mysql-operator helm release")
	}

	return nil
}

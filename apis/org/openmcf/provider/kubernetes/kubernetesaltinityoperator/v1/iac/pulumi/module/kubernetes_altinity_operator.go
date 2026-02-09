package module

import (
	"github.com/pkg/errors"
	kubernetesaltinityoperatorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesaltinityoperator/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the Altinity ClickHouse Operator Kubernetes add-on.
func Resources(ctx *pulumi.Context, stackInput *kubernetesaltinityoperatorv1.KubernetesAltinityOperatorStackInput) error {
	// initialize local values with computed data transformations
	locals := newLocals(stackInput)

	// set up kubernetes provider from the supplied cluster credential
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// deploy the operator via Helm
	if err := helmChart(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to deploy helm chart")
	}

	// export stack outputs
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// helmChart deploys the Altinity ClickHouse Operator Helm chart.
func helmChart(ctx *pulumi.Context, locals *locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	opts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
	}, namespaceDeps...)

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
			Values:          locals.HelmValues,
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install helm release")
	}

	return nil
}

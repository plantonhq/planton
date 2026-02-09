package module

import (
	"github.com/pkg/errors"
	kubernetesrookcephoperatorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesrookcephoperator/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the Rook Ceph Operator Kubernetes add-on.
func Resources(ctx *pulumi.Context, stackInput *kubernetesrookcephoperatorv1.KubernetesRookCephOperatorStackInput) error {
	// Initialize locals with computed values
	locals := initializeLocals(ctx, stackInput)

	// Set up kubernetes provider from the supplied cluster credential
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// --------------------------------------------------------------------
	// Deploy the Rook Ceph Operator via Helm
	// --------------------------------------------------------------------
	helmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
	}, namespaceDeps...)

	_, err = helm.NewRelease(ctx, locals.HelmReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.HelmReleaseName),
			Namespace:       pulumi.String(locals.Namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(locals.ChartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(300), // Rook operator may take longer to deploy
			Values:          pulumi.ToMap(locals.HelmValues),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install rook-ceph-operator helm release")
	}

	return nil
}

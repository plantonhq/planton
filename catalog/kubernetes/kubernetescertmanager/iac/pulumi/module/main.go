package module

import (
	"github.com/pkg/errors"
	kubernetescertmanagerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescertmanager/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs cert-manager from the official Helm chart as a real
// Helm release. The typed spec renders into chart values (values.go); the
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values].
//
// The chart owns ALL of cert-manager's Kubernetes objects, including the
// controller ServiceAccount (serviceAccount.create stays true; the workload
// identity annotation rides serviceAccount.annotations). The module itself
// creates only the optional anchor namespace.
func Resources(ctx *pulumi.Context, iacInput *kubernetescertmanagerv1alpha1.KubernetesCertManagerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ helm release --------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the whole install to be ready — including the
		// startupapicheck hook Job that proves the webhook actually
		// serves. A cert-manager whose webhook is not ready rejects
		// every Issuer/Certificate apply, so a premature "success"
		// would just move the failure downstream.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		WaitForJobs:   pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install cert-manager helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServiceAccountName, pulumi.String(locals.ServiceAccountName))
	ctx.Export(OpClusterResourceNamespace, pulumi.String(locals.ClusterResourceNamespace))

	return nil
}

package module

import (
	"github.com/pkg/errors"
	kubernetesseaweedfsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesseaweedfs/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs SeaweedFS from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); S3
// credentials stay chart-owned (`<name>-s3-secret`, generated once and kept
// on uninstall) or come from a referenced existing config Secret; the
// admin-console credentials materialize as the "<name>-admin-auth" Secret
// (admin_secret.go); the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesseaweedfsv1alpha1.KubernetesSeaweedFsIacInput) error {
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

	// --------------------------- admin credentials ------------------------
	createdAdminSecret, err := adminAuthSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return err
	}

	releaseDeps := namespaceDeps
	if createdAdminSecret != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdAdminSecret}))
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
		// Wait for every tier to become Ready — a store that never starts
		// (bad image, unschedulable pod, unbindable PVC) should fail THIS
		// deploy, not the first S3 request. The budget covers a
		// three-tier cold start plus the bucket-creation hook. SkipAwait
		// false is Helm --wait, stated explicitly to mirror the
		// Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install seaweedfs helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. Service names derive
// from the chart's componentName helper with fullnameOverride pinned to the
// resource name: `<name>-master`, `<name>-filer`, `<name>-s3`,
// `<name>-admin`. The `-s3` Service exists for the embedded and dedicated
// gateway shapes alike.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpS3Endpoint, pulumi.String(locals.S3Endpoint))
	ctx.Export(OpS3CredentialsSecretName, pulumi.String(locals.S3CredentialsSecretName))
	ctx.Export(OpFilerServiceName, pulumi.String(locals.FilerServiceName))
	ctx.Export(OpMasterServiceName, pulumi.String(locals.MasterServiceName))
	ctx.Export(OpAdminEndpoint, pulumi.String(locals.AdminEndpoint))
	ctx.Export(OpAdminAuthSecretName, pulumi.String(locals.AdminAuthSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

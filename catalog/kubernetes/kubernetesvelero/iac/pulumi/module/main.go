package module

import (
	"github.com/pkg/errors"
	kubernetesvelerov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesvelero/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Velero from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); the
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release with
// values = [typed, credentials, helm_values].
//
// The release name is FIXED ("velero"): Velero's CRDs and node-agent are
// cluster-scoped and one server owns the backup records in the store —
// one installation per cluster is an upstream constraint.
func Resources(ctx *pulumi.Context, iacInput *kubernetesvelerov1alpha1.KubernetesVeleroIacInput) error {
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

	var releaseDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ helm release --------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	// When the values carry actual secret material (S3 secret key, GCP
	// key JSON, Azure client secret — all flowing into
	// credentials.secretContents.cloud), mask the WHOLE values map in
	// state. Coarser than field-level masking, but nothing secret can
	// leak through a missed path — and the Terraform module masks the
	// same document with sensitive(). Never log mergedValues.
	var valuesInput pulumi.MapInput = pulumi.ToMap(mergedValues)
	if hasSecretMaterial(locals.Spec) {
		valuesInput = pulumi.ToSecret(pulumi.ToMap(mergedValues)).(pulumi.MapOutput)
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(vars.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: valuesInput,
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the server (and node-agent / upgradeCRDs job when
		// enabled) to become ready — a Velero that never comes up (a
		// ServiceMonitor rendered without the Prometheus operator CRDs, a
		// plugin image that cannot be pulled) should fail THIS deploy
		// with a readiness timeout, not surface later as backups that
		// silently never run. 600s covers the CRD-upgrade job plus plugin
		// pulls on slow registries.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, vars.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install velero helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.ReleaseName))
	// Chart-derived ("velero.serverServiceAccount" helper): with the
	// release fixed to "velero", velero.fullname is "velero" and the
	// server ServiceAccount is "velero-server" — the subject cloud-side
	// keyless bindings are written against.
	ctx.Export(OpServiceAccountName, pulumi.String(vars.ServerServiceAccountName))
	// The default BackupStorageLocation the module renders — what
	// Backup/Schedule resources reference through storageLocation.
	ctx.Export(OpBackupStorageLocationName, pulumi.String(vars.BackupStorageLocationName))

	return nil
}

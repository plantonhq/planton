package module

import (
	"github.com/pkg/errors"
	kubernetesperconamysqloperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesperconamysqloperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Percona Operator for MySQL (based on Percona
// XtraDB Cluster) from the official pxc-operator Helm chart as a single
// Helm release named after metadata.name. The operator reconciles
// PerconaXtraDBCluster custom resources (declared through KubernetesMysql)
// into Galera clusters with automated failover, HAProxy/ProxySQL routing,
// and scheduled XtraBackup backups.
//
// CRD LIFECYCLE: the chart ships the PerconaXtraDBCluster CRDs in its
// Helm-native crds/ directory — installed on first install, never upgraded
// or deleted by Helm. Uninstalling the release therefore NEVER
// cascade-deletes the database clusters (the upstream safety posture).
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release with
// values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesperconamysqloperatorv1alpha1.KubernetesPerconaMysqlOperatorIacInput) error {
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

	var operatorDeps []pulumi.Resource
	if createdNamespace != nil {
		operatorDeps = append(operatorDeps, createdNamespace)
	}

	// --------------------- validation webhook (widened watch) -------------
	// Rendered BEFORE the release so the operator's own registration
	// attempt finds it and only refreshes the CA bundle — see webhook.go
	// for the full lifecycle rationale.
	createdWebhook, err := validationWebhook(ctx, locals, kubernetesProvider, operatorDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create validation webhook configuration")
	}
	if createdWebhook != nil {
		operatorDeps = append(operatorDeps, createdWebhook)
	}

	// ------------------------------ operator release ----------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, &helmv3.ReleaseArgs{
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
		// Wait for the operator to become Available — an operator that
		// never becomes ready (an unpullable image from a private mirror
		// is the classic case) should fail THIS deploy with a readiness
		// timeout, not surface later as PerconaXtraDBCluster resources
		// that mysteriously never reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(operatorDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install pxc-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))

	return nil
}

// dependsOn wraps a possibly-empty dependency list into resource options
// (an empty DependsOn is a valid no-op).
func dependsOn(deps []pulumi.Resource) []pulumi.ResourceOption {
	if len(deps) == 0 {
		return nil
	}
	return []pulumi.ResourceOption{pulumi.DependsOn(deps)}
}

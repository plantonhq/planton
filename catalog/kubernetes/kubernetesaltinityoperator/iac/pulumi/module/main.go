package module

import (
	"github.com/pkg/errors"
	kubernetesaltinityoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesaltinityoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Altinity ClickHouse operator from the official
// altinity-clickhouse-operator Helm chart as a single Helm release named
// after metadata.name. The operator reconciles ClickHouseInstallation and
// ClickHouseKeeperInstallation custom resources (declared through
// KubernetesClickHouse) into running clusters with generated server
// configuration, rolling restarts and per-host StatefulSets.
//
// CRD LIFECYCLE: the module deliberately does NOT manage CRDs — the chart
// ships its four CRDs in the crds/ directory (Helm installs them on first
// install and never deletes them on uninstall, so removing the operator
// never cascade-deletes ClickHouseInstallation resources or their data)
// and its pre-install/pre-upgrade hook job server-side-applies them on
// every install and upgrade. This is the opposite posture from sibling
// operator modules that must own CRDs because their charts template them
// release-owned.
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release with
// values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesaltinityoperatorv1alpha1.KubernetesAltinityOperatorIacInput) error {
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
		// timeout, not surface later as ClickHouse clusters that
		// mysteriously never reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(operatorDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install altinity-clickhouse-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpDeploymentName, pulumi.String(locals.DeploymentName))
	ctx.Export(OpCredentialsSecretName, pulumi.String(locals.CredentialsSecretName))
	ctx.Export(OpMetricsEndpoint, pulumi.String(locals.MetricsEndpoint))

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

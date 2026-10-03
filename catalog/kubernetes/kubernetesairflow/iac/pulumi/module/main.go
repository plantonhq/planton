package module

import (
	"github.com/pkg/errors"
	kubernetesairflowv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesairflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Apache Airflow from the official Helm chart as a real
// Helm release. The typed spec renders into chart values (values.go);
// every credential travels through module-owned Secrets composed BEFORE
// the release (secrets.go) — connection URIs, security keys and the admin
// bootstrap password never appear in rendered values; the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release with
// values = [typed, helm_values, re-pins].
func Resources(ctx *pulumi.Context, iacInput *kubernetesairflowv1alpha1.KubernetesAirflowIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the chart's fullname budget: at the
	// default useStandardNaming=false the fullname IS the release name
	// and child names append fixed suffixes — the longest
	// ("-run-airflow-migrations", 23 chars) pushes names past the
	// Kubernetes 63-character limit when the resource name exceeds 40,
	// failing the deploy midway with API rejections. Twin: the
	// Terraform module's lifecycle precondition.
	if len(locals.ReleaseName) > vars.FullnameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the airflow chart appends child-name suffixes up to 23 characters (\"-run-airflow-migrations\") and Kubernetes caps names at 63, so the deploy would fail midway; use a name of at most %d characters",
			locals.ReleaseName, len(locals.ReleaseName), vars.FullnameBudget)
	}

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

	// ------------------------- module-owned secrets ----------------------
	// Every credential the chart consumes materializes BEFORE the
	// release: the chart's pods reference these Secrets by name at
	// startup (built-in secret env vars), so they must exist first.
	createdSecrets, err := airflowSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create module-owned secrets")
	}

	releaseDependencies := make([]pulumi.Resource, 0, len(createdSecrets)+1)
	if createdNamespace != nil {
		releaseDependencies = append(releaseDependencies, createdNamespace)
	}
	releaseDependencies = append(releaseDependencies, createdSecrets...)

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
		// Wait for the components to become Ready — an install whose
		// migration Job cannot reach the database, whose credential
		// Secrets are misnamed, or whose scheduler crash-loops should
		// fail THIS deploy, not the first pipeline run. The
		// post-install migration + create-user hook Jobs run inside
		// this budget too. SkipAwait false is Helm --wait, stated
		// explicitly to mirror the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(releaseDependencies) > 0 {
		opts = append(opts, pulumi.DependsOn(releaseDependencies))
	}

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install airflow helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The API server Service
// is `<name>-api-server` (the release name IS the chart fullname at the
// default naming scheme); the credential handles are Secret NAMES —
// values stay in-cluster.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	// The admin handle is honest: with the bootstrap user disabled no
	// admin credential exists, so the handle exports EMPTY rather than
	// a name that points at nothing (Terraform twin exports the same
	// empties).
	adminSecretName, adminSecretKey := "", ""
	if locals.AdminCreate {
		adminSecretName, adminSecretKey = locals.AdminSecretName, locals.AdminSecretKey
	}
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpApiServerService, pulumi.String(locals.ApiServerServiceName))
	ctx.Export(OpApiServerEndpoint, pulumi.String(locals.ApiServerEndpoint))
	ctx.Export(OpAdminPasswordSecretName, pulumi.String(adminSecretName))
	ctx.Export(OpAdminPasswordSecretKey, pulumi.String(adminSecretKey))
	ctx.Export(OpMetadataConnectionSecretName, pulumi.String(locals.MetadataConnSecretName))
	ctx.Export(OpBrokerUrlSecretName, pulumi.String(locals.BrokerUrlSecretName))
	ctx.Export(OpFernetKeySecretName, pulumi.String(locals.FernetKeySecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

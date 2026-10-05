package module

import (
	"github.com/pkg/errors"
	kubernetesvalkeyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesvalkey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Valkey from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); declared
// ACL passwords materialize as the "<name>-auth" Kubernetes Secret
// (secrets.go) the chart consumes via auth.usersExistingSecret; the
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values].
//
// The release is named after metadata.name (NOT a fixed chart name) and the
// chart's fullname is pinned to the same value: several Valkey instances
// coexist in one cluster, each rendering its own `<name>`,
// `<name>-headless`, and (replication) `<name>-read` Services.
func Resources(ctx *pulumi.Context, iacInput *kubernetesvalkeyv1alpha1.KubernetesValkeyIacInput) error {
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

	// ------------------------------ auth secret ---------------------------
	createdAuthSecret, err := authSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return err
	}

	releaseDeps := namespaceDeps
	if createdAuthSecret != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdAuthSecret}))
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
		// Wait for the workload to become Ready — a store that never
		// starts (bad image, unschedulable pod, unbindable volume) should
		// fail THIS deploy, not the first client connection. Replication
		// starts pods one at a time (OrderedReady) and each replica
		// full-syncs before Ready, so the budget is sized for a
		// multi-pod StatefulSet, not a single Deployment. SkipAwait
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
		return errors.Wrap(err, "failed to install valkey helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. Two are conditional on
// topology because the chart only renders those Services in replication
// mode: `<name>-read` needs replication with the read service enabled, and
// `<name>-headless` exists only alongside the StatefulSet (the standalone
// Deployment renders no headless Service) — both export empty otherwise.
// The credential handles export empty when auth is off (no Secret exists).
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.ReleaseName))

	readService := ""
	if locals.ReadServiceEnabled {
		readService = locals.ReleaseName + "-read"
	}
	ctx.Export(OpReadService, pulumi.String(readService))

	headlessService := ""
	if locals.ReplicationEnabled {
		headlessService = locals.ReleaseName + "-headless"
	}
	ctx.Export(OpHeadlessService, pulumi.String(headlessService))

	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))

	username := ""
	passwordSecretName := ""
	passwordSecretKey := ""
	if locals.AuthEnabled {
		// "default" is the user plain AUTH <password> maps to — the
		// application-facing credential; its Secret key is the username
		// (the auth Secret's one-key-per-user layout).
		username = "default"
		passwordSecretName = locals.AuthSecretName
		passwordSecretKey = "default"
	}
	ctx.Export(OpUsername, pulumi.String(username))
	ctx.Export(OpPasswordSecretName, pulumi.String(passwordSecretName))
	ctx.Export(OpPasswordSecretKey, pulumi.String(passwordSecretKey))
}

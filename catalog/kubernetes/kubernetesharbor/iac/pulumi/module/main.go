package module

import (
	"github.com/pkg/errors"
	kubernetesharborv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesharbor/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Harbor from the official chart as a real Helm
// release. The typed spec renders into chart values (values.go); every
// module-generated credential is materialized into module-owned Secrets
// BEFORE the release and referenced by NAME through the chart's
// existingSecret sites (auth_secrets.go) — the chart's publicly
// documented default credentials never ship.
//
// EXPOSURE COMPOSES: the module always renders one of the chart's
// in-cluster Service exposure types (ClusterIP/NodePort/LoadBalancer)
// with the chart's nginx front door terminating client traffic; the
// chart's ingress and Gateway API route types are never rendered —
// north-south exposure is a separate composed resource pointed at the
// exported front-door Service.
//
// DESTROY TRUTH (chart-verified at 1.19.1): with the default
// keep_volumes_on_uninstall the registry and jobservice PVCs carry
// `helm.sh/resource-policy: keep` and survive uninstall for a reinstall
// to adopt; the INTERNAL database and Redis volumes are
// StatefulSet-template PVCs Helm never deletes regardless. Retiring an
// install for good means sweeping those PVCs explicitly.
func Resources(ctx *pulumi.Context, iacInput *kubernetesharborv1alpha1.KubernetesHarborIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// NAME BUDGET (chart truth at 1.19.1): the chart truncates its
	// fullname at 63 and then APPENDS component suffixes — the longest,
	// `-jobservice-internal-tls` (24 chars), renders whenever
	// internalTLS runs in auto mode. The Terraform twin enforces the
	// same budget via a precondition.
	if len(locals.ReleaseName) > vars.MaxNameLength {
		return errors.Errorf(
			"metadata.name %q is %d characters; the Harbor chart derives object names by suffixing "+
				"up to 24 characters onto it and Kubernetes caps names at 63 — use a name of at most "+
				"%d characters",
			locals.ReleaseName, len(locals.ReleaseName), vars.MaxNameLength)
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

	var secretDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		secretDeps = append(secretDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// -------------------------- credential secrets ------------------------
	// Created BEFORE the release: the chart reads several of them at
	// install time (template-time lookups) and wires the rest as
	// secretKeyRef env — either way the Secrets must exist first.
	createdSecrets, gen, err := authSecrets(ctx, locals, kubernetesProvider, secretDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create credential secrets")
	}

	releaseDeps := make([]pulumi.Resource, 0, len(createdSecrets)+1)
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, createdNamespace)
	}
	releaseDeps = append(releaseDeps, createdSecrets...)

	// ------------------------------ helm release --------------------------
	mergedValues, err := buildHelmValues(locals, gen)
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
		// Wait for the rollout: every Harbor component self-readies
		// (core's startup probe budgets 60 minutes for first-boot
		// schema migrations — the 900s timeout below covers the normal
		// case; genuinely slow clusters surface honestly as a timeout).
		SkipAwait: pulumi.Bool(false),
		Atomic:    pulumi.Bool(false),
		Timeout:   pulumi.Int(900),
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(releaseDeps) > 0 {
		opts = append(opts, pulumi.DependsOn(releaseDeps))
	}

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install harbor helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

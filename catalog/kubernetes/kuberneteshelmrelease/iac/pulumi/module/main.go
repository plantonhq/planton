package module

import (
	"strings"

	"github.com/pkg/errors"
	kuberneteshelmreleasev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteshelmrelease/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sigs.k8s.io/yaml"
)

// Resources is the main entry point for the Pulumi module.
// It installs the chart as a REAL Helm release (helm.v3.Release): hooks run,
// the release secret is written, and `helm list` sees the release — the
// exact semantic twin of the Terraform module's helm_release resource.
// (The render-only helm.v3.Chart resource is deliberately NOT used: it
// template-renders client-side without creating a release, which silently
// skips hooks and leaves nothing for Helm tooling to manage.)
//
// CRD LIFECYCLE: Helm installs a chart's crds/ directory once and never
// upgrades or removes it. The module owns that surface through the
// catalog's derive-branch primitive (keptcrds): the pinned chart is
// rendered with exactly the values the release installs with, each
// CustomResourceDefinition from the crds/ directory is applied keyed by its
// own name as a kept resource (retained on destroy unless
// crds.keep_on_uninstall is false; re-adopted on reinstall; refused when
// the manifest lowers version below what the cluster carries), and the
// release installs with SkipCrds so Helm never touches them. The chart is
// arbitrary, so the module supplies no render override: CRDs the chart
// templates stay Helm's, refused unless the chart keeps them itself or
// crds.allow_helm_managed accepts them. A chart without CRDs (most charts)
// is ordinary: nothing is applied.
func Resources(ctx *pulumi.Context, iacInput *kuberneteshelmreleasev1alpha1.KubernetesHelmReleaseIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Create kubernetes provider from the credential in the iac-input
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

	spec := locals.Spec

	// ------------------------------ release values ------------------------
	// Built once and used twice: the release installs with them, and the
	// CRD render runs with them, so the derived CRDs can never see
	// different values than the install.
	mergedValues, err := buildHelmValues(spec)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}
	releaseValuesDocument, err := yaml.Marshal(mergedValues)
	if err != nil {
		return errors.Wrap(err, "failed to encode the release values for the CRD render")
	}

	// ------------------------------ CRDs ----------------------------------
	// Derived from the pinned chart and applied kept, ahead of the release
	// (see keptcrds for the mechanics and the failure vocabulary).
	// crds.install false means someone else owns the chart's crds/
	// directory: nothing is derived and the release still skips it.
	crds := spec.GetCrds()
	createdCrds, err := keptcrds.Apply(ctx, keptcrds.Args{
		Source: helmcrds.Source{
			Repository: spec.GetRepo(),
			Chart:      spec.GetChart(),
			Version:    spec.GetVersion(),
			Username:   spec.GetRepositoryUsername(),
			Password:   spec.GetRepositoryPassword(),
			Values:     []string{string(releaseValuesDocument)},
		},
		// The chart is the user's: it may carry no CRDs at all, and CRDs it
		// templates are Helm's unless the spec accepts that.
		Policy: helmcrds.Policy{
			ExpectCRDs:       false,
			AllowHelmManaged: crds != nil && crds.GetAllowHelmManaged(),
		},
		ReleaseName:     locals.ReleaseName,
		Namespace:       locals.Namespace,
		Install:         crds == nil || crds.Install == nil || crds.GetInstall(),
		KeepOnUninstall: crds == nil || crds.KeepOnUninstall == nil || crds.GetKeepOnUninstall(),
		ProviderConfig:  iacInput.ProviderConfig,
		ProviderName:    "kubernetes-crd-upsert",
	})
	if err != nil {
		return errors.Wrap(err, "failed to apply the chart's CRDs")
	}

	// The release depends on the namespace and on the module-owned CRDs
	// (Pulumi equivalent of Terraform depends_on).
	releaseDeps := append([]pulumi.Resource{}, createdCrds...)
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, createdNamespace)
	}
	var releaseDepOptions []pulumi.ResourceOption
	if len(releaseDeps) > 0 {
		releaseDepOptions = append(releaseDepOptions, pulumi.DependsOn(releaseDeps))
	}

	// ------------------------------ helm release --------------------------
	createdRelease, err := helmRelease(ctx, locals, mergedValues, kubernetesProvider, releaseDepOptions)
	if err != nil {
		return errors.Wrap(err, "failed to create helm release")
	}

	exportOutputs(ctx, locals, createdRelease)
	return nil
}

// helmRelease installs the chart with the merged values and the spec's
// lifecycle knobs. Every knob maps 1:1 onto a helm_release argument in the
// Terraform module — keep the two in lockstep.
func helmRelease(ctx *pulumi.Context, locals *Locals, mergedValues map[string]interface{},
	kubernetesProvider pulumi.ProviderResource,
	releaseDeps []pulumi.ResourceOption) (*helmv3.Release, error) {

	spec := locals.Spec

	// PARITY-EXCEPTION: take_ownership is Pulumi-inexpressible at the pinned
	// SDK (helm.v3.ReleaseArgs gained TakeOwnership only in a later
	// pulumi-kubernetes release). The Terraform module implements it. A set
	// field must never be silently dropped, so fail the deploy loudly and
	// route the user to the working engine. Dissolves at the next
	// pulumi-kubernetes SDK upgrade.
	if spec.GetTakeOwnership() {
		return nil, errors.New("take_ownership is not supported by the pulumi engine at the current pulumi-kubernetes SDK version; " +
			"deploy this release with the terraform provisioner, or drop take_ownership")
	}

	valuesInput := pulumi.MapInput(pulumi.ToMap(mergedValues))
	if len(spec.GetSetSensitive()) > 0 {
		// The merged map carries secret entries: mark the WHOLE values map
		// secret in Pulumi state. Coarser than Terraform's per-entry
		// set_sensitive masking, but errs on the safe side — documented
		// behavioral difference, not a parity exception (the release Helm
		// installs is identical).
		valuesInput = pulumi.ToSecret(pulumi.ToMap(mergedValues)).(pulumi.MapOutput)
	}

	// Chart resolution mirrors the Terraform provider exactly:
	// - OCI registry: the chart reference is <repo>/<chart>, no repository opts URL.
	// - HTTP(S) repository: chart name stays bare; the repo URL goes in
	//   repository opts.
	// Private-repo credentials ride repository opts in both forms.
	chartRef := spec.GetChart()
	repositoryOpts := &helmv3.RepositoryOptsArgs{}
	if strings.HasPrefix(spec.GetRepo(), "oci://") {
		chartRef = strings.TrimSuffix(spec.GetRepo(), "/") + "/" + spec.GetChart()
	} else {
		repositoryOpts.Repo = pulumi.String(spec.GetRepo())
	}
	if spec.GetRepositoryUsername() != "" {
		repositoryOpts.Username = pulumi.String(spec.GetRepositoryUsername())
		repositoryOpts.Password = pulumi.ToSecret(pulumi.String(spec.GetRepositoryPassword())).(pulumi.StringOutput)
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:            pulumi.String(locals.ReleaseName),
		Namespace:       pulumi.String(locals.Namespace),
		Chart:           pulumi.String(chartRef),
		Version:         pulumi.String(spec.GetVersion()),
		RepositoryOpts:  repositoryOpts,
		Values:          valuesInput,
		CreateNamespace: pulumi.Bool(false), // the module owns namespace creation (create_namespace flag)

		// Lifecycle knobs — 1:1 with the Terraform module's helm_release
		// arguments. SkipAwait inverts the TF `wait` flag: both engines
		// default to awaiting readiness.
		Atomic:        pulumi.Bool(spec.GetAtomic()),
		CleanupOnFail: pulumi.Bool(spec.GetCleanupOnFail()),
		SkipAwait:     pulumi.Bool(spec.GetSkipAwait()),
		WaitForJobs:   pulumi.Bool(spec.GetWaitForJobs()),
		Timeout:       pulumi.Int(locals.TimeoutSeconds),
		// The module owns the chart's crds/ directory (keptcrds above);
		// Helm never installs its own once-only copy.
		SkipCrds:                 pulumi.Bool(true),
		DependencyUpdate:         pulumi.Bool(spec.GetDependencyUpdate()),
		MaxHistory:               pulumi.Int(locals.MaxHistory),
		Replace:                  pulumi.Bool(spec.GetReplace()),
		ForceUpdate:              pulumi.Bool(spec.GetForceUpdate()),
		ReuseValues:              pulumi.Bool(spec.GetReuseValues()),
		ResetValues:              pulumi.Bool(spec.GetResetValues()),
		DisableWebhooks:          pulumi.Bool(spec.GetDisableWebhooks()),
		DisableOpenapiValidation: pulumi.Bool(spec.GetDisableOpenapiValidation()),
	}
	if spec.GetDescription() != "" {
		releaseArgs.Description = pulumi.String(spec.GetDescription())
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	createdRelease, err := helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to install helm release")
	}

	return createdRelease, nil
}

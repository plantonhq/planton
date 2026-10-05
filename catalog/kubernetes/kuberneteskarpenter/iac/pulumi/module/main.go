package module

import (
	"github.com/pkg/errors"
	kuberneteskarpenterv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskarpenter/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Karpenter from the official OCI-served Helm charts as
// TWO real Helm releases in the same namespace:
//
//  1. "karpenter-crd" (when spec.crds.install, default true) — the
//     companion CRD chart, upstream's supported mechanism for keeping CRDs
//     upgradable (Helm installs the copies bundled inside the main chart
//     once and NEVER upgrades them).
//  2. "karpenter" — the controller chart, installed with SkipCrds
//     unconditionally: the CRD release owns the CRDs, and skipping keeps
//     the controller release's shape deterministic whether or not
//     crds.install is on.
//
// Both release names are FIXED: Karpenter owns the cluster-wide
// karpenter.sh label domain, its CRDs, and node lifecycle — one
// installation per cluster is an upstream constraint.
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic twin
// of the Terraform module's helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kuberneteskarpenterv1alpha1.KubernetesKarpenterIacInput) error {
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

	var releaseDeps []pulumi.Resource
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, createdNamespace)
	}

	// ------------------------------ CRD release ---------------------------
	// OCI ENGINE ASYMMETRY: Pulumi's helm.v3.Release resolves oci://
	// registries through the CHART REFERENCE — the joined
	// "oci://public.ecr.aws/karpenter/<chart>" string with NO
	// RepositoryOpts. The Terraform provider instead takes
	// repository = "oci://public.ecr.aws/karpenter" plus the bare chart
	// name and joins them internally. Same chart bytes, different wiring —
	// keep both sides of the split in lockstep.
	crdReleaseName := ""
	if locals.CrdsInstall {
		crdRelease, err := helmv3.NewRelease(ctx, vars.CrdReleaseName, &helmv3.ReleaseArgs{
			Name:      pulumi.String(vars.CrdReleaseName),
			Namespace: pulumi.String(locals.Namespace),
			Chart:     pulumi.String(vars.HelmOciRepo + "/" + vars.CrdChartName),
			Version:   pulumi.String(locals.ChartVersion),
			Values:    pulumi.ToMap(buildCrdHelmValues(locals)),
			// The module owns namespace creation (create_namespace flag).
			CreateNamespace: pulumi.Bool(false),
			// Same atomic/wait posture as the controller release: a CRD
			// chart that fails to apply should fail THIS deploy cleanly,
			// never leave half the CRDs behind for the controller release
			// to trip over.
			Atomic:        pulumi.Bool(true),
			CleanupOnFail: pulumi.Bool(true),
			Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
		}, append([]pulumi.ResourceOption{
			pulumi.Provider(kubernetesProvider)},
			dependsOn(releaseDeps)...)...)
		if err != nil {
			return errors.Wrap(err, "failed to install karpenter-crd helm release")
		}
		crdReleaseName = vars.CrdReleaseName
		// The controller's pods reconcile NodePools/NodeClaims from the
		// moment they start — the CRDs must exist first.
		releaseDeps = append(releaseDeps, crdRelease)
	}

	// ------------------------------ controller release --------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	_, err = helmv3.NewRelease(ctx, vars.ReleaseName, &helmv3.ReleaseArgs{
		Name:      pulumi.String(vars.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		// Joined OCI reference, no RepositoryOpts — see the CRD release
		// comment for the engine asymmetry.
		Chart:   pulumi.String(vars.HelmOciRepo + "/" + vars.HelmChartName),
		Version: pulumi.String(locals.ChartVersion),
		Values:  pulumi.ToMap(mergedValues),
		// The CRD release owns the CRDs (upstream's upgradable path);
		// skipping the controller chart's bundled copies UNCONDITIONALLY
		// keeps this release's shape deterministic whether or not
		// crds.install is on. Identical skip_crds on the Terraform side.
		SkipCrds: pulumi.Bool(true),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the controller to become Available — a Karpenter that
		// never becomes ready (a ServiceMonitor rendered without the
		// Prometheus operator CRDs, a bad IRSA trust policy) should fail
		// THIS deploy with a readiness timeout, not surface later as pods
		// that stay Pending forever because no nodes ever appear.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(releaseDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install karpenter helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.ReleaseName))
	// Empty when this resource does not own the CRDs (crds.install=false).
	ctx.Export(OpCrdReleaseName, pulumi.String(crdReleaseName))
	// Fixed "karpenter": serviceAccount.name defaults to the chart fullname
	// template, which resolves to the release name because the release name
	// contains the chart name — see vars.ServiceAccountName.
	ctx.Export(OpServiceAccountName, pulumi.String(vars.ServiceAccountName))

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

package module

import (
	"github.com/pkg/errors"
	kubernetesexternaldnsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesexternaldns/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs ExternalDNS from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); declared
// provider credentials materialize as Kubernetes Secrets (secrets.go) that
// the chart's env/volume wiring consumes; the helm_values escape hatch
// merges last with Helm -f semantics — the exact semantic twin of the
// Terraform module's helm_release with values = [typed, helm_values].
//
// The release is named after metadata.name (NOT a fixed chart name):
// multiple ExternalDNS instances per cluster — one per DNS provider / zone
// set, separated by TXT owner IDs — are a first-class upstream pattern.
func Resources(ctx *pulumi.Context, iacInput *kubernetesexternaldnsv1alpha1.KubernetesExternalDnsIacInput) error {
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

	// ------------------------- credential secrets -------------------------
	createdSecrets, err := credentialSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return err
	}

	releaseDeps := namespaceDeps
	if len(createdSecrets) > 0 {
		releaseDeps = append(releaseDeps, pulumi.DependsOn(createdSecrets))
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
		// Wait for the controller Deployment to become Available — a
		// controller that never starts (bad image, unschedulable pod)
		// should fail THIS deploy, not the first record sync. Note the
		// controller validates provider CREDENTIALS at first sync, not at
		// startup, so a live install with wrong credentials still installs
		// green and surfaces in controller logs/records — by design
		// (matching upstream behavior).
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(300),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install external-dns helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServiceAccountName, pulumi.String(locals.ServiceAccountName))

	return nil
}

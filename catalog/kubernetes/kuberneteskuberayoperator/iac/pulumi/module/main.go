package module

import (
	"github.com/pkg/errors"
	kuberneteskuberayoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskuberayoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the KubeRay operator from the official
// kuberay-operator Helm chart as a single Helm release named after
// metadata.name. The operator reconciles RayCluster (declared with
// KubernetesRayCluster), RayJob and RayService custom resources into
// running Ray clusters.
//
// CRD LIFECYCLE: the chart ships its three ray.io CRDs (rayclusters,
// rayjobs, rayservices) from its crds/ DIRECTORY — Helm installs them
// once, never upgrades them on chart upgrades, and LEAVES them (and every
// Ray declaration) on uninstall (no release ownership metadata). That
// upstream posture is exactly the keep-on-uninstall this catalog wants
// for workload-bearing CRDs, so the module neither re-owns nor templates
// them — chart-version bumps that change CRDs are applied manually per
// the upstream release notes. NOTE the CRDs are large (~1MB each) and
// install server-side.
//
// NO WEBHOOK: the operator validates in its reconcile loop. There is no
// admission webhook, no certificate machinery, and no cert-manager
// dependency — a bad RayCluster surfaces on the CR's status conditions,
// not as an admission rejection.
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release values documents.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskuberayoperatorv1alpha1.KubernetesKubeRayOperatorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Fail-loud name budget: 63-char Kubernetes name limit minus the
	// chart's longest derived suffix, "-leader-election" (16 chars, the
	// leader-election Role/RoleBinding). The module pins fullnameOverride
	// to metadata.name, so the budget is exact. Twin of the Terraform
	// module's precondition.
	if len(locals.ReleaseName) > 47 {
		return errors.Errorf("metadata.name %q is %d characters — must be 47 or fewer: the chart derives \"<name>-leader-election\" (16-char suffix) and Kubernetes caps names at 63",
			locals.ReleaseName, len(locals.ReleaseName))
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

	// ------------------------------ operator release ----------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	var operatorDeps []pulumi.Resource
	if createdNamespace != nil {
		operatorDeps = append(operatorDeps, createdNamespace)
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
		// Wait for the operator Deployment to become Available — an
		// unpullable image, a missing ServiceMonitor CRD, or a broken
		// config should fail THIS deploy with a readiness timeout, not
		// surface later as RayClusters that mysteriously never
		// reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(operatorDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install kuberay-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpWatchedNamespaces, pulumi.ToStringArray(locals.WatchNamespaces))

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

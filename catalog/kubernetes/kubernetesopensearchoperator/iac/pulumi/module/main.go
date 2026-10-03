package module

import (
	"github.com/pkg/errors"
	kubernetesopensearchoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopensearchoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sigs.k8s.io/yaml"
)

// Resources installs the OpenSearch Kubernetes Operator from the official
// opensearch-operator Helm chart as a single Helm release named after
// metadata.name. The operator reconciles OpenSearchCluster custom
// resources (declared through KubernetesOpenSearch) into running search
// clusters with managed TLS, security bootstrap, rolling upgrades and
// Dashboards.
//
// CRD LIFECYCLE: the chart templates its CRDs as release-owned resources
// with NO keep-on-uninstall knob, so a Helm-owned install would
// cascade-delete every OpenSearchCluster (and its data) on uninstall. The
// module therefore OWNS the CRDs through the catalog's derive-branch
// primitive (keptcrds): the pinned chart is rendered with the release's own
// values plus installCRDs turned on, each CustomResourceDefinition is
// applied keyed by its own name as a kept resource (retained on destroy
// unless crds.keep_on_uninstall is false; re-adopted on reinstall; refused
// when the manifest lowers chart_version below what the cluster carries),
// and the release installs with skip_crds and installCRDs pinned false so
// Helm never touches them. The release depends on the CRDs so the operator
// never starts against an unregistered API group.
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release values documents.
func Resources(ctx *pulumi.Context, iacInput *kubernetesopensearchoperatorv1alpha1.KubernetesOpenSearchOperatorIacInput) error {
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

	// ------------------------------ release values ------------------------
	// Built once and used twice: the release installs with them, and the
	// CRD render runs with them (plus the CRD switch), so the derived CRDs
	// can never see different values than the install.
	mergedValues, err := buildHelmValues(locals)
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
	// crds.install false is the bring-your-own-CRDs arm: nothing is
	// applied and the release still skips CRDs.
	crds := iacInput.Target.Spec.GetCrds()
	createdCrds, err := keptcrds.Apply(ctx, keptcrds.Args{
		Source: helmcrds.Source{
			Repository:  vars.HelmChartRepo,
			Chart:       vars.HelmChartName,
			Version:     locals.ChartVersion,
			Values:      []string{string(releaseValuesDocument)},
			CRDOverride: vars.CrdRenderOverride,
		},
		// A typed kind knows its chart carries CRDs and pins the switch, so a
		// render that yields none is a failure and nothing is left to Helm.
		Policy:          helmcrds.Policy{ExpectCRDs: true},
		ReleaseName:     locals.ReleaseName,
		Namespace:       locals.Namespace,
		Install:         crds == nil || crds.Install == nil || crds.GetInstall(),
		KeepOnUninstall: crds == nil || crds.KeepOnUninstall == nil || crds.GetKeepOnUninstall(),
		ProviderConfig:  iacInput.ProviderConfig,
		ProviderName:    "kubernetes-crd-upsert",
	})
	if err != nil {
		return errors.Wrap(err, "failed to apply the opensearch operator CRDs")
	}
	operatorDeps := append([]pulumi.Resource{}, createdCrds...)
	if createdNamespace != nil {
		operatorDeps = append(operatorDeps, createdNamespace)
	}

	// ------------------------------ operator release ----------------------
	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns the CRDs (above); Helm must never install its
		// own copy of them, whichever way installCRDs is set.
		SkipCrds: pulumi.Bool(true),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the operator to become Available — an operator that
		// never becomes ready (an unpullable image from a private mirror
		// is the classic case) should fail THIS deploy with a readiness
		// timeout, not surface later as OpenSearch clusters that
		// mysteriously never reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(operatorDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install opensearch-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpDeploymentName, pulumi.String(locals.DeploymentName))

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

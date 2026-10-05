package module

import (
	"github.com/pkg/errors"
	kubernetesoteloperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesoteloperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sigs.k8s.io/yaml"
)

// Resources installs the OpenTelemetry Operator from the official
// opentelemetry-operator Helm chart as a single Helm release named after
// metadata.name. The operator reconciles OpenTelemetryCollector custom
// resources (declared through KubernetesOtelCollector) into running
// collector fleets, defaulting and validating them through its admission
// webhooks.
//
// CRD LIFECYCLE: the chart templates its opentelemetry.io CRDs as
// release-owned resources, so a Helm-owned install would cascade-delete
// every collector declaration on uninstall. The module therefore OWNS the
// CRDs through the catalog's derive-branch primitive (keptcrds): the
// pinned chart is rendered with the release's own values plus the CRD
// switch turned on, each CustomResourceDefinition is applied keyed by its
// own name as a kept resource (retained on destroy unless
// crds.keep_on_uninstall is false; re-adopted on reinstall; refused when
// the manifest lowers chart_version below what the cluster carries), and
// the release installs with skip_crds and crds.create pinned false so
// Helm never touches them. The release depends on the CRDs so the
// operator never starts against an unregistered API group.
//
// THE CONVERSION-TRUST COUPLING (why cert-manager is required): the
// collector CRD carries a version-conversion webhook and the
// cert-manager.io/inject-ca-from annotation, both rendered from THIS
// release's identity because the CRD render runs with the release's own
// values. Because the CRDs are retained past the release's lifetime,
// their conversion trust must be kept current by a RUNNING reconciler —
// cert-manager's CA injector — not by a certificate embedded once at
// install time.
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release values documents.
func Resources(ctx *pulumi.Context, iacInput *kubernetesoteloperatorv1alpha1.KubernetesOtelOperatorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Fail-loud name budget: 63-char Kubernetes name limit minus the
	// chart's longest derived suffix, "-controller-manager-service-cert"
	// (33 chars). The module pins fullnameOverride to metadata.name, so
	// the budget is exact. Twin of the Terraform module's precondition.
	if len(locals.ReleaseName) > 30 {
		return errors.Errorf("metadata.name %q is %d characters — must be 30 or fewer: the chart derives \"<name>-controller-manager-service-cert\" (33-char suffix) and Kubernetes caps names at 63",
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
			APIVersions: vars.CrdRenderApiVersions,
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
		return errors.Wrap(err, "failed to apply the opentelemetry operator CRDs")
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
		// own copy of them, whichever way crds.create is set.
		SkipCrds: pulumi.Bool(true),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the operator to become Available — the manager pod
		// mounts the cert-manager-issued webhook Secret, so an install
		// without a working cert-manager (or with an unpullable image)
		// should fail THIS deploy with a readiness timeout, not surface
		// later as collectors that mysteriously never reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(operatorDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install opentelemetry-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpWebhookService, pulumi.String(locals.WebhookService))
	ctx.Export(OpWebhookCertSecretName, pulumi.String(locals.WebhookCertSecretName))

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

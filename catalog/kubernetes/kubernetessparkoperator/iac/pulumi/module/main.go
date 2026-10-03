package module

import (
	"github.com/pkg/errors"
	kubernetessparkoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessparkoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Apache Spark Kubernetes Operator from the
// official ASF spark-kubernetes-operator Helm chart as a single Helm
// release named after metadata.name. The operator reconciles
// SparkApplication (per-job) and SparkCluster (long-lived) custom
// resources into running Spark workloads.
//
// CRD LIFECYCLE: the chart ships its two spark.apache.org CRDs from its
// crds/ DIRECTORY — Helm installs them once, never upgrades them on chart
// upgrades, and LEAVES them on uninstall (no release ownership metadata).
// That upstream posture is exactly the keep-on-uninstall this catalog
// wants for workload-bearing CRDs, so the module neither re-owns nor
// templates them — chart-version bumps that change CRDs are applied
// manually per the upstream release notes.
//
// NO WEBHOOK: this operator validates in its reconcile loop. There is no
// admission webhook, no certificate machinery, and no cert-manager
// dependency — one less lifecycle to manage and nothing that can
// fail-close the cluster's write path.
//
// MULTI-INSTANCE SAFETY (the RBAC name re-pins): the chart hardcodes all
// its cluster-scoped RBAC names as plain values ("spark-operator-
// clusterrole", …) — a second install anywhere on the cluster would
// collide by construction. buildHelmValues derives every RBAC name from
// the release identity instead, so instances coexist.
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release values documents.
func Resources(ctx *pulumi.Context, iacInput *kubernetessparkoperatorv1alpha1.KubernetesSparkOperatorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Fail-loud name budget: 63-char Kubernetes name limit minus the
	// module's longest derived suffix, "-workload-clusterrole"/
	// "-config-monitor-binding" (23 chars). The module pins
	// fullnameOverride AND every RBAC name to this identity, so the
	// budget is exact. Twin of the Terraform module's precondition.
	if len(locals.ReleaseName) > 40 {
		return errors.Errorf("metadata.name %q is %d characters — must be 40 or fewer: the module derives \"<name>-config-monitor-binding\" (23-char suffix) and Kubernetes caps names at 63",
			locals.ReleaseName, len(locals.ReleaseName))
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Workload namespaces (spec.workload.namespaces) are CHART-created and
	// chart-kept (helm.sh/resource-policy: keep) — deliberately not this
	// module's resources; only the installation namespace is.
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// ------------------------------ operator release ----------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	resourceOptions := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if createdNamespace != nil {
		resourceOptions = append(resourceOptions, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
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
		// Wait for the operator Deployment to become Available — a JVM
		// with a 30s-initial-delay startup probe; an unpullable image or
		// a broken config should fail THIS deploy with a readiness
		// timeout, not surface later as SparkApplications that
		// mysteriously never reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, resourceOptions...)
	if err != nil {
		return errors.Wrap(err, "failed to install spark-kubernetes-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpWorkloadServiceAccount, pulumi.String(locals.WorkloadServiceAccount))
	ctx.Export(OpWatchedNamespaces, pulumi.ToStringArray(locals.WorkloadNamespaces))

	return nil
}

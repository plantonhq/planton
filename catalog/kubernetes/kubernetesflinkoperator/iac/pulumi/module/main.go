package module

import (
	"github.com/pkg/errors"
	kubernetesflinkoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflinkoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Apache Flink Kubernetes Operator from the
// official ASF chart — served per version at
// https://downloads.apache.org/flink/flink-kubernetes-operator-<version>/
// — as a single Helm release named after metadata.name. The operator
// reconciles FlinkDeployment (declared with KubernetesFlinkDeployment)
// and the other flink.apache.org custom resources into running Flink
// clusters.
//
// CRD LIFECYCLE: the chart ships its four flink.apache.org CRDs from its
// crds/ DIRECTORY — Helm installs them once, never upgrades them on
// chart upgrades, and LEAVES them on uninstall (no release ownership
// metadata). That upstream posture is exactly the keep-on-uninstall this
// catalog wants for workload-bearing CRDs, so the module neither re-owns
// nor templates them — chart-version bumps that change CRDs are applied
// manually per the upstream release notes.
//
// THE WEBHOOK LIFECYCLE (chart truth at 1.15.0): with the webhook
// enabled (the upstream default this module keeps), the chart renders
// cert-manager Issuer/Certificate resources UNCONDITIONALLY —
// cert-manager is this kind's registry prerequisite, there is no
// self-signed fallback — and both webhook configurations are
// failurePolicy Fail. webhook_enabled=false removes the webhook, the
// certificate machinery, and the cert-manager dependency; the operator
// still validates in its reconcile loop.
//
// THE KEYSTORE PASSWORD (why this module generates a credential): the
// chart's default webhook keystore Secret is a HARDCODED PUBLIC PASSWORD
// — see keystore_secret.go. The module generates a random password,
// materializes it as a module-owned Secret, and wires
// webhook.keystore.passwordSecretRef at it; useDefaultPassword=false is
// additionally RE-PINNED after the escape-hatch merge (values.go).
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release values documents.
func Resources(ctx *pulumi.Context, iacInput *kubernetesflinkoperatorv1alpha1.KubernetesFlinkOperatorIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Fail-loud name budget: 63-char Kubernetes name limit minus the
	// spec's published 18-char suffix budget (the module's longest
	// derived child name is "<name>-webhook-keystore", a 17-char suffix
	// — the spec contract reserves one extra character). The chart's
	// webhook Service/certificate/issuer names are CHART-FIXED
	// ("flink-operator-webhook-service", "flink-operator-serving-cert",
	// "flink-operator-selfsigned-issuer") — not fullname-derived, so
	// they are excluded from the budget. Twin of the Terraform module's
	// precondition.
	if len(locals.ReleaseName) > 45 {
		return errors.Errorf("metadata.name %q is %d characters — must be 45 or fewer: the module derives \"<name>-webhook-keystore\" and Kubernetes caps names at 63",
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

	var namespaceDeps []pulumi.Resource
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, createdNamespace)
	}

	// Watch namespaces MUST exist before the Helm release: the chart
	// plants job ServiceAccount/Role/RoleBinding INTO each watched
	// namespace (templates/flink/*) and does NOT create the namespaces
	// itself — an absent watch namespace fails the install with
	// "namespaces \"…\" not found" (verified live). The module owns
	// them so a fenced install is self-contained.
	watchNs, err := watchNamespaces(ctx, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create watch namespaces")
	}
	for _, ns := range watchNs {
		namespaceDeps = append(namespaceDeps, ns)
	}

	// ------------------------ webhook keystore secret ---------------------
	// Module-generated credential; nil on the webhook-disabled arm — see
	// keystore_secret.go for the full posture.
	createdKeystoreSecret, err := webhookKeystoreSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create webhook keystore secret")
	}

	releaseDeps := namespaceDeps
	if createdKeystoreSecret != nil {
		releaseDeps = append(releaseDeps, createdKeystoreSecret)
	}

	// ------------------------------ operator release ----------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			// The repository URL carries the chart version — the chart
			// is served from a versioned Apache downloads directory.
			Repo: pulumi.String(locals.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the operator Deployment to become Available — a JVM
		// with a 30s-initial-delay startup probe, plus (webhook arm) a
		// cert-manager certificate the webhook container mounts; an
		// unpullable image, an absent cert-manager, or a broken config
		// should fail THIS deploy with a readiness timeout, not surface
		// later as FlinkDeployments that mysteriously never reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(releaseDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install flink-kubernetes-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpJobServiceAccount, pulumi.String(locals.JobServiceAccount))
	ctx.Export(OpWatchedNamespaces, pulumi.ToStringArray(locals.Spec.GetWatchNamespaces()))
	ctx.Export(OpWebhookService, pulumi.String(locals.WebhookService))

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

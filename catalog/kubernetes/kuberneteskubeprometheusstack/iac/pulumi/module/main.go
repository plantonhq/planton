package module

import (
	"github.com/pkg/errors"
	kuberneteskubeprometheusstackv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskubeprometheusstack/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the kube-prometheus-stack from the official Helm chart
// as a real Helm release. The typed spec renders into chart values
// (values.go); the CRDs ride the chart's crds subchart (install-once,
// keep-on-uninstall); declared remote-write usernames are materialized into
// a module-owned Secret (the Prometheus CRD reads both basic-auth halves
// from Secrets); typed notification credentials are materialized into a
// second module-owned Secret that Alertmanager mounts; the helm_values
// escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, stackInput *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// FAIL LOUDLY on names past the chart's fullname budget: the chart
	// SILENTLY truncates fullnameOverride at 26 characters (headroom for
	// its longest child name), which would break the `<name>-prometheus` /
	// `<name>-grafana` naming contract every exported output is built on.
	// Twin: the Terraform module's lifecycle precondition.
	if len(locals.ReleaseName) > vars.FullnameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the kube-prometheus-stack chart derives child names from it and silently truncates past %d characters, which would break the stack's naming contract; use a name of at most %d characters",
			locals.ReleaseName, len(locals.ReleaseName), vars.FullnameBudget, vars.FullnameBudget)
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var releaseDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------- remote-write username Secret ---------------------
	// Materialized only when a remote-write entry declares basic auth: the
	// Prometheus CRD reads the username from a Secret, but a username is
	// not a secret — the spec accepts it as a plain string and the module
	// owns this Secret (the declared-credentials pattern). Passwords stay
	// in the user's own Secrets.
	if locals.RemoteWriteAuthSecretName != "" {
		usernameData := map[string]string{}
		for i, rw := range locals.Spec.GetPrometheus().GetRemoteWrite() {
			if ba := rw.GetBasicAuth(); ba != nil {
				usernameData[remoteWriteUsernameKey(i)] = ba.GetUsername()
			}
		}
		createdSecret, err := kubernetescorev1.NewSecret(ctx,
			locals.RemoteWriteAuthSecretName,
			&kubernetescorev1.SecretArgs{
				Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(locals.RemoteWriteAuthSecretName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				}),
				StringData: pulumi.ToStringMap(usernameData),
			},
			append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)...)
		if err != nil {
			return errors.Wrap(err, "failed to create remote-write auth secret")
		}
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdSecret}))
	}

	// ------------------- notification credentials Secret ------------------
	// Created BEFORE the release: Alertmanager mounts it through
	// alertmanagerSpec.secrets, and a pod whose Secret volume is missing
	// never starts. The values arrive resolved from managed-secret
	// references; the provider stores Secret data as Pulumi secrets.
	if n := locals.Notifications; n != nil && len(n.SecretData) > 0 {
		createdSecret, err := kubernetescorev1.NewSecret(ctx,
			locals.NotificationsSecretName,
			&kubernetescorev1.SecretArgs{
				Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(locals.NotificationsSecretName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				}),
				StringData: pulumi.ToSecret(pulumi.ToStringMap(n.SecretData)).(pulumi.StringMapOutput),
			},
			append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)...)
		if err != nil {
			return errors.Wrap(err, "failed to create alertmanager notifications secret")
		}
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdSecret}))
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
		// Wait for the whole stack to become Ready — an operator that
		// never starts, an unschedulable Prometheus or an unbindable
		// volume should fail THIS deploy, not the first scrape. The
		// budget covers the operator reconciling the Prometheus and
		// Alertmanager StatefulSets after the release's own resources are
		// up. SkipAwait false is Helm --wait, stated explicitly to mirror
		// the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(900),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install kube-prometheus-stack helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. Every child service
// name derives from the pinned fullname (= the resource name); the
// Prometheus endpoint is the URL Grafana datasources and remote readers
// compose against; the Grafana admin Secret follows the credential arm
// (subchart-generated `<name>-grafana` or the referenced existing Secret).
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpPrometheusService, pulumi.String(locals.PrometheusService))
	ctx.Export(OpPrometheusEndpoint, pulumi.String(locals.PrometheusEndpoint))
	ctx.Export(OpAlertmanagerService, pulumi.String(locals.AlertmanagerService))
	ctx.Export(OpAlertmanagerEndpoint, pulumi.String(locals.AlertmanagerEndpoint))
	ctx.Export(OpGrafanaService, pulumi.String(locals.GrafanaService))
	ctx.Export(OpGrafanaEndpoint, pulumi.String(locals.GrafanaEndpoint))
	ctx.Export(OpGrafanaAdminSecretName, pulumi.String(locals.GrafanaAdminSecretName))
	ctx.Export(OpPrometheusPortForwardCommand, pulumi.String(locals.PrometheusPortForwardCommand))
	ctx.Export(OpGrafanaPortForwardCommand, pulumi.String(locals.GrafanaPortForwardCommand))
}

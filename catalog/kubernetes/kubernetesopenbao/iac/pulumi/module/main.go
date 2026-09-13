package module

import (
	"github.com/pkg/errors"
	kubernetesopenbaov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenbao/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs OpenBao from the official chart as a real Helm
// release. The typed spec renders into chart values (values.go) AND into
// the server's HCL configuration (bao_config.go) — the chart takes
// config as a raw string, so the module owns synthesizing the listener,
// storage, Raft retry_join, seal and telemetry stanzas from typed
// fields.
//
// THE SEAL LIFECYCLE (why this module never waits on the release): a
// fresh OpenBao server is UNINITIALIZED and SEALED, and the chart's
// readiness probe (`bao status`) deliberately fails for sealed servers —
// pod readiness IS the seal status. `bao operator init` + unseal are
// runtime API operations no deployment tool performs; until someone
// performs them the StatefulSet never reports ready. A Helm wait here
// would therefore hang on every fresh install (and atomic would roll it
// back) — the E2E verifier owns initialization and readiness instead.
// The chart keeps sealed pods addressable (publishNotReadyAddresses on
// every Service) exactly so init/unseal can reach them.
func Resources(ctx *pulumi.Context, stackInput *kubernetesopenbaov1alpha1.KubernetesOpenBaoStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// NAME BUDGET (chart truth at 0.28.6): the chart truncates its
	// fullname at 63 then APPENDS Service suffixes — `-internal` (9)
	// always, `-agent-injector-svc` (19) with the injector. Service
	// names cap at 63, so the budget depends on the injector arm. With
	// `backup` declared the CronJob `<name>-backup` must fit Kubernetes'
	// 52-character CronJob cap, which is tighter still. The Terraform
	// twin enforces the same budgets via preconditions.
	maxLen := vars.MaxNameLength
	if locals.Spec.GetInjector().GetEnabled() {
		maxLen = vars.MaxNameLengthWithInjector
	}
	if locals.BackupEnabled && vars.MaxNameLengthWithBackup < maxLen {
		maxLen = vars.MaxNameLengthWithBackup
	}
	if len(locals.ReleaseName) > maxLen {
		return errors.Errorf(
			"metadata.name %q is %d characters; the OpenBao chart derives Service names by suffixing "+
				"(up to 19 characters with the injector enabled) onto it, Kubernetes caps Service names "+
				"at 63 and CronJob names at 52 (the backup CronJob is <name>-backup) — use a name of at most %d characters",
			locals.ReleaseName, len(locals.ReleaseName), maxLen)
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

	var releaseDeps []pulumi.Resource
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, createdNamespace)
	}

	// -------------------- seal credentials secret ------------------------
	// Declared-credential auto-unseal arms materialize their material
	// into a module-owned Secret BEFORE the release; the chart wires
	// each key to the server as an environment variable
	// (extraSecretEnvironmentVars) — the config ConfigMap carries only
	// non-credential seal parameters.
	if sealData := sealSecretData(locals.Spec); sealData != nil {
		sealSecretArgs := &kubernetescorev1.SecretArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(
				&kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(locals.SealCredentialsSecretName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				}),
			StringData: pulumi.ToStringMap(sealData),
		}
		sealSecretOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
		if createdNamespace != nil {
			sealSecretOpts = append(sealSecretOpts, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
		}
		createdSealSecret, err := kubernetescorev1.NewSecret(ctx,
			locals.SealCredentialsSecretName, sealSecretArgs, sealSecretOpts...)
		if err != nil {
			return errors.Wrap(err, "failed to create seal credentials secret")
		}
		releaseDeps = append(releaseDeps, createdSealSecret)
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
		// NEVER wait, NEVER atomic: sealed/uninitialized pods are
		// NotReady BY DESIGN (see the module header) — a wait hangs
		// every fresh install and atomic would then roll it back. The
		// Terraform twin sets wait = false for the same reason.
		SkipAwait: pulumi.Bool(true),
		Atomic:    pulumi.Bool(false),
		Timeout:   pulumi.Int(600),
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(releaseDeps) > 0 {
		opts = append(opts, pulumi.DependsOn(releaseDeps))
	}

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install openbao helm release")
	}

	// ------------------------- backup and restore --------------------------
	// Module-owned, beside the release (never inside it): the chart's own
	// snapshot agent is S3-only, static-keys-only, and cannot restore, so
	// the module renders its own CronJob and Job on the official openbao
	// and rclone images (backup.go, restore.go). The release does not
	// wait on them and they do not wait on the release: a CronJob whose
	// first run fails its login until the operator runs the recipe is the
	// designed day-1 shape, taught by the run's own log.
	if locals.BackupEnabled {
		var backupDeps []pulumi.Resource
		if createdNamespace != nil {
			backupDeps = append(backupDeps, createdNamespace)
		}
		jobDeps, err := backupResources(ctx, locals, kubernetesProvider, backupDeps)
		if err != nil {
			return err
		}
		if locals.RestoreDeclared {
			if err := restoreJob(ctx, locals, kubernetesProvider, jobDeps); err != nil {
				return err
			}
		}
	}

	exportOutputs(ctx, locals)
	return nil
}

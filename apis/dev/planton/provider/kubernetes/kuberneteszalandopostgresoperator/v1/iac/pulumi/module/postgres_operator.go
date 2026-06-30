package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// postgresOperator deploys the Zalando Postgres‑Operator via Helm.
func postgresOperator(ctx *pulumi.Context, locals *Locals, kubernetesProvider *pulumikubernetes.Provider, namespaceDeps []pulumi.ResourceOption) error {
	namespace := locals.Namespace

	// 1. Create backup Secret and ConfigMap if backup_config is specified
	backupConfigMapName, err := createBackupResources(
		ctx,
		locals,
		locals.KubernetesZalandoPostgresOperator.Spec.BackupConfig,
		namespace,
		kubernetesProvider,
		locals.KubernetesLabels,
		namespaceDeps,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create backup resources")
	}

	// 2. Build Helm values with backup ConfigMap if configured
	helmValues := backupConfigMapName.ApplyT(func(cmName string) pulumi.Map {
		baseValues := pulumi.Map{
			"configKubernetes": pulumi.Map{
				"inherited_labels": pulumi.ToStringArray([]string{
					kuberneteslabelkeys.Resource,
					kuberneteslabelkeys.Organization,
					kuberneteslabelkeys.Environment,
					kuberneteslabelkeys.ResourceKind,
					kuberneteslabelkeys.ResourceId,
				}),
			},
		}

		// Add pod_environment_configmap if backup is configured
		if cmName != "" {
			baseValues["configKubernetes"].(pulumi.Map)["pod_environment_configmap"] = pulumi.String(cmName)
		}

		return baseValues
	}).(pulumi.MapOutput)

	// 3. Helm release
	helmOpts := append([]pulumi.ResourceOption{
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
		pulumi.Provider(kubernetesProvider),
	}, namespaceDeps...)
	_, err = helm.NewRelease(ctx,
		"postgres-operator",
		&helm.ReleaseArgs{
			Name:            pulumi.String(vars.HelmChartName),
			Namespace:       pulumi.String(namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(vars.HelmChartVersion),
			RepositoryOpts:  helm.RepositoryOptsArgs{Repo: pulumi.String(vars.HelmChartRepo)},
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(false),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values:          helmValues,
		},
		helmOpts...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create helm release")
	}

	// 4. Export stack‑output(s)
	ctx.Export(OpNamespace, pulumi.String(namespace))

	return nil
}

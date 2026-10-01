package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetespostgresv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespostgres/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one CloudNativePG-managed PostgreSQL cluster:
//
//  1. the namespace (optional),
//  2. declared-credential Secrets (external-cluster passwords, role
//     passwords, the superuser/app passwords, backup keys) — secrets always
//     travel via secretKeyRef, never inline in a custom resource,
//  3. the Barman Cloud ObjectStore resource(s) — the backup destination
//     and, for recovery bootstraps, the restore source,
//  4. the Cluster resource itself (the typed SDK catches field/structure
//     drift against the pinned CRD at compile time), archiving into a
//     backup series of its own,
//  5. the on-demand Backup the series starts from, and one ScheduledBackup
//     per declared schedule.
//
// Ordering matters only for the namespace (everything is namespaced) and
// for the ObjectStores (the instance pods' plugin sidecar resolves them at
// startup); the operator tolerates ScheduledBackups arriving with the
// Cluster.
func Resources(ctx *pulumi.Context, stackInput *kubernetespostgresv1alpha1.KubernetesPostgresStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	credentialSecrets, err := createCredentialSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create credential secrets")
	}

	objectStores, backupStore, err := createObjectStores(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create barman object stores")
	}

	backupSeries := pulumi.String("").ToStringOutput()
	if backupStore != nil {
		backupSeries = backupServerName(locals, backupStore)
	}

	// The Cluster waits for every satellite: credential Secrets must exist
	// before the operator reads them, and the plugin resolves the
	// ObjectStore at reconcile time.
	clusterDeps := namespaceDeps
	if len(credentialSecrets) > 0 || len(objectStores) > 0 {
		deps := make([]pulumi.Resource, 0, len(credentialSecrets)+len(objectStores))
		deps = append(deps, credentialSecrets...)
		deps = append(deps, objectStores...)
		clusterDeps = append(clusterDeps, pulumi.DependsOn(deps))
	}

	createdCluster, err := createCluster(ctx, locals, kubernetesProvider, clusterDeps, backupSeries)
	if err != nil {
		return errors.Wrap(err, "failed to create postgresql cluster")
	}

	afterCluster := append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdCluster}))
	if err := createSeriesStartBackup(ctx, locals, kubernetesProvider, afterCluster, backupSeries); err != nil {
		return err
	}
	if err := createScheduledBackups(ctx, locals, kubernetesProvider, afterCluster); err != nil {
		return errors.Wrap(err, "failed to create scheduled backups")
	}

	exportOutputs(ctx, locals, backupSeries)
	return nil
}

// exportOutputs publishes the composition handles. The username/password
// handles point at the EFFECTIVE application secret: the operator-generated
// `<name>-app` normally, or the module-provided secret when initdb declared
// an owner password (the operator adopts a provided bootstrap secret
// instead of generating one).
func exportOutputs(ctx *pulumi.Context, locals *Locals, backupSeries pulumi.StringOutput) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	ctx.Export(OpRwService, pulumi.String(locals.RwServiceName))
	ctx.Export(OpRoService, pulumi.String(locals.RoServiceName))
	ctx.Export(OpRService, pulumi.String(locals.RServiceName))
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(fmt.Sprintf(
		"kubectl port-forward svc/%s -n %s 5432:5432",
		locals.RwServiceName, locals.Namespace)))
	ctx.Export(OpUsernameSecretName, pulumi.String(locals.EffectiveAppSecretName))
	ctx.Export(OpUsernameSecretKey, pulumi.String("username"))
	ctx.Export(OpPasswordSecretName, pulumi.String(locals.EffectiveAppSecretName))
	ctx.Export(OpPasswordSecretKey, pulumi.String("password"))

	// Populated only when superuser access is enabled — the operator
	// deletes the secret (and blanks the password) otherwise.
	superuserSecretName := ""
	if locals.Spec.GetSuperuser().GetEnabled() {
		superuserSecretName = locals.OperatorSuperuserSecret
		if locals.Spec.GetSuperuser().GetPassword() != "" {
			superuserSecretName = locals.ProvidedSuperuserSecret
		}
	}
	ctx.Export(OpSuperuserSecretName, pulumi.String(superuserSecretName))

	// The backup series this install archives into; empty without backup.
	ctx.Export(OpBackupServerName, backupSeries)
}

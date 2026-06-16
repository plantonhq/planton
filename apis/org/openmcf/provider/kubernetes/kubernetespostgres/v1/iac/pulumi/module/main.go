package module

import (
	"github.com/pkg/errors"
	kubernetespostgresv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	zalandov1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/zalandooperator/kubernetes/acid/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetespostgresv1.KubernetesPostgresStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build namespace dependency list so all namespaced resources wait for the namespace
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Build restore (standby block + STANDBY_* env) and backup (WALG_*/AWS_* env)
	// configurations. Both credential paths create a Secret and reference it via
	// secretKeyRef, so credentials never land in plaintext in the postgresql CR.
	backupConfig := locals.KubernetesPostgres.Spec.BackupConfig

	var restoreConfig *kubernetespostgresv1.KubernetesPostgresRestoreConfig
	if backupConfig != nil {
		restoreConfig = backupConfig.RestoreConfig
	}

	standbyBlock, err := buildRestoreStandbyBlock(restoreConfig)
	if err != nil {
		return errors.Wrap(err, "failed to build restore standby block")
	}

	restoreEnvVars, err := buildRestoreEnvVars(ctx, kubernetesProvider, locals, namespaceDeps, restoreConfig)
	if err != nil {
		return errors.Wrap(err, "failed to build restore environment")
	}

	backupEnvVars, err := buildBackupEnvVars(ctx, kubernetesProvider, locals, namespaceDeps, backupConfig)
	if err != nil {
		return errors.Wrap(err, "failed to build backup environment")
	}

	// Standby env first, then backup env (preserves prior ordering).
	var allEnvVars pulumi.MapArrayInput
	mergedEnvVars := append(restoreEnvVars, backupEnvVars...)
	if len(mergedEnvVars) > 0 {
		allEnvVars = pulumi.MapArray(mergedEnvVars)
	}

	// Convert databases list to map[string]string for Zalando operator
	// The operator expects a map where key=database_name, value=owner_role
	var databasesMap pulumi.StringMapInput
	if len(locals.KubernetesPostgres.Spec.Databases) > 0 {
		databasesMapData := make(map[string]string)
		for _, db := range locals.KubernetesPostgres.Spec.Databases {
			databasesMapData[db.Name] = db.OwnerRole
		}
		databasesMap = pulumi.ToStringMap(databasesMapData)
	}

	// Convert users list to map[string][]string for Zalando operator
	var usersMap pulumi.StringArrayMapInput
	if len(locals.KubernetesPostgres.Spec.Users) > 0 {
		usersMapData := make(map[string][]string)
		for _, user := range locals.KubernetesPostgres.Spec.Users {
			usersMapData[user.Name] = user.Flags
		}
		usersMap = pulumi.ToStringArrayMap(usersMapData)
	}

	//create zalando postgresql resource
	postgresqlArgs := &zalandov1.PostgresqlArgs{
		Metadata: metav1.ObjectMetaArgs{
			// for zolando operator the name is required to be always prefixed by teamId
			// a kubernetes service with the same name is created by the operator
			Name:      pulumi.Sprintf("%s-%s", vars.TeamId, locals.KubernetesPostgres.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Spec: zalandov1.PostgresqlSpecArgs{
			NumberOfInstances: pulumi.Int(locals.KubernetesPostgres.Spec.Container.Replicas),
			Patroni:           zalandov1.PostgresqlSpecPatroniArgs{},
			PodAnnotations: pulumi.ToStringMap(map[string]string{
				"postgres-cluster-id": locals.KubernetesPostgres.Metadata.Name,
			}),
			Postgresql: zalandov1.PostgresqlSpecPostgresqlArgs{
				Version: pulumi.String(vars.PostgresVersion),
				Parameters: pulumi.StringMap{
					"max_connections": pulumi.String("200"),
				},
			},
			Resources: zalandov1.PostgresqlSpecResourcesArgs{
				Limits: zalandov1.PostgresqlSpecResourcesLimitsArgs{
					Cpu:    pulumi.String(locals.KubernetesPostgres.Spec.Container.Resources.Limits.Cpu),
					Memory: pulumi.String(locals.KubernetesPostgres.Spec.Container.Resources.Limits.Memory),
				},
				Requests: zalandov1.PostgresqlSpecResourcesRequestsArgs{
					Cpu:    pulumi.String(locals.KubernetesPostgres.Spec.Container.Resources.Requests.Cpu),
					Memory: pulumi.String(locals.KubernetesPostgres.Spec.Container.Resources.Requests.Memory),
				},
			},
			TeamId: pulumi.String(vars.TeamId),
			Volume: zalandov1.PostgresqlSpecVolumeArgs{
				Size: pulumi.String(locals.KubernetesPostgres.Spec.Container.DiskSize),
			},
			// Add standby block if restore is enabled (for disaster recovery)
			Standby: standbyBlock,
			// Merge backup and restore environment variables
			Env: allEnvVars,
			// Databases to create with their owner roles
			Databases: databasesMap,
			// Users/roles to create (must be declared before being used as database owners)
			Users: usersMap,
		},
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err = zalandov1.NewPostgresql(ctx,
		"database",
		postgresqlArgs,
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create postgresql")
	}

	if locals.KubernetesPostgres.Spec.Ingress == nil ||
		!locals.KubernetesPostgres.Spec.Ingress.Enabled ||
		locals.KubernetesPostgres.Spec.Ingress.Hostname == "" {
		//if ingress is not enabled, no load-balancer resource is required. so just exit the function.
		return nil
	}

	if err := ingress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create ingress")
	}
	return nil
}

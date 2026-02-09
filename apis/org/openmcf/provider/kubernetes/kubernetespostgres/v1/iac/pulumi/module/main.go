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

	// Build restore configuration (standby block + STANDBY_* env vars)
	// Pass operator-level bucket as fallback (if available from stack input)
	// TODO: Extract operator bucket from stackInput if operator config is available
	operatorBucket := ""
	var restoreConfig *kubernetespostgresv1.KubernetesPostgresRestoreConfig
	if locals.KubernetesPostgres.Spec.BackupConfig != nil {
		restoreConfig = locals.KubernetesPostgres.Spec.BackupConfig.Restore
	}
	standbyBlock, standbyEnvVars, err := buildRestoreConfig(restoreConfig, operatorBucket)
	if err != nil {
		return errors.Wrap(err, "failed to build restore configuration")
	}

	// Build backup environment variables (existing function)
	backupEnvVars := buildBackupEnvVars(locals.KubernetesPostgres.Spec.BackupConfig, locals.KubernetesPostgres.Metadata.Name)

	// Merge backup and restore environment variables
	var allEnvVars pulumi.MapArrayInput
	if standbyEnvVars != nil && backupEnvVars != nil {
		// Both sets of env vars exist, merge them
		backupArray, ok := backupEnvVars.(pulumi.MapArray)
		if ok {
			allEnvVars = pulumi.MapArray(append(standbyEnvVars, backupArray...))
		} else {
			allEnvVars = pulumi.MapArray(standbyEnvVars)
		}
	} else if standbyEnvVars != nil {
		// Only standby env vars
		allEnvVars = pulumi.MapArray(standbyEnvVars)
	} else {
		// Only backup env vars (or none)
		allEnvVars = backupEnvVars
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

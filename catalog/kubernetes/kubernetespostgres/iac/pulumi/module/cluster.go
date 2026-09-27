package module

import (
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetespostgresv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespostgres/v1alpha1"
	postgresqlv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/cloudnativepg/kubernetes/postgresql/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createCluster renders the postgresql.cnpg.io/v1 Cluster resource with the
// typed crd2pulumi SDK (field/structure drift against the pinned CRD fails
// at compile time). Unset optionals are omitted entirely so the apiserver
// applies the CRD's own defaults — presence discipline mirrors the
// Terraform module's null-prune rendering byte for byte.
func createCluster(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
) (pulumi.Resource, error) {
	spec := locals.Spec

	clusterSpec := postgresqlv1.ClusterSpecArgs{
		Instances: pulumi.Int(int(spec.GetInstances())),
		Storage:   buildStorage(spec.GetStorage()),
		// The operator stamps these on every object it creates (the
		// instance pods, their volumes, the Services), so the database's
		// pods name their organization and environment.
		InheritedMetadata: postgresqlv1.ClusterSpecInheritedMetadataArgs{
			Labels: pulumi.ToStringMap(locals.Labels),
		},
	}

	if spec.GetImageName() != "" {
		clusterSpec.ImageName = pulumi.String(spec.GetImageName())
	}

	if walStorage := spec.GetWalStorage(); walStorage != nil {
		clusterSpec.WalStorage = postgresqlv1.ClusterSpecWalStorageArgs{
			Size:               pulumi.String(walStorage.GetSize()),
			StorageClass:       storageClassOrNil(walStorage),
			ResizeInUseVolumes: resizeOrNil(walStorage),
		}
	}

	if resources := spec.GetResources(); resources != nil {
		// Absent quantities are OMITTED, never rendered empty: CNPG's
		// mutating webhook rejects "" with `quantities must match the
		// regular expression ...` (verified live against a spec
		// declaring only limits.memory).
		resourcesArgs := postgresqlv1.ClusterSpecResourcesArgs{}
		if limits := quantityMap(resources.GetLimits()); limits != nil {
			resourcesArgs.Limits = limits
		}
		if requests := quantityMap(resources.GetRequests()); requests != nil {
			resourcesArgs.Requests = requests
		}
		clusterSpec.Resources = resourcesArgs
	}

	if postgresql := buildPostgresql(spec.GetPostgresql()); postgresql != nil {
		clusterSpec.Postgresql = postgresql
	}

	if bootstrap := buildBootstrap(locals); bootstrap != nil {
		clusterSpec.Bootstrap = bootstrap
	}

	if externalClusters := buildExternalClusters(locals); len(externalClusters) > 0 {
		clusterSpec.ExternalClusters = externalClusters
	}

	// Superuser posture: the enable flag and (optionally) the provided
	// password secret. The operator blanks the postgres password and
	// deletes the secret whenever access is disabled.
	if superuser := spec.GetSuperuser(); superuser != nil && superuser.GetEnabled() {
		clusterSpec.EnableSuperuserAccess = pulumi.Bool(true)
		if superuser.GetPassword() != "" {
			clusterSpec.SuperuserSecret = postgresqlv1.ClusterSpecSuperuserSecretArgs{
				Name: pulumi.String(locals.ProvidedSuperuserSecret),
			}
		}
	}

	if roles := buildManagedRoles(locals); len(roles) > 0 {
		clusterSpec.Managed = postgresqlv1.ClusterSpecManagedArgs{
			Roles: roles,
		}
	}

	// The Barman Cloud plugin wiring: designating the plugin as the WAL
	// archiver is what starts continuous archiving into the ObjectStore.
	if spec.GetBackup() != nil {
		clusterSpec.Plugins = postgresqlv1.ClusterSpecPluginsArray{
			postgresqlv1.ClusterSpecPluginsArgs{
				Name:          pulumi.String(vars.BarmanCloudPluginName),
				IsWALArchiver: pulumi.Bool(true),
				Parameters: pulumi.StringMap{
					"barmanObjectName": pulumi.String(locals.BackupObjectStoreName),
				},
			},
		}
	}

	// Keyless cloud identity rides the instance ServiceAccount: the
	// operator creates one ServiceAccount per cluster (named after it),
	// and the template's annotations are what the cloud webhooks key on.
	if annotations := workloadIdentityAnnotations(spec.GetWorkloadIdentity()); len(annotations) > 0 {
		clusterSpec.ServiceAccountTemplate = postgresqlv1.ClusterSpecServiceAccountTemplateArgs{
			Metadata: postgresqlv1.ClusterSpecServiceAccountTemplateMetadataArgs{
				Annotations: pulumi.ToStringMap(annotations),
			},
		}
	}

	if certificates := spec.GetCertificates(); certificates != nil {
		certificatesArgs := postgresqlv1.ClusterSpecCertificatesArgs{}
		hasAny := false
		if certificates.GetServerTlsSecret().GetValue() != "" {
			certificatesArgs.ServerTLSSecret = pulumi.String(certificates.GetServerTlsSecret().GetValue())
			hasAny = true
		}
		if certificates.GetServerCaSecret() != "" {
			certificatesArgs.ServerCASecret = pulumi.String(certificates.GetServerCaSecret())
			hasAny = true
		}
		if len(certificates.GetServerAltDnsNames()) > 0 {
			certificatesArgs.ServerAltDNSNames = pulumi.ToStringArray(certificates.GetServerAltDnsNames())
			hasAny = true
		}
		if hasAny {
			clusterSpec.Certificates = certificatesArgs
		}
	}

	if monitoring := spec.GetMonitoring(); monitoring != nil {
		monitoringArgs := postgresqlv1.ClusterSpecMonitoringArgs{}
		hasAny := false
		if monitoring.GetTlsEnabled() {
			monitoringArgs.Tls = postgresqlv1.ClusterSpecMonitoringTlsArgs{
				Enabled: pulumi.Bool(true),
			}
			hasAny = true
		}
		if monitoring.GetDisableDefaultQueries() {
			monitoringArgs.DisableDefaultQueries = pulumi.Bool(true)
			hasAny = true
		}
		if hasAny {
			clusterSpec.Monitoring = monitoringArgs
		}
	}

	if affinity := buildAffinity(spec.GetScheduling()); affinity != nil {
		clusterSpec.Affinity = affinity
	}
	if spec.GetScheduling().GetPriorityClassName() != "" {
		clusterSpec.PriorityClassName = pulumi.String(spec.GetScheduling().GetPriorityClassName())
	}

	if updateStrategy := spec.GetUpdateStrategy(); updateStrategy != nil {
		if updateStrategy.GetPrimaryUpdateStrategy() != "" {
			clusterSpec.PrimaryUpdateStrategy = pulumi.String(updateStrategy.GetPrimaryUpdateStrategy())
		}
		if updateStrategy.GetPrimaryUpdateMethod() != "" {
			clusterSpec.PrimaryUpdateMethod = pulumi.String(updateStrategy.GetPrimaryUpdateMethod())
		}
	}

	// enable_pdb defaults to true both here and upstream; only an explicit
	// false needs rendering (presence-sensitive — an absent key already
	// means true to the operator).
	if spec.EnablePdb != nil && !spec.GetEnablePdb() {
		clusterSpec.EnablePDB = pulumi.Bool(false)
	}

	if len(spec.GetImagePullSecrets()) > 0 {
		pullSecrets := postgresqlv1.ClusterSpecImagePullSecretsArray{}
		for _, name := range spec.GetImagePullSecrets() {
			pullSecrets = append(pullSecrets, postgresqlv1.ClusterSpecImagePullSecretsArgs{
				Name: pulumi.String(name),
			})
		}
		clusterSpec.ImagePullSecrets = pullSecrets
	}

	return postgresqlv1.NewCluster(ctx, locals.ClusterName,
		&postgresqlv1.ClusterArgs{
			Metadata: kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.ClusterName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: clusterSpec,
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)...)
}

func buildStorage(storage *kubernetespostgresv1alpha1.KubernetesPostgresStorage) postgresqlv1.ClusterSpecStorageArgs {
	return postgresqlv1.ClusterSpecStorageArgs{
		Size:               pulumi.String(storage.GetSize()),
		StorageClass:       storageClassOrNil(storage),
		ResizeInUseVolumes: resizeOrNil(storage),
	}
}

func storageClassOrNil(storage *kubernetespostgresv1alpha1.KubernetesPostgresStorage) pulumi.StringPtrInput {
	if storage.GetStorageClass().GetValue() == "" {
		return nil
	}
	return pulumi.String(storage.GetStorageClass().GetValue())
}

// resizeOrNil renders resize_in_use_volumes only when it DIVERGES from the
// shared default (true) — the CRD default already covers the true case, and
// omitting it keeps the rendered resource minimal on both engines.
func resizeOrNil(storage *kubernetespostgresv1alpha1.KubernetesPostgresStorage) pulumi.BoolPtrInput {
	if storage.ResizeInUseVolumes != nil && !storage.GetResizeInUseVolumes() {
		return pulumi.Bool(false)
	}
	return nil
}

func buildPostgresql(config *kubernetespostgresv1alpha1.KubernetesPostgresServerConfig) postgresqlv1.ClusterSpecPostgresqlPtrInput {
	if config == nil {
		return nil
	}

	postgresqlArgs := postgresqlv1.ClusterSpecPostgresqlArgs{}
	hasAny := false

	if len(config.GetParameters()) > 0 {
		postgresqlArgs.Parameters = pulumi.ToStringMap(config.GetParameters())
		hasAny = true
	}
	if len(config.GetPgHba()) > 0 {
		postgresqlArgs.Pg_hba = pulumi.ToStringArray(config.GetPgHba())
		hasAny = true
	}
	if len(config.GetPgIdent()) > 0 {
		postgresqlArgs.Pg_ident = pulumi.ToStringArray(config.GetPgIdent())
		hasAny = true
	}
	if len(config.GetSharedPreloadLibraries()) > 0 {
		postgresqlArgs.Shared_preload_libraries = pulumi.ToStringArray(config.GetSharedPreloadLibraries())
		hasAny = true
	}
	if synchronous := config.GetSynchronous(); synchronous != nil {
		postgresqlArgs.Synchronous = postgresqlv1.ClusterSpecPostgresqlSynchronousArgs{
			Method:         pulumi.String(synchronous.GetMethod()),
			Number:         pulumi.Int(int(synchronous.GetNumber())),
			DataDurability: pulumi.String(synchronous.GetDataDurability()),
		}
		hasAny = true
	}
	if config.GetEnableAlterSystem() {
		postgresqlArgs.EnableAlterSystem = pulumi.Bool(true)
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return postgresqlArgs
}

func buildAffinity(scheduling *kubernetespostgresv1alpha1.KubernetesPostgresScheduling) postgresqlv1.ClusterSpecAffinityPtrInput {
	if scheduling == nil {
		return nil
	}

	affinityArgs := postgresqlv1.ClusterSpecAffinityArgs{}
	hasAny := false

	// The operator's anti-affinity is on by default; only the TYPE and the
	// topology key need rendering when the user tunes them.
	if scheduling.GetAntiAffinityType() != "" && scheduling.GetAntiAffinityType() != "preferred" {
		affinityArgs.PodAntiAffinityType = pulumi.String(scheduling.GetAntiAffinityType())
		hasAny = true
	}
	if scheduling.GetTopologyKey() != "" {
		affinityArgs.TopologyKey = pulumi.String(scheduling.GetTopologyKey())
		hasAny = true
	}
	if len(scheduling.GetNodeSelector()) > 0 {
		affinityArgs.NodeSelector = pulumi.ToStringMap(scheduling.GetNodeSelector())
		hasAny = true
	}
	if len(scheduling.GetTolerations()) > 0 {
		tolerations := postgresqlv1.ClusterSpecAffinityTolerationsArray{}
		for _, toleration := range scheduling.GetTolerations() {
			tolerationArgs := postgresqlv1.ClusterSpecAffinityTolerationsArgs{}
			if toleration.GetKey() != "" {
				tolerationArgs.Key = pulumi.String(toleration.GetKey())
			}
			if toleration.GetOperator() != "" {
				tolerationArgs.Operator = pulumi.String(toleration.GetOperator())
			}
			if toleration.GetValue() != "" {
				tolerationArgs.Value = pulumi.String(toleration.GetValue())
			}
			if toleration.GetEffect() != "" {
				tolerationArgs.Effect = pulumi.String(toleration.GetEffect())
			}
			if toleration.TolerationSeconds != nil {
				tolerationArgs.TolerationSeconds = pulumi.Int(int(toleration.GetTolerationSeconds()))
			}
			tolerations = append(tolerations, tolerationArgs)
		}
		affinityArgs.Tolerations = tolerations
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return affinityArgs
}

func buildManagedRoles(locals *Locals) postgresqlv1.ClusterSpecManagedRolesArray {
	roles := postgresqlv1.ClusterSpecManagedRolesArray{}
	for _, role := range locals.Spec.GetRoles() {
		roleArgs := postgresqlv1.ClusterSpecManagedRolesArgs{
			Name: pulumi.String(role.GetName()),
		}
		if role.GetComment() != "" {
			roleArgs.Comment = pulumi.String(role.GetComment())
		}
		if role.GetEnsure() != "" && role.GetEnsure() != "present" {
			roleArgs.Ensure = pulumi.String(role.GetEnsure())
		}
		if role.GetPassword() != "" {
			roleArgs.PasswordSecret = postgresqlv1.ClusterSpecManagedRolesPasswordSecretArgs{
				Name: pulumi.String(locals.ClusterName + "-role-" + role.GetName()),
			}
		}
		if role.GetDisablePassword() {
			roleArgs.DisablePassword = pulumi.Bool(true)
		}
		if role.GetLogin() {
			roleArgs.Login = pulumi.Bool(true)
		}
		if role.GetSuperuser() {
			roleArgs.Superuser = pulumi.Bool(true)
		}
		if role.GetCreatedb() {
			roleArgs.Createdb = pulumi.Bool(true)
		}
		if role.GetCreaterole() {
			roleArgs.Createrole = pulumi.Bool(true)
		}
		if role.GetReplication() {
			roleArgs.Replication = pulumi.Bool(true)
		}
		if role.GetBypassrls() {
			roleArgs.Bypassrls = pulumi.Bool(true)
		}
		if len(role.GetInRoles()) > 0 {
			roleArgs.InRoles = pulumi.ToStringArray(role.GetInRoles())
		}
		// -1 (unlimited) is the engine default; render only a real limit.
		if role.ConnectionLimit != nil && role.GetConnectionLimit() != -1 {
			roleArgs.ConnectionLimit = pulumi.Int(int(role.GetConnectionLimit()))
		}
		roles = append(roles, roleArgs)
	}
	return roles
}

// workloadIdentityAnnotations translates the shared workload-identity oneof
// into the exact ServiceAccount annotations each cloud's webhook expects —
// the same mapping every catalog addon uses.
func workloadIdentityAnnotations(workloadIdentity *kubernetesprovider.KubernetesWorkloadIdentity) map[string]string {
	if workloadIdentity == nil {
		return nil
	}
	annotations := map[string]string{}
	if gke := workloadIdentity.GetGke(); gke != nil {
		annotations["iam.gke.io/gcp-service-account"] = gke.GetServiceAccountEmail().GetValue()
	}
	if eks := workloadIdentity.GetEks(); eks != nil {
		annotations["eks.amazonaws.com/role-arn"] = eks.GetRoleArn().GetValue()
	}
	if aks := workloadIdentity.GetAks(); aks != nil {
		annotations["azure.workload.identity/client-id"] = aks.GetClientId().GetValue()
		if aks.GetTenantId() != "" {
			annotations["azure.workload.identity/tenant-id"] = aks.GetTenantId()
		}
	}
	return annotations
}

// quantityMap renders a cpu/memory block with absent quantities OMITTED —
// CNPG's mutating webhook rejects an empty-string quantity (`quantities
// must match the regular expression ...`, verified live against a spec
// declaring only limits.memory). Returns nil when the block is absent or
// carries no set quantity, so the key never renders at all.
func quantityMap(cpuMemory *kubernetesprovider.CpuMemory) pulumi.Map {
	if cpuMemory == nil {
		return nil
	}
	quantities := pulumi.Map{}
	if cpuMemory.GetCpu() != "" {
		quantities["cpu"] = pulumi.String(cpuMemory.GetCpu())
	}
	if cpuMemory.GetMemory() != "" {
		quantities["memory"] = pulumi.String(cpuMemory.GetMemory())
	}
	if len(quantities) == 0 {
		return nil
	}
	return quantities
}

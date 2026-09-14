package module

import (
	"github.com/pkg/errors"
	azuredataprotectionbackupinstancev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredataprotectionbackupinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dataprotection"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the Data Protection backup instance. Exactly one
// variant block is set in the spec (validated at admission); each
// variant creates its own provider resource -- ONE resource exists
// per deployment.
//
// The vault's managed identity must already hold the datasource roles
// Azure Backup requires (disk: "Disk Backup Reader" on the disk +
// "Disk Snapshot Contributor" on the snapshot resource group; blob
// and Data Lake: "Storage Account Backup Contributor" on the storage
// account; Kubernetes: the AKS Backup extension + trusted access;
// MySQL/PostgreSQL: the vault identity's backup roles on the server).
// Azure validates the grants at create time -- an authorization-class
// failure here means missing role assignments, not a module defect.
//
// Nearly everything is ForceNew; only backup_policy_id updates in
// place (and on the kubernetes_cluster variant even that replaces the
// instance -- the provider ships no update path for it).
func Resources(ctx *pulumi.Context, stackInput *azuredataprotectionbackupinstancev1alpha1.AzureDataProtectionBackupInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataProtectionBackupInstance.Spec

	switch {
	case spec.BlobStorage != nil:
		return blobStorageInstance(ctx, locals, azureProvider)
	case spec.Disk != nil:
		return diskInstance(ctx, locals, azureProvider)
	case spec.KubernetesCluster != nil:
		return kubernetesClusterInstance(ctx, locals, azureProvider)
	case spec.MysqlFlexibleServer != nil:
		return mysqlFlexibleServerInstance(ctx, locals, azureProvider)
	case spec.PostgresqlFlexibleServer != nil:
		return postgresqlFlexibleServerInstance(ctx, locals, azureProvider)
	case spec.DataLakeStorage != nil:
		return dataLakeStorageInstance(ctx, locals, azureProvider)
	}

	// Unreachable behind the spec's exactly-one CEL; loud beats silent
	// for a direct module invocation that skipped admission.
	return errors.New("no instance variant set -- exactly one of blob_storage, disk, kubernetes_cluster, mysql_flexible_server, postgresql_flexible_server or data_lake_storage is required")
}

// The classic SDK's resource token misspells Blob as "Blog"
// (BackupInstanceBlogStorage) -- a bridge artifact over the correctly
// named azurerm_data_protection_backup_instance_blob_storage. Do NOT
// "fix" the name: it is the SDK's own identifier and creates the same
// ARM object the Terraform module does.
func blobStorageInstance(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupInstance.Spec
	variant := spec.BlobStorage

	args := &dataprotection.BackupInstanceBlogStorageArgs{
		Name:             pulumi.String(spec.Name),
		Location:         pulumi.String(spec.Region),
		VaultId:          pulumi.String(locals.VaultId),
		BackupPolicyId:   pulumi.String(locals.BackupPolicyId),
		StorageAccountId: pulumi.String(variant.StorageAccountId.GetValue()),
	}

	// Sent only when non-empty: the provider omits the containers list
	// from the ARM body when absent (operational-only protection
	// covers the whole account). ONE-WAY once set -- the provider
	// ForceNews on clearing the list, never on changing it.
	if len(variant.StorageAccountContainerNames) > 0 {
		args.StorageAccountContainerNames = pulumi.ToStringArray(variant.StorageAccountContainerNames)
	}

	createdInstance, err := dataprotection.NewBackupInstanceBlogStorage(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create blob storage backup instance %s", spec.Name)
	}

	exportOutputs(ctx, createdInstance.ID(), spec.Name)
	return nil
}

func diskInstance(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupInstance.Spec
	variant := spec.Disk

	args := &dataprotection.BackupInstanceDiskArgs{
		Name:           pulumi.String(spec.Name),
		Location:       pulumi.String(spec.Region),
		VaultId:        pulumi.String(locals.VaultId),
		BackupPolicyId: pulumi.String(locals.BackupPolicyId),
		DiskId:         pulumi.String(variant.DiskId.GetValue()),
		// Where the incremental snapshots land.
		SnapshotResourceGroupName: pulumi.String(variant.SnapshotResourceGroupName.GetValue()),
	}

	// Unset means the vault's own subscription -- the provider's
	// default.
	if variant.SnapshotSubscriptionId != nil && *variant.SnapshotSubscriptionId != "" {
		args.SnapshotSubscriptionId = pulumi.String(*variant.SnapshotSubscriptionId)
	}

	createdInstance, err := dataprotection.NewBackupInstanceDisk(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create disk backup instance %s", spec.Name)
	}

	exportOutputs(ctx, createdInstance.ID(), spec.Name)
	return nil
}

// The one variant with NO update path at all -- every field including
// backup_policy_id replaces the instance when changed.
func kubernetesClusterInstance(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupInstance.Spec
	variant := spec.KubernetesCluster

	args := &dataprotection.BackupInstanceKubernetesClusterArgs{
		Name:                      pulumi.String(spec.Name),
		Location:                  pulumi.String(spec.Region),
		VaultId:                   pulumi.String(locals.VaultId),
		BackupPolicyId:            pulumi.String(locals.BackupPolicyId),
		KubernetesClusterId:       pulumi.String(variant.KubernetesClusterId.GetValue()),
		SnapshotResourceGroupName: pulumi.String(variant.SnapshotResourceGroupName.GetValue()),
	}

	if params := variant.BackupDatasourceParameters; params != nil {
		paramsArgs := &dataprotection.BackupInstanceKubernetesClusterBackupDatasourceParametersArgs{
			ClusterScopedResourcesEnabled: pulumi.Bool(params.GetClusterScopedResourcesEnabled()),
			VolumeSnapshotEnabled:         pulumi.Bool(params.GetVolumeSnapshotEnabled()),
		}
		if len(params.IncludedNamespaces) > 0 {
			paramsArgs.IncludedNamespaces = pulumi.ToStringArray(params.IncludedNamespaces)
		}
		if len(params.ExcludedNamespaces) > 0 {
			paramsArgs.ExcludedNamespaces = pulumi.ToStringArray(params.ExcludedNamespaces)
		}
		if len(params.IncludedResourceTypes) > 0 {
			paramsArgs.IncludedResourceTypes = pulumi.ToStringArray(params.IncludedResourceTypes)
		}
		if len(params.ExcludedResourceTypes) > 0 {
			paramsArgs.ExcludedResourceTypes = pulumi.ToStringArray(params.ExcludedResourceTypes)
		}
		if len(params.LabelSelectors) > 0 {
			paramsArgs.LabelSelectors = pulumi.ToStringArray(params.LabelSelectors)
		}
		args.BackupDatasourceParameters = paramsArgs
	}

	createdInstance, err := dataprotection.NewBackupInstanceKubernetesCluster(ctx,
		spec.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create kubernetes cluster backup instance %s", spec.Name)
	}

	exportOutputs(ctx, createdInstance.ID(), spec.Name)
	return nil
}

func mysqlFlexibleServerInstance(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupInstance.Spec
	variant := spec.MysqlFlexibleServer

	createdInstance, err := dataprotection.NewBackupInstanceMysqlFlexibleServer(ctx,
		spec.Name,
		&dataprotection.BackupInstanceMysqlFlexibleServerArgs{
			Name:           pulumi.String(spec.Name),
			Location:       pulumi.String(spec.Region),
			VaultId:        pulumi.String(locals.VaultId),
			BackupPolicyId: pulumi.String(locals.BackupPolicyId),
			ServerId:       pulumi.String(variant.ServerId.GetValue()),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mysql flexible server backup instance %s", spec.Name)
	}

	exportOutputs(ctx, createdInstance.ID(), spec.Name)
	return nil
}

func postgresqlFlexibleServerInstance(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupInstance.Spec
	variant := spec.PostgresqlFlexibleServer

	createdInstance, err := dataprotection.NewBackupInstancePostgresqlFlexibleServer(ctx,
		spec.Name,
		&dataprotection.BackupInstancePostgresqlFlexibleServerArgs{
			Name:           pulumi.String(spec.Name),
			Location:       pulumi.String(spec.Region),
			VaultId:        pulumi.String(locals.VaultId),
			BackupPolicyId: pulumi.String(locals.BackupPolicyId),
			ServerId:       pulumi.String(variant.ServerId.GetValue()),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create postgresql flexible server backup instance %s", spec.Name)
	}

	exportOutputs(ctx, createdInstance.ID(), spec.Name)
	return nil
}

// The one variant that names its vault and policy arguments
// differently (DataProtectionBackupVaultId /
// BackupPolicyDataLakeStorageId) -- same values, provider-side
// renames recorded in the parity manifest.
func dataLakeStorageInstance(ctx *pulumi.Context, locals *Locals, azureProvider pulumi.ProviderResource) error {
	spec := locals.AzureDataProtectionBackupInstance.Spec
	variant := spec.DataLakeStorage

	createdInstance, err := dataprotection.NewBackupInstanceDataLakeStorage(ctx,
		spec.Name,
		&dataprotection.BackupInstanceDataLakeStorageArgs{
			Name:                          pulumi.String(spec.Name),
			Location:                      pulumi.String(spec.Region),
			DataProtectionBackupVaultId:   pulumi.String(locals.VaultId),
			BackupPolicyDataLakeStorageId: pulumi.String(locals.BackupPolicyId),
			StorageAccountId:              pulumi.String(variant.StorageAccountId.GetValue()),
			StorageContainerNames:         pulumi.ToStringArray(variant.StorageContainerNames),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data lake storage backup instance %s", spec.Name)
	}

	exportOutputs(ctx, createdInstance.ID(), spec.Name)
	return nil
}

// exportOutputs keeps the six variant branches' output shapes
// identical -- one instance ID, one name, whichever variant ran.
func exportOutputs(ctx *pulumi.Context, instanceId pulumi.IDOutput, instanceName string) {
	ctx.Export(OpBackupInstanceId, instanceId)
	ctx.Export(OpBackupInstanceName, pulumi.String(instanceName))
}

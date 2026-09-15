package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// cluster provisions the managed database cluster and exports its outputs.
func cluster(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.DatabaseCluster, error) {
	spec := locals.DigitalOceanDatabaseCluster.Spec

	if spec.Engine == 0 {
		return nil, errors.Errorf("database engine is required")
	}

	if spec.StorageAutoscale != nil {
		return nil, errors.New("PARITY-EXCEPTION: spec.storage_autoscale is modeled and Terraform wires it; the Pulumi DigitalOcean SDK v4.49.0 has no storage_autoscale field on DatabaseCluster. Re-evaluate when the SDK exposes storage_autoscale.")
	}

	// User tags plus the standard Planton labels rendered as "key:value"
	// tags — the exact set the Terraform module applies.
	tagSet := map[string]bool{}
	var tagInputs pulumi.StringArray
	for _, t := range spec.Tags {
		if !tagSet[t] {
			tagSet[t] = true
			tagInputs = append(tagInputs, pulumi.String(t))
		}
	}
	for k, v := range locals.DigitalOceanLabels {
		t := k + ":" + v
		if !tagSet[t] {
			tagSet[t] = true
			tagInputs = append(tagInputs, pulumi.String(t))
		}
	}

	// Enum value names are exactly the DigitalOcean API slugs.
	clusterArgs := &digitalocean.DatabaseClusterArgs{
		Engine:    pulumi.String(spec.Engine.String()),
		Name:      pulumi.String(spec.ClusterName),
		Region:    pulumi.String(spec.Region.String()),
		Version:   pulumi.String(spec.EngineVersion),
		Size:      pulumi.String(spec.SizeSlug),
		NodeCount: pulumi.Int(int(spec.NodeCount)),
		Tags:      tagInputs,
	}

	// The provider's storage_size_mib is a string holding a bare MiB count;
	// the spec carries GiB for ergonomics.
	if spec.StorageGib != 0 {
		clusterArgs.StorageSizeMib = pulumi.String(fmt.Sprintf("%d", uint64(spec.StorageGib)*1024))
	}

	// Optional VPC attachment (create-only).
	if spec.Vpc != nil && spec.Vpc.GetValue() != "" {
		clusterArgs.PrivateNetworkUuid = pulumi.StringPtr(spec.Vpc.GetValue())
	}

	// Optional DigitalOcean project placement (create-only).
	if projectId := spec.GetProjectId().GetValue(); projectId != "" {
		clusterArgs.ProjectId = pulumi.StringPtr(projectId)
	}

	// Weekly maintenance window. The SDK models a list; a cluster has
	// exactly one window, so the spec carries a single message.
	if spec.MaintenanceWindow != nil {
		clusterArgs.MaintenanceWindows = digitalocean.DatabaseClusterMaintenanceWindowArray{
			digitalocean.DatabaseClusterMaintenanceWindowArgs{
				Day:  pulumi.String(spec.MaintenanceWindow.Day),
				Hour: pulumi.String(spec.MaintenanceWindow.Hour),
			},
		}
	}

	// Provision-from-backup. Consumed only at creation; never read back.
	if spec.BackupRestore != nil {
		backupRestoreArgs := digitalocean.DatabaseClusterBackupRestoreArgs{
			DatabaseName: pulumi.String(spec.BackupRestore.DatabaseName),
		}
		if spec.BackupRestore.BackupCreatedAt != "" {
			backupRestoreArgs.BackupCreatedAt = pulumi.StringPtr(spec.BackupRestore.BackupCreatedAt)
		}
		clusterArgs.BackupRestore = backupRestoreArgs
	}

	// Engine-conditional tuning: spec CEL rules enforce the engine pairing,
	// so these are simply passed through when set.
	if spec.EvictionPolicy != "" {
		clusterArgs.EvictionPolicy = pulumi.StringPtr(spec.EvictionPolicy)
	}
	if spec.SqlMode != "" {
		clusterArgs.SqlMode = pulumi.StringPtr(spec.SqlMode)
	}

	createdCluster, err := digitalocean.NewDatabaseCluster(
		ctx,
		"cluster",
		clusterArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean database cluster")
	}

	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpConnectionUri, createdCluster.Uri)
	ctx.Export(OpHost, createdCluster.Host)
	ctx.Export(OpPort, createdCluster.Port)
	ctx.Export(OpDatabaseUser, createdCluster.User)
	ctx.Export(OpDatabasePassword, createdCluster.Password)
	ctx.Export(OpPrivateHost, createdCluster.PrivateHost)
	ctx.Export(OpPrivateUri, createdCluster.PrivateUri)
	ctx.Export(OpDatabaseName, createdCluster.Database)
	ctx.Export(OpUiHost, createdCluster.UiHost)
	ctx.Export(OpUiPort, createdCluster.UiPort)
	ctx.Export(OpUiUri, createdCluster.UiUri)
	ctx.Export(OpUiDatabase, createdCluster.UiDatabase)
	ctx.Export(OpUiUser, createdCluster.UiUser)
	ctx.Export(OpUiPassword, createdCluster.UiPassword)

	return createdCluster, nil
}

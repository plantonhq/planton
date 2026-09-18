package module

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tagsCombinedBudget is DigitalOcean's cap on a database cluster's COMBINED
// tags -- the tag names joined by commas -- measured 2026-09-17 against
// `POST /v2/databases`: 255 characters pass, 256 fail with
// `422 combined tags cannot exceed 255 characters`; colons count as one
// character and the tag count matters only through the separators. The six
// Planton label tags carry metadata.name and metadata.id, so a long
// resource name spends the budget before any spec.tags entry does. Checked
// before anything renders, with the same number and the same message as
// the Terraform module's precondition (its twin).
const tagsCombinedBudget = 255

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

	// User tags plus the standard Planton labels rendered as "key:value"
	// tags — the exact set the Terraform module applies. Labels are added
	// in key order so the rendered list is deterministic on every apply.
	tagSet := map[string]bool{}
	var tags []string
	for _, t := range spec.Tags {
		if !tagSet[t] {
			tagSet[t] = true
			tags = append(tags, t)
		}
	}
	labelKeys := make([]string, 0, len(locals.DigitalOceanLabels))
	for k := range locals.DigitalOceanLabels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	var labelTags []string
	for _, k := range labelKeys {
		t := k + ":" + locals.DigitalOceanLabels[k]
		labelTags = append(labelTags, t)
		if !tagSet[t] {
			tagSet[t] = true
			tags = append(tags, t)
		}
	}

	// Fail loud on DigitalOcean's combined-tags budget before anything
	// renders (see tagsCombinedBudget; twin of the Terraform precondition).
	if combined := strings.Join(tags, ","); len(combined) > tagsCombinedBudget {
		return nil, errors.Errorf(
			"DigitalOcean caps a database cluster's combined tags (joined by commas) at %d characters; this cluster's %d tags join to %d characters. Shorten metadata.name or metadata.id, or remove entries from spec.tags -- the Planton label tags alone use %d characters here.",
			tagsCombinedBudget, len(tags), len(combined), len(strings.Join(labelTags, ",")))
	}

	tagInputs := make(pulumi.StringArray, 0, len(tags))
	for _, t := range tags {
		tagInputs = append(tagInputs, pulumi.String(t))
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
	if spec.ProjectId != "" {
		clusterArgs.ProjectId = pulumi.StringPtr(spec.ProjectId)
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

	// Automatic storage growth. Zero threshold/increment mean "DigitalOcean's
	// default" and are left null, never sent as 0 -- the Terraform module's
	// coalescing.
	if spec.StorageAutoscale != nil {
		autoscaleArgs := &digitalocean.DatabaseClusterStorageAutoscaleArgs{
			Enabled: pulumi.Bool(spec.StorageAutoscale.Enabled),
		}
		if spec.StorageAutoscale.ThresholdPercent > 0 {
			autoscaleArgs.ThresholdPercent = pulumi.IntPtr(int(spec.StorageAutoscale.ThresholdPercent))
		}
		if spec.StorageAutoscale.IncrementGib > 0 {
			autoscaleArgs.IncrementGib = pulumi.IntPtr(int(spec.StorageAutoscale.IncrementGib))
		}
		clusterArgs.StorageAutoscale = autoscaleArgs
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

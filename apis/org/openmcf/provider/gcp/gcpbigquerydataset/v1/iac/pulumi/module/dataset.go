package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/bigquery"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func dataset(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBigQueryDataset.Spec

	args := &bigquery.DatasetArgs{
		DatasetId: pulumi.String(spec.DatasetId),
		Project:   pulumi.StringPtr(spec.ProjectId.GetValue()),
		Location:  pulumi.StringPtr(spec.Location),
		Labels:    pulumi.ToStringMap(locals.GcpLabels),
	}

	// Set optional string fields.
	if spec.FriendlyName != "" {
		args.FriendlyName = pulumi.StringPtr(spec.FriendlyName)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.DefaultCollation != "" {
		args.DefaultCollation = pulumi.StringPtr(spec.DefaultCollation)
	}
	if spec.StorageBillingModel != "" {
		args.StorageBillingModel = pulumi.StringPtr(spec.StorageBillingModel)
	}

	// Set optional integer fields.
	if spec.DefaultTableExpirationMs > 0 {
		args.DefaultTableExpirationMs = pulumi.IntPtr(int(spec.DefaultTableExpirationMs))
	}
	if spec.DefaultPartitionExpirationMs > 0 {
		args.DefaultPartitionExpirationMs = pulumi.IntPtr(int(spec.DefaultPartitionExpirationMs))
	}
	if spec.MaxTimeTravelHours > 0 {
		args.MaxTimeTravelHours = pulumi.StringPtr(strconv.Itoa(int(spec.MaxTimeTravelHours)))
	}

	// Set optional boolean fields.
	if spec.IsCaseInsensitive {
		args.IsCaseInsensitive = pulumi.BoolPtr(true)
	}
	if spec.DeleteContentsOnDestroy {
		args.DeleteContentsOnDestroy = pulumi.BoolPtr(true)
	}

	// Set CMEK encryption configuration.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.DefaultEncryptionConfiguration = &bigquery.DatasetDefaultEncryptionConfigurationArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Map access control entries.
	if len(spec.Access) > 0 {
		accessArray := bigquery.DatasetAccessTypeArray{}
		for _, entry := range spec.Access {
			accessArgs := &bigquery.DatasetAccessTypeArgs{}

			if entry.Role != "" {
				accessArgs.Role = pulumi.StringPtr(entry.Role)
			}
			if entry.UserByEmail != "" {
				accessArgs.UserByEmail = pulumi.StringPtr(entry.UserByEmail)
			}
			if entry.GroupByEmail != "" {
				accessArgs.GroupByEmail = pulumi.StringPtr(entry.GroupByEmail)
			}
			if entry.Domain != "" {
				accessArgs.Domain = pulumi.StringPtr(entry.Domain)
			}
			if entry.SpecialGroup != "" {
				accessArgs.SpecialGroup = pulumi.StringPtr(entry.SpecialGroup)
			}
			if entry.IamMember != "" {
				accessArgs.IamMember = pulumi.StringPtr(entry.IamMember)
			}
			if entry.View != nil {
				accessArgs.View = &bigquery.DatasetAccessViewArgs{
					ProjectId: pulumi.String(entry.View.ProjectId),
					DatasetId: pulumi.String(entry.View.DatasetId),
					TableId:   pulumi.String(entry.View.TableId),
				}
			}

			accessArray = append(accessArray, accessArgs)
		}
		args.Accesses = accessArray
	}

	createdDataset, err := bigquery.NewDataset(ctx, "bigquery-dataset", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create bigquery dataset")
	}

	ctx.Export(OpDatasetId, createdDataset.DatasetId)
	ctx.Export(OpSelfLink, createdDataset.SelfLink)
	ctx.Export(OpProject, createdDataset.Project)
	ctx.Export(OpCreationTime, createdDataset.CreationTime)

	return nil
}

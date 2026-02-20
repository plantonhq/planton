package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudceninstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudceninstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cen"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudceninstancev1.AlicloudCenInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudCenInstance.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	cenArgs := &cen.InstanceArgs{
		CenInstanceName: pulumi.StringPtr(spec.CenInstanceName),
		Tags:            pulumi.ToStringMap(locals.Tags),
	}

	if spec.Description != "" {
		cenArgs.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.ProtectionLevel != nil && *spec.ProtectionLevel != "" {
		cenArgs.ProtectionLevel = pulumi.StringPtr(*spec.ProtectionLevel)
	}

	if spec.ResourceGroupId != "" {
		cenArgs.ResourceGroupId = pulumi.StringPtr(spec.ResourceGroupId)
	}

	cenInstance, err := cen.NewInstance(ctx, spec.CenInstanceName, cenArgs,
		pulumi.Provider(alicloudProvider),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create CEN instance %s", spec.CenInstanceName)
	}

	for i, attachment := range spec.Attachments {
		if err := cenAttachment(ctx, alicloudProvider, cenInstance, attachment, i); err != nil {
			return err
		}
	}

	ctx.Export(OpCenId, cenInstance.ID())
	ctx.Export(OpCenInstanceName, cenInstance.CenInstanceName)

	return nil
}

func cenAttachment(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	cenInstance *cen.Instance,
	attachment *alicloudceninstancev1.AlicloudCenAttachment,
	index int,
) error {
	childType := "VPC"
	if attachment.ChildInstanceType != nil && *attachment.ChildInstanceType != "" {
		childType = *attachment.ChildInstanceType
	}

	resourceName := fmt.Sprintf("attachment-%d-%s", index, childType)

	_, err := cen.NewInstanceAttachment(ctx, resourceName, &cen.InstanceAttachmentArgs{
		InstanceId:            cenInstance.ID(),
		ChildInstanceId:       pulumi.String(attachment.ChildInstanceId.GetValue()),
		ChildInstanceType:     pulumi.String(childType),
		ChildInstanceRegionId: pulumi.String(attachment.ChildInstanceRegionId),
	},
		pulumi.Provider(provider),
		pulumi.Parent(cenInstance),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create CEN attachment %d (%s)", index, childType)
	}

	return nil
}

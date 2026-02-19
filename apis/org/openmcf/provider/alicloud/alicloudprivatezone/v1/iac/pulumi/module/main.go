package module

import (
	"github.com/pkg/errors"
	alicloudprivatezonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudprivatezone/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/pvtz"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudprivatezonev1.AlicloudPrivateZoneStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudPrivateZone.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	zone, err := pvtz.NewZone(ctx, spec.ZoneName, &pvtz.ZoneArgs{
		ZoneName:        pulumi.String(spec.ZoneName),
		Remark:          optionalString(spec.Remark),
		ResourceGroupId: optionalString(spec.ResourceGroupId),
		Tags:            pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create private zone %s", spec.ZoneName)
	}

	vpcs := pvtz.ZoneAttachmentVpcArray{}
	for _, att := range spec.VpcAttachments {
		vpcEntry := pvtz.ZoneAttachmentVpcArgs{
			VpcId: pulumi.String(att.VpcId.GetValue()),
		}
		if att.RegionId != "" {
			vpcEntry.RegionId = pulumi.String(att.RegionId)
		}
		vpcs = append(vpcs, vpcEntry)
	}

	_, err = pvtz.NewZoneAttachment(ctx, spec.ZoneName+"-attachment", &pvtz.ZoneAttachmentArgs{
		ZoneId: zone.ID(),
		Vpcs:   vpcs,
	},
		pulumi.Provider(alicloudProvider),
		pulumi.Parent(zone),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to attach VPCs to private zone %s", spec.ZoneName)
	}

	for i, record := range spec.Records {
		if err := zoneRecord(ctx, alicloudProvider, zone, spec.ZoneName, i, record); err != nil {
			return err
		}
	}

	ctx.Export(OpZoneId, zone.ID())
	ctx.Export(OpZoneName, zone.ZoneName)
	ctx.Export(OpIsPtr, zone.IsPtr)
	ctx.Export(OpRecordCount, zone.RecordCount)

	return nil
}

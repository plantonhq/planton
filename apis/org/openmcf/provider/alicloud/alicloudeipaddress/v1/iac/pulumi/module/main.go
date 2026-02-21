package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudeipaddressv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudeipaddress/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudeipaddressv1.AliCloudEipAddressStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudEipAddress.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	resourceName := spec.AddressName
	if resourceName == "" {
		resourceName = stackInput.Target.Metadata.Name
	}

	eip, err := ecs.NewEipAddress(ctx, resourceName, &ecs.EipAddressArgs{
		AddressName:        optionalString(spec.AddressName),
		Description:        optionalString(spec.Description),
		Bandwidth:          pulumi.String(fmt.Sprintf("%d", bandwidth(spec))),
		InternetChargeType: pulumi.String(internetChargeType(spec)),
		Isp:                pulumi.String(isp(spec)),
		ResourceGroupId:    optionalString(spec.ResourceGroupId),
		Tags:               pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create EIP %s", resourceName)
	}

	ctx.Export(OpEipId, eip.ID())
	ctx.Export(OpIpAddress, eip.IpAddress)

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

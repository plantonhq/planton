package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudnatgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudnatgateway/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func snatEntry(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	natGateway *vpc.NatGateway,
	natGatewayName string,
	snatIp string,
	index int,
	entry *alicloudnatgatewayv1.AliCloudSnatEntry,
) error {
	entryName := entry.SnatEntryName
	if entryName == "" {
		entryName = fmt.Sprintf("%s-snat-%d", natGatewayName, index)
	}

	args := &vpc.SnatEntryArgs{
		SnatTableId:   natGateway.SnatTableIds,
		SnatIp:        pulumi.String(snatIp),
		SnatEntryName: pulumi.String(entryName),
	}

	if entry.SourceVswitchId != nil && entry.SourceVswitchId.GetValue() != "" {
		args.SourceVswitchId = pulumi.String(entry.SourceVswitchId.GetValue())
	}

	if entry.SourceCidr != "" {
		args.SourceCidr = pulumi.String(entry.SourceCidr)
	}

	_, err := vpc.NewSnatEntry(ctx, entryName, args,
		pulumi.Provider(provider),
		pulumi.Parent(natGateway),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create SNAT entry %s", entryName)
	}

	return nil
}

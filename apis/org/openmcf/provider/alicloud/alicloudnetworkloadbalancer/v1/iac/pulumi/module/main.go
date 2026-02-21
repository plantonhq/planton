package module

import (
	"github.com/pkg/errors"
	alicloudnetworkloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudnetworkloadbalancer/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/nlb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudnetworkloadbalancerv1.AliCloudNetworkLoadBalancerStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudNetworkLoadBalancer.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	zoneMappings := nlb.LoadBalancerZoneMappingArray{}
	for _, zm := range spec.ZoneMappings {
		zmArgs := nlb.LoadBalancerZoneMappingArgs{
			ZoneId:    pulumi.String(zm.ZoneId),
			VswitchId: pulumi.String(zm.VswitchId.GetValue()),
		}
		if zm.AllocationId != nil {
			zmArgs.AllocationId = pulumi.String(zm.AllocationId.GetValue())
		}
		zoneMappings = append(zoneMappings, zmArgs)
	}

	lbName := spec.LoadBalancerName
	if lbName == "" {
		lbName = locals.AliCloudNetworkLoadBalancer.Metadata.Name
	}

	lbArgs := &nlb.LoadBalancerArgs{
		LoadBalancerName: pulumi.String(lbName),
		VpcId:            pulumi.String(spec.VpcId.GetValue()),
		AddressType:      pulumi.String(addressType(spec)),
		LoadBalancerType: pulumi.String("Network"),
		PaymentType:      pulumi.String("PayAsYouGo"),
		CrossZoneEnabled: pulumi.Bool(crossZoneEnabled(spec)),
		ZoneMappings:     zoneMappings,
		Tags:             pulumi.ToStringMap(locals.Tags),
	}

	if spec.ResourceGroupId != "" {
		lbArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	lb, err := nlb.NewLoadBalancer(ctx, lbName, lbArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create NLB %s", lbName)
	}

	serverGroupIdByName := make(map[string]pulumi.IDOutput)
	serverGroupIdsMap := pulumi.StringMap{}

	for _, sg := range spec.ServerGroups {
		created, err := serverGroup(ctx, alicloudProvider, spec.VpcId.GetValue(), sg)
		if err != nil {
			return err
		}
		serverGroupIdByName[sg.Name] = created.ID()
		serverGroupIdsMap[sg.Name] = created.ID()
	}

	for _, l := range spec.Listeners {
		if err := listener(ctx, alicloudProvider, lb, serverGroupIdByName, l); err != nil {
			return err
		}
	}

	ctx.Export(OpLoadBalancerId, lb.ID())
	ctx.Export(OpDnsName, lb.DnsName)
	ctx.Export(OpServerGroupIds, serverGroupIdsMap)

	return nil
}

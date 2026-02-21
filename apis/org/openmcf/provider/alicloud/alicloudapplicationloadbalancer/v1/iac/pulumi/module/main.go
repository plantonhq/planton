package module

import (
	"github.com/pkg/errors"
	alicloudapplicationloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudapplicationloadbalancer/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/alb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudapplicationloadbalancerv1.AliCloudApplicationLoadBalancerStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudApplicationLoadBalancer.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	zoneMappings := alb.LoadBalancerZoneMappingArray{}
	for _, zm := range spec.ZoneMappings {
		zoneMappings = append(zoneMappings, alb.LoadBalancerZoneMappingArgs{
			ZoneId:    pulumi.String(zm.ZoneId),
			VswitchId: pulumi.String(zm.VswitchId.GetValue()),
		})
	}

	lbName := spec.LoadBalancerName
	if lbName == "" {
		lbName = locals.AliCloudApplicationLoadBalancer.Metadata.Name
	}

	lbArgs := &alb.LoadBalancerArgs{
		LoadBalancerName:    pulumi.String(lbName),
		VpcId:               pulumi.String(spec.VpcId.GetValue()),
		AddressType:         pulumi.String(addressType(spec)),
		LoadBalancerEdition: pulumi.String(loadBalancerEdition(spec)),
		LoadBalancerBillingConfig: alb.LoadBalancerLoadBalancerBillingConfigArgs{
			PayType: pulumi.String("PayAsYouGo"),
		},
		ZoneMappings: zoneMappings,
		Tags:         pulumi.ToStringMap(locals.Tags),
	}

	if spec.ResourceGroupId != "" {
		lbArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	if spec.AccessLogConfig != nil {
		lbArgs.AccessLogConfig = alb.LoadBalancerAccessLogConfigArgs{
			LogProject: pulumi.String(spec.AccessLogConfig.LogProject),
			LogStore:   pulumi.String(spec.AccessLogConfig.LogStore),
		}
	}

	lb, err := alb.NewLoadBalancer(ctx, lbName, lbArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create ALB %s", lbName)
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

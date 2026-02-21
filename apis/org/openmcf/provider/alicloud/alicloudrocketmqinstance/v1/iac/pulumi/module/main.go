package module

import (
	"github.com/pkg/errors"
	alicloudrocketmqinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrocketmqinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/rocketmq"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudrocketmqinstancev1.AliCloudRocketmqInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudRocketmqInstance.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	name := instanceName(locals)

	vpcInfo := rocketmq.RocketMQInstanceNetworkInfoVpcInfoArgs{
		VpcId: pulumi.String(spec.VpcId.GetValue()),
	}

	if spec.VswitchId != nil && spec.VswitchId.GetValue() != "" {
		vpcInfo.Vswitches = rocketmq.RocketMQInstanceNetworkInfoVpcInfoVswitchArray{
			rocketmq.RocketMQInstanceNetworkInfoVpcInfoVswitchArgs{
				VswitchId: pulumi.StringPtr(spec.VswitchId.GetValue()),
			},
		}
	}

	if spec.SecurityGroupId != "" {
		vpcInfo.SecurityGroupIds = pulumi.String(spec.SecurityGroupId)
	}

	internetInfo := rocketmq.RocketMQInstanceNetworkInfoInternetInfoArgs{
		InternetSpec: pulumi.String(internetSpec(spec)),
		FlowOutType:  pulumi.String(flowOutType(spec)),
	}

	if spec.InternetInfo != nil && spec.InternetInfo.FlowOutBandwidth != nil {
		internetInfo.FlowOutBandwidth = pulumi.IntPtr(int(*spec.InternetInfo.FlowOutBandwidth))
	}

	instanceArgs := &rocketmq.RocketMQInstanceArgs{
		InstanceName:  pulumi.StringPtr(name),
		SeriesCode:    pulumi.String(spec.SeriesCode),
		SubSeriesCode: pulumi.String(spec.SubSeriesCode),
		ServiceCode:   pulumi.String("rmq"),
		PaymentType:   pulumi.String(paymentType(spec)),
		CommodityCode: pulumi.StringPtr(commodityCode(spec)),
		NetworkInfo: rocketmq.RocketMQInstanceNetworkInfoArgs{
			VpcInfo:      vpcInfo,
			InternetInfo: internetInfo,
		},
		Tags: pulumi.ToStringMap(locals.Tags),
	}

	if spec.Remark != "" {
		instanceArgs.Remark = pulumi.StringPtr(spec.Remark)
	}

	if len(spec.IpWhitelists) > 0 {
		instanceArgs.IpWhitelists = pulumi.ToStringArray(spec.IpWhitelists)
	}

	if spec.ResourceGroupId != "" {
		instanceArgs.ResourceGroupId = pulumi.StringPtr(spec.ResourceGroupId)
	}

	instanceArgs.Period = optionalInt(spec.Period)
	instanceArgs.AutoRenew = optionalBool(spec.AutoRenew)
	instanceArgs.AutoRenewPeriod = optionalInt(spec.AutoRenewPeriod)

	if spec.PeriodUnit != nil && *spec.PeriodUnit != "" {
		instanceArgs.PeriodUnit = pulumi.StringPtr(*spec.PeriodUnit)
	}

	if spec.MsgProcessSpec != "" || spec.ProductInfo != nil {
		productInfo := rocketmq.RocketMQInstanceProductInfoArgs{
			MsgProcessSpec: pulumi.String(spec.MsgProcessSpec),
		}

		if spec.ProductInfo != nil {
			productInfo.MessageRetentionTime = optionalInt(spec.ProductInfo.MessageRetentionTime)
			productInfo.AutoScaling = optionalBool(spec.ProductInfo.AutoScaling)
			productInfo.TraceOn = optionalBool(spec.ProductInfo.TraceOn)
			productInfo.StorageEncryption = optionalBool(spec.ProductInfo.StorageEncryption)

			if spec.ProductInfo.StorageSecretKey != "" {
				productInfo.StorageSecretKey = pulumi.StringPtr(spec.ProductInfo.StorageSecretKey)
			}

			if spec.ProductInfo.SendReceiveRatio != nil {
				productInfo.SendReceiveRatio = pulumi.Float64Ptr(*spec.ProductInfo.SendReceiveRatio)
			}
		}

		instanceArgs.ProductInfo = productInfo
	}

	instance, err := rocketmq.NewRocketMQInstance(ctx, name, instanceArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create RocketMQ instance %s", name)
	}

	topicIdMap := pulumi.StringMap{}

	for _, t := range spec.Topics {
		created, err := topic(ctx, alicloudProvider, instance, t)
		if err != nil {
			return err
		}
		topicIdMap[t.TopicName] = created.ID()
	}

	consumerGroupIdMap := pulumi.StringMap{}

	for _, cg := range spec.ConsumerGroups {
		created, err := consumerGroup(ctx, alicloudProvider, instance, cg)
		if err != nil {
			return err
		}
		consumerGroupIdMap[cg.ConsumerGroupId] = created.ID()
	}

	ctx.Export(OpInstanceId, instance.ID())
	ctx.Export(OpTcpEndpoint, extractEndpointUrl(instance, "TCP_VPC"))
	ctx.Export(OpInternetEndpoint, extractEndpointUrl(instance, "TCP_INTERNET"))
	ctx.Export(OpTopicIds, topicIdMap)
	ctx.Export(OpConsumerGroupIds, consumerGroupIdMap)

	return nil
}

// extractEndpointUrl searches the instance's computed endpoints for a matching
// type and returns its URL, or an empty string if not found.
func extractEndpointUrl(instance *rocketmq.RocketMQInstance, endpointType string) pulumi.StringOutput {
	return instance.NetworkInfo.Endpoints().ApplyT(func(endpoints []rocketmq.RocketMQInstanceNetworkInfoEndpoint) string {
		for _, ep := range endpoints {
			if ep.EndpointType != nil && *ep.EndpointType == endpointType {
				if ep.EndpointUrl != nil {
					return *ep.EndpointUrl
				}
			}
		}
		return ""
	}).(pulumi.StringOutput)
}

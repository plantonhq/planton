package module

import (
	"github.com/pkg/errors"
	alicloudrocketmqinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrocketmqinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/rocketmq"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func consumerGroup(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	instance *rocketmq.RocketMQInstance,
	cg *alicloudrocketmqinstancev1.AliCloudRocketmqConsumerGroup,
) (*rocketmq.ConsumerGroup, error) {
	retryPolicyArgs := rocketmq.ConsumerGroupConsumeRetryPolicyArgs{
		RetryPolicy: pulumi.StringPtr(retryPolicy(cg)),
	}

	if cg.ConsumeRetryPolicy != nil {
		retryPolicyArgs.MaxRetryTimes = optionalInt(cg.ConsumeRetryPolicy.MaxRetryTimes)

		if cg.ConsumeRetryPolicy.DeadLetterTargetTopic != "" {
			retryPolicyArgs.DeadLetterTargetTopic = pulumi.StringPtr(cg.ConsumeRetryPolicy.DeadLetterTargetTopic)
		}
	}

	args := &rocketmq.ConsumerGroupArgs{
		InstanceId:         instance.ID(),
		ConsumerGroupId:    pulumi.String(cg.ConsumerGroupId),
		ConsumeRetryPolicy: retryPolicyArgs,
	}

	if cg.DeliveryOrderType != nil && *cg.DeliveryOrderType != "" {
		args.DeliveryOrderType = pulumi.StringPtr(*cg.DeliveryOrderType)
	}

	if cg.Remark != "" {
		args.Remark = pulumi.StringPtr(cg.Remark)
	}

	args.MaxReceiveTps = optionalInt(cg.MaxReceiveTps)

	created, err := rocketmq.NewConsumerGroup(ctx, cg.ConsumerGroupId, args,
		pulumi.Provider(provider),
		pulumi.Parent(instance),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create consumer group %s", cg.ConsumerGroupId)
	}

	return created, nil
}

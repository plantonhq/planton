package module

import (
	"strings"

	alicloudrocketmqinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrocketmqinstance/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudRocketmqInstance *alicloudrocketmqinstancev1.AlicloudRocketmqInstance
	Tags                     map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudrocketmqinstancev1.AlicloudRocketmqInstanceStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudRocketmqInstance = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudRocketmqInstance.String()),
	}

	if target.Metadata.Id != "" {
		locals.Tags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Tags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Tags["environment"] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.Tags[k] = v
	}

	return locals
}

func instanceName(locals *Locals) string {
	if locals.AlicloudRocketmqInstance.Spec.InstanceName != "" {
		return locals.AlicloudRocketmqInstance.Spec.InstanceName
	}
	return locals.AlicloudRocketmqInstance.Metadata.Name
}

func paymentType(spec *alicloudrocketmqinstancev1.AlicloudRocketmqInstanceSpec) string {
	if spec.PaymentType != nil && *spec.PaymentType != "" {
		return *spec.PaymentType
	}
	return "PayAsYouGo"
}

// commodityCode derives the billing commodity code from payment_type and
// sub_series_code. This is an implementation detail hidden from users.
func commodityCode(spec *alicloudrocketmqinstancev1.AlicloudRocketmqInstanceSpec) string {
	if spec.SubSeriesCode == "serverless" {
		return "ons_rmqsrvlesspost_public_cn"
	}
	if paymentType(spec) == "Subscription" {
		return "ons_rmqsub_public_cn"
	}
	return "ons_rmqpost_public_cn"
}

func internetSpec(spec *alicloudrocketmqinstancev1.AlicloudRocketmqInstanceSpec) string {
	if spec.InternetInfo != nil && spec.InternetInfo.Enabled != nil && *spec.InternetInfo.Enabled {
		return "enable"
	}
	return "disable"
}

func flowOutType(spec *alicloudrocketmqinstancev1.AlicloudRocketmqInstanceSpec) string {
	if spec.InternetInfo == nil || spec.InternetInfo.Enabled == nil || !*spec.InternetInfo.Enabled {
		return "uninvolved"
	}
	if spec.InternetInfo.FlowOutType != nil && *spec.InternetInfo.FlowOutType != "" {
		return *spec.InternetInfo.FlowOutType
	}
	return "payByTraffic"
}

func messageType(t *alicloudrocketmqinstancev1.AlicloudRocketmqTopic) string {
	if t.MessageType != nil && *t.MessageType != "" {
		return *t.MessageType
	}
	return "NORMAL"
}

func retryPolicy(cg *alicloudrocketmqinstancev1.AlicloudRocketmqConsumerGroup) string {
	if cg.ConsumeRetryPolicy != nil && cg.ConsumeRetryPolicy.RetryPolicy != nil && *cg.ConsumeRetryPolicy.RetryPolicy != "" {
		return *cg.ConsumeRetryPolicy.RetryPolicy
	}
	return "DefaultRetryPolicy"
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

func optionalBool(b *bool) pulumi.BoolPtrInput {
	if b == nil {
		return nil
	}
	return pulumi.Bool(*b)
}

func optionalInt(i *int32) pulumi.IntPtrInput {
	if i == nil {
		return nil
	}
	return pulumi.Int(int(*i))
}

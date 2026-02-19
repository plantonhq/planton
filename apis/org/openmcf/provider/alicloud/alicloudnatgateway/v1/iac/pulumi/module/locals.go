package module

import (
	"strings"

	alicloudnatgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudnatgateway/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudNatGateway *alicloudnatgatewayv1.AlicloudNatGateway
	Tags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudnatgatewayv1.AlicloudNatGatewayStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudNatGateway = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudNatGateway.String()),
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

func natType(spec *alicloudnatgatewayv1.AlicloudNatGatewaySpec) string {
	if spec.NatType != nil {
		return *spec.NatType
	}
	return "Enhanced"
}

func paymentType(spec *alicloudnatgatewayv1.AlicloudNatGatewaySpec) string {
	if spec.PaymentType != nil {
		return *spec.PaymentType
	}
	return "PayAsYouGo"
}

func internetChargeType(spec *alicloudnatgatewayv1.AlicloudNatGatewaySpec) string {
	if spec.InternetChargeType != nil {
		return *spec.InternetChargeType
	}
	return "PayByLcu"
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

package module

import (
	"strings"

	alicloudvpngatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudvpngateway/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudVpnGateway *alicloudvpngatewayv1.AlicloudVpnGateway
	Tags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudvpngatewayv1.AlicloudVpnGatewayStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudVpnGateway = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudVpnGateway.String()),
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

func paymentType(spec *alicloudvpngatewayv1.AlicloudVpnGatewaySpec) string {
	if spec.PaymentType != nil {
		return *spec.PaymentType
	}
	return "PayAsYouGo"
}

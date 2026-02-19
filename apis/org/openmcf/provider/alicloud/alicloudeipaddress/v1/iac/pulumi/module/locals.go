package module

import (
	"strings"

	alicloudeipaddressv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudeipaddress/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudEipAddress *alicloudeipaddressv1.AlicloudEipAddress
	Tags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudeipaddressv1.AlicloudEipAddressStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudEipAddress = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudEipAddress.String()),
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

func bandwidth(spec *alicloudeipaddressv1.AlicloudEipAddressSpec) int32 {
	if spec.Bandwidth != nil {
		return *spec.Bandwidth
	}
	return 5
}

func internetChargeType(spec *alicloudeipaddressv1.AlicloudEipAddressSpec) string {
	if spec.InternetChargeType != nil {
		return *spec.InternetChargeType
	}
	return "PayByTraffic"
}

func isp(spec *alicloudeipaddressv1.AlicloudEipAddressSpec) string {
	if spec.Isp != nil {
		return *spec.Isp
	}
	return "BGP"
}

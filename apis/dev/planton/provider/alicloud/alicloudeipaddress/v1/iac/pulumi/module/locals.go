package module

import (
	"strings"

	alicloudeipaddressv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudeipaddress/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudEipAddress *alicloudeipaddressv1.AliCloudEipAddress
	Tags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudeipaddressv1.AliCloudEipAddressStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudEipAddress = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudEipAddress.String()),
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

func bandwidth(spec *alicloudeipaddressv1.AliCloudEipAddressSpec) int32 {
	if spec.Bandwidth != nil {
		return *spec.Bandwidth
	}
	return 5
}

func internetChargeType(spec *alicloudeipaddressv1.AliCloudEipAddressSpec) string {
	if spec.InternetChargeType != nil {
		return *spec.InternetChargeType
	}
	return "PayByTraffic"
}

func isp(spec *alicloudeipaddressv1.AliCloudEipAddressSpec) string {
	if spec.Isp != nil {
		return *spec.Isp
	}
	return "BGP"
}

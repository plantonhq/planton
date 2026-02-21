package module

import (
	"strings"

	aliclouddnszonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/aliclouddnszone/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudDnsZone *aliclouddnszonev1.AliCloudDnsZone
	Tags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *aliclouddnszonev1.AliCloudDnsZoneStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudDnsZone = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudDnsZone.String()),
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
